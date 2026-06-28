package app

import (
	"fmt"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/exportfmt"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

type ExportRequest struct {
	Selector string
	Tags     []string
	All      bool
	Format   string
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
	switch req.Format {
	case "json":
		return exportfmt.JSON(paths, st.Data.Secrets)
	case "env":
		return exportfmt.Env(paths, st.Data.Secrets)
	case "shell":
		return exportfmt.Shell(paths, st.Data.Secrets)
	default:
		return "", fmt.Errorf("unsupported format: %s", req.Format)
	}
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
