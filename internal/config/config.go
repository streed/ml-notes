package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/streed/ml-notes/internal/constants"
)

type Config struct {
	// Legacy fields for backward compatibility
	DatabasePath  string `json:"database_path,omitempty"`
	DataDirectory string `json:"data_directory,omitempty"`

	// Project system fields
	ProjectsDirectory string `json:"projects_directory"`
	CurrentProject    string `json:"current_project"`

	// Global settings
	OllamaEndpoint      string `json:"ollama_endpoint"`
	EmbeddingModel      string `json:"embedding_model"`
	VectorDimensions    int    `json:"vector_dimensions"`
	EnableVectorSearch  bool   `json:"enable_vector_search"`
	Debug               bool   `json:"debug"`
	VectorConfigVersion string `json:"vector_config_version,omitempty"`
	SummarizationModel  string `json:"summarization_model,omitempty"`
	EnableSummarization bool   `json:"enable_summarization"`
	Editor              string `json:"editor,omitempty"`
	EnableAutoTagging   bool   `json:"enable_auto_tagging"`
	AutoTagModel        string `json:"auto_tag_model,omitempty"`
	MaxAutoTags         int    `json:"max_auto_tags"`
	WebUITheme          string `json:"webui_theme,omitempty"`
	WebUICustomCSS      string `json:"webui_custom_css,omitempty"`
	GitHubOwner         string `json:"github_owner,omitempty"`
	GitHubRepo          string `json:"github_repo,omitempty"`
}

// Project represents a single project with its own database
type Project struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// getDefaultConfig returns a fresh copy of the default configuration
func getDefaultConfig() Config {
	return Config{
		// Legacy fields for backward compatibility
		DatabasePath:  "", // Will be set to DataDirectory/notes.db
		DataDirectory: "", // Will be set to ~/.local/share/ml-notes

		// Project system defaults
		ProjectsDirectory: "",        // Will be set to ~/.local/share/ml-notes/projects
		CurrentProject:    "default", // Default project name

		// Global settings
		OllamaEndpoint:      "http://localhost:11434",
		EmbeddingModel:      "nomic-embed-text",
		VectorDimensions:    384,
		EnableVectorSearch:  true,
		Debug:               false,
		SummarizationModel:  "llama3.2:latest",
		EnableSummarization: true,
		Editor:              "", // Empty means auto-detect editor
		EnableAutoTagging:   true,
		AutoTagModel:        "", // Empty means use SummarizationModel
		MaxAutoTags:         5,
		WebUITheme:          "dark",     // Default theme
		WebUICustomCSS:      "",         // Path to custom CSS file
		GitHubOwner:         "streed",   // Default GitHub owner for updates
		GitHubRepo:          "ml-notes", // Default GitHub repository for updates
	}
}

func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}
	return filepath.Join(configDir, "ml-notes", "config.json"), nil
}

func GetDefaultDataDirectory() string {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".", ".ml-notes")
		}
		dataDir = filepath.Join(homeDir, ".local", "share")
	}
	return filepath.Join(dataDir, "ml-notes")
}

