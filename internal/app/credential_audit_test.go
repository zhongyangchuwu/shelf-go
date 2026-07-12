package app

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func TestCredentialAuditClassifiesHealthWithoutSecretValues(t *testing.T) {
	vault := vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{
		auditCredentialItem("expired", "Expired", rawJSON("2026-07-11"), vaultmodel.FieldTypeDate),
		auditCredentialItem("due_today", "Due today", rawJSON("2026-07-12"), vaultmodel.FieldTypeDate),
		auditCredentialItem("due_soon", "Due soon", rawJSON("2026-08-11"), vaultmodel.FieldTypeDate),
		auditCredentialItem("healthy", "Healthy", rawJSON("2026-08-12"), vaultmodel.FieldTypeDate),
		auditCredentialItem("missing", "Missing", nil, vaultmodel.FieldTypeDate),
		auditCredentialItem("malformed", "Malformed", rawJSON("not-a-date"), vaultmodel.FieldTypeString),
		{ID: "generic", Title: "Generic", Template: vaultmodel.TemplateGenericSecret, Fields: []vaultmodel.Field{{ID: "value", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, Purpose: vaultmodel.PurposePrimarySecret, Value: rawJSON("fake-generic-secret")}}},
	}}

	report := credentialAuditForVault(vault, 30, time.Date(2026, 7, 12, 16, 30, 0, 0, time.FixedZone("local", -7*60*60)))
	if got, want := len(report.Items), 6; got != want {
		t.Fatalf("audit item count = %d, want %d", got, want)
	}
	statuses := make(map[string]CredentialHealthStatus, len(report.Items))
	for _, item := range report.Items {
		statuses[item.ID] = item.Status
	}
	for id, want := range map[string]CredentialHealthStatus{
		"expired":   CredentialHealthExpired,
		"due_today": CredentialHealthExpiring,
		"due_soon":  CredentialHealthExpiring,
		"healthy":   CredentialHealthHealthy,
		"missing":   CredentialHealthMissing,
		"malformed": CredentialHealthMalformed,
	} {
		if got := statuses[id]; got != want {
			t.Fatalf("status for %s = %q, want %q", id, got, want)
		}
	}
	if report.HasExpired() != true || report.HasWarnings() != true {
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

func auditCredentialItem(id, title string, expires json.RawMessage, expiryType vaultmodel.FieldType) vaultmodel.Item {
	return vaultmodel.Item{
		ID:       id,
		Title:    title,
		Template: vaultmodel.TemplateAPICredential,
		Fields: []vaultmodel.Field{
			{ID: "credential", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, Purpose: vaultmodel.PurposePrimarySecret, Value: rawJSON("fake-credential-secret")},
			{ID: "expires_at", Type: expiryType, Sensitivity: vaultmodel.SensitivityPublic, Purpose: vaultmodel.PurposeExpiry, Value: expires},
		},
	}
}

func TestCredentialAuditReportsPersistedMalformedCustomExpiryWithoutValue(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeConfig(t, vaultPath)
	template := vaultmodel.Template{
		ID:   "custom_credential",
		Name: "Custom Credential",
		Fields: []vaultmodel.FieldTemplate{
			{ID: "token", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, Purpose: vaultmodel.PurposePrimarySecret, Required: true},
			{ID: "renewal_date", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Purpose: vaultmodel.PurposeExpiry},
		},
	}
	vault := vaultmodel.Vault{Version: vaultmodel.Version, Templates: []vaultmodel.Template{template}, Items: []vaultmodel.Item{{
		ID:       "custom",
		Title:    "Custom credential",
		Template: template.ID,
		Fields: []vaultmodel.Field{
			{ID: "token", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, Purpose: vaultmodel.PurposePrimarySecret, Value: rawJSON("fake-custom-secret")},
			{ID: "renewal_date", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Purpose: vaultmodel.PurposeExpiry, Value: rawJSON("not-a-date")},
		},
	}}}
	if err := vaultjson.NewRepository(vaultPath).Save(vault); err != nil {
		t.Fatalf("save custom vault: %v", err)
	}
	report, err := testApp().CredentialAudit(configPath, vaultPath, CredentialAuditRequest{WithinDays: DefaultCredentialAuditWithinDays})
	if err != nil {
		t.Fatalf("CredentialAudit: %v", err)
	}
	if got := report.Items; len(got) != 1 || got[0].Status != CredentialHealthMalformed || got[0].Expires != "" {
		t.Fatalf("malformed report = %#v", got)
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	if strings.Contains(string(encoded), "fake-custom-secret") || strings.Contains(string(encoded), "not-a-date") {
		t.Fatalf("malformed report leaked value: %s", encoded)
	}
}
