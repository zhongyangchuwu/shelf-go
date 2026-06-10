package store

import "encoding/json"

const CurrentVersion = 1

type Data struct {
	Version int               `json:"version"`
	Secrets map[string]Secret `json:"secrets"`
}

type Secret struct {
	Value       json.RawMessage `json:"value"`
	Env         string          `json:"env,omitempty"`
	Description string          `json:"description,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
}

type Info struct {
	Path        string   `json:"path"`
	GroupPath   string   `json:"group_path"`
	Key         string   `json:"key"`
	ValueSet    bool     `json:"value_set"`
	Env         string   `json:"env,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags"`
}

func NewData() Data {
	return Data{Version: CurrentVersion, Secrets: map[string]Secret{}}
}
