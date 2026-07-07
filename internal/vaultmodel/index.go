package vaultmodel

import "fmt"

type Index struct {
	itemsByID    map[string]Item
	itemsByTitle map[string][]Item
	itemsByAlias map[string][]Item
	fieldsByEnv  map[string]FieldRef
}

type FieldRef struct {
	Item  Item
	Field Field
}

func NewIndex(v Vault) (Index, error) {
	if err := Validate(v); err != nil {
		return Index{}, err
	}
	idx := Index{
		itemsByID:    make(map[string]Item, len(v.Items)),
		itemsByTitle: make(map[string][]Item, len(v.Items)),
		itemsByAlias: make(map[string][]Item),
		fieldsByEnv:  make(map[string]FieldRef),
	}
	for _, item := range v.Items {
		idx.itemsByID[item.ID] = item
		idx.itemsByTitle[item.Title] = append(idx.itemsByTitle[item.Title], item)
		for _, alias := range item.Aliases {
			if alias == "" {
				continue
			}
			idx.itemsByAlias[alias] = append(idx.itemsByAlias[alias], item)
		}
		for _, field := range item.Fields {
			for _, binding := range field.Bindings {
				if binding.Type == BindingTypeEnv {
					idx.fieldsByEnv[binding.Name] = FieldRef{Item: item, Field: field}
				}
			}
		}
	}
	return idx, nil
}

func (idx Index) ItemByID(id string) (Item, bool) {
	item, ok := idx.itemsByID[id]
	return item, ok
}

func (idx Index) ItemsByTitle(title string) []Item {
	return cloneItems(idx.itemsByTitle[title])
}

func (idx Index) ItemsByAlias(alias string) []Item {
	return cloneItems(idx.itemsByAlias[alias])
}

func (idx Index) FieldByEnv(name string) (FieldRef, bool) {
	ref, ok := idx.fieldsByEnv[name]
	return ref, ok
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
