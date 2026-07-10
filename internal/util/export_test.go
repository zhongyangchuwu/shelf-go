package util

import (
	"strings"
	"testing"
)

func TestEnvBindingsRejectsLineBreakingValues(t *testing.T) {
	_, err := EnvBindings([]Binding{{EnvName: "TOKEN", Value: "one\ntwo"}})
	if err == nil {
		t.Fatal("expected newline value to be rejected")
	}
	if !strings.Contains(err.Error(), "TOKEN") || !strings.Contains(err.Error(), "--format shell") {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(err.Error(), "one") || strings.Contains(err.Error(), "two") {
		t.Fatalf("error leaked value: %v", err)
	}
}

func TestEnvBindingsRendersSimpleValues(t *testing.T) {
	got, err := EnvBindings([]Binding{{EnvName: "TOKEN", Value: "abc=123"}})
	if err != nil {
		t.Fatalf("env bindings: %v", err)
	}
	want := "TOKEN=abc=123\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
