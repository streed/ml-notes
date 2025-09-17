package preferences

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/streed/ml-notes/internal/logger"
)

// Preference represents a key-value preference stored in SQLite
type Preference struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Type      string    `json:"type"` // "string", "json", "bool", "number"
	UpdatedAt time.Time `json:"updated_at"`
}

// PreferencesRepository handles preference storage and retrieval
type PreferencesRepository struct {
	db *sql.DB
}

// NewPreferencesRepository creates a new preferences repository
func NewPreferencesRepository(db *sql.DB) *PreferencesRepository {
	repo := &PreferencesRepository{db: db}
	if err := repo.createTable(); err != nil {
		// Log error but don't fail - preferences are optional
		fmt.Printf("Warning: Failed to create preferences table: %v\n", err)
	}
	return repo
}

// createTable creates the preferences table if it doesn't exist
func (r *PreferencesRepository) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS preferences (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		type TEXT NOT NULL DEFAULT 'string',
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_preferences_type ON preferences(type);
	CREATE INDEX IF NOT EXISTS idx_preferences_updated_at ON preferences(updated_at);
	`

	_, err := r.db.Exec(query)
	return err
}

// Set stores a preference value
func (r *PreferencesRepository) Set(key, value, valueType string) error {
	query := `
	INSERT OR REPLACE INTO preferences (key, value, type, updated_at)
	VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err := r.db.Exec(query, key, value, valueType)
	return err
}

// Get retrieves a preference value
func (r *PreferencesRepository) Get(key string) (*Preference, error) {
	query := `
	SELECT key, value, type, updated_at
	FROM preferences
	WHERE key = ?
	`

	row := r.db.QueryRow(query, key)

	var pref Preference
	err := row.Scan(&pref.Key, &pref.Value, &pref.Type, &pref.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &pref, nil
}

// GetAll retrieves all preferences
func (r *PreferencesRepository) GetAll() ([]*Preference, error) {
	query := `
	SELECT key, value, type, updated_at
	FROM preferences
	ORDER BY key
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Debug("Failed to close rows: %v", err)
		}
	}()

	var prefs []*Preference
	for rows.Next() {
		var pref Preference
		err := rows.Scan(&pref.Key, &pref.Value, &pref.Type, &pref.UpdatedAt)
		if err != nil {
			return nil, err
		}
		prefs = append(prefs, &pref)
	}

	return prefs, rows.Err()
}

// Delete removes a preference
func (r *PreferencesRepository) Delete(key string) error {
	query := `DELETE FROM preferences WHERE key = ?`
	_, err := r.db.Exec(query, key)
	return err
}

// SetString stores a string preference
func (r *PreferencesRepository) SetString(key, value string) error {
	return r.Set(key, value, "string")
}

// GetString retrieves a string preference
func (r *PreferencesRepository) GetString(key, defaultValue string) string {
	pref, err := r.Get(key)
	if err != nil {
		return defaultValue
	}
	return pref.Value
}

// SetBool stores a boolean preference
func (r *PreferencesRepository) SetBool(key string, value bool) error {
	valueStr := "false"
	if value {
		valueStr = "true"
	}
	return r.Set(key, valueStr, "bool")
}

// GetBool retrieves a boolean preference
func (r *PreferencesRepository) GetBool(key string, defaultValue bool) bool {
	pref, err := r.Get(key)
	if err != nil {
		return defaultValue
	}
	return pref.Value == "true"
}

// SetJSON stores a JSON preference (marshals the object)
func (r *PreferencesRepository) SetJSON(key string, value interface{}) error {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Set(key, string(jsonBytes), "json")
}

// GetJSON retrieves and unmarshals a JSON preference
func (r *PreferencesRepository) GetJSON(key string, target interface{}) error {
	pref, err := r.Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(pref.Value), target)
}

// HasKey checks if a preference key exists
func (r *PreferencesRepository) HasKey(key string) bool {
	_, err := r.Get(key)
	return err == nil
}
