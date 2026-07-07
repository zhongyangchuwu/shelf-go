// Package vaultmodel defines the v2 in-memory vault model, built-in templates,
// validation, and runtime index construction.
//
// It has no dependencies on CLI, app, project, or v1 vault packages.
package vaultmodel

import "encoding/json"

// Version is the v2 vault format version.
const Version = 2

// CurrentVersion is an alias for Version, retained for compatibility.
const CurrentVersion = Version

// Template identifiers for built-in templates.
const (
	TemplateGenericSecret = "generic_secret"
	TemplateAPICredential = "api_credential"
)

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

// Binding type constants.
const BindingTypeEnv = "env"

// Purpose constants for built-in template fields.
const (
	PurposePrimarySecret = "primary_secret"
	PurposeProvider      = "provider/source"
	PurposeWebsite       = "website"
	PurposeExpiry        = "expiry"
	PurposeScopes        = "scopes"
)

// Vault is the top-level v2 vault document.
type Vault struct {
	Version   int        `json:"version"`
	Items     []Item     `json:"items"`
	Templates []Template `json:"templates,omitempty"`
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
	Bindings    []Binding       `json:"bindings,omitempty"`
}

// Binding attaches a field value to an environment variable name.
type Binding struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// Template describes a reusable field layout for items.
type Template struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
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
	Required    bool        `json:"required,omitempty"`
	Repeatable  bool        `json:"repeatable,omitempty"`
}
