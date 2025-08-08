package renderer

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ao/tfprettyplan/pkg/config"
	"github.com/ao/tfprettyplan/pkg/models"
	"github.com/fatih/color"
)

// Renderer is responsible for rendering Terraform plan summaries in ASCII format
type Renderer struct {
	colorEnabled bool
	config       *config.Config
	tableConfig  *config.TableConfig
}

// Option is a functional option for configuring the renderer
type Option func(*Renderer)

// WithColor enables or disables color output
func WithColor(enabled bool) Option {
	return func(r *Renderer) {
		r.colorEnabled = enabled
	}
}

// WithConfig sets the configuration for the renderer
func WithConfig(cfg *config.Config) Option {
	return func(r *Renderer) {
		r.config = cfg
		r.tableConfig = cfg.GetTableConfig()
	}
}

// New creates a new Renderer with the provided options
func New(opts ...Option) *Renderer {
	// Create default configuration
	defaultConfig := config.DefaultConfig()

	r := &Renderer{
		colorEnabled: true, // Enable color by default
		config:       defaultConfig,
		tableConfig:  defaultConfig.GetTableConfig(),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Render renders a plan summary to the provided writer
func (r *Renderer) Render(w io.Writer, summary *models.PlanSummary) {
	r.renderSummaryTable(w, summary)
	r.renderResourceChanges(w, summary)
}

// renderSummaryTable renders a summary table with counts of resource changes
func (r *Renderer) renderSummaryTable(w io.Writer, summary *models.PlanSummary) {
	fmt.Fprintln(w, "Terraform Plan Summary")
	fmt.Fprintln(w, "=====================")
	fmt.Fprintln(w)

	// Create a simple table manually
	fmt.Fprintln(w, "+--------+-------+")
	fmt.Fprintln(w, "| ACTION | COUNT |")
	fmt.Fprintln(w, "+--------+-------+")

	// Add rows with colored output if enabled
	addRow := func(action string, count int, colorFunc func(format string, a ...interface{}) string) {
		if count > 0 {
			if r.colorEnabled {
				fmt.Fprintf(w, "| %-6s | %5d |\n", colorFunc(action), count)
			} else {
				fmt.Fprintf(w, "| %-6s | %5d |\n", action, count)
			}
		}
	}

	addRow("Create", summary.AddCount, color.GreenString)
	addRow("Update", summary.ChangeCount, color.YellowString)
	addRow("Delete", summary.DeleteCount, color.RedString)
	addRow("No-op", summary.NoOpCount, color.BlueString)

	total := summary.AddCount + summary.ChangeCount + summary.DeleteCount + summary.NoOpCount
	fmt.Fprintf(w, "| %-6s | %5d |\n", "Total", total)
	fmt.Fprintln(w, "+--------+-------+")
	fmt.Fprintln(w)
}

// renderResourceChanges renders detailed information about each resource change
func (r *Renderer) renderResourceChanges(w io.Writer, summary *models.PlanSummary) {
	// Group changes by type
	creates := filterByChangeType(summary.ResourceChanges, models.Create)
	updates := filterByChangeType(summary.ResourceChanges, models.Update)
	deletes := filterByChangeType(summary.ResourceChanges, models.Delete)

	// Render each group
	if len(creates) > 0 {
		r.renderChangeGroup(w, "Resources to Create", creates, color.GreenString)
	}

	if len(updates) > 0 {
		r.renderChangeGroup(w, "Resources to Update", updates, color.YellowString)
	}

	if len(deletes) > 0 {
		r.renderChangeGroup(w, "Resources to Delete", deletes, color.RedString)
	}
}

// renderChangeGroup renders a group of resource changes with the same change type
func (r *Renderer) renderChangeGroup(w io.Writer, title string, changes []models.ResourceChange, colorFunc func(format string, a ...interface{}) string) {
	if r.colorEnabled {
		fmt.Fprintln(w, colorFunc(title))
		fmt.Fprintln(w, colorFunc(strings.Repeat("=", len(title))))
	} else {
		fmt.Fprintln(w, title)
		fmt.Fprintln(w, strings.Repeat("=", len(title)))
	}
	fmt.Fprintln(w)

	// Sort changes by address for consistent output
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Address < changes[j].Address
	})

	for _, change := range changes {
		r.renderResourceChange(w, &change, colorFunc)
	}
}

// renderResourceChange renders details of a single resource change
func (r *Renderer) renderResourceChange(w io.Writer, change *models.ResourceChange, colorFunc func(format string, a ...interface{}) string) {
	// Display resource address and type
	address := change.Address
	if r.colorEnabled {
		address = colorFunc(address)
	}
	fmt.Fprintf(w, "• %s (%s)\n", address, change.Type)

	// For updates, show what's changing
	if change.ChangeType == models.Update {
		r.renderAttributeChanges(w, change)
	}

	fmt.Fprintln(w)
}

