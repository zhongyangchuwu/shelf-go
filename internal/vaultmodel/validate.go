package vaultmodel

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"time"
)

var envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func Validate(v Vault) error {
	registry := NewRegistryWithDefinitions(v.Templates, v.Profiles)
	if err := validateTemplates(v.Templates); err != nil {
		return err
	}
	if err := validateProfiles(v.Profiles, registry); err != nil {
		return err
	}
	return validateWithRegistry(v, registry)
}

func ValidateWithRegistry(v Vault, registry Registry) error {
	return validateWithRegistry(v, registry)
}

func validateWithRegistry(v Vault, registry Registry) error {
	if v.Version != Version {
		return fmt.Errorf("unsupported vault version: %d", v.Version)
	}
	seenItems := make(map[string]struct{}, len(v.Items))
	for i, item := range v.Items {
		if item.ID == "" {
			return fmt.Errorf("items[%d].id is empty", i)
		}
		if _, exists := seenItems[item.ID]; exists {
			return fmt.Errorf("duplicate item id: %s", item.ID)
		}
		seenItems[item.ID] = struct{}{}
		if err := validateItem(i, item, registry); err != nil {
			return err
		}
	}
	return nil
}

func validateItem(index int, item Item, registry Registry) error {
	if item.Title == "" {
		return fmt.Errorf("items[%d].title is empty", index)
	}
	tmpl, ok := registry.Template(item.Template)
	if !ok {
		return fmt.Errorf("items[%d].template unknown: %s", index, item.Template)
	}
	templateFields := make(map[string]FieldTemplate, len(tmpl.Fields))
	for _, field := range tmpl.Fields {
		templateFields[field.ID] = field
	}

	seenFields := make(map[string]struct{}, len(item.Fields))
	fieldsByID := make(map[string]Field, len(item.Fields))
	for j, field := range item.Fields {
		if field.ID == "" {
			return fmt.Errorf("items[%d].fields[%d].id is empty", index, j)
		}
		if _, exists := seenFields[field.ID]; exists {
			return fmt.Errorf("items[%d] duplicate field id: %s", index, field.ID)
		}
		definition, ok := templateFields[field.ID]
		if !ok {
			return fmt.Errorf("items[%d].fields[%d].id unknown: %s", index, j, field.ID)
		}
		seenFields[field.ID] = struct{}{}
		fieldsByID[field.ID] = field
		if err := validateFieldValue(index, j, field.Value, definition); err != nil {
			return err
		}
	}

	for _, templateField := range tmpl.Fields {
		if templateField.Required {
			if _, ok := fieldsByID[templateField.ID]; !ok {
				return fmt.Errorf("items[%d] missing required field: %s", index, templateField.ID)
			}
		}
		if target, validity := IsValidityFieldID(templateField.ID); validity {
			field, ok := fieldsByID[templateField.ID]
			if !ok || len(field.Value) == 0 {
				return fmt.Errorf("items[%d] missing validity for field: %s", index, target)
			}
		}
	}
	for fieldID := range fieldsByID {
		if target, env := IsEnvFieldID(fieldID); env {
			if _, ok := fieldsByID[target]; !ok {
				return fmt.Errorf("items[%d].field %s env target missing: %s", index, fieldID, target)
			}
		}
		if target, validity := IsValidityFieldID(fieldID); validity {
			if _, ok := fieldsByID[target]; !ok {
				return fmt.Errorf("items[%d].field %s validity target missing: %s", index, fieldID, target)
			}
		}
	}
	return nil
}

func validateFieldValue(itemIndex, fieldIndex int, value json.RawMessage, definition FieldTemplate) error {
	if len(value) == 0 {
		return nil
	}
	if _, validity := IsValidityFieldID(definition.ID); validity {
		return validateValidityValue(itemIndex, fieldIndex, value, definition.ID)
	}
	switch definition.Type {
	case FieldTypeString, FieldTypeConcealed:
		var text string
		if err := json.Unmarshal(value, &text); err != nil {
			return fmt.Errorf("items[%d].fields[%d].value must be string: %w", itemIndex, fieldIndex, err)
		}
	case FieldTypeURL:
		var text string
		if err := json.Unmarshal(value, &text); err != nil {
			return fmt.Errorf("items[%d].fields[%d].value must be url string: %w", itemIndex, fieldIndex, err)
		}
		u, err := url.Parse(text)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("items[%d].fields[%d].value invalid url: %s", itemIndex, fieldIndex, text)
		}
	case FieldTypeDate:
		var text string
		if err := json.Unmarshal(value, &text); err != nil {
			return fmt.Errorf("items[%d].fields[%d].value must be date string: %w", itemIndex, fieldIndex, err)
		}
		if _, err := time.Parse(time.DateOnly, text); err != nil {
			return fmt.Errorf("items[%d].fields[%d].value invalid date: %s", itemIndex, fieldIndex, text)
		}
	case FieldTypeStringList:
		var text []string
		if err := json.Unmarshal(value, &text); err != nil {
			return fmt.Errorf("items[%d].fields[%d].value must be string array: %w", itemIndex, fieldIndex, err)
		}
	case FieldTypeJSON:
		if !json.Valid(value) {
			return fmt.Errorf("items[%d].fields[%d].value invalid json", itemIndex, fieldIndex)
		}
	default:
		return fmt.Errorf("items[%d].fields[%d].type unsupported: %s", itemIndex, fieldIndex, definition.Type)
	}
	if _, env := IsEnvFieldID(definition.ID); env {
		if err := validateEnvNameValue(itemIndex, fieldIndex, value); err != nil {
			return err
		}
	}
	return nil
}

