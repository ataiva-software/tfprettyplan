package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ao/tfprettyplan/pkg/models"
	"github.com/ao/tfprettyplan/pkg/parser"
	"github.com/ao/tfprettyplan/pkg/renderer"
)

// displayProviderError formats and displays Terraform provider errors in a user-friendly way
func displayProviderError(err error) {
	fmt.Fprintf(os.Stderr, "\nTerraform Provider Error Detected\n")
	fmt.Fprintf(os.Stderr, "===========================\n\n")
	fmt.Fprintf(os.Stderr, "%v\n\n", err)
	fmt.Fprintf(os.Stderr, "For more information on resolving provider errors, see: docs/terraform-workflow.md\n\n")
	
	// Provide specific guidance based on the error
	if strings.Contains(err.Error(), "plugin schemas") || strings.Contains(err.Error(), "unavailable provider") {
		fmt.Fprintf(os.Stderr, "Quick Fix: Generate the plan JSON in the same directory as your Terraform configuration:\n\n")
		fmt.Fprintf(os.Stderr, "  cd /path/to/your/terraform/project\n")
		fmt.Fprintf(os.Stderr, "  terraform init\n")
		fmt.Fprintf(os.Stderr, "  terraform plan -out=plan.tfplan\n")
		fmt.Fprintf(os.Stderr, "  terraform show -json plan.tfplan > plan.json\n")
		fmt.Fprintf(os.Stderr, "  tfprettyplan plan.json\n\n")
	}
}

func main() {
	// Define command-line flags
	var (
		planFile    string
		noColor     bool
		showVersion bool
		version     = "0.1.0" // Hardcoded for now, could be set during build
	)

	flag.StringVar(&planFile, "file", "", "Path to Terraform plan JSON file")
	flag.StringVar(&planFile, "f", "", "Path to Terraform plan JSON file (shorthand)")
	flag.BoolVar(&noColor, "no-color", false, "Disable color output")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information (shorthand)")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "TFPrettyPlan - A tool to visualize Terraform plan files in a readable format\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [plan-file]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "If plan-file is provided without the -file flag, it will be used as the input file.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s plan.json\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  %s -file=plan.json\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  terraform show -json plan.tfplan | %s\n", filepath.Base(os.Args[0]))
	}

	flag.Parse()

	// Show version and exit if requested
	if showVersion {
		fmt.Printf("TFPrettyPlan v%s\n", version)
		os.Exit(0)
	}

	// Check for a positional argument if no file flag was provided
	if planFile == "" && flag.NArg() > 0 {
		planFile = flag.Arg(0)
	}

	// Determine if we're reading from stdin or a file
	var err error
	var planData []byte

	if planFile == "" {
		// Check if stdin has data
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			// No data on stdin and no file specified
			flag.Usage()
			os.Exit(1)
		}

		// Read from stdin
		planData, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
	}

	// Create a new parser
	p := parser.New()

	// Parse the plan
	var summary *models.PlanSummary
	if planFile != "" {
		summary, err = p.ParseFile(planFile)
		if err != nil {
			// Check for provider errors and display them more prominently
			if strings.Contains(err.Error(), "provider error") || 
			   strings.Contains(err.Error(), "plugin schemas") || 
			   strings.Contains(err.Error(), "unavailable provider") {
				displayProviderError(err)
			} else {
				fmt.Fprintf(os.Stderr, "Error parsing plan file: %v\n", err)
			}
			os.Exit(1)
		}
	} else {
		summary, err = p.ParseJSON(planData)
		if err != nil {
			// Check for provider errors and display them more prominently
			if strings.Contains(err.Error(), "provider error") || 
			   strings.Contains(err.Error(), "plugin schemas") || 
			   strings.Contains(err.Error(), "unavailable provider") {
				displayProviderError(err)
			} else {
				fmt.Fprintf(os.Stderr, "Error parsing plan JSON: %v\n", err)
			}
			os.Exit(1)
		}
	}

	// Create a renderer with the appropriate color setting
	r := renderer.New(renderer.WithColor(!noColor))

	// Render the plan summary to stdout
	r.Render(os.Stdout, summary)
}
