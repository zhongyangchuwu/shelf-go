package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/project"
	"github.com/zhongyangchuwu/shelf-go/internal/util"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

type ProjectAddRequest struct {
	Selector string
	Field    string
	Env      string
	Optional bool
	Tags     []string
}

type ProjectConfigureState struct {
	EnvFiles    []ProjectConfigureEnvFile
	SecretPaths []string
	Entries     []ProjectConfigureEntry
}

type ProjectConfigureEnvFile struct {
	Name  string
	Vars  []ProjectConfigureEnvVar
	Error string
}

type ProjectConfigureEnvVar struct {
	Name  string
	Empty bool
	Bound bool
}

type ProjectConfigureEntry struct {
	Key      string
	Kind     string
	Selector string
	Env      string
	Required bool
}

type ProjectConfigureActionKind string

const (
	ProjectConfigureAddPath ProjectConfigureActionKind = "add_path"
	ProjectConfigureRemove  ProjectConfigureActionKind = "remove"
)

type ProjectConfigureAction struct {
	Kind ProjectConfigureActionKind
	Env  string
	Path string
	Key  string
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

	resolvedEntries, diagnostics, err := a.resolveProjectBindings(configPathFlag, vaultPathFlag, manifest)
	if err != nil {
		return out.String(), err
	}
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

func (a *App) ProjectConfigureState(configPathFlag, vaultPathFlag string) (ProjectConfigureState, error) {
	root, manifest, err := loadProjectManifest()
	if errors.Is(err, os.ErrNotExist) {
		return ProjectConfigureState{}, fmt.Errorf("%s not found in %s; run `shelf project init`", project.FileName, root)
	}
	if err != nil {
		return ProjectConfigureState{}, err
	}
	var resolvedEntries []project.Binding
	var secretPaths []string
	if manifest.IsV2() {
		vault, err := a.loadV2VaultForRead(configPathFlag, vaultPathFlag)
		if err != nil {
			return ProjectConfigureState{}, err
		}
		resolvedEntries, _ = project.ResolveV2Bindings(manifest, vault)
		secretPaths = v2ProjectSecretSelectors(vault)
	} else {
		_, st, err := a.LoadRuntime(configPathFlag, vaultPathFlag)
		if err != nil {
			return ProjectConfigureState{}, err
		}
		resolvedEntries, _ = project.ResolveEntries(manifest, st)
		secretPaths = st.List("")
	}
	envFiles := projectConfigureEnvFiles(project.EnvFileDetails(root, resolvedEnvNameSet(resolvedEntries)))
	return ProjectConfigureState{
		EnvFiles:    envFiles,
		SecretPaths: secretPaths,
		Entries:     projectConfigureEntries(manifest),
	}, nil
}

func (a *App) ApplyProjectConfigureActions(configPathFlag, vaultPathFlag string, actions []ProjectConfigureAction) (string, error) {
	root, manifest, err := loadProjectManifest()
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("%s not found in %s; run `shelf project init`", project.FileName, root)
	}
	if err != nil {
		return "", err
	}
	var v1Store *vault.Store
	var v2Vault vaultmodel.Vault
	if manifest.IsV2() {
		v2Vault, err = a.loadV2VaultForRead(configPathFlag, vaultPathFlag)
	} else {
		_, v1Store, err = a.LoadRuntime(configPathFlag, vaultPathFlag)
	}
	if err != nil {
		return "", err
	}
	for _, action := range actions {
		switch action.Kind {
		case ProjectConfigureAddPath:
			if manifest.IsV2() {
				binding, err := buildV2ProjectBinding(v2Vault, ProjectAddRequest{Selector: action.Path, Env: action.Env})
				if err == nil {
					err = manifest.AddV2Binding(binding)
				}
			} else {
				var entry project.Entry
				manifest, entry, err = project.AddEntry(manifest, v1Store, project.AddEntryRequest{Selector: action.Path, Env: action.Env})
				_ = entry
			}
		case ProjectConfigureRemove:
			removed := manifest.RemoveEntry(action.Key)
			if manifest.IsV2() {
				removed = manifest.RemoveV2Binding(action.Key)
			}
			if !removed {
				err = fmt.Errorf("entry not found: %s", action.Key)
			}
		default:
			err = fmt.Errorf("unsupported project configure action: %s", action.Kind)
		}
		if err != nil {
			return "", err
		}
	}
	if err := project.Save(filepath.Join(root, project.FileName), manifest); err != nil {
		return "", err
	}
	return fmt.Sprintf("applied %d change(s)\n", len(actions)), nil
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
	manifest, err := project.Load(manifestPath)
	if err != nil {
		return "", err
	}
	if manifest.IsV2() {
		vault, err := a.loadV2VaultForRead(configPathFlag, vaultPathFlag)
		if err != nil {
			return "", err
		}
		binding, err := buildV2ProjectBinding(vault, req)
		if err != nil {
			return "", err
		}
		if err := manifest.AddV2Binding(binding); err != nil {
			return "", err
		}
		if err := project.Save(manifestPath, manifest); err != nil {
			return "", err
		}
		return fmt.Sprintf("added %s\n", binding.Key()), nil
	}
	_, st, err := a.LoadRuntime(configPathFlag, vaultPathFlag)
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

func (a *App) ProjectAddCompletions(configPathFlag, vaultPathFlag string) ([]string, error) {
	_, manifest, err := loadProjectManifest()
	if err != nil {
		return nil, err
	}
	if manifest.IsV2() {
		vault, err := a.loadV2VaultForRead(configPathFlag, vaultPathFlag)
		if err != nil {
			return nil, err
		}
		return v2ProjectSecretSelectors(vault), nil
	}
	return a.AllSecretPaths(configPathFlag, vaultPathFlag)
}

func ProjectRm(key string) (string, error) {
	root, manifest, err := loadProjectManifest()
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("%s not found", project.FileName)
	}
	if err != nil {
		return "", err
	}
	removed := manifest.RemoveEntry(key)
	if manifest.IsV2() {
		removed = manifest.RemoveV2Binding(key)
	}
	if !removed {
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
	if manifest.IsV2() {
		for _, binding := range manifest.Bindings {
			req := "required"
			if !binding.IsRequired() {
				req = "optional"
			}
			fmt.Fprintf(&out, "item   %s.%s -> %s (%s)\n", binding.Item, binding.Field, binding.Env, req)
		}
		return out.String(), nil
	}
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
	resolvedEntries, diagnostics, err := a.resolveProjectBindings(configPathFlag, vaultPathFlag, manifest)
	if err != nil {
		return ProjectExportResult{}, err
	}
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

func projectConfigureEnvFiles(files []project.EnvFileDetail) []ProjectConfigureEnvFile {
	converted := make([]ProjectConfigureEnvFile, 0, len(files))
	for _, file := range files {
		converted = append(converted, ProjectConfigureEnvFile{
			Name:  file.Name,
			Vars:  projectConfigureEnvVars(file.Vars),
			Error: file.Error,
		})
	}
	return converted
}

func projectConfigureEnvVars(vars []project.EnvVarStatus) []ProjectConfigureEnvVar {
	converted := make([]ProjectConfigureEnvVar, 0, len(vars))
	for _, variable := range vars {
		converted = append(converted, ProjectConfigureEnvVar{
			Name:  variable.Name,
			Empty: variable.Empty,
			Bound: variable.Bound,
		})
	}
	return converted
}

func projectConfigureEntries(manifest project.Manifest) []ProjectConfigureEntry {
	if manifest.IsV2() {
		entries := make([]ProjectConfigureEntry, 0, len(manifest.Bindings))
		for _, binding := range manifest.Bindings {
			entries = append(entries, ProjectConfigureEntry{Kind: "item", Key: binding.Key(), Selector: binding.Key(), Env: binding.Env, Required: binding.IsRequired()})
		}
		return entries
	}
	entries := make([]ProjectConfigureEntry, 0, len(manifest.Secrets))
	for _, entry := range manifest.Secrets {
		item := ProjectConfigureEntry{Key: entry.Key(), Required: entry.IsRequired()}
		switch {
		case entry.IsPrefix():
			item.Kind = "prefix"
			item.Selector = entry.Prefix
		case entry.IsTag():
			item.Kind = "tag"
			item.Selector = entry.Key()
		default:
			item.Kind = "path"
			item.Selector = entry.Path
			item.Env = entry.Env
		}
		entries = append(entries, item)
	}
	return entries
}

func ProjectEntryCompletions(toComplete string) ([]string, error) {
	_, manifest, err := loadProjectManifest()
	if err != nil {
		return nil, err
	}
	if manifest.IsV2() {
		completions := make([]string, 0, len(manifest.Bindings))
		for _, binding := range manifest.Bindings {
			key := binding.Key()
			if strings.HasPrefix(key, toComplete) {
				completions = append(completions, key)
			}
		}
		return completions, nil
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

// resolveProjectBindings resolves entries for v1 or v2 manifests.
func (a *App) resolveProjectBindings(configPathFlag, vaultPathFlag string, manifest project.Manifest) ([]project.Binding, []project.Diagnostic, error) {
	if manifest.IsV2() {
		vault, err := a.loadV2VaultForRead(configPathFlag, vaultPathFlag)
		if err != nil {
			return nil, nil, err
		}
		entries, diagnostics := project.ResolveV2Bindings(manifest, vault)
		return entries, diagnostics, nil
	}
	_, st, err := a.LoadRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return nil, nil, err
	}
	entries, diagnostics := project.ResolveEntries(manifest, st)
	return entries, diagnostics, nil
}

// buildV2ProjectBinding builds a V2Binding from a ProjectAddRequest and a v2 vault.
func buildV2ProjectBinding(vault vaultmodel.Vault, req ProjectAddRequest) (project.V2Binding, error) {
	selector := req.Selector
	if selector == "" {
		return project.V2Binding{}, fmt.Errorf("selector is required")
	}
	itemSelector := selector
	fieldID := req.Field
	if before, after, ok := strings.Cut(selector, "."); ok {
		itemSelector = before
		if fieldID == "" {
			fieldID = after
		}
	}
	item, err := resolveV2ProjectItem(vault, itemSelector)
	if err != nil {
		return project.V2Binding{}, err
	}
	if fieldID == "" {
		fieldID, err = primaryV2ProjectField(vault, item)
		if err != nil {
			return project.V2Binding{}, err
		}
	}
	field, ok := v2FieldByID(item, fieldID)
	if !ok {
		return project.V2Binding{}, fmt.Errorf("field not found: %s.%s", item.ID, fieldID)
	}
	if field.Sensitivity != vaultmodel.SensitivitySecret {
		return project.V2Binding{}, fmt.Errorf("field is not secret: %s.%s", item.ID, fieldID)
	}
	envName := req.Env
	if envName == "" {
		envName = v2EnvNameForField(vault, item, fieldID)
	}
	if envName == "" {
		return project.V2Binding{}, fmt.Errorf("env name is required for %s.%s", item.ID, fieldID)
	}
	binding := project.V2Binding{Item: item.ID, Field: fieldID, Env: envName}
	if req.Optional {
		binding.Required = new(bool)
	}
	return binding, nil
}

// v2ProjectSecretSelectors returns item.field selectors for all secret fields in the vault.
func v2ProjectSecretSelectors(vault vaultmodel.Vault) []string {
	var selectors []string
	for _, item := range vault.Items {
		for _, field := range item.Fields {
			if field.Sensitivity == vaultmodel.SensitivitySecret {
				selectors = append(selectors, item.ID+"."+field.ID)
			}
		}
	}
	return selectors
}

func resolveV2ProjectItem(vault vaultmodel.Vault, selector string) (vaultmodel.Item, error) {
	var matches []vaultmodel.Item
	for _, item := range vault.Items {
		if item.ID == selector || item.Title == selector || hasString(item.Aliases, selector) {
			matches = append(matches, item)
		}
	}
	if len(matches) == 0 {
		return vaultmodel.Item{}, fmt.Errorf("item not found: %s", selector)
	}
	if len(matches) > 1 {
		var b strings.Builder
		fmt.Fprintf(&b, "selector matched multiple items: %s", selector)
		for _, item := range matches {
			fmt.Fprintf(&b, "\n- %s %s", item.ID, item.Title)
		}
		return vaultmodel.Item{}, fmt.Errorf("%s", b.String())
	}
	return matches[0], nil
}

func primaryV2ProjectField(vault vaultmodel.Vault, item vaultmodel.Item) (string, error) {
	registry := vaultmodel.NewRegistryWithDefinitions(vault.Templates, vault.Profiles)
	tmpl, ok := registry.Template(item.Template)
	if !ok {
		return "", fmt.Errorf("template not found: %s", item.Template)
	}
	for _, field := range tmpl.Fields {
		if field.Purpose == vaultmodel.PurposePrimarySecret {
			return field.ID, nil
		}
	}
	return "", fmt.Errorf("template has no primary secret field: %s", tmpl.ID)
}

func v2FieldByID(item vaultmodel.Item, fieldID string) (vaultmodel.Field, bool) {
	for _, field := range item.Fields {
		if field.ID == fieldID {
			return field, true
		}
	}
	return vaultmodel.Field{}, false
}

func v2EnvNameForField(vault vaultmodel.Vault, item vaultmodel.Item, fieldID string) string {
	registry := vaultmodel.NewRegistryWithDefinitions(vault.Templates, vault.Profiles)
	tmpl, ok := registry.Template(item.Template)
	if !ok {
		return ""
	}
	for _, templateField := range tmpl.Fields {
		if templateField.Purpose != vaultmodel.PurposeEnvName || templateField.EnvFor != fieldID {
			continue
		}
		field, ok := v2FieldByID(item, templateField.ID)
		if !ok || len(field.Value) == 0 {
			return ""
		}
		value, err := stringValue(field.Value)
		if err != nil {
			return ""
		}
		return value
	}
	return ""
}
