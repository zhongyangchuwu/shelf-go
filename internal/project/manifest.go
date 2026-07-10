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
	Version  int         `json:"version"`
	Bindings []V2Binding `json:"bindings,omitempty"`
}

type V2Binding struct {
	Item     string `json:"item"`
	Field    string `json:"field"`
	Env      string `json:"env"`
	Required *bool  `json:"required,omitempty"`
	Title    string `json:"title,omitempty"`
}

func (m Manifest) MarshalJSON() ([]byte, error) {
	type itemManifest struct {
		Version  int         `json:"version"`
		Bindings []V2Binding `json:"bindings"`
	}
	bindings := m.Bindings
	if bindings == nil {
		bindings = []V2Binding{}
	}
	return json.Marshal(itemManifest{Version: m.Version, Bindings: bindings})
}

func New() Manifest {
	return Manifest{Version: CurrentVersion, Bindings: []V2Binding{}}
}

func NewV2() Manifest {
	return New()
}

func (m Manifest) IsV2() bool {
	return true
}

func (b V2Binding) IsRequired() bool {
	return b.Required == nil || *b.Required
}

func (b V2Binding) Key() string {
	return b.Item + "." + b.Field
}

func (m *Manifest) AddV2Binding(binding V2Binding) error {
	for _, existing := range m.Bindings {
		if existing.Item == binding.Item && existing.Field == binding.Field {
			return fmt.Errorf("binding with item %q and field %q already exists", binding.Item, binding.Field)
		}
	}
	m.Bindings = append(m.Bindings, binding)
	return nil
}

func (m *Manifest) RemoveV2Binding(key string) bool {
	for i, binding := range m.Bindings {
		if binding.Key() == key {
			m.Bindings = append(m.Bindings[:i], m.Bindings[i+1:]...)
			return true
		}
	}
	return false
}

func (m Manifest) FindV2Binding(key string) (V2Binding, bool) {
	for _, binding := range m.Bindings {
		if binding.Key() == key {
			return binding, true
		}
	}
	return V2Binding{}, false
}
