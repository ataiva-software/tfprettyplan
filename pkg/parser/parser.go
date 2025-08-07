package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ao/tfprettyplan/pkg/models"
)

// Parser is responsible for parsing Terraform plan files
type Parser struct{}

// New creates a new Parser
func New() *Parser {
	return &Parser{}
}

// validateJSON does basic validation of JSON data before parsing
func (p *Parser) validateJSON(data []byte) error {
	// Check for empty input
	if len(data) == 0 {
		return fmt.Errorf("empty input: no JSON data provided")
	}

	// Trim whitespace
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return fmt.Errorf("empty input: JSON data contains only whitespace")
	}

	// Check if it starts with { and ends with }
	if trimmed[0] != '{' || trimmed[len(trimmed)-1] != '}' {
		return fmt.Errorf("malformed JSON: input does not appear to be a valid JSON object")
	}

	return nil
}

// checkForTerraformProviderErrors checks if the JSON data contains Terraform provider errors
func (p *Parser) checkForTerraformProviderErrors(data []byte) error {
	// Check for common Terraform provider error messages in the JSON data
	if bytes.Contains(data, []byte("Failed to load plugin schemas")) {
		return fmt.Errorf("detected Terraform provider error: failed to load plugin schemas. " +
			"Please ensure you're running this command in the directory where the Terraform configuration exists " +
			"and that 'terraform init' has been run. See docs/terraform-workflow.md for more information")
	}

	if bytes.Contains(data, []byte("Error: No value for required variable")) {
		return fmt.Errorf("detected Terraform error: missing required variables. " +
			"Please provide all required variables when generating the plan")
	}
	
	if bytes.Contains(data, []byte("unavailable provider")) {
		return fmt.Errorf("detected Terraform provider error: unavailable provider. " +
			"Please run 'terraform init' in the directory where the Terraform configuration exists " +
			"before generating the plan JSON. See docs/terraform-workflow.md for more information")
	}

	if bytes.Contains(data, []byte("Could not load the schema for provider")) {
		return fmt.Errorf("detected Terraform provider schema error. " +
			"Please ensure you're running this command in the directory where the Terraform configuration exists " +
			"and that 'terraform init' has been run. See docs/terraform-workflow.md for more information")
	}

	if bytes.Contains(data, []byte("Error: Could not load plugin")) {
		return fmt.Errorf("detected Terraform plugin error. " +
			"Please ensure you have the required provider plugins installed with 'terraform init'. " +
			"See docs/terraform-workflow.md for more information")
	}

	if bytes.Contains(data, []byte("Error: Provider configuration not present")) {
		return fmt.Errorf("detected Terraform provider configuration error. " +
			"Provider configuration is missing or incomplete. " +
			"Please ensure your Terraform configuration includes the necessary provider blocks")
	}

	if bytes.Contains(data, []byte("Error: Invalid provider configuration")) {
		return fmt.Errorf("detected invalid Terraform provider configuration. " +
			"Please check your provider configuration for syntax errors or invalid settings")
	}

	// Check for general Terraform errors that might appear in the output
	if bytes.Contains(data, []byte("Error: ")) && !bytes.Contains(data, []byte("{")) {
		// This might be a Terraform error message rather than valid JSON
		return fmt.Errorf("detected Terraform error output instead of valid JSON plan. " +
			"Please follow the workflow in docs/terraform-workflow.md to generate a valid plan JSON file")
	}

	return nil
}

// ParseFile parses a Terraform plan file and returns a PlanSummary
func (p *Parser) ParseFile(path string) (*models.PlanSummary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read plan file: %w", err)
	}

	// Check file size
	if len(data) == 0 {
		return nil, fmt.Errorf("empty plan file: %s. Please ensure the file contains valid Terraform plan JSON", path)
	}

	// Check for Terraform provider errors
	if err := p.checkForTerraformProviderErrors(data); err != nil {
		return nil, err
	}

	// Validate JSON before parsing
	if err := p.validateJSON(data); err != nil {
		return nil, fmt.Errorf("invalid plan file: %s. %w", path, err)
	}

	return p.ParseJSON(data)
}

