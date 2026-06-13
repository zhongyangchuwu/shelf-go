package cli

import (
	"strings"
	"testing"
)

func TestCompletionCommandGeneratesZsh(t *testing.T) {
	out, err := runShelf(t, "completion", "zsh")
	if err != nil {
		t.Fatalf("completion zsh: %v", err)
	}
	if !strings.Contains(out, "#compdef shelf") {
		t.Fatalf("unexpected zsh completion output")
	}
}
