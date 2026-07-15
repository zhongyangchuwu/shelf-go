package vaultmodel

import (
	"encoding/json"
	"fmt"
)

type Index struct {
	itemsByID    map[string]Item
	itemsByTitle map[string][]Item
	fieldsByEnv  map[string][]FieldRef
}

type FieldRef struct {
	Item     Item
	Field    ResolvedField
	EnvField ResolvedField
}

func NewIndex(v Vault) (Index, error) {
	registry := NewRegistryWithDefinitions(v.Templates, v.Profiles)
	if err := ValidateWithRegistry(v, registry); err != nil {
		return Index{}, err
	}
	return newIndexWithRegistry(v, registry), nil
}

func newIndexWithRegistry(v Vault, registry Registry) Index {
	idx := Index{
		itemsByID:    make(map[string]Item, len(v.Items)),
		itemsByTitle: make(map[string][]Item, len(v.Items)),
		fieldsByEnv:  make(map[string][]FieldRef),
	}
	for _, item := range v.Items {
		idx.itemsByID[item.ID] = item
		idx.itemsByTitle[item.Title] = append(idx.itemsByTitle[item.Title], item)
		template, _ := registry.Template(item.Template)
		templateFields := make(map[string]FieldTemplate, len(template.Fields))
		for _, field := range template.Fields {
			templateFields[field.ID] = field
		}
		fields := make(map[string]Field, len(item.Fields))
		for _, field := range item.Fields {
			fields[field.ID] = field
		}
		for _, definition := range template.Fields {
			targetID, env := IsEnvFieldID(definition.ID)
			if !env {
				continue
			}
			envField, exists := fields[definition.ID]
			if !exists || len(envField.Value) == 0 {
				continue
			}
			var envName string
			if err := json.Unmarshal(envField.Value, &envName); err != nil || envName == "" {
				continue
			}
			target, exists := fields[targetID]
			if !exists {
				continue
			}
			targetDefinition := templateFields[targetID]
			idx.fieldsByEnv[envName] = append(idx.fieldsByEnv[envName], FieldRef{
				Item:     item,
				Field:    resolvedField(targetDefinition, target),
				EnvField: resolvedField(definition, envField),
			})
		}
	}
	return idx
}

func (idx Index) ItemByID(id string) (Item, bool) {
	item, ok := idx.itemsByID[id]
	return item, ok
}

func (idx Index) ItemsByTitle(title string) []Item {
	return cloneItems(idx.itemsByTitle[title])
}

func (idx Index) FieldsByEnv(name string) []FieldRef {
	refs := idx.fieldsByEnv[name]
	if len(refs) == 0 {
		return nil
	}
	out := make([]FieldRef, len(refs))
	copy(out, refs)
	return out
}

func (idx Index) UniqueTitle(title string) (Item, error) {
	items := idx.itemsByTitle[title]
	if len(items) == 0 {
		return Item{}, fmt.Errorf("item title not found: %s", title)
	}
	if len(items) > 1 {
		return Item{}, fmt.Errorf("item title is ambiguous: %s", title)
	}
	return items[0], nil
}

func cloneItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}
	out := make([]Item, len(items))
	copy(out, items)
	return out
}
