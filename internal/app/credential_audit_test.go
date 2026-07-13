package app

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func TestCredentialAuditClassifiesV3ValidityWithoutValues(t *testing.T) {
	vault := vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{
		auditCredentialItem("expired", "Expired", `"2026-07-11"`),
		auditCredentialItem("due_today", "Due today", `"2026-07-12"`),
		auditCredentialItem("due_soon", "Due soon", `"2026-08-11"`),
		auditCredentialItem("healthy", "Healthy", `"2026-08-12"`),
		auditCredentialItem("unknown", "Unknown", `"unknown"`),
		auditCredentialItem("permanent", "Permanent", `"permanent"`),
		auditCredentialItem("invalid", "Invalid", `"not-a-date"`),
		{ID: "generic", Title: "Generic", Template: vaultmodel.TemplateGenericSecret, Fields: []vaultmodel.Field{{ID: "value", Value: json.RawMessage(`"fake-generic-secret"`)}}},
	}}

	report := credentialAuditForVault(vault, 30, time.Date(2026, 7, 12, 16, 30, 0, 0, time.FixedZone("local", -7*60*60)))
	if got, want := len(report.Items), 7; got != want {
		t.Fatalf("audit item count = %d, want %d", got, want)
	}
	statuses := make(map[string]CredentialHealthStatus, len(report.Items))
	for _, item := range report.Items {
		statuses[item.ID] = item.Status
	}
	for id, want := range map[string]CredentialHealthStatus{
		"expired": CredentialHealthExpired, "due_today": CredentialHealthExpiring, "due_soon": CredentialHealthExpiring,
		"healthy": CredentialHealthHealthy, "unknown": CredentialHealthUnknown, "permanent": CredentialHealthPermanent, "invalid": CredentialHealthInvalid,
	} {
		if got := statuses[id]; got != want {
			t.Fatalf("status for %s = %q, want %q", id, got, want)
		}
	}
	if !report.HasExpired() || !report.HasWarnings() {
		t.Fatalf("report severity = expired %t warnings %t", report.HasExpired(), report.HasWarnings())
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	if strings.Contains(string(encoded), "fake-generic-secret") || strings.Contains(string(encoded), "fake-credential-secret") || strings.Contains(string(encoded), "not-a-date") {
		t.Fatalf("audit report leaked value: %s", encoded)
	}
}

func TestCredentialAuditHorizonRejectsNegativeValues(t *testing.T) {
	_, err := testApp().credentialAuditAt("", "", CredentialAuditRequest{WithinDays: -1}, time.Now())
	if err == nil || !strings.Contains(err.Error(), "horizon") {
		t.Fatalf("negative horizon error = %v", err)
	}
}

func auditCredentialItem(id, title, validity string) vaultmodel.Item {
	return vaultmodel.Item{
		ID: id, Title: title, Template: vaultmodel.TemplateAPICredential,
		Fields: []vaultmodel.Field{
			{ID: "credential", Value: json.RawMessage(`"fake-credential-secret"`)},
			{ID: vaultmodel.ValidityFieldID("credential"), Value: json.RawMessage(validity)},
		},
	}
}
