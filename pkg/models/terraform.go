package models

// ChangeType represents the type of change for a resource
type ChangeType string

const (
	// Create represents a resource that will be created
	Create ChangeType = "create"
	// Update represents a resource that will be updated
	Update ChangeType = "update"
	// Delete represents a resource that will be deleted
	Delete ChangeType = "delete"
	// NoOp represents a resource with no changes
	NoOp ChangeType = "no-op"
)

// ResourceChange represents a change to a Terraform resource
type ResourceChange struct {
	Address      string            // Resource address (e.g., aws_instance.example)
	Type         string            // Resource type (e.g., aws_instance)
	Name         string            // Resource name (e.g., example)
	ChangeType   ChangeType        // Type of change (create, update, delete)
	Before       map[string]any    // Resource state before change
	After        map[string]any    // Resource state after change
	BeforeValues map[string]string // Formatted values before change
	AfterValues  map[string]string // Formatted values after change
	Module       string            // Module path if applicable
}

// PlanSummary represents a summary of all changes in a Terraform plan
type PlanSummary struct {
	ResourceChanges []ResourceChange
	AddCount        int // Number of resources to be created
	ChangeCount     int // Number of resources to be modified
	DeleteCount     int // Number of resources to be deleted
	NoOpCount       int // Number of resources with no changes
}

// TerraformPlan represents the structure of a Terraform plan JSON file
type TerraformPlan struct {
	FormatVersion    string                   `json:"format_version"`
	TerraformVersion string                   `json:"terraform_version"`
	Variables        map[string]any           `json:"variables"`
	PlannedValues    map[string]any           `json:"planned_values"`
	ResourceChanges  []map[string]interface{} `json:"resource_changes"`
	Configuration    map[string]any           `json:"configuration"`
}
