package vaultmodel

import (
	"encoding/json"
	"fmt"
)

func BuiltInTemplates() []Template {
	return []Template{
		{
			ID:          TemplateGenericSecret,
			Name:        "Generic Secret",
			Aliases:     []string{"generic"},
			Description: "A single concealed secret value.",
			Fields: []FieldTemplate{
				{ID: "value", Label: "Value", Type: FieldTypeConcealed, Sensitivity: SensitivitySecret, Purpose: PurposePrimarySecret, Required: true},
				{ID: "value_env", Label: "Value Env", Type: FieldTypeString, Sensitivity: SensitivityPublic, Purpose: PurposeEnvName, EnvFor: "value"},
			},
		},
		{
			ID:          TemplateAPICredential,
			Name:        "API Credential",
			Aliases:     []string{"api"},
			Description: "An API credential with value-free provider, website, expiration, and scope metadata.",
			Fields: []FieldTemplate{
				{ID: "credential", Label: "Credential", Type: FieldTypeConcealed, Sensitivity: SensitivitySecret, Purpose: PurposePrimarySecret, Required: true},
				{ID: "credential_env", Label: "Credential Env", Type: FieldTypeString, Sensitivity: SensitivityPublic, Purpose: PurposeEnvName, EnvFor: "credential"},
				{ID: "provider", Label: "Provider", Type: FieldTypeString, Sensitivity: SensitivityPublic, Purpose: PurposeProvider},
				{ID: "website", Label: "Website", Type: FieldTypeURL, Sensitivity: SensitivityPublic, Purpose: PurposeWebsite},
				{ID: "expires_at", Label: "Expires", Type: FieldTypeDate, Sensitivity: SensitivityPublic, Purpose: PurposeExpiry},
				{ID: "scopes", Label: "Scopes", Type: FieldTypeStringList, Sensitivity: SensitivityPublic, Purpose: PurposeScopes},
			},
		},
	}
}

func BuiltInProfiles() []Profile {
	return []Profile{
		{
			ID:       ProfileOpenAI,
			Name:     "OpenAI",
			Aliases:  []string{"openai"},
			Template: TemplateAPICredential,
			Defaults: map[string]json.RawMessage{
				"provider":       json.RawMessage(`"OpenAI"`),
				"website":        json.RawMessage(`"https://platform.openai.com"`),
				"credential_env": json.RawMessage(`"OPENAI_API_KEY"`),
			},
		},
	}
}

type Registry struct {
	templates       map[string]Template
	templateAliases map[string]string
	profiles        map[string]Profile
	profileAliases  map[string]string
}

func NewRegistry(templates ...Template) Registry {
	return NewRegistryWithDefinitions(templates, nil)
}

func NewRegistryWithDefinitions(templates []Template, profiles []Profile) Registry {
	allTemplates := BuiltInTemplates()
	allTemplates = append(allTemplates, templates...)
	allProfiles := BuiltInProfiles()
	allProfiles = append(allProfiles, profiles...)

	out := Registry{
		templates:       make(map[string]Template, len(allTemplates)),
		templateAliases: make(map[string]string),
		profiles:        make(map[string]Profile, len(allProfiles)),
		profileAliases:  make(map[string]string),
	}
	for _, tmpl := range allTemplates {
		out.templates[tmpl.ID] = tmpl
		for _, alias := range tmpl.Aliases {
			out.templateAliases[alias] = tmpl.ID
		}
	}
	for _, profile := range allProfiles {
		out.profiles[profile.ID] = profile
		for _, alias := range profile.Aliases {
			out.profileAliases[alias] = profile.ID
		}
	}
	return out
}

func NewDefaultRegistry() Registry {
	return NewRegistryWithDefinitions(nil, nil)
}

func (r Registry) Template(id string) (Template, bool) {
	if r.templates == nil {
		r = NewDefaultRegistry()
	}
	if tmpl, ok := r.templates[id]; ok {
		return tmpl, true
	}
	if canonical, ok := r.templateAliases[id]; ok {
		tmpl, ok := r.templates[canonical]
		return tmpl, ok
	}
	return Template{}, false
}

func (r Registry) Profile(id string) (Profile, bool) {
	if r.profiles == nil {
		r = NewDefaultRegistry()
	}
	if profile, ok := r.profiles[id]; ok {
		return profile, true
	}
	if canonical, ok := r.profileAliases[id]; ok {
		profile, ok := r.profiles[canonical]
		return profile, ok
	}
	return Profile{}, false
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
