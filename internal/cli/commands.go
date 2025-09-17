package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
)

func init() {
	// Register all CLI commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(tagsCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(autotagCmd)
	rootCmd.AddCommand(updateCmd)
}

// Init Command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ml-notes configuration",
	Long: `Initialize ml-notes configuration interactively or with flags.
This command sets up the configuration file and creates necessary directories.`,
	RunE: runInit,
}

var (
	initDataDir             string
	initOllamaEndpoint      string
	initInteractive         bool
	initSummarizationModel  string
	initEnableSummarization bool
)

func init() {
	initCmd.Flags().StringVar(&initDataDir, "data-dir", "", "Data directory for storing notes database")
	initCmd.Flags().StringVar(&initOllamaEndpoint, "ollama-endpoint", "", "Ollama API endpoint (e.g., http://localhost:11434)")
	initCmd.Flags().StringVar(&initSummarizationModel, "summarization-model", "", "Model to use for summarization (e.g., llama3.2:latest)")
	initCmd.Flags().BoolVar(&initEnableSummarization, "enable-summarization", true, "Enable AI summarization features")
	initCmd.Flags().BoolVarP(&initInteractive, "interactive", "i", false, "Run interactive setup")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if config already exists
	configPath, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Configuration already exists at: %s\n", configPath)
		fmt.Print("Do you want to overwrite it? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Configuration initialization cancelled.")
			return nil
		}
	}

	if initInteractive {
		return runInteractiveInit()
	}

	// Non-interactive initialization
	cfg, err := config.InitializeConfigWithSummarization(
		initDataDir,
		initOllamaEndpoint,
		initSummarizationModel,
		initEnableSummarization,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	fmt.Printf("Configuration initialized successfully!\n")
	fmt.Printf("Config file: %s\n", configPath)
	fmt.Printf("Data directory: %s\n", cfg.DataDirectory)
	fmt.Printf("Database: %s\n", cfg.GetDatabasePath())

	return nil
}

func runInteractiveInit() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== ML Notes Interactive Setup ===")
	fmt.Println()

	// Data directory
	fmt.Printf("Data directory (default: %s): ", config.GetDefaultDataDirectory())
	dataDir, _ := reader.ReadString('\n')
	dataDir = strings.TrimSpace(dataDir)

	// Ollama endpoint
	fmt.Print("Ollama endpoint (default: http://localhost:11434): ")
	ollamaEndpoint, _ := reader.ReadString('\n')
	ollamaEndpoint = strings.TrimSpace(ollamaEndpoint)

	// Summarization model
	fmt.Print("Summarization model (default: llama3.2:latest): ")
	summarizationModel, _ := reader.ReadString('\n')
	summarizationModel = strings.TrimSpace(summarizationModel)

	// Enable summarization
	fmt.Print("Enable AI summarization? (Y/n): ")
	enableSummarization, _ := reader.ReadString('\n')
	enableSummarization = strings.TrimSpace(strings.ToLower(enableSummarization))
	enableSummarizationBool := enableSummarization != "n" && enableSummarization != "no"

	// Initialize configuration
	cfg, err := config.InitializeConfigWithSummarization(
		dataDir,
		ollamaEndpoint,
		summarizationModel,
		enableSummarizationBool,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	configPath, _ := config.GetConfigPath()
	fmt.Printf("\n✅ Configuration initialized successfully!\n")
	fmt.Printf("Config file: %s\n", configPath)
	fmt.Printf("Data directory: %s\n", cfg.DataDirectory)
	fmt.Printf("Database: %s\n", cfg.GetDatabasePath())

	return nil
}

// Config Command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long:  `View and modify ml-notes configuration settings.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE:  runConfigShow,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Data Directory: %s\n", cfg.DataDirectory)
	fmt.Printf("  Database Path: %s\n", cfg.GetDatabasePath())
	fmt.Printf("  Ollama Endpoint: %s\n", cfg.OllamaEndpoint)
	fmt.Printf("  Lil-Rag URL: %s\n", cfg.LilRagURL)
	fmt.Printf("  Debug: %t\n", cfg.Debug)
	fmt.Printf("  Enable Summarization: %t\n", cfg.EnableSummarization)
	fmt.Printf("  Summarization Model: %s\n", cfg.SummarizationModel)
	fmt.Printf("  Enable Auto-tagging: %t\n", cfg.EnableAutoTagging)
	fmt.Printf("  Auto-tag Model: %s\n", cfg.AutoTagModel)
	fmt.Printf("  Max Auto Tags: %d\n", cfg.MaxAutoTags)

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	switch key {
	case "data-directory":
		cfg.DataDirectory = value
	case "ollama-endpoint":
		cfg.OllamaEndpoint = value
	case "lilrag-url":
		cfg.LilRagURL = value
	case "debug":
		cfg.Debug = value == "true"
	case "enable-summarization":
		cfg.EnableSummarization = value == "true"
	case "summarization-model":
		cfg.SummarizationModel = value
	case "enable-auto-tagging":
		cfg.EnableAutoTagging = value == "true"
	case "auto-tag-model":
		cfg.AutoTagModel = value
	case "max-auto-tags":
		if maxTags, err := strconv.Atoi(value); err == nil {
			cfg.MaxAutoTags = maxTags
		} else {
			return fmt.Errorf("invalid number for max-auto-tags: %s", value)
		}
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Configuration updated: %s = %s\n", key, value)
	return nil
}

