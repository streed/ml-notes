package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Project represents a ml-notes project
type Project struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Path         string    `json:"path"`          // Directory where project files are stored
	DatabasePath string    `json:"database_path"` // Path to project's database
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GetID returns the project ID
func (p *Project) GetID() string {
	return p.ID
}

// GetName returns the project name
func (p *Project) GetName() string {
	return p.Name
}

// GetDescription returns the project description
func (p *Project) GetDescription() string {
	return p.Description
}

// GetDatabasePath returns the project database path
func (p *Project) GetDatabasePath() string {
	return p.DatabasePath
}

// ProjectManager handles project operations
type ProjectManager struct {
	configDir    string
	dataDir      string
	projectsFile string
	projects     map[string]*Project
}

// NewProjectManager creates a new project manager
func NewProjectManager(configDir, dataDir string) (*ProjectManager, error) {
	projectsFile := filepath.Join(configDir, "projects.json")

	pm := &ProjectManager{
		configDir:    configDir,
		dataDir:      dataDir,
		projectsFile: projectsFile,
		projects:     make(map[string]*Project),
	}

	if err := pm.loadProjects(); err != nil {
		return nil, fmt.Errorf("failed to load projects: %w", err)
	}

	return pm, nil
}

// loadProjects loads projects from the projects.json file
func (pm *ProjectManager) loadProjects() error {
	if _, err := os.Stat(pm.projectsFile); os.IsNotExist(err) {
		// Create default project if no projects exist
		return pm.createDefaultProject()
	}

	data, err := os.ReadFile(pm.projectsFile)
	if err != nil {
		return fmt.Errorf("failed to read projects file: %w", err)
	}

	var projects []*Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return fmt.Errorf("failed to parse projects file: %w", err)
	}

	for _, project := range projects {
		pm.projects[project.ID] = project
	}

	return nil
}

// saveProjects saves projects to the projects.json file
func (pm *ProjectManager) saveProjects() error {
	// Ensure config directory exists
	if err := os.MkdirAll(pm.configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	projects := make([]*Project, 0, len(pm.projects))
	for _, project := range pm.projects {
		projects = append(projects, project)
	}

	// Sort projects by name for consistent output
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal projects: %w", err)
	}

	return os.WriteFile(pm.projectsFile, data, 0o644)
}

// createDefaultProject creates a default "default" project
func (pm *ProjectManager) createDefaultProject() error {
	project := &Project{
		ID:           "default",
		Name:         "Default",
		Description:  "Default project for ml-notes",
		Path:         filepath.Join(pm.dataDir, "projects", "default"),
		DatabasePath: filepath.Join(pm.dataDir, "projects", "default", "notes.db"),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Create project directory
	if err := os.MkdirAll(project.Path, 0o755); err != nil {
		return fmt.Errorf("failed to create default project directory: %w", err)
	}

	pm.projects[project.ID] = project
	return pm.saveProjects()
}

// CreateProject creates a new project
func (pm *ProjectManager) CreateProject(name, description string) (*Project, error) {
	// Generate ID from name (lowercase, replace spaces with dashes)
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "_", "-")
	// Remove any non-alphanumeric characters except dashes
	var cleanID strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleanID.WriteRune(r)
		}
	}
	id = cleanID.String()

	// Ensure ID is unique
	originalID := id
	counter := 1
	for pm.projects[id] != nil {
		id = fmt.Sprintf("%s-%d", originalID, counter)
		counter++
	}

	project := &Project{
		ID:           id,
		Name:         name,
		Description:  description,
		Path:         filepath.Join(pm.dataDir, "projects", id),
		DatabasePath: filepath.Join(pm.dataDir, "projects", id, "notes.db"),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Create project directory
	if err := os.MkdirAll(project.Path, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create project directory: %w", err)
	}

	pm.projects[project.ID] = project
	if err := pm.saveProjects(); err != nil {
		return nil, err
	}

	return project, nil
}

// GetProject returns a project by ID
func (pm *ProjectManager) GetProject(id string) (*Project, error) {
	project, exists := pm.projects[id]
	if !exists {
		return nil, fmt.Errorf("project '%s' not found", id)
	}
	return project, nil
}

// ListProjects returns all projects
func (pm *ProjectManager) ListProjects() []*Project {
	projects := make([]*Project, 0, len(pm.projects))
	for _, project := range pm.projects {
		projects = append(projects, project)
	}

	// Sort by name
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects
}

// ListProjects returns all projects as interface{} slice for migration compatibility
func (pm *ProjectManager) ListProjectsForMigration() []interface{} {
	projects := pm.ListProjects()
	result := make([]interface{}, len(projects))
	for i, project := range projects {
		result[i] = project
	}
	return result
}

// DeleteProject deletes a project
func (pm *ProjectManager) DeleteProject(id string) error {
	_, exists := pm.projects[id]
	if !exists {
		return fmt.Errorf("project '%s' not found", id)
	}

	// Cannot delete the default project
	if id == "default" {
		return fmt.Errorf("cannot delete the default project")
	}

	// Remove project files (optional - could be made configurable)
	// For safety, we'll leave the files and only remove from the registry

	delete(pm.projects, id)
	return pm.saveProjects()
}

// UpdateProject updates project metadata
func (pm *ProjectManager) UpdateProject(id string, name, description string) error {
	project, exists := pm.projects[id]
	if !exists {
		return fmt.Errorf("project '%s' not found", id)
	}

	if name != "" {
		project.Name = name
	}
	if description != "" {
		project.Description = description
	}

	project.UpdatedAt = time.Now()
	return pm.saveProjects()
}

// GetProjectDatabasePath returns the database path for a project
func (pm *ProjectManager) GetProjectDatabasePath(id string) (string, error) {
	project, err := pm.GetProject(id)
	if err != nil {
		return "", err
	}
	return project.DatabasePath, nil
}

// MigrateFromLegacyDatabase migrates an existing database to the default project
func (pm *ProjectManager) MigrateFromLegacyDatabase(legacyDBPath string) error {
	// Check if legacy database exists
	if _, err := os.Stat(legacyDBPath); os.IsNotExist(err) {
		return nil // Nothing to migrate
	}

	defaultProject, err := pm.GetProject("default")
	if err != nil {
		return err
	}

	// Check if the target database already exists
	if _, err := os.Stat(defaultProject.DatabasePath); err == nil {
		// Target already exists, don't overwrite
		return nil
	}

	// Copy legacy database to default project
	return copyFile(legacyDBPath, defaultProject.DatabasePath)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Write to destination
	return os.WriteFile(dst, data, 0o644)
}
