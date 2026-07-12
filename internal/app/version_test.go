package app

import (
	"strings"
	"testing"
)

func TestStringReportsCurrentReleaseVersion(t *testing.T) {
	if got := String(); !strings.HasPrefix(got, "v0.2.0 ") {
		t.Fatalf("version = %q, want v0.2.0 prefix", got)
	}
}
