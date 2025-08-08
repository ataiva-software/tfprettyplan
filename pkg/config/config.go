package config

// OutputFormat represents the format of the output
type OutputFormat string

const (
	// StandardFormat is the default output format with fixed column widths
	StandardFormat OutputFormat = "standard"
	// WideFormat is an expanded output format with wider columns
	WideFormat OutputFormat = "wide"
)

// Config holds the configuration for the application
type Config struct {
	// OutputFormat specifies the format of the output (standard, wide, owide)
	OutputFormat OutputFormat
	// NoColor disables color output
	NoColor bool
	// MaxWidth is the maximum width of the terminal
	MaxWidth int
	// AutoDetectWidth enables automatic detection of terminal width
	AutoDetectWidth bool
}

// TableConfig holds the configuration for table rendering
type TableConfig struct {
	// MaxAttributeWidth is the maximum width for attribute names
	MaxAttributeWidth int
	// MaxValueWidth is the maximum width for attribute values
	MaxValueWidth int
	// MinValueWidth is the minimum width for attribute values
	MinValueWidth int
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		OutputFormat:    StandardFormat,
		NoColor:         false,
		MaxWidth:        80,
		AutoDetectWidth: true,
	}
}

// GetTableConfig returns the table configuration based on the output format and terminal width
func (c *Config) GetTableConfig() *TableConfig {
	tc := &TableConfig{
		MaxAttributeWidth: 13, // Default from current implementation
		MinValueWidth:     10,
	}

	// Adjust column widths based on output format and terminal width
	switch c.OutputFormat {
	case WideFormat:
		tc.MaxValueWidth = 32
	default:
		tc.MaxValueWidth = 16 // Default from current implementation
	}

	// If auto-detect is enabled and we have terminal width, adjust dynamically
	if c.AutoDetectWidth && c.MaxWidth > 0 {
		// Calculate available width after accounting for table borders and padding
		// Table format: | ATTRIBUTE | OLD VALUE | NEW VALUE |
		// Borders and padding: 2 + 2 + 2 + 2 + 2 = 10 characters
		availableWidth := c.MaxWidth - 10

		// Attribute column gets 30% of space, each value column gets 35%
		if availableWidth > 60 { // Only adjust if we have reasonable space
			tc.MaxAttributeWidth = (availableWidth * 30) / 100
			tc.MaxValueWidth = (availableWidth * 35) / 100
		}
	}

	return tc
}
