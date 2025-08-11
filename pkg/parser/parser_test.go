package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ao/tfprettyplan/pkg/models"
)

func TestParseFile(t *testing.T) {
	// Create a temporary test file with valid content
	tempDir := t.TempDir()
	validPlanPath := filepath.Join(tempDir, "valid-plan.json")

	// Create a sample plan based on examples/sample-plan.json
	samplePlan := createSamplePlan()
	planData, err := json.MarshalIndent(samplePlan, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal sample plan: %v", err)
	}

	if err := os.WriteFile(validPlanPath, planData, 0644); err != nil {
		t.Fatalf("Failed to write test plan file: %v", err)
	}

	// Create an invalid JSON file
	invalidPlanPath := filepath.Join(tempDir, "invalid-plan.json")
	if err := os.WriteFile(invalidPlanPath, []byte("this is not valid JSON"), 0644); err != nil {
		t.Fatalf("Failed to write invalid test plan file: %v", err)
	}

	// Create an empty file
	emptyPlanPath := filepath.Join(tempDir, "empty-plan.json")
	if err := os.WriteFile(emptyPlanPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to write empty test plan file: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "Valid plan file",
			filePath: validPlanPath,
			wantErr:  false,
		},
		{
			name:        "Invalid JSON file",
			filePath:    invalidPlanPath,
			wantErr:     true,
			errContains: "invalid plan file",
		},
		{
			name:        "Empty file",
			filePath:    emptyPlanPath,
			wantErr:     true,
			errContains: "empty plan file",
		},
		{
			name:        "Non-existent file",
			filePath:    "non-existent-file.json",
			wantErr:     true,
			errContains: "failed to read plan file",
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, err := p.ParseFile(tt.filePath)

			// Check if error is expected
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseFile() expected error but got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ParseFile() error = %v, want it to contain %v", err, tt.errContains)
				}
				return
			}

			// No error expected
			if err != nil {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For valid plan, check that the summary contains expected resources
			if summary == nil {
				t.Errorf("ParseFile() returned nil summary")
				return
			}

			// Check resource counts
			expectedCounts := struct {
				create int
				update int
				delete int
			}{2, 1, 1} // Based on sample-plan.json

			if summary.AddCount != expectedCounts.create {
				t.Errorf("ParseFile() summary.AddCount = %v, want %v", summary.AddCount, expectedCounts.create)
			}
			if summary.ChangeCount != expectedCounts.update {
				t.Errorf("ParseFile() summary.ChangeCount = %v, want %v", summary.ChangeCount, expectedCounts.update)
			}
			if summary.DeleteCount != expectedCounts.delete {
				t.Errorf("ParseFile() summary.DeleteCount = %v, want %v", summary.DeleteCount, expectedCounts.delete)
			}
		})
	}
}

func TestParseJSON(t *testing.T) {
	// Create a sample plan based on examples/sample-plan.json
	samplePlan := createSamplePlan()
	validPlanData, err := json.MarshalIndent(samplePlan, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal sample plan: %v", err)
	}

	tests := []struct {
		name        string
		data        []byte
		wantErr     bool
		errContains string
	}{
		{
			name:    "Valid JSON data",
			data:    validPlanData,
			wantErr: false,
		},
		{
			name:        "Invalid JSON data",
			data:        []byte("this is not valid JSON"),
			wantErr:     true,
			errContains: "invalid JSON input",
		},
		{
			name:        "Empty JSON data",
			data:        []byte{},
			wantErr:     true,
			errContains: "empty JSON input",
		},
		{
			name:        "Whitespace only",
			data:        []byte("   \n   "),
			wantErr:     true,
			errContains: "empty input",
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, err := p.ParseJSON(tt.data)

			// Check if error is expected
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseJSON() expected error but got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ParseJSON() error = %v, want it to contain %v", err, tt.errContains)
				}
				return
			}

			// No error expected
			if err != nil {
				t.Errorf("ParseJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For valid plan, check that the summary contains expected resources
			if summary == nil {
				t.Errorf("ParseJSON() returned nil summary")
				return
			}

			// Check resource counts
			expectedCounts := struct {
				create int
				update int
				delete int
			}{2, 1, 1} // Based on sample-plan.json

			if summary.AddCount != expectedCounts.create {
				t.Errorf("ParseJSON() summary.AddCount = %v, want %v", summary.AddCount, expectedCounts.create)
			}
			if summary.ChangeCount != expectedCounts.update {
				t.Errorf("ParseJSON() summary.ChangeCount = %v, want %v", summary.ChangeCount, expectedCounts.update)
			}
			if summary.DeleteCount != expectedCounts.delete {
				t.Errorf("ParseJSON() summary.DeleteCount = %v, want %v", summary.DeleteCount, expectedCounts.delete)
			}
		})
	}
}

