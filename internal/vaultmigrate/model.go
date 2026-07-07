package vaultmigrate

import "github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"

const ProjectManifestVersion = 2

type PathMapping struct {
	OldPath string
	ItemID  string
	FieldID string
	EnvName string
	Title   string
	Labels  []string `json:"-"`
}

type VaultResult struct {
	Vault vaultmodel.Vault
	Paths map[string]PathMapping
}

type ProjectManifest struct {
	Version  int              `json:"version"`
	Bindings []ProjectBinding `json:"bindings"`
}

type ProjectBinding struct {
	Item     string `json:"item"`
	Field    string `json:"field"`
	Env      string `json:"env"`
	Required bool   `json:"required"`
	Title    string `json:"title,omitempty"`
}

type ProjectResult struct {
	Manifest ProjectManifest
	Warnings []Warning
}

type Warning struct {
	Selector string
	Message  string
}
