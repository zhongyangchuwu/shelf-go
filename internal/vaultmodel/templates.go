package vaultmodel

import "fmt"

func BuiltInTemplates() []Template {
	return []Template{
		{
			ID:          TemplateGenericSecret,
			Name:        "Generic Secret",
			Description: "A single concealed secret value.",
			Fields: []FieldTemplate{
				{ID: "value", Label: "Value", Type: FieldTypeConcealed, Sensitivity: SensitivitySecret, Purpose: PurposePrimarySecret, Required: true},
			},
		},
		{
			ID:          TemplateAPICredential,
			Name:        "API Credential",
			Description: "An API credential with value-free provider, website, expiration, and scope metadata.",
			Fields: []FieldTemplate{
				{ID: "credential", Label: "Credential", Type: FieldTypeConcealed, Sensitivity: SensitivitySecret, Purpose: PurposePrimarySecret, Required: true},
				{ID: "provider", Label: "Provider", Type: FieldTypeString, Sensitivity: SensitivityPublic, Purpose: PurposeProvider},
				{ID: "website", Label: "Website", Type: FieldTypeURL, Sensitivity: SensitivityPublic, Purpose: PurposeWebsite},
				{ID: "expires_at", Label: "Expires", Type: FieldTypeDate, Sensitivity: SensitivityPublic, Purpose: PurposeExpiry},
				{ID: "scopes", Label: "Scopes", Type: FieldTypeStringList, Sensitivity: SensitivityPublic, Purpose: PurposeScopes},
			},
		},
	}
}

type Registry struct {
	templates map[string]Template
}

func NewRegistry(templates ...Template) Registry {
	all := BuiltInTemplates()
	all = append(all, templates...)
	out := Registry{templates: make(map[string]Template, len(all))}
	for _, tmpl := range all {
		out.templates[tmpl.ID] = tmpl
	}
	return out
}

func (r Registry) Template(id string) (Template, bool) {
	if r.templates == nil {
		r = NewRegistry()
	}
	tmpl, ok := r.templates[id]
	return tmpl, ok
}

func (r Registry) RequiredFields(id string) ([]FieldTemplate, error) {
	tmpl, ok := r.Template(id)
	if !ok {
		return nil, fmt.Errorf("unknown template: %s", id)
	}
	fields := make([]FieldTemplate, 0, len(tmpl.Fields))
	for _, field := range tmpl.Fields {
		if field.Required {
			fields = append(fields, field)
		}
	}
	return fields, nil
}
