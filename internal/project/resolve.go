package project

import (
	"fmt"
	"io"

	"errors"

	"github.com/zhongyangchuwu/shelf-go/internal/source"
	"github.com/zhongyangchuwu/shelf-go/internal/util"
)

type Binding struct {
	Path    string
	EnvName string
	Value   string
}

type Diagnostic struct {
	Status  string
	Path    string
	Message string
}

func ResolveEntries(m Manifest, reader source.Reader) ([]Binding, []Diagnostic) {
	entries := make([]Binding, 0)
	diagnostics := make([]Diagnostic, 0)
	envOwners := map[string]string{}

	for _, entry := range m.Secrets {
		if entry.IsPrefix() {
			matches, err := reader.List(entry.Prefix)
			if err != nil {
				diagnostics = append(diagnostics, Diagnostic{Status: "fail", Path: entry.Prefix + " (prefix)", Message: err.Error()})
				continue
			}
			if len(matches) == 0 {
				path := entry.Prefix + " (prefix)"
				if entry.IsRequired() {
					diagnostics = append(diagnostics, Diagnostic{Status: "fail", Path: path, Message: "no matches required"})
				} else {
					diagnostics = append(diagnostics, Diagnostic{Status: "warn", Path: path, Message: "no matches optional"})
				}
				continue
			}
			for _, path := range matches {
				secret, err := reader.Get(path)
				if errors.Is(err, source.ErrNotFound) {
					continue
				}
				if err != nil {
					diagnostics = append(diagnostics, Diagnostic{Status: "fail", Path: path, Message: err.Error()})
					continue
				}
				appendResolvedEntry(&entries, &diagnostics, envOwners, path, secret, "")
			}
			continue
		}

		if entry.IsTag() {
			matches, err := reader.ListByTags("", entry.Tags)
			if err != nil {
				diagnostics = append(diagnostics, Diagnostic{Status: "fail", Path: entry.Key() + " (tags)", Message: err.Error()})
				continue
			}
			if len(matches) == 0 {
				path := entry.Key() + " (tags)"
				if entry.IsRequired() {
					diagnostics = append(diagnostics, Diagnostic{Status: "fail", Path: path, Message: "no matches required"})
				} else {
					diagnostics = append(diagnostics, Diagnostic{Status: "warn", Path: path, Message: "no matches optional"})
				}
				continue
			}
			for _, path := range matches {
				secret, err := reader.Get(path)
				if errors.Is(err, source.ErrNotFound) {
					continue
				}
				if err != nil {
					diagnostics = append(diagnostics, Diagnostic{Status: "fail", Path: path, Message: err.Error()})
					continue
				}
				appendResolvedEntry(&entries, &diagnostics, envOwners, path, secret, "")
			}
			continue
		}

		secret, err := reader.Get(entry.Path)
		if errors.Is(err, source.ErrNotFound) {
			if entry.IsRequired() {
				diagnostics = append(diagnostics, Diagnostic{Status: "fail", Path: entry.Path, Message: "missing required"})
			} else {
				diagnostics = append(diagnostics, Diagnostic{Status: "warn", Path: entry.Path, Message: "missing optional"})
			}
			continue
		}
		if err != nil {
			diagnostics = append(diagnostics, Diagnostic{Status: "fail", Path: entry.Path, Message: err.Error()})
			continue
		}
		appendResolvedEntry(&entries, &diagnostics, envOwners, entry.Path, secret, entry.Env)
	}

	return entries, diagnostics
}

func appendResolvedEntry(entries *[]Binding, diagnostics *[]Diagnostic, envOwners map[string]string, path string, secret source.Secret, envOverride string) {
	envName := envOverride
	if envName == "" {
		var err error
		envName, err = source.EnvName(path, secret)
		if err != nil {
			*diagnostics = append(*diagnostics, Diagnostic{Status: "fail", Path: path, Message: err.Error()})
			return
		}
	}
	if owner, exists := envOwners[envName]; exists {
		*diagnostics = append(*diagnostics, Diagnostic{Status: "fail", Path: path, Message: fmt.Sprintf("env name %s conflicts with %s", envName, owner)})
		return
	}
	envOwners[envName] = path
	*entries = append(*entries, Binding{Path: path, EnvName: envName, Value: secret.Value})
}

func HasFailures(diagnostics []Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Status == "fail" {
			return true
		}
	}
	return false
}

func RenderDiagnostics(w io.Writer, diagnostics []Diagnostic) {
	for _, diagnostic := range diagnostics {
		fmt.Fprintf(w, "%s %s %s\n", diagnostic.Status, diagnostic.Path, diagnostic.Message)
	}
}

func BindingsForRender(entries []Binding) []util.Binding {
	bindings := make([]util.Binding, 0, len(entries))
	for _, entry := range entries {
		bindings = append(bindings, util.Binding{EnvName: entry.EnvName, Value: entry.Value})
	}
	return bindings
}
