package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/autotag"
	interrors "github.com/streed/ml-notes/internal/errors"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
	"github.com/streed/ml-notes/internal/search"
)

var importCmd = &cobra.Command{
	Use:   "import <url>",
	Short: "Import a website as a new note",
	Long: `Import a website by URL and create a new note with the page title and content converted to markdown.

This command uses a headless browser to load the webpage, waiting for dynamic content to load,
then extracts the title and converts the body content to markdown format.

Examples:
  ml-notes import https://example.com
  ml-notes import https://blog.example.com/article --auto-tag
  ml-notes import https://docs.example.com --tags "docs,reference"`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

var (
	importTags    []string
	importAutoTag bool
	waitTimeout   time.Duration
)

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringSliceVarP(&importTags, "tags", "T", []string{}, "Tags for the imported note (comma-separated)")
	importCmd.Flags().BoolVar(&importAutoTag, "auto-tag", false, "Automatically generate tags using AI")
	importCmd.Flags().DurationVar(&waitTimeout, "timeout", 30*time.Second, "Timeout for page loading (default: 30s)")
}

func runImport(cmd *cobra.Command, args []string) error {
	pageURL := args[0]

	// Validate URL
	if _, err := url.Parse(pageURL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	fmt.Printf("🌐 Importing from: %s\n", pageURL)

	// Extract page content using headless browser
	title, content, err := extractPageContent(pageURL)
	if err != nil {
		return fmt.Errorf("failed to extract page content: %w", err)
	}

	if title == "" {
		title = pageURL // Use URL as fallback title
	}

	if content == "" {
		return interrors.ErrEmptyContent
	}

	fmt.Printf("📄 Page title: %s\n", title)
	fmt.Printf("📝 Content extracted (%d characters)\n", len(content))

	// Create note with tags if provided
	var note *models.Note
	if len(importTags) > 0 {
		note, err = noteRepo.CreateWithTags(title, content, importTags)
	} else {
		note, err = noteRepo.Create(title, content)
	}
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	// Auto-tag if requested
	if importAutoTag {
		fmt.Println("🤖 Generating AI tags...")
		autoTagger := autotag.NewAutoTagger(appConfig)

		if autoTagger.IsAvailable() {
			suggestedTags, err := autoTagger.SuggestTags(note)
			if err != nil {
				fmt.Printf("⚠️  Auto-tagging failed: %v\n", err)
			} else if len(suggestedTags) > 0 {
				// Merge with existing tags
				allTags := note.Tags
				tagSet := make(map[string]bool)
				for _, tag := range allTags {
					tagSet[tag] = true
				}
				for _, tag := range suggestedTags {
					if !tagSet[tag] {
						allTags = append(allTags, tag)
					}
				}

				// Update note with auto-generated tags
				if err := noteRepo.UpdateTags(note.ID, allTags); err != nil {
					fmt.Printf("⚠️  Failed to apply auto-tags: %v\n", err)
				} else {
					note.Tags = allTags // Update for display
					fmt.Printf("🏷️  Auto-generated tags: %s\n", strings.Join(suggestedTags, ", "))
				}
			} else {
				fmt.Println("🏷️  No auto-tags generated")
			}
		} else {
			fmt.Printf("⚠️  Auto-tagging unavailable. Please ensure summarization is enabled and Ollama is running.\n")
		}
	}

	// Index the note for semantic search
	fullText := title + " " + content

	// Use namespace-aware indexing if available
	if lilragSearch, ok := vectorSearch.(*search.LilRagSearch); ok {
		namespace := getCurrentProjectNamespace()
		if err := lilragSearch.IndexNoteWithNamespace(note.ID, fullText, namespace, "default"); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to index note for semantic search: %v\n", err)
		}
	} else {
		if err := vectorSearch.IndexNote(note.ID, fullText); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to index note for semantic search: %v\n", err)
		}
	}

	fmt.Printf("\n✅ Note imported successfully!\n")
	fmt.Printf("ID: %d\n", note.ID)
	fmt.Printf("Title: %s\n", note.Title)
	if len(note.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(note.Tags, ", "))
	}
	fmt.Printf("Created: %s\n", note.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Source: %s\n", pageURL)

	return nil
}

