package cli

import (
	"strings"
	"testing"
)

func TestTUIRejectsNonInteractiveStreams(t *testing.T) {
	_, err := runShelf(t, "tui")
	if err == nil {
		t.Fatal("shelf tui with buffered streams returned nil error")
	}
	if !strings.Contains(err.Error(), "requires an interactive terminal") {
		t.Fatalf("non-TTY error = %q", err)
	}
}
