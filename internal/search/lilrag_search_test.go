package search

import (
	"testing"

	"github.com/streed/ml-notes/internal/config"
)

func TestCreateNamespace(t *testing.T) {
	// Create a mock repository and config for testing
	cfg := &config.Config{}
	lrs := NewLilRagSearch(nil, cfg)

	tests := []struct {
		name      string
		namespace string
		expected  string
	}{
		{
			name:      "empty namespace",
			namespace: "",
			expected:  "ml-notes",
		},
		{
			name:      "project namespace",
			namespace: "myproject",
			expected:  "ml-notes-myproject",
		},
		{
			name:      "complex project name",
			namespace: "my-awesome-project",
			expected:  "ml-notes-my-awesome-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lrs.createNamespace(tt.namespace)
			if result != tt.expected {
				t.Errorf("createNamespace(%q) = %q, want %q", tt.namespace, result, tt.expected)
			}
		})
	}
}
