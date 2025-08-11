package renderer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ao/tfprettyplan/pkg/config"
	"github.com/ao/tfprettyplan/pkg/models"
)

func TestRenderer_RenderWithDifferentFormats(t *testing.T) {
	// Create a test summary with some resource changes
	summary := createTestSummary()

	tests := []struct {
		name         string
		outputFormat config.OutputFormat
		wantWidth    int // The expected value width in output
	}{
		{
			name:         "Standard format",
			outputFormat: config.StandardFormat,
			wantWidth:    16, // Standard format should use narrower columns
		},
		{
			name:         "Wide format",
			outputFormat: config.WideFormat,
			wantWidth:    32, // Wide format should use wider columns
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with the specified output format
			cfg := config.DefaultConfig()
			cfg.OutputFormat = tt.outputFormat
			cfg.NoColor = true // Disable colors for consistent output
			cfg.AutoDetectWidth = false // Disable auto-width for consistent tests

			// Create renderer with this config
			r := New(
				WithColor(false),
				WithConfig(cfg),
			)

			// Render to a buffer
			var buf bytes.Buffer
			r.Render(&buf, summary)
			output := buf.String()

			// Verify the summary table contains all expected actions
			if !strings.Contains(output, "Create") ||
			   !strings.Contains(output, "Update") ||
			   !strings.Contains(output, "Delete") ||
			   !strings.Contains(output, "No-op") {
				t.Errorf("Summary table missing expected actions")
			}

			// Verify that resources to delete section is present
			if !strings.Contains(output, "Resources to Delete") {
				t.Errorf("Missing 'Resources to Delete' section")
			}

			// For update resources, verify table width
			if strings.Contains(output, "Resources to Update") {
				// Check that the table has the expected width
				lines := strings.Split(output, "\n")
				
				// Find the table header line
				var headerLine string
				for _, line := range lines {
					if strings.Contains(line, "ATTRIBUTE") && strings.Contains(line, "OLD VALUE") {
						headerLine = line
						break
					}
				}
				
				if headerLine == "" {
					t.Fatalf("Could not find table header in output")
				}
				
				// Check the width of the value columns
				// The header format is now "  │ ATTRIBUTE │ OLD VALUE │ NEW VALUE │"
				// We're looking at the space allocated for OLD VALUE and NEW VALUE
				
				parts := strings.Split(headerLine, "│")
				if len(parts) != 5 {
					t.Fatalf("Unexpected header format: %s", headerLine)
				}
				
				oldValuePart := parts[2]
				oldValueWidth := len(oldValuePart) - 2 // Subtract 2 for the spaces
				
				// Verify the width matches our expectation
				if oldValueWidth != tt.wantWidth {
					t.Errorf("Value column width = %d, want %d for %s", 
						oldValueWidth, tt.wantWidth, tt.outputFormat)
				}
				
				// Also check if the output contains wide values for wide format
				if tt.outputFormat == config.WideFormat {
					// In wide format, longer values should be displayed without truncation
					if !strings.Contains(output, "This is a longer description") {
						t.Errorf("Wide format should show longer values without truncation")
					}
				}
			}
		})
	}
}

// TestRenderer_RenderDeletedAttributes tests that resources to be deleted are properly displayed
func TestRenderer_RenderDeletedAttributes(t *testing.T) {
	// Create a test summary with a resource to delete
	summary := &models.PlanSummary{
		DeleteCount: 1,
		ResourceChanges: []models.ResourceChange{
			{
				Address:    "aws_iam_role.test",
				Type:       "aws_iam_role",
				Name:       "test",
				ChangeType: models.Delete,
				Before: map[string]any{
					"name":              "test-role",
					"assume_role_policy": "{\"Version\":\"2012-10-17\"}",
					"tags": map[string]any{
						"Name":        "Test Role",
						"Environment": "dev",
					},
				},
				After:        nil,
				BeforeValues: map[string]string{
					"name":              "test-role",
					"assume_role_policy": "{\"Version\":\"2012-10-17\"}",
					"tags.Name":        "Test Role",
					"tags.Environment": "dev",
				},
				AfterValues:  map[string]string{},
			},
		},
	}

	// Create renderer with default config
	r := New(WithColor(false))

	// Render to a buffer
	var buf bytes.Buffer
	r.Render(&buf, summary)
	output := buf.String()

	// Verify the output contains expected elements
	expectedElements := []string{
		"Resources to Delete",
		"aws_iam_role.test",
		"CURRENT VALUE (WILL BE DESTROYED)",
		"test-role",
		"tags.Name",
		"Test Role",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', but it didn't", expected)
		}
	}
}

