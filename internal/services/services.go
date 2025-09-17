package services

import (
	"github.com/streed/ml-notes/internal/autotag"
	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/models"
	"github.com/streed/ml-notes/internal/preferences"
	"github.com/streed/ml-notes/internal/search"
	"github.com/streed/ml-notes/internal/summarize"
)

// Services contains all the service dependencies
type Services struct {
	Config      *config.Config
	Notes       *NotesService
	Tags        *TagsService
	Search      *SearchService
	AutoTag     *AutoTagService
	Analyze     *AnalyzeService
	Preferences *PreferencesService
}

// NewServices creates a new services container
func NewServices(
	cfg *config.Config,
	noteRepo *models.NoteRepository,
	prefsRepo *preferences.PreferencesRepository,
	vectorSearch search.SearchProvider,
) *Services {
	// Initialize individual services
	notesService := NewNotesService(noteRepo, vectorSearch)
	tagsService := NewTagsService(noteRepo)
	searchService := NewSearchService(noteRepo, vectorSearch)
	autoTagService := NewAutoTagService(cfg, noteRepo)
	analyzeService := NewAnalyzeService(cfg, noteRepo)
	preferencesService := NewPreferencesService(prefsRepo)

	return &Services{
		Config:      cfg,
		Notes:       notesService,
		Tags:        tagsService,
		Search:      searchService,
		AutoTag:     autoTagService,
		Analyze:     analyzeService,
		Preferences: preferencesService,
	}
}

// Close cleans up any resources
func (s *Services) Close() error {
	// Add any cleanup logic here
	return nil
}

// NotesService handles note operations
type NotesService struct {
	repo         *models.NoteRepository
	vectorSearch search.SearchProvider
}

func NewNotesService(repo *models.NoteRepository, vectorSearch search.SearchProvider) *NotesService {
	return &NotesService{
		repo:         repo,
		vectorSearch: vectorSearch,
	}
}

func (s *NotesService) GetByID(id int) (*models.Note, error) {
	return s.repo.GetByID(id)
}

func (s *NotesService) List(limit, offset int) ([]*models.Note, error) {
	return s.repo.List(limit, offset)
}

func (s *NotesService) Create(title, content string, tags []string) (*models.Note, error) {
	var note *models.Note
	var err error

	if len(tags) > 0 {
		note, err = s.repo.CreateWithTags(title, content, tags)
	} else {
		note, err = s.repo.Create(title, content)
	}

	if err != nil {
		return nil, err
	}

	// Index for vector search
	if s.vectorSearch != nil {
		fullText := note.Title + " " + note.Content
		if err := s.vectorSearch.IndexNote(note.ID, fullText); err != nil {
			// Log error but don't fail the creation
		}
	}

	return note, nil
}

func (s *NotesService) Update(note *models.Note) error {
	err := s.repo.Update(note)
	if err != nil {
		return err
	}

	// Re-index for vector search
	if s.vectorSearch != nil {
		fullText := note.Title + " " + note.Content
		if err := s.vectorSearch.IndexNote(note.ID, fullText); err != nil {
			// Log error but don't fail the update
		}
	}

	return nil
}

func (s *NotesService) Delete(id int) error {
	return s.repo.Delete(id)
}

// TagsService handles tag operations
type TagsService struct {
	repo *models.NoteRepository
}

func NewTagsService(repo *models.NoteRepository) *TagsService {
	return &TagsService{repo: repo}
}

func (s *TagsService) GetAll() ([]models.Tag, error) {
	return s.repo.GetAllTags()
}

func (s *TagsService) UpdateNoteTags(noteID int, tags []string) error {
	return s.repo.UpdateTags(noteID, tags)
}

// SearchService handles search operations
type SearchService struct {
	repo         *models.NoteRepository
	vectorSearch search.SearchProvider
}

func NewSearchService(repo *models.NoteRepository, vectorSearch search.SearchProvider) *SearchService {
	return &SearchService{
		repo:         repo,
		vectorSearch: vectorSearch,
	}
}

func (s *SearchService) SearchNotes(query string, useVector bool, limit int) ([]*models.Note, error) {
	if useVector && s.vectorSearch != nil {
		return s.vectorSearch.SearchSimilar(query, limit)
	}
	return s.repo.Search(query)
}

func (s *SearchService) SearchByTags(tags []string) ([]*models.Note, error) {
	return s.repo.SearchByTags(tags)
}

// AutoTagService handles auto-tagging operations
type AutoTagService struct {
	autoTagger *autotag.AutoTagger
	repo       *models.NoteRepository
}

func NewAutoTagService(cfg *config.Config, repo *models.NoteRepository) *AutoTagService {
	return &AutoTagService{
		autoTagger: autotag.NewAutoTagger(cfg),
		repo:       repo,
	}
}

func (s *AutoTagService) IsAvailable() bool {
	return s.autoTagger.IsAvailable()
}

func (s *AutoTagService) SuggestTags(note *models.Note) ([]string, error) {
	return s.autoTagger.SuggestTags(note)
}

// AnalyzeService handles note analysis operations
type AnalyzeService struct {
	summarizer *summarize.Summarizer
	repo       *models.NoteRepository
}

func NewAnalyzeService(cfg *config.Config, repo *models.NoteRepository) *AnalyzeService {
	return &AnalyzeService{
		summarizer: summarize.NewSummarizer(cfg),
		repo:       repo,
	}
}

func (s *AnalyzeService) AnalyzeNote(note *models.Note, prompt string) (*summarize.SummaryResult, error) {
	if prompt != "" {
		return s.summarizer.SummarizeNoteWithPrompt(note, prompt)
	}
	return s.summarizer.SummarizeNote(note)
}

// PreferencesService handles user preferences
type PreferencesService struct {
	repo *preferences.PreferencesRepository
}

func NewPreferencesService(repo *preferences.PreferencesRepository) *PreferencesService {
	return &PreferencesService{repo: repo}
}

func (s *PreferencesService) GetString(key, defaultValue string) string {
	return s.repo.GetString(key, defaultValue)
}

func (s *PreferencesService) SetString(key, value string) error {
	return s.repo.SetString(key, value)
}

func (s *PreferencesService) GetBool(key string, defaultValue bool) bool {
	return s.repo.GetBool(key, defaultValue)
}

func (s *PreferencesService) SetBool(key string, value bool) error {
	return s.repo.SetBool(key, value)
}

func (s *PreferencesService) GetJSON(key string, target interface{}) error {
	return s.repo.GetJSON(key, target)
}

func (s *PreferencesService) SetJSON(key string, value interface{}) error {
	return s.repo.SetJSON(key, value)
}
