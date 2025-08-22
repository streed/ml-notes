package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
)

// Embed the web assets at compile time
//go:embed web/templates
var templateFS embed.FS

//go:embed web/static
var staticFS embed.FS

// EmbeddedAssetProvider implements the AssetProvider interface using embedded assets
type EmbeddedAssetProvider struct{}

// GetTemplates returns parsed templates from embedded assets
func (e *EmbeddedAssetProvider) GetTemplates() (*template.Template, error) {
	return template.ParseFS(templateFS, "web/templates/*.html")
}

// GetStaticHandler returns an HTTP handler for serving static assets
func (e *EmbeddedAssetProvider) GetStaticHandler() http.Handler {
	// Get the static subdirectory from the embedded filesystem
	staticSubFS, err := fs.Sub(staticFS, "web/static")
	if err != nil {
		panic(err) // This should never happen with properly embedded assets
	}
	
	return http.FileServer(http.FS(staticSubFS))
}

// GetStaticFS returns the embedded static filesystem
func (e *EmbeddedAssetProvider) GetStaticFS() fs.FS {
	staticSubFS, err := fs.Sub(staticFS, "web/static")
	if err != nil {
		panic(err) // This should never happen with properly embedded assets
	}
	return staticSubFS
}

// HasEmbeddedAssets returns true if assets are embedded (always true in this implementation)
func (e *EmbeddedAssetProvider) HasEmbeddedAssets() bool {
	return true
}