func TestProcessResourceChange(t *testing.T) {
	p := New()

	tests := []struct {
		name         string
		resourceData map[string]interface{}
		want         models.ChangeType
		wantErr      bool
	}{
		{
			name: "Create action",
			resourceData: map[string]interface{}{
				"address": "aws_instance.example",
				"type":    "aws_instance",
				"change": map[string]interface{}{
					"actions": []interface{}{"create"},
					"before":  nil,
					"after":   map[string]interface{}{"ami": "ami-123"},
				},
			},
			want:    models.Create,
			wantErr: false,
		},
		{
			name: "Update action",
			resourceData: map[string]interface{}{
				"address": "aws_instance.example",
				"type":    "aws_instance",
				"change": map[string]interface{}{
					"actions": []interface{}{"update"},
					"before":  map[string]interface{}{"ami": "ami-123"},
					"after":   map[string]interface{}{"ami": "ami-456"},
				},
			},
			want:    models.Update,
			wantErr: false,
		},
		{
			name: "Delete action",
			resourceData: map[string]interface{}{
				"address": "aws_instance.example",
				"type":    "aws_instance",
				"change": map[string]interface{}{
					"actions": []interface{}{"delete"},
					"before":  map[string]interface{}{"ami": "ami-123"},
					"after":   nil,
				},
			},
			want:    models.Delete,
			wantErr: false,
		},
		{
			name: "No-op action",
			resourceData: map[string]interface{}{
				"address": "aws_instance.example",
				"type":    "aws_instance",
				"change": map[string]interface{}{
					"actions": []interface{}{"no-op"},
					"before":  map[string]interface{}{"ami": "ami-123"},
					"after":   map[string]interface{}{"ami": "ami-123"},
				},
			},
			want:    models.NoOp,
			wantErr: false,
		},
		{
			name: "Missing address",
			resourceData: map[string]interface{}{
				"type": "aws_instance",
				"change": map[string]interface{}{
					"actions": []interface{}{"create"},
				},
			},
			want:    models.NoOp,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			change, err := p.processResourceChange(tt.resourceData)

			if tt.wantErr {
				if err == nil {
					t.Errorf("processResourceChange() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("processResourceChange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if change.ChangeType != tt.want {
				t.Errorf("processResourceChange() changeType = %v, want %v", change.ChangeType, tt.want)
			}
		})
	}
}

// Helper function to create a sample plan similar to examples/sample-plan.json
func createSamplePlan() map[string]interface{} {
	return map[string]interface{}{
		"format_version":    "1.0",
		"terraform_version": "1.5.0",
		"variables": map[string]interface{}{
			"environment": map[string]interface{}{"value": "dev"},
			"region":      map[string]interface{}{"value": "us-west-2"},
		},
		"resource_changes": []interface{}{
			map[string]interface{}{
				"address":       "aws_instance.example",
				"mode":          "managed",
				"type":          "aws_instance",
				"name":          "example",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": map[string]interface{}{
					"actions": []interface{}{"create"},
					"before":  nil,
					"after": map[string]interface{}{
						"ami":          "ami-0c55b159cbfafe1f0",
						"instance_type": "t2.micro",
						"tags": map[string]interface{}{
							"Name":        "Example Instance",
							"Environment": "dev",
						},
					},
				},
			},
			map[string]interface{}{
				"address":       "aws_s3_bucket.logs",
				"mode":          "managed",
				"type":          "aws_s3_bucket",
				"name":          "logs",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": map[string]interface{}{
					"actions": []interface{}{"update"},
					"before": map[string]interface{}{
						"acl":           "private",
						"bucket":        "example-logs-bucket",
						"force_destroy": false,
						"tags": map[string]interface{}{
							"Name":        "Log Bucket",
							"Environment": "dev",
						},
					},
					"after": map[string]interface{}{
						"acl":           "public-read",
						"bucket":        "example-logs-bucket",
						"force_destroy": true,
						"tags": map[string]interface{}{
							"Name":        "Logs Bucket",
							"Environment": "dev",
						},
					},
				},
			},
			map[string]interface{}{
				"address":       "aws_security_group.allow_ssh",
				"mode":          "managed",
				"type":          "aws_security_group",
				"name":          "allow_ssh",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": map[string]interface{}{
					"actions": []interface{}{"create"},
					"before":  nil,
					"after": map[string]interface{}{
						"description": "Allow SSH inbound traffic",
						"name":        "allow_ssh",
						"ingress": []interface{}{
							map[string]interface{}{
								"description": "SSH from VPC",
								"from_port":   22,
								"to_port":     22,
								"protocol":    "tcp",
								"cidr_blocks": []interface{}{"10.0.0.0/16"},
							},
						},
						"tags": map[string]interface{}{
							"Name": "allow_ssh",
						},
					},
				},
			},
			map[string]interface{}{
				"address":       "aws_iam_role.lambda",
				"mode":          "managed",
				"type":          "aws_iam_role",
				"name":          "lambda",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": map[string]interface{}{
					"actions": []interface{}{"delete"},
					"before": map[string]interface{}{
						"name":              "lambda-execution-role",
						"assume_role_policy": "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"lambda.amazonaws.com\"},\"Action\":\"sts:AssumeRole\"}]}",
						"tags": map[string]interface{}{
							"Name":        "Lambda Execution Role",
							"Environment": "dev",
						},
					},
					"after": nil,
				},
			},
		},
		"configuration": map[string]interface{}{},
	}
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

