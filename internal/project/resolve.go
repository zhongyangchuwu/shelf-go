package project

import (
	"fmt"
	"io"

	"github.com/zhongyangchuwu/shelf-go/internal/util"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
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

func ResolveEntries(m Manifest, st *vault.Store) ([]Binding, []Diagnostic) {
	entries := make([]Binding, 0)
	diagnostics := make([]Diagnostic, 0)
	envOwners := map[string]string{}

	for _, entry := range m.Secrets {
		if entry.IsPrefix() {
			matches := list(st, entry.Prefix)
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
				secret, ok := get(st, path)
				if !ok {
					continue
				}
				appendResolvedEntry(&entries, &diagnostics, envOwners, path, secret, "")
			}
			continue
		}

		if entry.IsTag() {
			matches := listByTags(st, "", entry.Tags)
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
				secret, ok := get(st, path)
				if !ok {
					continue
				}
				appendResolvedEntry(&entries, &diagnostics, envOwners, path, secret, "")
			}
			continue
		}

		secret, ok := get(st, entry.Path)
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

func get(st *vault.Store, path string) (vault.Secret, bool) {
	if st == nil {
		return vault.Secret{}, false
	}
	return st.Get(path)
}

func list(st *vault.Store, prefix string) []string {
	if st == nil {
		return nil
	}
	return st.List(prefix)
}

func listByTags(st *vault.Store, prefix string, tags []string) []string {
	if st == nil {
		return nil
	}
	return st.ListByTags(prefix, tags)
}

func appendResolvedEntry(entries *[]Binding, diagnostics *[]Diagnostic, envOwners map[string]string, path string, secret vault.Secret, envOverride string) {
	envName := envOverride
	if envName == "" {
		var err error
		envName, err = vault.EnvName(path, secret)
		if err != nil {
			*diagnostics = append(*diagnostics, Diagnostic{Status: "fail", Path: path, Message: err.Error()})
			return
		}
	}
	value, err := util.ValueString(secret.Value)
	if err != nil {
		*diagnostics = append(*diagnostics, Diagnostic{Status: "fail", Path: path, Message: err.Error()})
		return
	}
	if owner, exists := envOwners[envName]; exists {
		*diagnostics = append(*diagnostics, Diagnostic{Status: "fail", Path: path, Message: fmt.Sprintf("env name %s conflicts with %s", envName, owner)})
		return
	}
	envOwners[envName] = path
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

func BindingsForRender(entries []Binding) []util.Binding {
	bindings := make([]util.Binding, 0, len(entries))
	for _, entry := range entries {
		bindings = append(bindings, util.Binding{EnvName: entry.EnvName, Value: entry.Value})
	}
	return bindings
}