// ParseJSON parses Terraform plan JSON data and returns a PlanSummary
func (p *Parser) ParseJSON(data []byte) (*models.PlanSummary, error) {
	// Check for empty input
	if len(data) == 0 {
		return nil, fmt.Errorf("empty JSON input. Please provide valid Terraform plan JSON data")
	}

	// Check for Terraform provider errors
	if err := p.checkForTerraformProviderErrors(data); err != nil {
		return nil, err
	}

	// Validate JSON before parsing
	if err := p.validateJSON(data); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}

	var plan models.TerraformPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		// Provide more context for common JSON parsing errors
		if strings.Contains(err.Error(), "unexpected end of JSON input") {
			return nil, fmt.Errorf("failed to parse JSON: unexpected end of JSON input. The JSON data appears to be truncated or incomplete. " +
				"Please ensure the Terraform plan was generated correctly. See docs/terraform-workflow.md for more information")
		}
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate required fields
	if len(plan.ResourceChanges) == 0 {
		// Still create an empty summary rather than failing
		return &models.PlanSummary{
			ResourceChanges: []models.ResourceChange{},
		}, nil
	}

	summary := &models.PlanSummary{
		ResourceChanges: make([]models.ResourceChange, 0, len(plan.ResourceChanges)),
	}

	for _, rc := range plan.ResourceChanges {
		resourceChange, err := p.processResourceChange(rc)
		if err != nil {
			// Log the error but continue processing other resources
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			continue
		}

		if resourceChange != nil {
			summary.ResourceChanges = append(summary.ResourceChanges, *resourceChange)

			// Update counters
			switch resourceChange.ChangeType {
			case models.Create:
				summary.AddCount++
			case models.Update:
				summary.ChangeCount++
			case models.Delete:
				summary.DeleteCount++
			case models.NoOp:
				summary.NoOpCount++
			}
		}
	}

	return summary, nil
}

// processResourceChange converts a raw resource change from the JSON into our ResourceChange model
func (p *Parser) processResourceChange(raw map[string]interface{}) (*models.ResourceChange, error) {
	// Check for required fields
	address, ok := raw["address"].(string)
	if !ok || address == "" {
		return nil, fmt.Errorf("missing or invalid resource address")
	}

	typeName, _ := raw["type"].(string)
	if typeName == "" {
		// Try to extract type from address if not explicitly provided
		parts := strings.Split(address, ".")
		if len(parts) > 0 {
			typeName = parts[0]
		}
	}

	// Extract the name from the address
	name := ""
	parts := strings.Split(address, ".")
	if len(parts) > 1 {
		name = parts[len(parts)-1]
	}

	// Extract module path if present
	module := ""
	if strings.HasPrefix(address, "module.") {
		moduleEnd := strings.LastIndex(address, ".")
		if moduleEnd > 0 {
			module = address[:moduleEnd]
		}
	}

	// Determine change type
	changeType := models.NoOp
	beforeMap := make(map[string]any)
	afterMap := make(map[string]any)
	beforeValues := make(map[string]string)
	afterValues := make(map[string]string)

	if change, ok := raw["change"].(map[string]interface{}); ok {
		// Extract actions
		actions, ok := change["actions"].([]interface{})
		if ok && len(actions) > 0 {
			action, _ := actions[0].(string)
			switch action {
			case "create":
				changeType = models.Create
			case "update":
				changeType = models.Update
			case "delete":
				changeType = models.Delete
			case "no-op":
				changeType = models.NoOp
			default:
				// Default to NoOp if action is unknown
				changeType = models.NoOp
			}
		}

		// Extract before/after values safely
		before, _ := change["before"].(map[string]interface{})
		after, _ := change["after"].(map[string]interface{})

		// Convert before/after to our model
		for k, v := range before {
			beforeMap[k] = v
			beforeValues[k] = fmt.Sprintf("%v", v)
		}

		for k, v := range after {
			afterMap[k] = v
			afterValues[k] = fmt.Sprintf("%v", v)
		}

		return &models.ResourceChange{
			Address:      address,
			Type:         typeName,
			Name:         name,
			ChangeType:   changeType,
			Before:       beforeMap,
			After:        afterMap,
			BeforeValues: beforeValues,
			AfterValues:  afterValues,
			Module:       module,
		}, nil
	}

	// If we can't determine the change type, still return a resource with NoOp
	return &models.ResourceChange{
		Address:      address,
		Type:         typeName,
		Name:         name,
		ChangeType:   models.NoOp,
		Before:       beforeMap,
		After:        afterMap,
		BeforeValues: beforeValues,
		AfterValues:  afterValues,
		Module:       module,
	}, nil
}