// truncateValue truncates a string value if it's longer than maxWidth
// Uses smart truncation to preserve important parts of the value
func (r *Renderer) truncateValue(value string, maxWidth int) string {
	if len(value) <= maxWidth {
		return value
	}

	// If the value is a path-like string with slashes, preserve the beginning and end
	if strings.Contains(value, "/") {
		parts := strings.Split(value, "/")
		if len(parts) > 2 {
			// Keep first and last part, truncate middle
			firstPart := parts[0]
			lastPart := parts[len(parts)-1]

			// Calculate how much space we have for the middle
			remainingSpace := maxWidth - len(firstPart) - len(lastPart) - 5 // 5 for "/.../"

			if remainingSpace > 0 {
				// We can show some of the middle parts
				middleParts := parts[1 : len(parts)-1]
				middle := ""

				for _, part := range middleParts {
					if len(middle)+len(part)+1 <= remainingSpace {
						if middle != "" {
							middle += "/"
						}
						middle += part
					} else {
						break
					}
				}

				if middle != "" {
					return firstPart + "/" + middle + "/.../" + lastPart
				}
				return firstPart + "/.../" + lastPart
			}
		}
	}

	// For JSON-like values with braces, preserve structure
	if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
		return value[:maxWidth-3] + "...}"
	}

	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		return value[:maxWidth-3] + "...]"
	}

	// For long strings without special structure, truncate middle
	if len(value) > maxWidth && maxWidth > 6 {
		halfWidth := (maxWidth - 3) / 2
		return value[:halfWidth] + "..." + value[len(value)-halfWidth:]
	}

	// Default truncation
	return value[:maxWidth-3] + "..."
}

// renderAttributeChanges renders a table showing attribute changes for updated resources
func (r *Renderer) renderAttributeChanges(w io.Writer, change *models.ResourceChange) {
	// Find attributes that have changed
	changedAttrs := make(map[string]struct{})
	for k := range change.BeforeValues {
		if after, exists := change.AfterValues[k]; exists {
			if after != change.BeforeValues[k] {
				changedAttrs[k] = struct{}{}
			}
		} else {
			changedAttrs[k] = struct{}{}
		}
	}

	for k := range change.AfterValues {
		if _, exists := change.BeforeValues[k]; !exists {
			changedAttrs[k] = struct{}{}
		}
	}

	// If no changes, don't render anything
	if len(changedAttrs) == 0 {
		return
	}

	// Convert to slice and sort
	attrs := make([]string, 0, len(changedAttrs))
	for k := range changedAttrs {
		attrs = append(attrs, k)
	}
	sort.Strings(attrs)

	// Create table header with dynamic widths
	attrWidth := r.tableConfig.MaxAttributeWidth
	valueWidth := r.tableConfig.MaxValueWidth

	// Calculate total width of the table (for future use)
	_ = attrWidth + valueWidth*2 + 7 // 7 for borders and padding

	// Create the top border
	fmt.Fprintf(w, "  +%s+%s+%s+\n",
		strings.Repeat("-", attrWidth+2),
		strings.Repeat("-", valueWidth+2),
		strings.Repeat("-", valueWidth+2))

	// Create the header row
	fmt.Fprintf(w, "  | %-*s | %-*s | %-*s |\n",
		attrWidth, "ATTRIBUTE",
		valueWidth, "OLD VALUE",
		valueWidth, "NEW VALUE")

	// Create the separator
	fmt.Fprintf(w, "  +%s+%s+%s+\n",
		strings.Repeat("-", attrWidth+2),
		strings.Repeat("-", valueWidth+2),
		strings.Repeat("-", valueWidth+2))

	// Add rows for each changed attribute
	for _, attr := range attrs {
		oldVal := change.BeforeValues[attr]
		newVal := change.AfterValues[attr]

		if oldVal == "" {
			oldVal = "(none)"
		}
		if newVal == "" {
			newVal = "(none)"
		}

		// Truncate values if they're too long
		oldVal = r.truncateValue(oldVal, valueWidth)
		newVal = r.truncateValue(newVal, valueWidth)

		fmt.Fprintf(w, "  | %-*s | %-*s | %-*s |\n",
			attrWidth, attr,
			valueWidth, oldVal,
			valueWidth, newVal)
	}

	// Create the bottom border
	fmt.Fprintf(w, "  +%s+%s+%s+\n",
		strings.Repeat("-", attrWidth+2),
		strings.Repeat("-", valueWidth+2),
		strings.Repeat("-", valueWidth+2))
}

// filterByChangeType returns a slice of resource changes filtered by the given change type
func filterByChangeType(changes []models.ResourceChange, changeType models.ChangeType) []models.ResourceChange {
	var filtered []models.ResourceChange
	for _, change := range changes {
		if change.ChangeType == changeType {
			filtered = append(filtered, change)
		}
	}
	return filtered
}

// RenderToString renders a plan summary to a string
func (r *Renderer) RenderToString(summary *models.PlanSummary) string {
	var buf bytes.Buffer
	r.Render(&buf, summary)
	return buf.String()
}
