package manifest

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
