package project

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	CurrentVersion = 1
	V2Version      = 2
	FileName       = ".shelf.json"
)

type Manifest struct {
	Version  int         `json:"version"`
	Secrets  []Entry     `json:"secrets"`
	Bindings []V2Binding `json:"bindings,omitempty"`
}

type Entry struct {
	Path     string   `json:"path,omitempty"`
	Prefix   string   `json:"prefix,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Env      string   `json:"env,omitempty"`
	Required *bool    `json:"required,omitempty"`
}

type V2Binding struct {
	Item     string `json:"item"`
	Field    string `json:"field"`
	Env      string `json:"env"`
	Required *bool  `json:"required,omitempty"`
	Title    string `json:"title,omitempty"`
}

func (m Manifest) MarshalJSON() ([]byte, error) {
	switch m.Version {
	case CurrentVersion:
		type v1Manifest struct {
			Version int     `json:"version"`
			Secrets []Entry `json:"secrets"`
		}
		secrets := m.Secrets
		if secrets == nil {
			secrets = []Entry{}
		}
		return json.Marshal(v1Manifest{Version: m.Version, Secrets: secrets})
	case V2Version:
		type v2Manifest struct {
			Version  int         `json:"version"`
			Bindings []V2Binding `json:"bindings"`
		}
		bindings := m.Bindings
		if bindings == nil {
			bindings = []V2Binding{}
		}
		return json.Marshal(v2Manifest{Version: m.Version, Bindings: bindings})
	default:
		type rawManifest Manifest
		return json.Marshal(rawManifest(m))
	}
}

func New() Manifest {
	return Manifest{Version: CurrentVersion, Secrets: []Entry{}}
}

func NewV2() Manifest {
	return Manifest{Version: V2Version, Bindings: []V2Binding{}}
}

func (m Manifest) IsV2() bool {
	return m.Version == V2Version
}

func (e Entry) IsRequired() bool {
	return e.Required == nil || *e.Required
}

func (b V2Binding) IsRequired() bool {
	return b.Required == nil || *b.Required
}

// Key returns the unique identifier for this entry: path, prefix, or comma-joined tags.
func (e Entry) Key() string {
	if e.Path != "" {
		return e.Path
	}
	if e.Prefix != "" {
		return e.Prefix
	}
	return strings.Join(e.Tags, ",")
}

func (b V2Binding) Key() string {
	return b.Item + "." + b.Field
}

// IsPrefix returns true if this is a prefix entry.
func (e Entry) IsPrefix() bool {
	return e.Prefix != ""
}

// IsTag returns true if this is a tag-selected entry.
func (e Entry) IsTag() bool {
	return len(e.Tags) > 0
}

// AddEntry appends an entry. Returns error if an entry with the same path, prefix, or tags already exists.
func (m *Manifest) AddEntry(entry Entry) error {
	for _, existing := range m.Secrets {
		if entry.Path != "" && existing.Path == entry.Path {
			return fmt.Errorf("entry with path %q already exists", entry.Path)
		}
		if entry.Prefix != "" && existing.Prefix == entry.Prefix {
			return fmt.Errorf("entry with prefix %q already exists", entry.Prefix)
		}
		if entry.IsTag() && existing.IsTag() && entry.Key() == existing.Key() {
			return fmt.Errorf("entry with tags %q already exists", entry.Key())
		}
	}
	m.Secrets = append(m.Secrets, entry)
	return nil
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

// RemoveEntry removes an entry by path, prefix, or comma-joined tags. Returns false if not found.
func (m *Manifest) RemoveEntry(key string) bool {
	for i, entry := range m.Secrets {
		if entry.Key() == key {
			m.Secrets = append(m.Secrets[:i], m.Secrets[i+1:]...)
			return true
		}
	}
	return false
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

// FindEntry looks up an entry by path, prefix, or comma-joined tags.
func (m *Manifest) FindEntry(key string) (Entry, bool) {
	for _, entry := range m.Secrets {
		if entry.Key() == key {
			return entry, true
		}
	}
	return Entry{}, false
}

func (m *Manifest) FindV2Binding(key string) (V2Binding, bool) {
	for _, binding := range m.Bindings {
		if binding.Key() == key {
			return binding, true
		}
	}
	return V2Binding{}, false
}
