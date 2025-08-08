package terminal

import (
	"os"

	"golang.org/x/term"
)

// DefaultWidth is the default terminal width if detection fails
const DefaultWidth = 80

// GetWidth returns the width of the terminal
// If detection fails, it returns the default width
func GetWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return DefaultWidth
	}
	return width
}

// IsTerminal returns true if stdout is a terminal
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