func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		cfg := getDefaultConfig()
		if cfg.DataDirectory == "" {
			cfg.DataDirectory = GetDefaultDataDirectory()
		}
		if cfg.DatabasePath == "" {
			cfg.DatabasePath = filepath.Join(cfg.DataDirectory, "notes.db")
		}
		return &cfg, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults for empty fields
	defaults := getDefaultConfig()

	// Handle backward compatibility - migrate from old data directory structure
	if cfg.DataDirectory == "" {
		cfg.DataDirectory = GetDefaultDataDirectory()
	}
	if cfg.DatabasePath == "" {
		cfg.DatabasePath = filepath.Join(cfg.DataDirectory, "notes.db")
	}

	// Set up project system defaults
	if cfg.ProjectsDirectory == "" {
		cfg.ProjectsDirectory = filepath.Join(cfg.DataDirectory, "projects")
	}
	if cfg.CurrentProject == "" {
		cfg.CurrentProject = "default"
	}

	// Set other defaults
	if cfg.OllamaEndpoint == "" {
		cfg.OllamaEndpoint = defaults.OllamaEndpoint
	}
	if cfg.EmbeddingModel == "" {
		cfg.EmbeddingModel = defaults.EmbeddingModel
	}
	if cfg.VectorDimensions == 0 {
		cfg.VectorDimensions = defaults.VectorDimensions
	}
	if cfg.GitHubOwner == "" {
		cfg.GitHubOwner = defaults.GitHubOwner
	}
	if cfg.GitHubRepo == "" {
		cfg.GitHubRepo = defaults.GitHubRepo
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create data directory if it doesn't exist
	if cfg.DataDirectory != "" {
		if err := os.MkdirAll(cfg.DataDirectory, 0755); err != nil {
			return fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file with secure permissions
	if err := os.WriteFile(configPath, data, constants.ConfigFileMode); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func InitializeConfig(dataDir, ollamaEndpoint string) (*Config, error) {
	// Get a fresh copy of the default configuration
	cfg := getDefaultConfig()

	// Set custom values if provided
	if dataDir != "" {
		cfg.DataDirectory = dataDir
	} else {
		cfg.DataDirectory = GetDefaultDataDirectory()
	}

	cfg.DatabasePath = filepath.Join(cfg.DataDirectory, "notes.db")

	if ollamaEndpoint != "" {
		cfg.OllamaEndpoint = ollamaEndpoint
	}

	// Save the configuration
	if err := Save(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// InitializeConfigWithSummarization creates a new config with summarization settings
func InitializeConfigWithSummarization(dataDir, ollamaEndpoint, summarizationModel string, enableSummarization bool) (*Config, error) {
	// Get a fresh copy of the default configuration
	cfg := getDefaultConfig()

	// Set custom values if provided
	if dataDir != "" {
		cfg.DataDirectory = dataDir
	} else {
		cfg.DataDirectory = GetDefaultDataDirectory()
	}

	cfg.DatabasePath = filepath.Join(cfg.DataDirectory, "notes.db")

	if ollamaEndpoint != "" {
		cfg.OllamaEndpoint = ollamaEndpoint
	}

	// Set summarization settings
	cfg.EnableSummarization = enableSummarization
	if summarizationModel != "" {
		cfg.SummarizationModel = summarizationModel
	}

	// Save the configuration
	if err := Save(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) GetDatabasePath() string {
	if c.DatabasePath != "" {
		return c.DatabasePath
	}
	return filepath.Join(c.DataDirectory, "notes.db")
}

func (c *Config) GetOllamaAPIURL(endpoint string) string {
	return fmt.Sprintf("%s/api/%s", c.OllamaEndpoint, endpoint)
}

func (c *Config) GetVectorConfigHash() string {
	// Create a hash of vector-related configuration
	return fmt.Sprintf("%s-%d-%v", c.EmbeddingModel, c.VectorDimensions, c.EnableVectorSearch)
}

func (c *Config) NeedsReindex(oldHash string) bool {
	return c.GetVectorConfigHash() != oldHash
}

// Project Management Functions

// GetCurrentProjectPath returns the path to the current project's directory
func (c *Config) GetCurrentProjectPath() string {
	return filepath.Join(c.ProjectsDirectory, c.CurrentProject)
}

// GetCurrentProjectDatabasePath returns the path to the current project's database
func (c *Config) GetCurrentProjectDatabasePath() string {
	return filepath.Join(c.GetCurrentProjectPath(), "notes.db")
}

// GetProjectPath returns the path to a specific project's directory
func (c *Config) GetProjectPath(projectName string) string {
	return filepath.Join(c.ProjectsDirectory, projectName)
}

// GetProjectDatabasePath returns the path to a specific project's database
func (c *Config) GetProjectDatabasePath(projectName string) string {
	return filepath.Join(c.GetProjectPath(projectName), "notes.db")
}

// IsValidProjectName checks if a project name is valid (alphanumeric, dashes, underscores only)
// titleCase converts a string to title case (capitalizes first letter of each word)
func titleCase(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	result.Grow(len(s))

	capitalize := true
	for _, r := range s {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			result.WriteRune(r)
			capitalize = true
		} else if capitalize && unicode.IsLetter(r) {
			result.WriteRune(unicode.ToUpper(r))
			capitalize = false
		} else {
			result.WriteRune(unicode.ToLower(r))
			capitalize = false
		}
	}

	return result.String()
}

func IsValidProjectName(name string) bool {
	if name == "" || len(name) > 50 {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

// ListProjects returns a list of all projects
func (c *Config) ListProjects() ([]*Project, error) {
	projects := []*Project{}

	// Ensure projects directory exists
	if err := os.MkdirAll(c.ProjectsDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create projects directory: %w", err)
	}

	// Read projects directory
	entries, err := os.ReadDir(c.ProjectsDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read projects directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			projectPath := filepath.Join(c.ProjectsDirectory, entry.Name())
			project, err := c.LoadProject(entry.Name())
			if err != nil {
				// If project.json doesn't exist, create a basic project entry
				project = &Project{
					Name:        entry.Name(),
					DisplayName: titleCase(strings.ReplaceAll(entry.Name(), "_", " ")),
					CreatedAt:   entry.Name(), // Use folder name as fallback
					UpdatedAt:   entry.Name(),
				}

				// Try to get actual creation time from folder
				if info, err := entry.Info(); err == nil {
					project.CreatedAt = info.ModTime().Format(time.RFC3339)
					project.UpdatedAt = info.ModTime().Format(time.RFC3339)
				}
			}

			// Verify the project has a database file
			dbPath := filepath.Join(projectPath, "notes.db")
			if _, err := os.Stat(dbPath); err == nil {
				projects = append(projects, project)
			}
		}
	}

	return projects, nil
}

// LoadProject loads a project's metadata
func (c *Config) LoadProject(projectName string) (*Project, error) {
	projectPath := filepath.Join(c.ProjectsDirectory, projectName, "project.json")

	data, err := os.ReadFile(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project file: %w", err)
	}

	var project Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("failed to parse project file: %w", err)
	}

	return &project, nil
}

// SaveProject saves a project's metadata
func (c *Config) SaveProject(project *Project) error {
	projectDir := filepath.Join(c.ProjectsDirectory, project.Name)

	// Create project directory if it doesn't exist
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Update timestamp
	project.UpdatedAt = time.Now().Format(time.RFC3339)

	// Save project metadata
	projectPath := filepath.Join(projectDir, "project.json")
	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal project: %w", err)
	}

	if err := os.WriteFile(projectPath, data, constants.ConfigFileMode); err != nil {
		return fmt.Errorf("failed to write project file: %w", err)
	}

	return nil
}

// CreateProject creates a new project
func (c *Config) CreateProject(name, displayName, description string) (*Project, error) {
	if !IsValidProjectName(name) {
		return nil, fmt.Errorf("invalid project name: %s", name)
	}

	projectDir := filepath.Join(c.ProjectsDirectory, name)

	// Check if project already exists
	if _, err := os.Stat(projectDir); err == nil {
		return nil, fmt.Errorf("project already exists: %s", name)
	}

	// Create project directory
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create project metadata
	project := &Project{
		Name:        name,
		DisplayName: displayName,
		Description: description,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	// Save project metadata
	if err := c.SaveProject(project); err != nil {
		// Clean up on failure
		os.RemoveAll(projectDir)
		return nil, err
	}

	return project, nil
}

// SwitchProject changes the current project
func (c *Config) SwitchProject(projectName string) error {
	// Verify project exists
	projectPath := filepath.Join(c.ProjectsDirectory, projectName)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("project does not exist: %s", projectName)
	}

	// Update current project
	c.CurrentProject = projectName

	// Save configuration
	return Save(c)
}

// DeleteProject removes a project (with confirmation safeguards)
func (c *Config) DeleteProject(projectName string) error {
	if projectName == "default" {
		return fmt.Errorf("cannot delete default project")
	}

	if projectName == c.CurrentProject {
		return fmt.Errorf("cannot delete currently active project")
	}

	projectPath := filepath.Join(c.ProjectsDirectory, projectName)

	// Verify project exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("project does not exist: %s", projectName)
	}

	// Remove project directory
	return os.RemoveAll(projectPath)
}
