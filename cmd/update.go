package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/updater"
)

var (
	updateForce      bool
	updatePrerelease bool
	updateVersion    string
	updateDryRun     bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update ml-notes to the latest version",
	Long: `Update ml-notes to the latest version from GitHub releases.

This command checks for newer releases and downloads the appropriate binary
for your platform. The update is performed atomically to ensure safety.

Examples:
  ml-notes update                    # Update to latest stable release
  ml-notes update --force            # Force update even if already latest
  ml-notes update --prerelease       # Include pre-release versions
  ml-notes update --version v1.2.3   # Update to specific version
  ml-notes update --dry-run          # Check for updates without installing

The updater will:
1. Check your current version against GitHub releases
2. Download the appropriate binary for your platform (${GOOS}/${GOARCH})
3. Verify the download integrity
4. Atomically replace the current binary
5. Preserve the original binary as a backup

Platform Detection:
- OS: ` + runtime.GOOS + `
- Architecture: ` + runtime.GOARCH + `

Safety Features:
- Backup creation before update
- Atomic replacement (download to temp, then move)
- Checksum verification (if available)
- Rollback capability if update fails`,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVar(&updateForce, "force", false, "Force update even if already on latest version")
	updateCmd.Flags().BoolVar(&updatePrerelease, "prerelease", false, "Include pre-release versions")
	updateCmd.Flags().StringVar(&updateVersion, "version", "", "Update to specific version (e.g., v1.2.3)")
	updateCmd.Flags().BoolVar(&updateDryRun, "dry-run", false, "Check for updates without installing")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	logger.Info("Checking for ml-notes updates...")

	// Initialize updater with config values
	u := updater.New(&updater.Config{
		GitHubOwner:       appConfig.GitHubOwner,
		GitHubRepo:        appConfig.GitHubRepo,
		CurrentVersion:    Version,
		IncludePrerelease: updatePrerelease,
		Platform:          runtime.GOOS,
		Architecture:      runtime.GOARCH,
	})

	// Check for updates
	updateInfo, err := u.CheckForUpdate()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	// Handle specific version request
	if updateVersion != "" {
		updateInfo, err = u.CheckSpecificVersion(updateVersion)
		if err != nil {
			return fmt.Errorf("failed to check version %s: %w", updateVersion, err)
		}
	}

	// Display current status
	fmt.Printf("Current version: %s\n", Version)

	if updateInfo == nil {
		fmt.Printf("✅ You are already running the latest version!\n")
		if !updateForce {
			return nil
		}
		fmt.Printf("🔄 Force update requested, proceeding anyway...\n")
		// Get latest release info for force update
		updateInfo, err = u.GetLatestRelease()
		if err != nil {
			return fmt.Errorf("failed to get latest release: %w", err)
		}
	}

	fmt.Printf("Latest version:  %s\n", updateInfo.Version)
	fmt.Printf("Release URL:     %s\n", updateInfo.ReleaseURL)
	fmt.Printf("Published:       %s\n", updateInfo.PublishedAt.Format("2006-01-02 15:04:05"))

	if updateInfo.Prerelease {
		fmt.Printf("⚠️  This is a pre-release version\n")
	}

	if updateDryRun {
		fmt.Printf("🔍 Dry run mode - would update to %s\n", updateInfo.Version)
		return nil
	}

	// Confirm update
	if !updateForce {
		fmt.Printf("\nDo you want to update to %s? [y/N]: ", updateInfo.Version)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			fmt.Println("Update cancelled.")
			return nil
		}
	}

	// Perform update
	fmt.Printf("\n🚀 Updating ml-notes to %s...\n", updateInfo.Version)

	progress := make(chan updater.ProgressInfo, 1)
	go func() {
		for p := range progress {
			switch p.Stage {
			case updater.StageDownload:
				fmt.Printf("📥 Downloading... %.1f%% (%s)\n", p.Percent, formatBytes(p.BytesDownloaded))
			case updater.StageVerify:
				fmt.Printf("🔍 Verifying download...\n")
			case updater.StageBackup:
				fmt.Printf("💾 Creating backup...\n")
			case updater.StageReplace:
				fmt.Printf("🔄 Installing update...\n")
			case updater.StageComplete:
				fmt.Printf("✅ Update complete!\n")
			}
		}
	}()

	if err := u.PerformUpdate(updateInfo, progress); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Printf("\n🎉 Successfully updated to ml-notes %s!\n", updateInfo.Version)
	fmt.Printf("💡 Run 'ml-notes --version' to verify the new version.\n")

	// Show what's new if available
	if updateInfo.ReleaseNotes != "" {
		fmt.Printf("\n📋 What's new in %s:\n", updateInfo.Version)
		fmt.Printf("%s\n", updateInfo.ReleaseNotes)
	}

	return nil
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
