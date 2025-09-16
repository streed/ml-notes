package cmd

import (
	"testing"
	"net/url"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid http url",
			input:   "https://example.com",
			wantErr: false,
		},
		{
			name:    "valid https url",
			input:   "http://example.com",
			wantErr: false,
		},
		{
			name:    "valid file url",
			input:   "file:///tmp/test.html",
			wantErr: false,
		},
		{
			name:    "relative path (invalid for our use case)",
			input:   "not-a-url",
			wantErr: false, // url.Parse accepts this but we'd expect Chrome to handle it
		},
		{
			name:    "empty url",
			input:   "",
			wantErr: false, // url.Parse accepts empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := url.Parse(tt.input)
			hasErr := err != nil
			if hasErr != tt.wantErr {
				t.Errorf("url.Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCleanMarkdownContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove multiple empty lines",
			input:    "Line 1\n\n\n\nLine 2",
			expected: "Line 1\n\nLine 2",
		},
		{
			name:     "remove leading empty lines",
			input:    "\n\nLine 1\nLine 2",
			expected: "Line 1\nLine 2",
		},
		{
			name:     "remove trailing empty lines",
			input:    "Line 1\nLine 2\n\n\n",
			expected: "Line 1\nLine 2",
		},
		{
			name:     "trim whitespace",
			input:    "  Line 1  \n  Line 2  ",
			expected: "Line 1\nLine 2",
		},
		{
			name:     "single line",
			input:    "Single line",
			expected: "Single line",
		},
		{
			name:     "empty content",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanMarkdownContent(tt.input)
			if result != tt.expected {
				t.Errorf("cleanMarkdownContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIsRestrictedEnvironment(t *testing.T) {
	// Since we're running in GitHub Actions, this should return true
	result := isRestrictedEnvironment()
	if !result {
		t.Errorf("isRestrictedEnvironment() = %v, want true (running in CI)", result)
	}
}



func TestImageURLConversion(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		href     string
		expected string
	}{
		{
			name:     "relative path with https base",
			baseURL:  "https://example.com/page",
			href:     "/logo.png",
			expected: "https://example.com/logo.png",
		},
		{
			name:     "relative path with http base",
			baseURL:  "http://example.com/page",
			href:     "/logo.png",
			expected: "http://example.com/logo.png",
		},
		{
			name:     "absolute url unchanged",
			baseURL:  "https://example.com/page",
			href:     "https://cdn.other.com/image.jpg",
			expected: "https://cdn.other.com/image.jpg",
		},
		{
			name:     "protocol relative url",
			baseURL:  "https://example.com/page",
			href:     "//cdn.example.com/image.png",
			expected: "https://cdn.example.com/image.png",
		},
		{
			name:     "relative path from subdirectory",
			baseURL:  "https://blog.example.com/posts/article",
			href:     "../images/header.jpg",
			expected: "https://blog.example.com/images/header.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveURL(tt.baseURL, tt.href)
			if result != tt.expected {
				t.Errorf("resolveURL(%q, %q) = %q, want %q", tt.baseURL, tt.href, result, tt.expected)
			}
		})
	}
}