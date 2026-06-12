package manifest

import "fmt"

const (
	CurrentVersion = 1
	FileName       = ".shelf.json"
)

type Manifest struct {
	Version int     `json:"version"`
	Secrets []Entry `json:"secrets"`
}

type Entry struct {
	Path     string `json:"path,omitempty"`
	Prefix   string `json:"prefix,omitempty"`
	Env      string `json:"env,omitempty"`
	Required *bool  `json:"required,omitempty"`
}

func New() Manifest {
	return Manifest{Version: CurrentVersion, Secrets: []Entry{}}
}

func (e Entry) IsRequired() bool {
	return e.Required == nil || *e.Required
}

// Key returns the unique identifier for this entry: path if set, otherwise prefix.
func (e Entry) Key() string {
	if e.Path != "" {
		return e.Path
	}
	return e.Prefix
}

// IsPrefix returns true if this is a prefix entry.
func (e Entry) IsPrefix() bool {
	return e.Prefix != ""
}

// AddEntry appends an entry. Returns error if an entry with the same path or prefix already exists.
func (m *Manifest) AddEntry(entry Entry) error {
	for _, existing := range m.Secrets {
		if entry.Path != "" && existing.Path == entry.Path {
			return fmt.Errorf("entry with path %q already exists", entry.Path)
		}
		if entry.Prefix != "" && existing.Prefix == entry.Prefix {
			return fmt.Errorf("entry with prefix %q already exists", entry.Prefix)
		}
	}
	m.Secrets = append(m.Secrets, entry)
	return nil
}

// RemoveEntry removes an entry by path or prefix. Returns false if not found.
func (m *Manifest) RemoveEntry(key string) bool {
	for i, entry := range m.Secrets {
		if entry.Path == key || entry.Prefix == key {
			m.Secrets = append(m.Secrets[:i], m.Secrets[i+1:]...)
			return true
		}
	}
	return false
}

// FindEntry looks up an entry by path or prefix.
func (m *Manifest) FindEntry(key string) (Entry, bool) {
	for _, entry := range m.Secrets {
		if entry.Path == key || entry.Prefix == key {
			return entry, true
		}
	}
	return Entry{}, false
}
