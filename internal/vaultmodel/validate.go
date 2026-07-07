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
	return ValidateWithRegistry(v, NewRegistry())
}

func ValidateWithRegistry(v Vault, registry Registry) error {
	if v.Version != Version {
		return fmt.Errorf("unsupported vault version: %d", v.Version)
	}
	seenItems := make(map[string]struct{}, len(v.Items))
	seenEnv := map[string]string{}
	for i, item := range v.Items {
		if item.ID == "" {
			return fmt.Errorf("items[%d].id is empty", i)
		}
		if _, exists := seenItems[item.ID]; exists {
			return fmt.Errorf("duplicate item id: %s", item.ID)
		}
		seenItems[item.ID] = struct{}{}
		if err := validateItem(i, item, registry, seenEnv); err != nil {
			return err
		}
	}
	return nil
}

func validateItem(index int, item Item, registry Registry, seenEnv map[string]string) error {
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
		if err := validateField(index, j, item, field, templateFields[field.ID], seenEnv); err != nil {
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
	return nil
}

func validateField(itemIndex, fieldIndex int, item Item, field Field, templateField FieldTemplate, seenEnv map[string]string) error {
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
	if len(field.Bindings) > 0 && field.Sensitivity != SensitivitySecret {
		return fmt.Errorf("items[%d].fields[%d] public field must not carry bindings: %s", itemIndex, fieldIndex, field.ID)
	}
	if err := validateFieldValue(itemIndex, fieldIndex, field); err != nil {
		return err
	}
	for k, binding := range field.Bindings {
		if binding.Type != BindingTypeEnv {
			return fmt.Errorf("items[%d].fields[%d].bindings[%d].type unsupported: %s", itemIndex, fieldIndex, k, binding.Type)
		}
		if !envNamePattern.MatchString(binding.Name) {
			return fmt.Errorf("items[%d].fields[%d].bindings[%d].name invalid: %s", itemIndex, fieldIndex, k, binding.Name)
		}
		owner := item.ID + ":" + field.ID
		if existing, ok := seenEnv[binding.Name]; ok {
			return fmt.Errorf("duplicate env binding %s on %s and %s", binding.Name, existing, owner)
		}
		seenEnv[binding.Name] = owner
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
