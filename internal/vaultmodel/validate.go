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
		seenFields[field.ID] = struct{}{}
		fieldsByID[field.ID] = field
		if err := validateField(index, j, field, templateFields[field.ID]); err != nil {
			return err
		}
	}

	for _, templateField := range tmpl.Fields {
		if !templateField.Required {
			continue
		}
		if _, ok := fieldsByID[templateField.ID]; !ok {
			return fmt.Errorf("items[%d] missing required field: %s", index, templateField.ID)
		}
	}

	for _, templateField := range tmpl.Fields {
		if templateField.Purpose != PurposeEnvName {
			continue
		}
		if _, ok := fieldsByID[templateField.EnvFor]; !ok {
			return fmt.Errorf("items[%d].field %s env_for target missing: %s", index, templateField.ID, templateField.EnvFor)
		}
	}
	return nil
}

func validateField(itemIndex, fieldIndex int, field Field, templateField FieldTemplate) error {
	if templateField.ID != "" {
		if field.Type != templateField.Type {
			return fmt.Errorf("items[%d].fields[%d].type for %s = %s, want %s", itemIndex, fieldIndex, field.ID, field.Type, templateField.Type)
		}
		if field.Sensitivity != templateField.Sensitivity {
			return fmt.Errorf("items[%d].fields[%d].sensitivity for %s = %s, want %s", itemIndex, fieldIndex, field.ID, field.Sensitivity, templateField.Sensitivity)
		}
		if templateField.Purpose != "" && field.Purpose != templateField.Purpose {
			return fmt.Errorf("items[%d].fields[%d].purpose for %s = %s, want %s", itemIndex, fieldIndex, field.ID, field.Purpose, templateField.Purpose)
		}
	}
	if field.Type == "" {
		return fmt.Errorf("items[%d].fields[%d].type is empty", itemIndex, fieldIndex)
	}
	if field.Sensitivity != SensitivityPublic && field.Sensitivity != SensitivitySecret {
		return fmt.Errorf("items[%d].fields[%d].sensitivity invalid: %s", itemIndex, fieldIndex, field.Sensitivity)
	}
	if err := validateFieldValue(itemIndex, fieldIndex, field); err != nil {
		return err
	}
	if templateField.Purpose == PurposeEnvName || field.Purpose == PurposeEnvName {
		if field.Type != FieldTypeString {
			return fmt.Errorf("items[%d].fields[%d].env name field must be string: %s", itemIndex, fieldIndex, field.ID)
		}
		if field.Sensitivity != SensitivityPublic {
			return fmt.Errorf("items[%d].fields[%d].env name field must be public: %s", itemIndex, fieldIndex, field.ID)
		}
		if err := validateEnvNameFieldValue(itemIndex, fieldIndex, field); err != nil {
			return err
		}
	}
	return nil
}

func validateFieldValue(itemIndex, fieldIndex int, field Field) error {
	if len(field.Value) == 0 {
		return nil
	}
	switch field.Type {
	case FieldTypeString, FieldTypeConcealed:
		var value string
		if err := json.Unmarshal(field.Value, &value); err != nil {
			return fmt.Errorf("items[%d].fields[%d].value must be string: %w", itemIndex, fieldIndex, err)
		}
	case FieldTypeURL:
		var value string
		if err := json.Unmarshal(field.Value, &value); err != nil {
			return fmt.Errorf("items[%d].fields[%d].value must be url string: %w", itemIndex, fieldIndex, err)
		}
		u, err := url.Parse(value)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("items[%d].fields[%d].value invalid url: %s", itemIndex, fieldIndex, value)
		}
	case FieldTypeDate:
		var value string
		if err := json.Unmarshal(field.Value, &value); err != nil {
			return fmt.Errorf("items[%d].fields[%d].value must be date string: %w", itemIndex, fieldIndex, err)
		}
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return fmt.Errorf("items[%d].fields[%d].value invalid date: %s", itemIndex, fieldIndex, value)
		}
	case FieldTypeStringList:
		var value []string
		if err := json.Unmarshal(field.Value, &value); err != nil {
			return fmt.Errorf("items[%d].fields[%d].value must be string array: %w", itemIndex, fieldIndex, err)
		}
	case FieldTypeJSON:
		if !json.Valid(field.Value) {
			return fmt.Errorf("items[%d].fields[%d].value invalid json", itemIndex, fieldIndex)
		}
	default:
		return fmt.Errorf("items[%d].fields[%d].type unsupported: %s", itemIndex, fieldIndex, field.Type)
	}
	return nil
}