// Add Command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a new note",
	Long:  `Create a new note with the specified title, content, and tags.`,
	RunE:  runAdd,
}

var (
	addTitle   string
	addContent string
	addTags    string
	addFile    string
)

func init() {
	addCmd.Flags().StringVarP(&addTitle, "title", "t", "", "Note title")
	addCmd.Flags().StringVarP(&addContent, "content", "c", "", "Note content")
	addCmd.Flags().StringVar(&addTags, "tags", "", "Comma-separated tags")
	addCmd.Flags().StringVarP(&addFile, "file", "f", "", "Read content from file")
}

func runAdd(cmd *cobra.Command, args []string) error {
	content := addContent

	// Read from file if specified
	if addFile != "" {
		data, err := os.ReadFile(addFile)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", addFile, err)
		}
		content = string(data)
	}

	// Read from stdin if no content and no file
	if content == "" && addFile == "" {
		fmt.Print("Enter note content (Ctrl+D to finish):\n")
		data, err := os.ReadFile("/dev/stdin")
		if err == nil {
			content = string(data)
		}
	}

	if addTitle == "" {
		return fmt.Errorf("title is required (use -t or --title)")
	}

	var tags []string
	if addTags != "" {
		tags = strings.Split(addTags, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}

	note, err := noteRepo.CreateWithTags(addTitle, content, tags)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	fmt.Printf("Note created successfully!\n")
	fmt.Printf("ID: %d\n", note.ID)
	fmt.Printf("Title: %s\n", note.Title)
	if len(note.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(note.Tags, ", "))
	}

	return nil
}

// List Command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List notes",
	Long:  `List notes with optional pagination and formatting.`,
	RunE:  runList,
}

var (
	listLimit  int
	listOffset int
	listShort  bool
)

func init() {
	listCmd.Flags().IntVar(&listLimit, "limit", 20, "Maximum number of notes to show")
	listCmd.Flags().IntVar(&listOffset, "offset", 0, "Number of notes to skip")
	listCmd.Flags().BoolVar(&listShort, "short", false, "Show only ID and title")
}

func runList(cmd *cobra.Command, args []string) error {
	notes, err := noteRepo.List(listLimit, listOffset)
	if err != nil {
		return fmt.Errorf("failed to list notes: %w", err)
	}

	if len(notes) == 0 {
		fmt.Println("No notes found.")
		return nil
	}

	for _, note := range notes {
		if listShort {
			fmt.Printf("%d: %s\n", note.ID, note.Title)
		} else {
			fmt.Printf("ID: %d\n", note.ID)
			fmt.Printf("Title: %s\n", note.Title)
			if len(note.Tags) > 0 {
				fmt.Printf("Tags: %s\n", strings.Join(note.Tags, ", "))
			}
			fmt.Printf("Created: %s\n", note.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Updated: %s\n", note.UpdatedAt.Format("2006-01-02 15:04:05"))

			// Show preview of content
			preview := note.Content
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			fmt.Printf("Content: %s\n", preview)
			fmt.Println("---")
		}
	}

	return nil
}

// Get Command
var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a note by ID",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid note ID: %s", args[0])
	}

	note, err := noteRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	fmt.Printf("ID: %d\n", note.ID)
	fmt.Printf("Title: %s\n", note.Title)
	if len(note.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(note.Tags, ", "))
	}
	fmt.Printf("Created: %s\n", note.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", note.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("\nContent:\n%s\n", note.Content)

	return nil
}

// Edit Command
var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a note",
	Args:  cobra.ExactArgs(1),
	RunE:  runEdit,
}

var (
	editTitle   string
	editContent string
	editTags    string
)

func init() {
	editCmd.Flags().StringVarP(&editTitle, "title", "t", "", "New note title")
	editCmd.Flags().StringVarP(&editContent, "content", "c", "", "New note content")
	editCmd.Flags().StringVar(&editTags, "tags", "", "New comma-separated tags")
}

func runEdit(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid note ID: %s", args[0])
	}

	note, err := noteRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	// Update fields if provided
	if editTitle != "" {
		note.Title = editTitle
	}
	if editContent != "" {
		note.Content = editContent
	}
	if editTags != "" {
		tags := strings.Split(editTags, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		note.Tags = tags
	}

	if err := noteRepo.Update(note); err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	fmt.Printf("Note updated successfully!\n")
	fmt.Printf("ID: %d\n", note.ID)
	fmt.Printf("Title: %s\n", note.Title)
	if len(note.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(note.Tags, ", "))
	}

	return nil
}

