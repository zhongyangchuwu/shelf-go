package app

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

const DefaultCredentialAuditWithinDays = 30

type CredentialAuditRequest struct {
	WithinDays int
}

type CredentialHealthStatus string

const (
	CredentialHealthHealthy   CredentialHealthStatus = "healthy"
	CredentialHealthExpiring  CredentialHealthStatus = "expiring"
	CredentialHealthExpired   CredentialHealthStatus = "expired"
	CredentialHealthMissing   CredentialHealthStatus = "missing"
	CredentialHealthMalformed CredentialHealthStatus = "malformed"
)

type CredentialHealthItem struct {
	ID      string                 `json:"id"`
	Title   string                 `json:"title"`
	Field   string                 `json:"field"`
	Expires string                 `json:"expires,omitempty"`
	Status  CredentialHealthStatus `json:"status"`
}

type CredentialAuditReport struct {
	WithinDays int                    `json:"within_days"`
	Items      []CredentialHealthItem `json:"items"`
}

func (a *App) CredentialAudit(configPathFlag, vaultPathFlag string, req CredentialAuditRequest) (CredentialAuditReport, error) {
	return a.credentialAuditAt(configPathFlag, vaultPathFlag, req, time.Now())
}

func (a *App) credentialAuditAt(configPathFlag, vaultPathFlag string, req CredentialAuditRequest, now time.Time) (CredentialAuditReport, error) {
	if req.WithinDays < 0 {
		return CredentialAuditReport{}, fmt.Errorf("audit horizon must be zero or greater")
	}
	vault, err := a.loadVaultForRead(configPathFlag, vaultPathFlag)
	if err != nil {
		return CredentialAuditReport{}, err
	}
	return credentialAuditForVault(vault, req.WithinDays, now), nil
}

func credentialAuditForVault(vault vaultmodel.Vault, withinDays int, now time.Time) CredentialAuditReport {
	today := dateOnlyUTC(now)
	deadline := today.AddDate(0, 0, withinDays)
	registry := vaultmodel.NewRegistryWithDefinitions(vault.Templates, vault.Profiles)
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	report := CredentialAuditReport{WithinDays: withinDays, Items: make([]CredentialHealthItem, 0)}
	for _, item := range vault.Items {
		primary, err := resolver.PrimarySecretField(item)
		if err != nil {
			continue
		}
		template, ok := registry.Template(item.Template)
		if !ok {
			continue
		}
		for _, expiryTemplate := range template.Fields {
			if expiryTemplate.Purpose != vaultmodel.PurposeExpiry {
				continue
			}
			expiryField := vaultmodel.Field{ID: expiryTemplate.ID, Type: expiryTemplate.Type, Sensitivity: expiryTemplate.Sensitivity, Purpose: expiryTemplate.Purpose}
			for _, field := range item.Fields {
				if field.ID == expiryTemplate.ID {
					expiryField = field
					break
				}
			}
			report.Items = append(report.Items, auditExpiryField(item, primary.ID, expiryField, today, deadline))
		}
	}
	sort.Slice(report.Items, func(i, j int) bool {
		left, right := report.Items[i], report.Items[j]
		if credentialHealthRank(left.Status) != credentialHealthRank(right.Status) {
			return credentialHealthRank(left.Status) < credentialHealthRank(right.Status)
		}
		if left.Expires != right.Expires {
			return left.Expires < right.Expires
		}
		if left.Title != right.Title {
			return left.Title < right.Title
		}
		return left.Field < right.Field
	})
	return report
}

func auditExpiryField(item vaultmodel.Item, secretFieldID string, expiryField vaultmodel.Field, today, deadline time.Time) CredentialHealthItem {
	record := CredentialHealthItem{ID: item.ID, Title: item.Title, Field: secretFieldID}
	if len(expiryField.Value) == 0 {
		record.Status = CredentialHealthMissing
		return record
	}
	var expires string
	if err := json.Unmarshal(expiryField.Value, &expires); err != nil {
		record.Status = CredentialHealthMalformed
		return record
	}
	expiry, err := time.Parse(time.DateOnly, expires)
	if err != nil {
		record.Status = CredentialHealthMalformed
		return record
	}
	record.Expires = expires
	switch {
	case expiry.Before(today):
		record.Status = CredentialHealthExpired
	case !expiry.After(deadline):
		record.Status = CredentialHealthExpiring
	default:
		record.Status = CredentialHealthHealthy
	}
	return record
}

func credentialHealthByItem(report CredentialAuditReport) map[string]CredentialHealthItem {
	health := make(map[string]CredentialHealthItem, len(report.Items))
	for _, candidate := range report.Items {
		current, exists := health[candidate.ID]
		if !exists || credentialHealthRank(candidate.Status) < credentialHealthRank(current.Status) {
			health[candidate.ID] = candidate
		}
	}
	return health
}
func (r CredentialAuditReport) HasExpired() bool {
	for _, item := range r.Items {
		if item.Status == CredentialHealthExpired {
			return true
		}
	}
	return false
}

func (r CredentialAuditReport) HasWarnings() bool {
	for _, item := range r.Items {
		if item.Status == CredentialHealthExpiring || item.Status == CredentialHealthMissing || item.Status == CredentialHealthMalformed {
			return true
		}
	}
	return false
}

func dateOnlyUTC(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func credentialHealthRank(status CredentialHealthStatus) int {
	switch status {
	case CredentialHealthExpired:
		return 0
	case CredentialHealthExpiring:
		return 1
	case CredentialHealthMissing:
		return 2
	case CredentialHealthMalformed:
		return 3
	default:
		return 4
	}
}
