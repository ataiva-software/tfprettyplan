package terminal

import (
	"os"
	"testing"

	"golang.org/x/term"
)

func TestGetWidth(t *testing.T) {
	// This test is limited because it depends on the actual terminal environment
	// where the test is running. We can at least verify that the function returns
	// a reasonable value.
	width := GetWidth()

	// The width should be either the actual terminal width or the default width
	// if terminal detection fails
	if width <= 0 {
		t.Errorf("GetWidth() returned an invalid width: %d", width)
	}

	// Check if the width is the default width when terminal detection fails
	// This is a bit tricky to test directly without mocking the term package,
	// so we'll just verify that the function returns a reasonable value.
	if width < DefaultWidth/2 || width > 1000 {
		t.Errorf("GetWidth() returned an unreasonable width: %d", width)
	}
}

func TestIsTerminal(t *testing.T) {
	// This test is also limited because it depends on the actual terminal environment.
	// We can at least verify that the function returns a boolean value.
	isTerminal := IsTerminal()

	// The expected result depends on whether the test is running in a terminal
	// or not, which we can't easily control in a test environment.
	// We can verify that the function matches the behavior of term.IsTerminal
	expectedIsTerminal := term.IsTerminal(int(os.Stdout.Fd()))
	if isTerminal != expectedIsTerminal {
		t.Errorf("IsTerminal() = %v, want %v", isTerminal, expectedIsTerminal)
	}
}

// TestDefaultWidth verifies that the DefaultWidth constant is set to a reasonable value
func TestDefaultWidth(t *testing.T) {
	if DefaultWidth <= 0 {
		t.Errorf("DefaultWidth is invalid: %d", DefaultWidth)
	}

	// Most terminals are at least 80 characters wide, so this is a reasonable default
	if DefaultWidth < 80 {
		t.Errorf("DefaultWidth is too small: %d", DefaultWidth)
	}

	// DefaultWidth should not be unreasonably large
	if DefaultWidth > 200 {
		t.Errorf("DefaultWidth is too large: %d", DefaultWidth)
	}
}

// This is a more sophisticated test that could be implemented if we refactor the
// terminal package to be more testable by accepting a file descriptor as a parameter.
// For now, we'll leave this commented out as a suggestion for future improvements.
/*
func TestGetWidthWithMock(t *testing.T) {
	// Create a pipe to simulate a terminal
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Save the original stdout and restore it after the test
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Replace stdout with our pipe
	os.Stdout = w

	// Test with a mock that always fails
	// In this case, GetWidth should return the default width
	// This would require modifying the GetWidth function to accept a file descriptor
	// as a parameter, or to use a mockable interface for term.GetSize.
	// width := GetWidthWithFd(int(r.Fd()))
	// if width != DefaultWidth {
	// 	t.Errorf("GetWidth() with failing term.GetSize = %d, want %d", width, DefaultWidth)
	// }
}
*/

