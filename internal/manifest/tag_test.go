package manifest

import (
	"strings"
	"testing"
)

func TestValidateAcceptsTagEntry(t *testing.T) {
	m := Manifest{Version: CurrentVersion, Secrets: []Entry{{Tags: []string{"ai", "prod"}, Required: boolPtr(false)}}}
	if err := Validate(m); err != nil {
		t.Fatalf("expected valid tag entry, got: %v", err)
	}
	entry := m.Secrets[0]
	if !entry.IsTag() || entry.Key() != "ai,prod" || entry.IsRequired() {
		t.Fatalf("unexpected tag entry helpers: key=%q isTag=%v required=%v", entry.Key(), entry.IsTag(), entry.IsRequired())
	}
}

func TestValidateRejectsInvalidTagEntries(t *testing.T) {
	tests := []struct {
		name    string
		entry   Entry
		wantErr string
	}{
		{name: "path and tags", entry: Entry{Path: "app:token", Tags: []string{"ai"}}, wantErr: "mutually exclusive"},
		{name: "prefix and tags", entry: Entry{Prefix: "app", Tags: []string{"ai"}}, wantErr: "mutually exclusive"},
		{name: "tag env", entry: Entry{Tags: []string{"ai"}, Env: "APP_TOKEN"}, wantErr: "tag entries must not carry env"},
		{name: "empty tag", entry: Entry{Tags: []string{""}}, wantErr: "tag must not be empty"},
		{name: "bad tag", entry: Entry{Tags: []string{"open ai"}}, wantErr: "unsupported characters"},
		{name: "duplicate entry tag", entry: Entry{Tags: []string{"ai", "ai"}}, wantErr: "duplicate tag"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(Manifest{Version: CurrentVersion, Secrets: []Entry{tt.entry}})
			if err == nil {
				t.Fatalf("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateRejectsDuplicateTagSelector(t *testing.T) {
	err := Validate(Manifest{Version: CurrentVersion, Secrets: []Entry{{Tags: []string{"ai", "prod"}}, {Tags: []string{"ai", "prod"}}}})
	if err == nil {
		t.Fatalf("expected duplicate tag selector to fail")
	}
	if !strings.Contains(err.Error(), "duplicate secrets entry tags") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveEntryHandlesTagEntries(t *testing.T) {
	m := Manifest{Version: CurrentVersion, Secrets: []Entry{{Path: "app:token"}, {Tags: []string{"ai", "prod"}}}}
	if !m.RemoveEntry("ai,prod") {
		t.Fatalf("expected tag entry removed")
	}
	if len(m.Secrets) != 1 || m.Secrets[0].Path != "app:token" {
		t.Fatalf("unexpected manifest after remove: %+v", m.Secrets)
	}
}