// createTestSummary creates a test plan summary with various resource changes
func createTestSummary() *models.PlanSummary {
	summary := &models.PlanSummary{
		AddCount:    1,
		ChangeCount: 1,
		DeleteCount: 1,
		ResourceChanges: []models.ResourceChange{
			{
				Address:    "aws_instance.example",
				Type:       "aws_instance",
				Name:       "example",
				ChangeType: models.Create,
				Before:     nil,
				After: map[string]any{
					"ami":          "ami-123456",
					"instance_type": "t2.micro",
				},
				BeforeValues: map[string]string{},
				AfterValues: map[string]string{
					"ami":          "ami-123456",
					"instance_type": "t2.micro",
				},
			},
			{
				Address:    "aws_s3_bucket.logs",
				Type:       "aws_s3_bucket",
				Name:       "logs",
				ChangeType: models.Update,
				Before: map[string]any{
					"acl":           "private",
					"force_destroy": false,
					"description":   "This is a short description",
				},
				After: map[string]any{
					"acl":           "public-read",
					"force_destroy": true,
					"description":   "This is a longer description that should be truncated in standard mode but visible in wide mode",
				},
				BeforeValues: map[string]string{
					"acl":           "private",
					"force_destroy": "false",
					"description":   "This is a short description",
				},
				AfterValues: map[string]string{
					"acl":           "public-read",
					"force_destroy": "true",
					"description":   "This is a longer description that should be truncated in standard mode but visible in wide mode",
				},
			},
			{
				Address:    "aws_iam_role.lambda",
				Type:       "aws_iam_role",
				Name:       "lambda",
				ChangeType: models.Delete,
				Before: map[string]any{
					"name": "lambda-role",
				},
				After:        nil,
				BeforeValues: map[string]string{"name": "lambda-role"},
				AfterValues:  map[string]string{},
			},
		},
	}
	return summary
}

func TestTruncateValue(t *testing.T) {
	r := New() // Use default config

	tests := []struct {
		name      string
		value     string
		maxWidth  int
		want      string
		wantWidth int
	}{
		{
			name:      "Short value not truncated",
			value:     "short",
			maxWidth:  10,
			want:      "short",
			wantWidth: 5,
		},
		{
			name:      "Long value truncated in middle",
			value:     "this is a very long value that should be truncated",
			maxWidth:  20,
			want:      "this is a...runcated",
			wantWidth: 20,
		},
		{
			name:      "Path value smart truncation",
			value:     "/very/long/path/with/many/nested/directories/file.txt",
			maxWidth:  25,
			want:      "/very/long/.../file.txt",
			wantWidth: 25,
		},
		{
			name:      "JSON-like value truncation",
			value:     "{\"key\":\"value\",\"nested\":{\"prop\":\"too long to display fully\"}}",
			maxWidth:  20,
			want:      "{\"key\":\"value\"...}}",
			wantWidth: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.truncateValue(tt.value, tt.maxWidth)
			
			if got != tt.want {
				t.Errorf("truncateValue() got = %v, want %v", got, tt.want)
			}
			
			if len(got) > tt.maxWidth {
				t.Errorf("truncateValue() returned value longer than maxWidth: len=%d, maxWidth=%d", 
					len(got), tt.maxWidth)
			}
		})
	}
}