// Delete Command
var deleteCmd = &cobra.Command{
	Use:   "delete <id> [id2] [id3] ...",
	Short: "Delete notes by ID",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
	var ids []int
	for _, arg := range args {
		id, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("invalid note ID: %s", arg)
		}
		ids = append(ids, id)
	}

	for _, id := range ids {
		if err := noteRepo.Delete(id); err != nil {
			logger.Error("Failed to delete note %d: %v", id, err)
		} else {
			fmt.Printf("Note %d deleted successfully.\n", id)
		}
	}

	return nil
}

// Search Command
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search notes",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSearch,
}

var (
	searchVector bool
	searchTags   string
	searchLimit  int
	searchShort  bool
)

func init() {
	searchCmd.Flags().BoolVar(&searchVector, "vector", true, "Use vector/semantic search")
	searchCmd.Flags().StringVar(&searchTags, "tags", "", "Search by tags (comma-separated)")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 10, "Maximum number of results")
	searchCmd.Flags().BoolVar(&searchShort, "short", false, "Show only ID and title")
}

func runSearch(cmd *cobra.Command, args []string) error {
	var query string
	if len(args) > 0 {
		query = args[0]
	}

	var notes []*models.Note
	var err error

	if searchTags != "" {
		// Search by tags
		tags := strings.Split(searchTags, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		notes, err = noteRepo.SearchByTags(tags)
	} else if query != "" {
		// Search by content
		if searchVector {
			notes, err = vectorSearch.SearchSimilar(query, searchLimit)
		} else {
			notes, err = noteRepo.Search(query)
			if len(notes) > searchLimit {
				notes = notes[:searchLimit]
			}
		}
	} else {
		return fmt.Errorf("please provide either a search query or use --tags to search by tags")
	}

	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(notes) == 0 {
		fmt.Println("No notes found.")
		return nil
	}

	fmt.Printf("Found %d note(s):\n\n", len(notes))

	for _, note := range notes {
		if searchShort {
			fmt.Printf("%d: %s\n", note.ID, note.Title)
		} else {
			fmt.Printf("ID: %d\n", note.ID)
			fmt.Printf("Title: %s\n", note.Title)
			if len(note.Tags) > 0 {
				fmt.Printf("Tags: %s\n", strings.Join(note.Tags, ", "))
			}
			// Show preview of content
			preview := note.Content
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			fmt.Printf("Content: %s\n", preview)
			fmt.Println("---")
		}
	}

	return nil
}

// Tags Command
var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage note tags",
	Long:  `List and manage tags for notes.`,
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags",
	RunE:  runTagsList,
}

func init() {
	tagsCmd.AddCommand(tagsListCmd)
}

func runTagsList(cmd *cobra.Command, args []string) error {
	tags, err := noteRepo.GetAllTags()
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	if len(tags) == 0 {
		fmt.Println("No tags found.")
		return nil
	}

	fmt.Printf("Tags (%d):\n", len(tags))
	for _, tag := range tags {
		fmt.Printf("  %s\n", tag.Name)
	}

	return nil
}

// Analyze Command
var analyzeCmd = &cobra.Command{
	Use:   "analyze <id>",
	Short: "Analyze a note with AI",
	Args:  cobra.ExactArgs(1),
	RunE:  runAnalyze,
}

var analyzePrompt string

func init() {
	analyzeCmd.Flags().StringVarP(&analyzePrompt, "prompt", "p", "", "Custom analysis prompt")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid note ID: %s", args[0])
	}

	note, err := noteRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	// For now, just show the note info as analysis isn't fully implemented in this simplified version
	fmt.Printf("Analysis for note %d:\n", note.ID)
	fmt.Printf("Title: %s\n", note.Title)
	fmt.Printf("Content length: %d characters\n", len(note.Content))
	fmt.Printf("Tags: %s\n", strings.Join(note.Tags, ", "))
	fmt.Printf("Created: %s\n", note.CreatedAt.Format("2006-01-02 15:04:05"))

	if analyzePrompt != "" {
		fmt.Printf("Custom prompt: %s\n", analyzePrompt)
	}

	return nil
}

// Autotag Command
var autotagCmd = &cobra.Command{
	Use:   "autotag <id>",
	Short: "Auto-tag a note with AI",
	Args:  cobra.ExactArgs(1),
	RunE:  runAutotag,
}

func runAutotag(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid note ID: %s", args[0])
	}

	note, err := noteRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	// For now, just show the note info as auto-tagging isn't fully implemented in this simplified version
	fmt.Printf("Auto-tagging for note %d:\n", note.ID)
	fmt.Printf("Title: %s\n", note.Title)
	fmt.Printf("Current tags: %s\n", strings.Join(note.Tags, ", "))
	fmt.Printf("Note: Auto-tagging feature requires AI service configuration.\n")

	return nil
}

// Update Command
var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an existing note",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

var (
	updateTitle   string
	updateContent string
	updateTags    string
)

func init() {
	updateCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "New note title")
	updateCmd.Flags().StringVarP(&updateContent, "content", "c", "", "New note content")
	updateCmd.Flags().StringVar(&updateTags, "tags", "", "New comma-separated tags")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	return runEdit(cmd, args) // Update is the same as edit
}
