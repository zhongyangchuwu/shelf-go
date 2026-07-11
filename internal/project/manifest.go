package project

import (
	"encoding/json"
	"fmt"
)

const (
	CurrentVersion = 1
	FileName       = ".shelf.json"
)

type Manifest struct {
	Version  int       `json:"version"`
	Bindings []Binding `json:"bindings,omitempty"`
}

type Binding struct {
	Item     string `json:"item"`
	Field    string `json:"field"`
	Env      string `json:"env"`
	Required *bool  `json:"required,omitempty"`
	Title    string `json:"title,omitempty"`
}

func (m Manifest) MarshalJSON() ([]byte, error) {
	type itemManifest struct {
		Version  int       `json:"version"`
		Bindings []Binding `json:"bindings"`
	}
	bindings := m.Bindings
	if bindings == nil {
		bindings = []Binding{}
	}
	return json.Marshal(itemManifest{Version: m.Version, Bindings: bindings})
}

func New() Manifest {
	return Manifest{Version: CurrentVersion, Bindings: []Binding{}}
}

func (b Binding) IsRequired() bool {
	return b.Required == nil || *b.Required
}

func (b Binding) Key() string {
	return b.Item + "." + b.Field
}

func (m *Manifest) AddBinding(binding Binding) error {
	for _, existing := range m.Bindings {
		if existing.Item == binding.Item && existing.Field == binding.Field {
			return fmt.Errorf("binding with item %q and field %q already exists", binding.Item, binding.Field)
		}
	}
	m.Bindings = append(m.Bindings, binding)
	return nil
}

func (m *Manifest) RemoveBinding(key string) bool {
	for i, binding := range m.Bindings {
		if binding.Key() == key {
			m.Bindings = append(m.Bindings[:i], m.Bindings[i+1:]...)
			return true
		}
	}
	return false
}

func (m Manifest) FindBinding(key string) (Binding, bool) {
	for _, binding := range m.Bindings {
		if binding.Key() == key {
			return binding, true
		}
	}
	return Binding{}, false
}
