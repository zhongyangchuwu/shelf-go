// Package vaultmodel defines the in-memory item-vault model, built-in templates,
// validation, and runtime index construction.
//
// It has no dependencies on CLI, app, or project packages.
package vaultmodel

import "encoding/json"

// Version is the current item-vault schema version.
const Version = 2

// CurrentVersion names the schema version written by Shelf.
const CurrentVersion = Version

// Template identifiers for built-in templates.
const (
	TemplateGenericSecret = "generic_secret"
	TemplateAPICredential = "api_credential"
)

// Profile identifiers for built-in profiles.
const ProfileOpenAI = "openai"

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

// Purpose constants for built-in template fields.
const (
	PurposePrimarySecret = "primary_secret"
	PurposeEnvName       = "env_name"
	PurposeProvider      = "provider/source"
	PurposeWebsite       = "website"
	PurposeExpiry        = "expiry"
	PurposeScopes        = "scopes"
)

// Vault is the top-level item-vault document.
type Vault struct {
	Version   int        `json:"version"`
	Items     []Item     `json:"items"`
	Templates []Template `json:"templates,omitempty"`
	Profiles  []Profile  `json:"profiles,omitempty"`
}

// Item represents a single secret item inside a vault.
type Item struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Template  string   `json:"template"`
	Fields    []Field  `json:"fields"`
	Labels    []string `json:"labels,omitempty"`
	Aliases   []string `json:"aliases,omitempty"`
	Notes     string   `json:"notes,omitempty"`
	CreatedAt string   `json:"created_at,omitempty"`
	UpdatedAt string   `json:"updated_at,omitempty"`
}

// Field represents a single typed field within an item.
type Field struct {
	ID          string          `json:"id"`
	Label       string          `json:"label"`
	Type        FieldType       `json:"type"`
	Sensitivity Sensitivity     `json:"sensitivity"`
	Purpose     string          `json:"purpose,omitempty"`
	Value       json.RawMessage `json:"value,omitempty"`
}

// Template describes a reusable field layout for items.
type Template struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Aliases     []string        `json:"aliases,omitempty"`
	Description string          `json:"description,omitempty"`
	Fields      []FieldTemplate `json:"fields"`
}

// FieldTemplate describes a single field slot in a template.
type FieldTemplate struct {
	ID          string      `json:"id"`
	Label       string      `json:"label"`
	Type        FieldType   `json:"type"`
	Sensitivity Sensitivity `json:"sensitivity"`
	Purpose     string      `json:"purpose,omitempty"`
	EnvFor      string      `json:"env_for,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Repeatable  bool        `json:"repeatable,omitempty"`
}

// Profile describes default field values over a template.
type Profile struct {
	ID       string                     `json:"id"`
	Name     string                     `json:"name"`
	Aliases  []string                   `json:"aliases,omitempty"`
	Template string                     `json:"template"`
	Defaults map[string]json.RawMessage `json:"defaults,omitempty"`
}
