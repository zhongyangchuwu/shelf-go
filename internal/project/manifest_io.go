package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func Load(path string) (Manifest, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	dec := json.NewDecoder(bytes.NewReader(content))
	dec.DisallowUnknownFields()
	var manifest Manifest
	if err := dec.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("invalid project manifest JSON: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return Manifest{}, fmt.Errorf("invalid project manifest JSON: trailing content")
		}
		return Manifest{}, fmt.Errorf("invalid project manifest JSON: %w", err)
	}
	if err := Validate(manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func Save(path string, manifest Manifest) error {
	if err := Validate(manifest); err != nil {
		return err
	}
	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return vault.Write(path, content, vault.Options{FileMode: 0o644, DirMode: 0o700, Sync: true})
}
