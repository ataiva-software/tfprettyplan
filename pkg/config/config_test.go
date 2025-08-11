package config

import (
	"testing"
)

func TestGetTableConfig(t *testing.T) {
	tests := []struct {
		name            string
		outputFormat    OutputFormat
		autoDetectWidth bool
		maxWidth        int
		wantAttrWidth   int
		wantValueWidth  int
	}{
		{
			name:            "Standard format",
			outputFormat:    StandardFormat,
			autoDetectWidth: false,
			maxWidth:        80,
			wantAttrWidth:   13,
			wantValueWidth:  16,
		},
		{
			name:            "Wide format",
			outputFormat:    WideFormat,
			autoDetectWidth: false,
			maxWidth:        120,
			wantAttrWidth:   13,
			wantValueWidth:  32,
		},
		{
			name:            "Auto-detect width with standard format",
			outputFormat:    StandardFormat,
			autoDetectWidth: true,
			maxWidth:        100,
			wantAttrWidth:   27,
			wantValueWidth:  31,
		},
		{
			name:            "Auto-detect width with wide format",
			outputFormat:    WideFormat,
			autoDetectWidth: true,
			maxWidth:        100,
			wantAttrWidth:   27,
			wantValueWidth:  31,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				OutputFormat:    tt.outputFormat,
				AutoDetectWidth: tt.autoDetectWidth,
				MaxWidth:        tt.maxWidth,
			}
			
			tableConfig := cfg.GetTableConfig()
			
			if tableConfig.MaxAttributeWidth != tt.wantAttrWidth {
				t.Errorf("GetTableConfig().MaxAttributeWidth = %v, want %v", 
					tableConfig.MaxAttributeWidth, tt.wantAttrWidth)
			}
			
			if tableConfig.MaxValueWidth != tt.wantValueWidth {
				t.Errorf("GetTableConfig().MaxValueWidth = %v, want %v", 
					tableConfig.MaxValueWidth, tt.wantValueWidth)
			}
		})
	}
}