func validateValidityValue(itemIndex, fieldIndex int, value json.RawMessage, fieldID string) error {
	var text string
	if err := json.Unmarshal(value, &text); err != nil {
		return fmt.Errorf("items[%d].fields[%d].%s must be a string: %w", itemIndex, fieldIndex, fieldID, err)
	}
	if text == "unknown" || text == "permanent" {
		return nil
	}
	if _, err := time.Parse(time.DateOnly, text); err == nil {
		return nil
	}
	return fmt.Errorf("items[%d].fields[%d].%s must be unknown, permanent, or YYYY-MM-DD", itemIndex, fieldIndex, fieldID)
}

func validateEnvNameValue(itemIndex, fieldIndex int, value json.RawMessage) error {
	var text string
	if err := json.Unmarshal(value, &text); err != nil {
		return fmt.Errorf("items[%d].fields[%d].env name value must be string: %w", itemIndex, fieldIndex, err)
	}
	if text == "" {
		return nil
	}
	if !envNamePattern.MatchString(text) {
		return fmt.Errorf("items[%d].fields[%d].env name invalid: %s", itemIndex, fieldIndex, text)
	}
	return nil
}

func validateTemplates(templates []Template) error {
	seen := map[string]struct{}{}
	for i, tmpl := range templates {
		if err := validateTemplate(fmt.Sprintf("templates[%d]", i), tmpl, seen); err != nil {
			return err
		}
	}
	return nil
}

func validateTemplate(prefix string, tmpl Template, seen map[string]struct{}) error {
	if tmpl.ID == "" {
		return fmt.Errorf("%s.id is empty", prefix)
	}
	if _, ok := seen[tmpl.ID]; ok {
		return fmt.Errorf("duplicate template id: %s", tmpl.ID)
	}
	seen[tmpl.ID] = struct{}{}
	if tmpl.Name == "" {
		return fmt.Errorf("%s.name is empty", prefix)
	}
	fields := make(map[string]FieldTemplate, len(tmpl.Fields))
	for j, field := range tmpl.Fields {
		if field.ID == "" {
			return fmt.Errorf("%s.fields[%d].id is empty", prefix, j)
		}
		if _, ok := fields[field.ID]; ok {
			return fmt.Errorf("%s duplicate field id: %s", prefix, field.ID)
		}
		if field.Type == "" {
			return fmt.Errorf("%s.field %s type is empty", prefix, field.ID)
		}
		if field.Sensitivity != SensitivityPublic && field.Sensitivity != SensitivitySecret {
			return fmt.Errorf("%s.field %s sensitivity invalid: %s", prefix, field.ID, field.Sensitivity)
		}
		fields[field.ID] = field
	}
	primary, ok := fields[tmpl.PrimarySecret]
	if tmpl.PrimarySecret == "" || !ok || primary.Sensitivity != SensitivitySecret {
		return fmt.Errorf("%s.primary_secret must name a secret field", prefix)
	}
	for _, field := range tmpl.Fields {
		if target, env := IsEnvFieldID(field.ID); env {
			if IsStructuralFieldID(target) {
				return fmt.Errorf("%s.field %s env target must be atomic: %s", prefix, field.ID, target)
			}
			targetField, ok := fields[target]
			if !ok {
				return fmt.Errorf("%s.field %s env target unknown: %s", prefix, field.ID, target)
			}
			if field.Type != FieldTypeString || field.Sensitivity != SensitivityPublic {
				return fmt.Errorf("%s.field %s env field must be public string", prefix, field.ID)
			}
			if !IsProjectableType(targetField.Type) {
				return fmt.Errorf("%s.field %s env target type is not projectable: %s", prefix, field.ID, targetField.Type)
			}
		}
		if target, validity := IsValidityFieldID(field.ID); validity {
			if IsStructuralFieldID(target) {
				return fmt.Errorf("%s.field %s validity target must be atomic: %s", prefix, field.ID, target)
			}
			targetField, ok := fields[target]
			if !ok {
				return fmt.Errorf("%s.field %s validity target unknown: %s", prefix, field.ID, target)
			}
			if targetField.Sensitivity != SensitivitySecret {
				return fmt.Errorf("%s.field %s validity target must be secret: %s", prefix, field.ID, target)
			}
			if field.Type != FieldTypeString || field.Sensitivity != SensitivityPublic || !field.Required {
				return fmt.Errorf("%s.field %s validity field must be required public string", prefix, field.ID)
			}
		}
	}
	return nil
}

func validateProfiles(profiles []Profile, registry Registry) error {
	seen := map[string]struct{}{}
	for i, profile := range profiles {
		if profile.ID == "" {
			return fmt.Errorf("profiles[%d].id is empty", i)
		}
		if _, ok := seen[profile.ID]; ok {
			return fmt.Errorf("duplicate profile id: %s", profile.ID)
		}
		seen[profile.ID] = struct{}{}
		if profile.Name == "" {
			return fmt.Errorf("profiles[%d].name is empty", i)
		}
		tmpl, ok := registry.Template(profile.Template)
		if !ok {
			return fmt.Errorf("profiles[%d].template unknown: %s", i, profile.Template)
		}
		for fieldID, value := range profile.Defaults {
			definition, ok, err := registry.FieldTemplate(tmpl.ID, fieldID)
			if err != nil || !ok {
				return fmt.Errorf("profiles[%d].defaults.%s field unknown", i, fieldID)
			}
			if definition.Sensitivity == SensitivitySecret {
				return fmt.Errorf("profiles[%d].defaults.%s must not set secret field defaults", i, fieldID)
			}
			if err := validateFieldValue(-1, i, value, definition); err != nil {
				return fmt.Errorf("profiles[%d].defaults.%s invalid: %w", i, fieldID, err)
			}
		}
	}
	return nil
}