// extractPageContent uses chromedp to extract title and content from a webpage
func extractPageContent(pageURL string) (title, content string, err error) {
	// Configure Chrome options with security considerations
	// Start with default options that include necessary headless browser settings
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// Only add minimal flags needed for CI/container environments
		// NoSandbox is only used if we detect we're in a restricted environment
		chromedp.DisableGPU, // Safe to disable GPU in headless mode
		// Use a realistic user agent for better compatibility
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	// Only disable sandbox if we're in a restricted environment (CI/containers)
	// This is detected by checking if we can create user namespaces
	if isRestrictedEnvironment() {
		opts = append(opts, chromedp.NoSandbox)
	}

	// Create allocator context
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Create chromedp context
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	var htmlContent string

	// Run the tasks
	err = chromedp.Run(ctx,
		// Navigate to the page
		chromedp.Navigate(pageURL),
		// Wait for the page to load
		chromedp.WaitVisible("body", chromedp.ByQuery),
		// Additional wait for dynamic content
		chromedp.Sleep(2*time.Second),
		// Extract title
		chromedp.Title(&title),
		// Extract the main content (prefer article, main, or body)
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Try to find main content areas in order of preference
			selectors := []string{
				"article",
				"main",
				"[role='main']",
				".main-content",
				".content",
				".post-content",
				".entry-content",
				"body",
			}

			for _, selector := range selectors {
				var exists bool
				err := chromedp.Evaluate(fmt.Sprintf(`document.querySelector('%s') !== null`, selector), &exists).Do(ctx)
				if err == nil && exists {
					return chromedp.InnerHTML(selector, &htmlContent, chromedp.ByQuery).Do(ctx)
				}
			}

			// Fallback to body
			return chromedp.InnerHTML("body", &htmlContent, chromedp.ByQuery).Do(ctx)
		}),
	)

	if err != nil {
		return "", "", fmt.Errorf("failed to extract page content: %w", err)
	}

	// Convert HTML to markdown
	// Use empty domain initially and we'll handle URL resolution manually afterward
	converter := md.NewConverter("", true, nil)
	
	// Configure converter options for better markdown output
	converter.AddRules(
		// Remove scripts and styles
		md.Rule{
			Filter: []string{"script", "style", "noscript"},
			Replacement: func(content string, selection *goquery.Selection, opt *md.Options) *string {
				text := ""
				return &text
			},
		},
		// Handle navigation and sidebar content
		md.Rule{
			Filter: []string{"nav", "aside", ".sidebar", ".navigation", ".menu"},
			Replacement: func(content string, selection *goquery.Selection, opt *md.Options) *string {
				text := ""
				return &text
			},
		},
		// Clean up footer content
		md.Rule{
			Filter: []string{"footer", ".footer"},
			Replacement: func(content string, selection *goquery.Selection, opt *md.Options) *string {
				text := ""
				return &text
			},
		},
		// Custom image handling to preserve original URLs
		md.Rule{
			Filter: []string{"img"},
			Replacement: func(content string, selection *goquery.Selection, opt *md.Options) *string {
				src, exists := selection.Attr("src")
				if !exists {
					text := ""
					return &text
				}
				
				// Resolve relative URLs to absolute URLs
				absoluteSrc := resolveURL(pageURL, src)
				
				alt, _ := selection.Attr("alt")
				if alt == "" {
					alt = "Image"
				}
				
				result := fmt.Sprintf("![%s](%s)", alt, absoluteSrc)
				return &result
			},
		},
	)

	markdown, err := converter.ConvertString(htmlContent)
	if err != nil {
		return "", "", fmt.Errorf("failed to convert HTML to markdown: %w", err)
	}

	// Clean up the markdown content
	content = cleanMarkdownContent(markdown)

	logger.Debug("Extracted content from %s: title='%s', content_length=%d", pageURL, title, len(content))

	return title, content, nil
}

// cleanMarkdownContent removes excessive whitespace and cleans up the markdown
func cleanMarkdownContent(content string) string {
	lines := strings.Split(content, "\n")
	var cleanLines []string
	
	previousLineEmpty := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Skip empty lines that follow other empty lines (collapse multiple empty lines)
		if trimmed == "" {
			if !previousLineEmpty {
				cleanLines = append(cleanLines, "")
				previousLineEmpty = true
			}
			continue
		}
		
		previousLineEmpty = false
		cleanLines = append(cleanLines, trimmed)
	}
	
	// Remove leading and trailing empty lines
	for len(cleanLines) > 0 && cleanLines[0] == "" {
		cleanLines = cleanLines[1:]
	}
	for len(cleanLines) > 0 && cleanLines[len(cleanLines)-1] == "" {
		cleanLines = cleanLines[:len(cleanLines)-1]
	}
	
	return strings.Join(cleanLines, "\n")
}

// isRestrictedEnvironment checks if we're running in a restricted environment
// where Chrome sandbox needs to be disabled (CI, containers, etc.)
func isRestrictedEnvironment() bool {
	// Check for common CI environment variables
	ciEnvVars := []string{
		"CI", "CONTINUOUS_INTEGRATION", "BUILD_NUMBER", "GITHUB_ACTIONS",
		"GITLAB_CI", "JENKINS_URL", "TRAVIS", "CIRCLECI", "BUILDKITE",
	}
	
	for _, envVar := range ciEnvVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}
	
	// Check if we're running in a container
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	
	// Check for AppArmor restrictions (common in Ubuntu 23.10+)
	if _, err := os.Stat("/proc/sys/kernel/apparmor_restrict_unprivileged_userns"); err == nil {
		return true
	}
	
	return false
}

// resolveURL resolves a potentially relative URL against a base URL
func resolveURL(baseURL, href string) string {
	// Parse the base URL
	base, err := url.Parse(baseURL)
	if err != nil {
		return href // Return original if we can't parse base
	}
	
	// Parse the href
	ref, err := url.Parse(href)
	if err != nil {
		return href // Return original if we can't parse href
	}
	
	// Resolve the reference against the base
	resolved := base.ResolveReference(ref)
	return resolved.String()
}