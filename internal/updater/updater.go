package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/streed/ml-notes/internal/logger"
)

// Config holds updater configuration
type Config struct {
	GitHubOwner       string
	GitHubRepo        string
	CurrentVersion    string
	IncludePrerelease bool
	Platform          string
	Architecture      string
	Timeout           time.Duration
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	Version      string    `json:"version"`
	DownloadURL  string    `json:"download_url"`
	ReleaseURL   string    `json:"release_url"`
	ReleaseNotes string    `json:"release_notes"`
	PublishedAt  time.Time `json:"published_at"`
	Prerelease   bool      `json:"prerelease"`
	Size         int64     `json:"size"`
	Checksum     string    `json:"checksum,omitempty"`
}

// ProgressStage represents the current stage of the update process
type ProgressStage int

const (
	StageDownload ProgressStage = iota
	StageVerify
	StageBackup
	StageReplace
	StageComplete
)

// ProgressInfo provides progress updates during the update process
type ProgressInfo struct {
	Stage           ProgressStage
	Percent         float64
	BytesDownloaded int64
	TotalBytes      int64
	Message         string
}

// Updater handles application updates
type Updater struct {
	config     *Config
	httpClient *http.Client
}

// GitHubRelease represents a GitHub release response
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
	Prerelease  bool      `json:"prerelease"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// New creates a new updater instance
func New(config *Config) *Updater {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Updater{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// CheckForUpdate checks if a newer version is available
func (u *Updater) CheckForUpdate() (*UpdateInfo, error) {
	releases, err := u.fetchReleases()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found")
	}

	// Find the latest appropriate release
	var latestRelease *GitHubRelease
	for _, release := range releases {
		// Skip pre-releases unless explicitly requested
		if release.Prerelease && !u.config.IncludePrerelease {
			continue
		}

		// This is the latest appropriate release
		latestRelease = &release
		break
	}

	if latestRelease == nil {
		return nil, fmt.Errorf("no suitable releases found")
	}

	// Check if we need to update
	if !u.isNewer(latestRelease.TagName, u.config.CurrentVersion) {
		return nil, nil // No update needed
	}

	// Find the appropriate asset for our platform
	downloadURL, size, err := u.findAssetForPlatform(latestRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to find asset for platform: %w", err)
	}

	return &UpdateInfo{
		Version:      latestRelease.TagName,
		DownloadURL:  downloadURL,
		ReleaseURL:   latestRelease.HTMLURL,
		ReleaseNotes: latestRelease.Body,
		PublishedAt:  latestRelease.PublishedAt,
		Prerelease:   latestRelease.Prerelease,
		Size:         size,
	}, nil
}

// CheckSpecificVersion checks for a specific version
func (u *Updater) CheckSpecificVersion(version string) (*UpdateInfo, error) {
	// Ensure version has 'v' prefix
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	releases, err := u.fetchReleases()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}

	// Find the specific version
	for _, release := range releases {
		if release.TagName == version {
			downloadURL, size, err := u.findAssetForPlatform(&release)
			if err != nil {
				return nil, fmt.Errorf("failed to find asset for platform: %w", err)
			}

			return &UpdateInfo{
				Version:      release.TagName,
				DownloadURL:  downloadURL,
				ReleaseURL:   release.HTMLURL,
				ReleaseNotes: release.Body,
				PublishedAt:  release.PublishedAt,
				Prerelease:   release.Prerelease,
				Size:         size,
			}, nil
		}
	}

	return nil, fmt.Errorf("version %s not found", version)
}

// GetLatestRelease gets the latest release info (for force updates)
func (u *Updater) GetLatestRelease() (*UpdateInfo, error) {
	releases, err := u.fetchReleases()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found")
	}

	// Get the first (latest) release
	release := releases[0]
	downloadURL, size, err := u.findAssetForPlatform(&release)
	if err != nil {
		return nil, fmt.Errorf("failed to find asset for platform: %w", err)
	}

	return &UpdateInfo{
		Version:      release.TagName,
		DownloadURL:  downloadURL,
		ReleaseURL:   release.HTMLURL,
		ReleaseNotes: release.Body,
		PublishedAt:  release.PublishedAt,
		Prerelease:   release.Prerelease,
		Size:         size,
	}, nil
}

// PerformUpdate downloads and installs the update
func (u *Updater) PerformUpdate(updateInfo *UpdateInfo, progress chan<- ProgressInfo) error {
	defer close(progress)

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create temporary directory for download
	tempDir, err := os.MkdirTemp("", "ml-notes-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			logger.Debug("Failed to remove temp directory: %v", err)
		}
	}()

	// Download the update
	progress <- ProgressInfo{Stage: StageDownload, Message: "Starting download..."}
	downloadPath, err := u.downloadUpdate(updateInfo, tempDir, progress)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	// Extract if needed and verify
	progress <- ProgressInfo{Stage: StageVerify, Message: "Verifying download..."}
	binaryPath, err := u.extractAndVerify(downloadPath, tempDir)
	if err != nil {
		return fmt.Errorf("failed to extract/verify update: %w", err)
	}

	// Create backup
	progress <- ProgressInfo{Stage: StageBackup, Message: "Creating backup..."}
	backupPath := execPath + ".backup"
	if err := u.createBackup(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Replace binary atomically
	progress <- ProgressInfo{Stage: StageReplace, Message: "Installing update..."}
	if err := u.replaceBinary(binaryPath, execPath); err != nil {
		// Try to restore backup
		logger.Error("Update failed, attempting to restore backup...")
		if restoreErr := u.restoreBackup(backupPath, execPath); restoreErr != nil {
			logger.Error("Failed to restore backup: %v", restoreErr)
		}
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	progress <- ProgressInfo{Stage: StageComplete, Message: "Update complete"}
	return nil
}

// fetchReleases fetches releases from GitHub API
func (u *Updater) fetchReleases() ([]GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", u.config.GitHubOwner, u.config.GitHubRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("ml-notes/%s", u.config.CurrentVersion))

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Debug("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	return releases, nil
}

// isNewer checks if version1 is newer than version2
func (u *Updater) isNewer(version1, version2 string) bool {
	// Simple version comparison - in production you'd want a proper semver library
	v1 := strings.TrimPrefix(version1, "v")
	v2 := strings.TrimPrefix(version2, "v")

	// For now, just do string comparison (works for most cases)
	// TODO: Implement proper semantic version comparison
	return v1 > v2
}

// findAssetForPlatform finds the appropriate download asset for the current platform
func (u *Updater) findAssetForPlatform(release *GitHubRelease) (string, int64, error) {
	// Platform mapping
	platformMap := map[string]string{
		"darwin":  "darwin",
		"linux":   "linux",
		"windows": "windows",
	}

	// Architecture mapping
	archMap := map[string]string{
		"amd64": "amd64",
		"arm64": "arm64",
		"386":   "386",
	}

	platform := platformMap[u.config.Platform]
	arch := archMap[u.config.Architecture]

	if platform == "" || arch == "" {
		return "", 0, fmt.Errorf("unsupported platform: %s/%s", u.config.Platform, u.config.Architecture)
	}

	// Expected filename patterns
	var expectedPatterns []string
	if u.config.Platform == "windows" {
		expectedPatterns = []string{
			fmt.Sprintf("ml-notes-%s-%s-%s.zip", release.TagName, platform, arch),
			fmt.Sprintf("ml-notes-%s-%s.zip", platform, arch),
		}
	} else {
		expectedPatterns = []string{
			fmt.Sprintf("ml-notes-%s-%s-%s.tar.gz", release.TagName, platform, arch),
			fmt.Sprintf("ml-notes-%s-%s.tar.gz", platform, arch),
		}
	}

	// Find matching asset
	for _, asset := range release.Assets {
		for _, pattern := range expectedPatterns {
			if asset.Name == pattern {
				return asset.BrowserDownloadURL, asset.Size, nil
			}
		}
	}

	return "", 0, fmt.Errorf("no asset found for platform %s/%s", platform, arch)
}

// downloadUpdate downloads the update file with progress reporting
func (u *Updater) downloadUpdate(updateInfo *UpdateInfo, tempDir string, progress chan<- ProgressInfo) (string, error) {
	resp, err := u.httpClient.Get(updateInfo.DownloadURL)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Debug("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Determine filename from URL
	filename := filepath.Base(updateInfo.DownloadURL)
	downloadPath := filepath.Join(tempDir, filename)

	file, err := os.Create(downloadPath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Debug("Failed to close file: %v", err)
		}
	}()

	// Copy with progress reporting
	var downloaded int64
	totalSize := resp.ContentLength
	if totalSize <= 0 {
		totalSize = updateInfo.Size
	}

	buffer := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := file.Write(buffer[:n]); writeErr != nil {
				return "", writeErr
			}
			downloaded += int64(n)

			if totalSize > 0 {
				percent := float64(downloaded) / float64(totalSize) * 100
				progress <- ProgressInfo{
					Stage:           StageDownload,
					Percent:         percent,
					BytesDownloaded: downloaded,
					TotalBytes:      totalSize,
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	return downloadPath, nil
}

// extractAndVerify extracts the binary from archive and verifies it
func (u *Updater) extractAndVerify(downloadPath, tempDir string) (string, error) {
	ext := filepath.Ext(downloadPath)
	var binaryPath string
	var err error

	switch ext {
	case ".gz":
		// tar.gz file
		binaryPath, err = u.extractTarGz(downloadPath, tempDir)
	case ".zip":
		// zip file
		binaryPath, err = u.extractZip(downloadPath, tempDir)
	default:
		// Assume it's already a binary
		binaryPath = downloadPath
	}

	if err != nil {
		return "", err
	}

	// Verify the binary is executable
	if err := u.verifyBinary(binaryPath); err != nil {
		return "", err
	}

	return binaryPath, nil
}

// extractTarGz extracts a tar.gz file and returns the binary path
func (u *Updater) extractTarGz(archivePath, destDir string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Debug("Failed to close file: %v", err)
		}
	}()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := gzr.Close(); err != nil {
			logger.Debug("Failed to close gzip reader: %v", err)
		}
	}()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if header.Typeflag == tar.TypeReg {
			// Check if this looks like our binary
			filename := filepath.Base(header.Name)
			if strings.HasPrefix(filename, "ml-notes") && !strings.Contains(filename, ".") {
				// Extract this file
				extractPath := filepath.Join(destDir, filename)
				outFile, err := os.Create(extractPath)
				if err != nil {
					return "", err
				}

				if _, err := io.Copy(outFile, tr); err != nil {
					_ = outFile.Close()
					return "", err
				}
				if err := outFile.Close(); err != nil {
					return "", fmt.Errorf("failed to close extracted file: %w", err)
				}

				// Make executable
				if err := os.Chmod(extractPath, 0o755); err != nil {
					return "", err
				}

				return extractPath, nil
			}
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

// extractZip extracts a zip file and returns the binary path
func (u *Updater) extractZip(archivePath, destDir string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := r.Close(); err != nil {
			logger.Debug("Failed to close zip reader: %v", err)
		}
	}()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		// Check if this looks like our binary
		filename := filepath.Base(f.Name)
		if strings.HasPrefix(filename, "ml-notes") {
			// Extract this file
			extractPath := filepath.Join(destDir, filename)

			rc, err := f.Open()
			if err != nil {
				return "", err
			}

			outFile, err := os.Create(extractPath)
			if err != nil {
				_ = rc.Close()
				return "", err
			}

			_, err = io.Copy(outFile, rc)
			_ = rc.Close()
			if closeErr := outFile.Close(); closeErr != nil {
				return "", fmt.Errorf("failed to close extracted file: %w", closeErr)
			}

			if err != nil {
				return "", err
			}

			// Make executable on Unix systems
			if runtime.GOOS != "windows" {
				if err := os.Chmod(extractPath, 0o755); err != nil {
					return "", err
				}
			}

			return extractPath, nil
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

// verifyBinary performs basic verification of the binary
func (u *Updater) verifyBinary(binaryPath string) error {
	// Check if file exists and is executable
	info, err := os.Stat(binaryPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("binary path is a directory")
	}

	// On Unix systems, check execute permission
	if runtime.GOOS != "windows" {
		if info.Mode()&0o111 == 0 {
			return fmt.Errorf("binary is not executable")
		}
	}

	return nil
}

// createBackup creates a backup of the current binary
func (u *Updater) createBackup(srcPath, backupPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := src.Close(); err != nil {
			logger.Debug("Failed to close source file: %v", err)
		}
	}()

	dst, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := dst.Close(); err != nil {
			logger.Debug("Failed to close destination file: %v", err)
		}
	}()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := src.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(backupPath, srcInfo.Mode())
}

// replaceBinary atomically replaces the current binary
func (u *Updater) replaceBinary(newBinaryPath, targetPath string) error {
	// Create a temporary name for the new binary in the same directory
	tempPath := targetPath + ".new"

	// Copy the new binary to the temp location first
	if err := u.copyFile(newBinaryPath, tempPath); err != nil {
		return fmt.Errorf("failed to copy binary to temp location: %w", err)
	}
	defer func() {
		if err := os.Remove(tempPath); err != nil {
			logger.Debug("Failed to remove temp file: %v", err)
		}
	}() // Clean up temp file if something goes wrong

	if runtime.GOOS == "windows" {
		// On Windows, move current binary to backup location
		backupPath := targetPath + ".old"
		if err := os.Rename(targetPath, backupPath); err != nil {
			return fmt.Errorf("failed to move current binary: %w", err)
		}

		// Move new binary to target location
		if err := os.Rename(tempPath, targetPath); err != nil {
			// Try to restore original
			if restoreErr := os.Rename(backupPath, targetPath); restoreErr != nil {
				logger.Error("Failed to restore backup binary: %v", restoreErr)
			}
			return fmt.Errorf("failed to install new binary: %w", err)
		}
	} else {
		// On Unix systems, we can't replace a running binary directly
		// Instead, we move the current binary to a backup location first
		backupPath := targetPath + ".old"

		// Move current binary to backup (this works even if it's running)
		if err := os.Rename(targetPath, backupPath); err != nil {
			return fmt.Errorf("failed to move current binary to backup: %w", err)
		}

		// Now move the new binary into place
		if err := os.Rename(tempPath, targetPath); err != nil {
			// Try to restore original
			if restoreErr := os.Rename(backupPath, targetPath); restoreErr != nil {
				logger.Error("Failed to restore backup binary: %v", restoreErr)
			}
			return fmt.Errorf("failed to install new binary: %w", err)
		}

		// Remove the old backup (the process will continue running from the old location)
		if err := os.Remove(backupPath); err != nil {
			logger.Debug("Failed to remove old backup: %v", err)
		}
	}

	return nil
}

// copyFile copies a file from src to dst, preserving permissions
func (u *Updater) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := srcFile.Close(); err != nil {
			logger.Debug("Failed to close source file: %v", err)
		}
	}()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := dstFile.Close(); err != nil {
			logger.Debug("Failed to close destination file: %v", err)
		}
	}()

	// Copy file contents
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// restoreBackup restores the backup in case of failure
func (u *Updater) restoreBackup(backupPath, targetPath string) error {
	return os.Rename(backupPath, targetPath)
}
