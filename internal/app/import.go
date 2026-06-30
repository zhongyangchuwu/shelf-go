package app

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/importer/gopass"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultfile"
)

type GopassImportOptions struct {
	Prefix  string
	Command string
	Force   bool
	DryRun  bool
}

type ImportSkip struct {
	Path   string
	Reason string
}

type ImportResult struct {
	Imported        []string
	SkippedExisting []string
	SkippedInvalid  []ImportSkip
	DryRun          bool
}

type gopassImportClient interface {
	ListFlat(prefix string) ([]string, error)
	ShowPassword(path string) (string, error)
}

func ImportGopassForRuntime(configPathFlag, vaultPathFlag string, opts GopassImportOptions) (ImportResult, error) {
	_, v, err := LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return ImportResult{}, err
	}
	client := gopass.NewClient(opts.Command)
	return ImportGopassToVault(v, client, opts)
}

func ImportGopassToVault(v *vaultfile.Vault, client gopassImportClient, opts GopassImportOptions) (ImportResult, error) {
	if v == nil {
		return ImportResult{}, fmt.Errorf("vault is required")
	}
	if client == nil {
		return ImportResult{}, fmt.Errorf("gopass client is required")
	}
	result := ImportResult{DryRun: opts.DryRun}
	entries, err := planGopassImport(client, strings.Trim(opts.Prefix, "/"), &result)
	if err != nil {
		return result, err
	}
	if opts.DryRun {
		for _, entry := range entries {
			result.Imported = append(result.Imported, entry.shelfPath)
		}
		sort.Strings(result.Imported)
		return result, nil
	}
	if len(entries) == 0 {
		return result, nil
	}
	if err := v.Update(func(st *vault.Store) error {
		for _, entry := range entries {
			if _, exists := st.Get(entry.shelfPath); exists && !opts.Force {
				result.SkippedExisting = append(result.SkippedExisting, entry.shelfPath)
				continue
			}
			secret := vault.Secret{Value: entry.value}
			if err := st.Set(entry.shelfPath, secret, opts.Force); err != nil {
				return err
			}
			result.Imported = append(result.Imported, entry.shelfPath)
		}
		return nil
	}); err != nil {
		return result, err
	}
	sort.Strings(result.Imported)
	sort.Strings(result.SkippedExisting)
	return result, nil
}

type plannedGopassImport struct {
	gopassPath string
	shelfPath  string
	value      json.RawMessage
}

func planGopassImport(client gopassImportClient, prefix string, result *ImportResult) ([]plannedGopassImport, error) {
	paths, err := client.ListFlat(prefix)
	if err != nil {
		return nil, err
	}
	entries := make([]plannedGopassImport, 0, len(paths))
	seen := map[string]string{}
	for _, gopassPath := range paths {
		shelfPath, ok, reason := MapGopassPath(gopassPath)
		if !ok {
			result.SkippedInvalid = append(result.SkippedInvalid, ImportSkip{Path: gopassPath, Reason: reason})
			continue
		}
		if previous, exists := seen[shelfPath]; exists {
			result.SkippedInvalid = append(result.SkippedInvalid, ImportSkip{Path: gopassPath, Reason: "maps to same Shelf path as " + previous})
			continue
		}
		seen[shelfPath] = gopassPath
		password, err := client.ShowPassword(gopassPath)
		if err != nil {
			return nil, fmt.Errorf("read gopass secret %s: %w", gopassPath, err)
		}
		value, err := json.Marshal(password)
		if err != nil {
			return nil, fmt.Errorf("encode gopass secret %s: %w", gopassPath, err)
		}
		entries = append(entries, plannedGopassImport{gopassPath: gopassPath, shelfPath: shelfPath, value: value})
	}
	return entries, nil
}

func MapGopassPath(path string) (string, bool, string) {
	path = strings.Trim(path, "/")
	idx := strings.LastIndex(path, "/")
	if idx <= 0 || idx == len(path)-1 {
		return "", false, "gopass path must contain group and key separated by '/'"
	}
	mapped := path[:idx] + ":" + path[idx+1:]
	if err := vault.ValidatePath(mapped); err != nil {
		return "", false, err.Error()
	}
	return mapped, true, ""
}