func validateEnvNameFieldValue(itemIndex, fieldIndex int, field Field) error {
	if len(field.Value) == 0 {
		return nil
	}
	var value string
	if err := json.Unmarshal(field.Value, &value); err != nil {
		return fmt.Errorf("items[%d].fields[%d].env name value must be string: %w", itemIndex, fieldIndex, err)
	}
	if value == "" {
		return nil
	}
	if !envNamePattern.MatchString(value) {
		return fmt.Errorf("items[%d].fields[%d].env name invalid: %s", itemIndex, fieldIndex, value)
	}
	return nil
}

func validateTemplates(templates []Template) error {
	seen := map[string]struct{}{}
	for i, tmpl := range templates {
		if tmpl.ID == "" {
			return fmt.Errorf("templates[%d].id is empty", i)
		}
		if _, ok := seen[tmpl.ID]; ok {
			return fmt.Errorf("duplicate template id: %s", tmpl.ID)
		}
		seen[tmpl.ID] = struct{}{}
		if tmpl.Name == "" {
			return fmt.Errorf("templates[%d].name is empty", i)
		}
		fields := map[string]FieldTemplate{}
		for j, field := range tmpl.Fields {
			if field.ID == "" {
				return fmt.Errorf("templates[%d].fields[%d].id is empty", i, j)
			}
			if _, ok := fields[field.ID]; ok {
				return fmt.Errorf("templates[%d] duplicate field id: %s", i, field.ID)
			}
			fields[field.ID] = field
		}
		for _, field := range tmpl.Fields {
			if field.Purpose != PurposeEnvName {
				continue
			}
			if field.Type != FieldTypeString {
				return fmt.Errorf("templates[%d].field %s env name field must be string", i, field.ID)
			}
			if field.Sensitivity != SensitivityPublic {
				return fmt.Errorf("templates[%d].field %s env name field must be public", i, field.ID)
			}
			if field.EnvFor == "" {
				return fmt.Errorf("templates[%d].field %s env_for is empty", i, field.ID)
			}
			target, ok := fields[field.EnvFor]
			if !ok {
				return fmt.Errorf("templates[%d].field %s env_for target unknown: %s", i, field.ID, field.EnvFor)
			}
			if target.Sensitivity != SensitivitySecret {
				return fmt.Errorf("templates[%d].field %s env_for target must be secret: %s", i, field.ID, field.EnvFor)
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
		fields := make(map[string]FieldTemplate, len(tmpl.Fields))
		for _, field := range tmpl.Fields {
			fields[field.ID] = field
		}
		for fieldID, value := range profile.Defaults {
			templateField, ok := fields[fieldID]
			if !ok {
				return fmt.Errorf("profiles[%d].defaults.%s field unknown", i, fieldID)
			}
			if templateField.Sensitivity == SensitivitySecret {
				return fmt.Errorf("profiles[%d].defaults.%s must not set secret field defaults", i, fieldID)
			}
			field := Field{ID: fieldID, Label: templateField.Label, Type: templateField.Type, Sensitivity: templateField.Sensitivity, Purpose: templateField.Purpose, Value: value}
			if err := validateField(-1, i, field, templateField); err != nil {
				return fmt.Errorf("profiles[%d].defaults.%s invalid: %w", i, fieldID, err)
			}
		}
	}
	return nil
}
