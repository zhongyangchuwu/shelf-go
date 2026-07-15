package vaultmodel

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Resolver struct {
	vault    Vault
	registry Registry
	index    Index
}

// ResolvedField combines an atomic item value with the immutable field schema
// from its owning template.
type ResolvedField struct {
	ID          string
	Label       string
	Type        FieldType
	Sensitivity Sensitivity
	Value       json.RawMessage
}

func NewResolver(vault Vault) (Resolver, error) {
	registry := NewRegistryWithDefinitions(vault.Templates, vault.Profiles)
	if err := ValidateWithRegistry(vault, registry); err != nil {
		return Resolver{}, err
	}
	return newResolverWithRegistry(vault, registry), nil
}

// NewResolverFromValidatedVault constructs a resolver for a vault already
// validated by the caller, such as a vaultjson repository load.
func NewResolverFromValidatedVault(vault Vault) Resolver {
	registry := NewRegistryWithDefinitions(vault.Templates, vault.Profiles)
	return newResolverWithRegistry(vault, registry)
}

func newResolverWithRegistry(vault Vault, registry Registry) Resolver {
	return Resolver{
		vault:    vault,
		registry: registry,
		index:    newIndexWithRegistry(vault, registry),
	}
}

func (r Resolver) ResolveItem(selector string) (int, Item, error) {
	if selector == "" {
		return 0, Item{}, fmt.Errorf("selector is required")
	}
	match := -1
	ambiguous := false
	for i, item := range r.vault.Items {
		if item.ID != selector && item.Title != selector {
			continue
		}
		if match >= 0 {
			ambiguous = true
			continue
		}
		match = i
	}
	if match < 0 {
		return 0, Item{}, fmt.Errorf("item not found: %s", selector)
	}
	if ambiguous {
		return 0, Item{}, r.ambiguousItemError(selector)
	}
	return match, r.vault.Items[match], nil
}

func (r Resolver) ItemByID(id string) (Item, bool) {
	return r.index.ItemByID(id)
}

func (r Resolver) ResolveEnv(name string) (FieldRef, error) {
	refs := r.index.fieldsByEnv[name]
	if len(refs) == 0 {
		return FieldRef{}, fmt.Errorf("env selector not found: %s", name)
	}
	if len(refs) > 1 {
		var b strings.Builder
		fmt.Fprintf(&b, "env selector matched multiple fields: %s", name)
		for _, ref := range refs {
			fmt.Fprintf(&b, "\n- %s %s %s", ref.Item.ID, ref.Item.Title, ref.Field.ID)
		}
		return FieldRef{}, fmt.Errorf("%s", b.String())
	}
	return refs[0], nil
}

func (r Resolver) Field(item Item, fieldID string) (int, ResolvedField, error) {
	definition, ok, err := r.registry.FieldTemplate(item.Template, fieldID)
	if err != nil {
		return 0, ResolvedField{}, err
	}
	if !ok {
		return 0, ResolvedField{}, fmt.Errorf("field not found: %s", fieldID)
	}
	for i, field := range item.Fields {
		if field.ID == fieldID {
			return i, resolvedField(definition, field), nil
		}
	}
	return 0, ResolvedField{}, fmt.Errorf("field not found: %s", fieldID)
}

func (r Resolver) DefaultField(item Item, fieldID string) (int, ResolvedField, error) {
	if fieldID != "" {
		return r.Field(item, fieldID)
	}
	primary, err := r.PrimarySecretField(item)
	if err != nil {
		return 0, ResolvedField{}, err
	}
	return r.Field(item, primary.ID)
}

func (r Resolver) PrimarySecretField(item Item) (FieldTemplate, error) {
	return r.registry.PrimarySecretField(item.Template)
}

func (r Resolver) EnvName(item Item, fieldID string) (string, bool, error) {
	envFieldID := EnvFieldID(fieldID)
	if _, ok, err := r.registry.FieldTemplate(item.Template, envFieldID); err != nil || !ok {
		return "", false, err
	}
	for _, envField := range item.Fields {
		if envField.ID != envFieldID {
			continue
		}
		if len(envField.Value) == 0 {
			return "", false, nil
		}
		var value string
		if err := json.Unmarshal(envField.Value, &value); err != nil {
			return "", false, err
		}
		if value == "" {
			return "", false, nil
		}
		return value, true, nil
	}
	return "", false, nil

}

// IsProjectableField reports whether fieldID is an atomic field whose type can
// be represented in an environment variable. Item default env mappings are
// optional because project bindings name their target environment explicitly.
func (r Resolver) IsProjectableField(item Item, fieldID string) bool {
	definition, ok, err := r.registry.FieldTemplate(item.Template, fieldID)
	return err == nil && ok && !IsStructuralFieldID(fieldID) && IsProjectableType(definition.Type)
}

func (r Resolver) SecretSelectors() []string {
	selectors := make([]string, 0)
	for _, item := range r.vault.Items {
		tmpl, ok := r.registry.Template(item.Template)
		if !ok {
			continue
		}
		for _, field := range tmpl.Fields {
			if field.Sensitivity == SensitivitySecret && !IsStructuralFieldID(field.ID) {
				selectors = append(selectors, item.ID+"."+field.ID)
			}
		}
	}
	return selectors
}

func (r Resolver) ambiguousItemError(selector string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "selector matched multiple items: %s", selector)
	for _, item := range r.vault.Items {
		if item.ID == selector || item.Title == selector {
			fmt.Fprintf(&b, "\n- %s %s", item.ID, item.Title)
		}
	}
	return fmt.Errorf("%s", b.String())
}

func resolvedField(definition FieldTemplate, field Field) ResolvedField {
	return ResolvedField{ID: field.ID, Label: definition.Label, Type: definition.Type, Sensitivity: definition.Sensitivity, Value: field.Value}
}

func hasString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
