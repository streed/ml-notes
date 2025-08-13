package constants

// Boolean string values
const (
	BoolTrue  = "true"
	BoolFalse = "false"
	BoolYes   = "yes"
	BoolNo    = "no"
	BoolOne   = "1"
	BoolZero  = "0"
)

// Magic numbers for various operations
const (
	// Display limits
	DefaultSearchLimit  = 10
	DefaultListLimit    = 20
	DefaultExportLimit  = 50
	MaxSummaryNotes     = 20
	TotalSummaryNotes   = 100
	
	// Text truncation lengths
	PreviewLength       = 100
	SearchPreviewLength = 150
	ContextBefore       = 50
	ContextAfter        = 100
	ShortPreviewLength  = 80
	
	// Time calculations
	HoursPerDay = 24
	
	// Embedding calculations
	BytesPerFloat32 = 4
	HashMultiplier  = 31
	HashModulo      = 100
)

// File permissions
const (
	ConfigFileMode = 0600 // Secure file permissions for config
)