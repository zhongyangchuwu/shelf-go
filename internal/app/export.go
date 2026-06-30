package app

import (
	"fmt"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/util"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

type ExportRequest struct {
	Selector string
	Tags     []string
	All      bool
	Format   string
}

func (a *App) ExportSecretsForRuntime(configPathFlag, vaultPathFlag string, req ExportRequest) (string, error) {
	_, st, err := a.LoadRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return "", err
	}
	return ExportSecrets(st, req)
}

func ExportSecrets(st *vault.Store, req ExportRequest) (string, error) {
	var paths []string
	if req.Selector != "" {
		if len(req.Tags) == 0 {
			if _, ok := st.Get(req.Selector); ok {
				paths = []string{req.Selector}
			} else {
				paths = st.List(req.Selector)
			}
		} else if secret, ok := st.Get(req.Selector); ok {
			if vault.HasTags(secret, req.Tags) {
				paths = []string{req.Selector}
			}
		} else {
			paths = st.ListByTags(req.Selector, req.Tags)
		}
	} else if len(req.Tags) > 0 {
		paths = st.ListByTags("", req.Tags)
	} else {
		return "", fmt.Errorf("path, prefix, or --tag is required")
	}
	if !req.All {
		filtered := make([]string, 0, len(paths))
		for _, path := range paths {
			secret, ok := st.Get(path)
			if ok && secret.Env != "" {
				filtered = append(filtered, path)
			}
		}
		paths = filtered
	}
	if len(paths) == 0 {
		return "", fmt.Errorf("no secrets matched: %s", exportSelector(req.Selector, req.Tags))
	}
	bindings, err := secretBindings(paths, st.Data.Secrets)
	if err != nil {
		return "", err
	}
	switch req.Format {
	case "json":
		return util.JSONBindings(bindings)
	case "env":
		return util.EnvBindings(bindings)
	case "shell":
		return util.ShellBindings(bindings)
	default:
		return "", fmt.Errorf("unsupported format: %s", req.Format)
	}
}

func secretBindings(paths []string, secrets map[string]vault.Secret) ([]util.Binding, error) {
	bindings := make([]util.Binding, 0, len(paths))
	for _, path := range paths {
		secret := secrets[path]
		envName, err := vault.EnvName(path, secret)
		if err != nil {
			return nil, err
		}
		value, err := util.ValueString(secret.Value)
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, util.Binding{EnvName: envName, Value: value})
	}
	return bindings, nil
}

func exportSelector(selector string, tags []string) string {
	if selector == "" {
		return "tag " + strings.Join(tags, ",")
	}
	if len(tags) == 0 {
		return selector
	}
	return selector + " with tag " + strings.Join(tags, ",")
}
