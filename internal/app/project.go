package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/project"
	"github.com/zhongyangchuwu/shelf-go/internal/util"
)

type ProjectAddRequest struct {
	Selector string
	Env      string
	Optional bool
	Tags     []string
}

type ProjectExportResult struct {
	Diagnostics string
	Output      string
}

func ProjectID() (string, error) {
	return project.ID()
}

func ProjectInit(force bool) (string, error) {
	root, err := project.Root()
	if err != nil {
		return "", err
	}
	manifestPath := filepath.Join(root, project.FileName)
	existed := false
	if _, err := os.Stat(manifestPath); err == nil {
		existed = true
		if !force {
			return "", fmt.Errorf("%s already exists (use --force to overwrite)", project.FileName)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err := project.Save(manifestPath, project.New()); err != nil {
		return "", err
	}
	label := map[bool]string{true: "overwritten", false: "created"}
	var out strings.Builder
	fmt.Fprintf(&out, "manifest: %s (%s)\n", manifestPath, label[existed])
	renderProjectEnvFileStatuses(&out, root, nil)
	return out.String(), nil
}

func (a *App) ProjectStatus(configPathFlag, vaultPathFlag string, parentEnv []string) (string, error) {
	root, manifest, err := loadProjectManifest()
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("%s not found in %s; run `shelf project init`", project.FileName, root)
	}
	if err != nil {
		return "", err
	}
	var out strings.Builder
	fmt.Fprintf(&out, "project: %s\n", project.IDBestEffort(root))
	fmt.Fprintf(&out, "root:    %s\n", root)
	fmt.Fprintf(&out, "config:  %s\n\n", project.FileName)

	_, st, err := a.LoadRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return out.String(), err
	}
	resolvedEntries, diagnostics := project.ResolveEntries(manifest, st)
	writeProjectDiagnostics(&out, diagnostics)
	for _, entry := range resolvedEntries {
		fmt.Fprintf(&out, "ok   %s -> %s\n", entry.Path, entry.EnvName)
	}
	for _, warning := range project.EnvOverrideWarnings(resolvedEntries, parentEnv) {
		fmt.Fprintln(&out, warning)
	}
	renderProjectEnvFileStatuses(&out, root, resolvedEnvNameSet(resolvedEntries))
	if project.HasFailures(diagnostics) {
		return out.String(), fmt.Errorf("project manifest check failed")
	}
	return out.String(), nil
}

func (a *App) ProjectAdd(configPathFlag, vaultPathFlag string, req ProjectAddRequest) (string, error) {
	root, err := project.Root()
	if err != nil {
		return "", err
	}
	manifestPath := filepath.Join(root, project.FileName)
	if _, err := os.Stat(manifestPath); errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("%s not found; run `shelf project init` first", project.FileName)
	} else if err != nil {
		return "", err
	}
	_, st, err := a.LoadRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return "", err
	}

	manifest, err := project.Load(manifestPath)
	if err != nil {
		return "", err
	}
	manifest, entry, err := project.AddEntry(manifest, st, project.AddEntryRequest{Selector: req.Selector, Env: req.Env, Optional: req.Optional, Tags: req.Tags})
	if err != nil {
		return "", err
	}
	if err := project.Save(manifestPath, manifest); err != nil {
		return "", err
	}
	return fmt.Sprintf("added %s\n", entry.Key()), nil
}

func ProjectRm(key string) (string, error) {
	root, manifest, err := loadProjectManifest()
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("%s not found", project.FileName)
	}
	if err != nil {
		return "", err
	}
	if !manifest.RemoveEntry(key) {
		return "", fmt.Errorf("entry not found: %s", key)
	}
	if err := project.Save(filepath.Join(root, project.FileName), manifest); err != nil {
		return "", err
	}
	return fmt.Sprintf("removed %s\n", key), nil
}

func ProjectList() (string, error) {
	_, manifest, err := loadProjectManifest()
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("%s not found", project.FileName)
	}
	if err != nil {
		return "", err
	}

	var out strings.Builder
	for _, entry := range manifest.Secrets {
		req := "required"
		if !entry.IsRequired() {
			req = "optional"
		}
		if entry.IsPrefix() {
			fmt.Fprintf(&out, "prefix %s (%s)\n", entry.Prefix, req)
		} else if entry.IsTag() {
			fmt.Fprintf(&out, "tag    %s (%s)\n", entry.Key(), req)
		} else if entry.Env != "" {
			fmt.Fprintf(&out, "path   %s -> %s (%s)\n", entry.Path, entry.Env, req)
		} else {
			fmt.Fprintf(&out, "path   %s (%s)\n", entry.Path, req)
		}
	}
	return out.String(), nil
}

func (a *App) ProjectExport(configPathFlag, vaultPathFlag, format string) (ProjectExportResult, error) {
	_, manifest, err := loadProjectManifest()
	if errors.Is(err, os.ErrNotExist) {
		return ProjectExportResult{}, fmt.Errorf("%s not found", project.FileName)
	}
	if err != nil {
		return ProjectExportResult{}, err
	}
	_, st, err := a.LoadRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return ProjectExportResult{}, err
	}

	resolvedEntries, diagnostics := project.ResolveEntries(manifest, st)
	var diagnosticsOut strings.Builder
	writeProjectDiagnostics(&diagnosticsOut, diagnostics)
	result := ProjectExportResult{Diagnostics: diagnosticsOut.String()}
	if project.HasFailures(diagnostics) {
		return result, fmt.Errorf("project export failed")
	}
	if len(resolvedEntries) == 0 {
		return result, fmt.Errorf("no secrets to export")
	}

	bindings := project.BindingsForRender(resolvedEntries)
	switch format {
	case "env":
		result.Output, err = util.EnvBindings(bindings)
	case "shell":
		result.Output, err = util.ShellBindings(bindings)
	case "json":
		result.Output, err = util.JSONBindings(bindings)
	default:
		return result, fmt.Errorf("unsupported format: %s", format)
	}
	if err != nil {
		return result, err
	}
	return result, nil
}

func ProjectEntryCompletions(toComplete string) ([]string, error) {
	_, manifest, err := loadProjectManifest()
	if err != nil {
		return nil, err
	}
	completions := make([]string, 0, len(manifest.Secrets))
	for _, entry := range manifest.Secrets {
		key := entry.Key()
		if strings.HasPrefix(key, toComplete) {
			completions = append(completions, key)
		}
	}
	return completions, nil
}

func loadProjectManifest() (string, project.Manifest, error) {
	root, err := project.Root()
	if err != nil {
		return "", project.Manifest{}, err
	}
	manifest, err := project.Load(filepath.Join(root, project.FileName))
	return root, manifest, err
}

func writeProjectDiagnostics(out *strings.Builder, diagnostics []project.Diagnostic) {
	for _, diagnostic := range diagnostics {
		fmt.Fprintf(out, "%s %s %s\n", diagnostic.Status, diagnostic.Path, diagnostic.Message)
	}
}

func renderProjectEnvFileStatuses(out *strings.Builder, root string, boundEnvNames map[string]struct{}) {
	if out.Len() > 0 {
		fmt.Fprintln(out)
	}
	project.RenderEnvFileStatuses(out, project.EnvFileStatuses(root, boundEnvNames))
}

func resolvedEnvNameSet(entries []project.Binding) map[string]struct{} {
	names := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		names[entry.EnvName] = struct{}{}
	}
	return names
}
