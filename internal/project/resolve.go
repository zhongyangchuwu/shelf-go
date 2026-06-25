package project

import (
	"fmt"
	"io"

	"github.com/zhongyangchuwu/shelf-go/internal/manifest"
	"github.com/zhongyangchuwu/shelf-go/internal/render"
	"github.com/zhongyangchuwu/shelf-go/internal/store"
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

func ResolveEntries(m manifest.Manifest, st *store.Store) ([]Binding, []Diagnostic) {
	entries := make([]Binding, 0)
	diagnostics := make([]Diagnostic, 0)
	envOwners := map[string]string{}

	for _, entry := range m.Secrets {
		if entry.IsPrefix() {
			matches := st.List(entry.Prefix)
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
				secret, ok := st.Get(path)
				if !ok {
					continue
				}
				appendResolvedEntry(&entries, &diagnostics, envOwners, path, secret, "")
			}
			continue
		}

		secret, ok := st.Get(entry.Path)
		if !ok {
			if entry.IsRequired() {
				diagnostics = append(diagnostics, Diagnostic{Status: "fail", Path: entry.Path, Message: "missing required"})
			} else {
				diagnostics = append(diagnostics, Diagnostic{Status: "warn", Path: entry.Path, Message: "missing optional"})
			}
			continue
		}
		appendResolvedEntry(&entries, &diagnostics, envOwners, entry.Path, secret, entry.Env)
	}

	return entries, diagnostics
}

func appendResolvedEntry(entries *[]Binding, diagnostics *[]Diagnostic, envOwners map[string]string, path string, secret store.Secret, envOverride string) {
	envName := envOverride
	if envName == "" {
		var err error
		envName, err = render.EnvName(path, secret)
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
	value, err := render.ValueString(secret.Value)
	if err != nil {
		*diagnostics = append(*diagnostics, Diagnostic{Status: "fail", Path: path, Message: err.Error()})
		return
	}
	*entries = append(*entries, Binding{Path: path, EnvName: envName, Value: value})
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

func BindingsForRender(entries []Binding) []render.Binding {
	bindings := make([]render.Binding, 0, len(entries))
	for _, entry := range entries {
		bindings = append(bindings, render.Binding{EnvName: entry.EnvName, Value: entry.Value})
	}
	return bindings
}
