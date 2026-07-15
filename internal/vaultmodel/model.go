// Package vaultmodel defines the in-memory item-vault model, built-in templates,
// validation, and runtime index construction.
//
// It has no dependencies on CLI, app, or project packages.
package vaultmodel

import (
	"encoding/json"
	"strings"
)

// Version is the current item-vault schema version.
const Version = 4

// CurrentVersion names the schema version written by Shelf.
const CurrentVersion = Version

// Template identifiers for built-in templates.
const (
	TemplateGenericSecret = "generic_secret"
	TemplateAPICredential = "api_credential"
)

// Profile identifiers for built-in profiles.
const ProfileOpenAI = "openai"

const (
	envFieldPrefix      = "env_"
	validityFieldPrefix = "validity_"
)

// EnvFieldID returns the structural field ID for an item's default environment
// projection of fieldID.
func EnvFieldID(fieldID string) string {
	return envFieldPrefix + fieldID
}

// ValidityFieldID returns the structural field ID for lifecycle metadata for
// fieldID.
func ValidityFieldID(fieldID string) string {
	return validityFieldPrefix + fieldID
}

// IsEnvFieldID reports whether id is an environment projection field and, if
// it is, returns the atomic field it projects.
func IsEnvFieldID(id string) (string, bool) {
	target := strings.TrimPrefix(id, envFieldPrefix)
	return target, target != id && target != ""
}

// IsValidityFieldID reports whether id is a lifecycle metadata field and, if
// it is, returns the tracked secret field.
func IsValidityFieldID(id string) (string, bool) {
	target := strings.TrimPrefix(id, validityFieldPrefix)
	return target, target != id && target != ""
}

// IsStructuralFieldID reports whether id is an environment or lifecycle
// companion rather than an atomic business field.
func IsStructuralFieldID(id string) bool {
	_, env := IsEnvFieldID(id)
	_, validity := IsValidityFieldID(id)
	return env || validity
}

// IsProjectableType reports whether a field type has an unambiguous string
// representation suitable for an environment variable.
func IsProjectableType(fieldType FieldType) bool {
	switch fieldType {
	case FieldTypeString, FieldTypeConcealed, FieldTypeURL:
		return true
	default:
		return false
	}
}

// FieldType enumerates the supported field value types.
type FieldType string

const (
	FieldTypeString     FieldType = "string"
	FieldTypeConcealed  FieldType = "concealed"
	FieldTypeURL        FieldType = "url"
	FieldTypeDate       FieldType = "date"
	FieldTypeStringList FieldType = "string_list"
	FieldTypeJSON       FieldType = "json"
)

// Sensitivity describes whether a field value may appear in non-materializing output.
type Sensitivity string

const (
	SensitivityPublic Sensitivity = "public"
	SensitivitySecret Sensitivity = "secret"
)

// Vault is the top-level item-vault document.
type Vault struct {
	Version   int        `json:"version"`
	Items     []Item     `json:"items"`
	Templates []Template `json:"templates,omitempty"`
	Profiles  []Profile  `json:"profiles,omitempty"`
}

// Item represents a single service, integration, or account inside a vault.
type Item struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Template    string   `json:"template"`
	Fields      []Field  `json:"fields"`
	Labels      []string `json:"labels,omitempty"`
	Description string   `json:"description,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
}

// Field is an atomic item value. Its type, label, sensitivity, and requiredness
// belong to the owning template.
type Field struct {
	ID    string          `json:"id"`
	Value json.RawMessage `json:"value,omitempty"`
}

// Template describes a reusable field layout for items.
type Template struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Aliases       []string        `json:"aliases,omitempty"`
	Description   string          `json:"description,omitempty"`
	PrimarySecret string          `json:"primary_secret"`
	Fields        []FieldTemplate `json:"fields"`
}

// FieldTemplate describes a single atomic or structural field slot in a template.
type FieldTemplate struct {
	ID          string      `json:"id"`
	Label       string      `json:"label"`
	Type        FieldType   `json:"type"`
	Sensitivity Sensitivity `json:"sensitivity"`
	Required    bool        `json:"required,omitempty"`
}

// Profile describes default field values over a template.
type Profile struct {
	ID       string                     `json:"id"`
	Name     string                     `json:"name"`
	Aliases  []string                   `json:"aliases,omitempty"`
	Template string                     `json:"template"`
	Defaults map[string]json.RawMessage `json:"defaults,omitempty"`
}
