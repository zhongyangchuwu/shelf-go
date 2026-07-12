package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/project"
	"github.com/zhongyangchuwu/shelf-go/internal/util"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

type ProjectAddRequest struct {
	Selector string
	Field    string
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
	return fmt.Sprintf("manifest: %s (%s)\n", manifestPath, label[existed]), nil
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
	healthReport, err := a.CredentialAudit(configPathFlag, vaultPathFlag, CredentialAuditRequest{WithinDays: DefaultCredentialAuditWithinDays})
	if err != nil {
		return out.String(), err
	}
	healthByItem := credentialHealthByItem(healthReport)
	writeProjectDiagnostics(&out, diagnostics)
	for _, entry := range resolvedEntries {
		fmt.Fprintf(&out, "ok   %s -> %s", entry.Path, entry.EnvName)
		if itemID, fieldID, ok := strings.Cut(entry.Path, "."); ok {
			if health, exists := healthByItem[itemID]; exists && health.Field == fieldID {
				fmt.Fprintf(&out, " health:%s", health.Status)
				if health.Expires != "" {
					fmt.Fprintf(&out, " expires:%s", health.Expires)
				}
			}
		}
		fmt.Fprintln(&out)
	}
	for _, warning := range project.EnvOverrideWarnings(resolvedEntries, parentEnv) {
		fmt.Fprintln(&out, warning)
	}
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
	manifest, err := project.Load(manifestPath)
	if err != nil {
		return "", err
	}
	vault, err := a.loadVaultForRead(configPathFlag, vaultPathFlag)
	if err != nil {
		return "", err
	}
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	binding, err := buildProjectBinding(resolver, req)
	if err != nil {
		return "", err
	}
	if err := manifest.AddBinding(binding); err != nil {
		return "", err
	}
	if err := project.Save(manifestPath, manifest); err != nil {
		return "", err
	}
	return fmt.Sprintf("added %s\n", binding.Key()), nil
}

func (a *App) ProjectAddCompletions(configPathFlag, vaultPathFlag string) ([]string, error) {
	vault, err := a.loadVaultForRead(configPathFlag, vaultPathFlag)
	if err != nil {
		return nil, err
	}
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	return resolver.SecretSelectors(), nil
}

func ProjectRm(key string) (string, error) {
	root, manifest, err := loadProjectManifest()
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("%s not found", project.FileName)
	}
	if err != nil {
		return "", err
	}
	if !manifest.RemoveBinding(key) {
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
	for _, binding := range manifest.Bindings {
		req := "required"
		if !binding.IsRequired() {
			req = "optional"
		}
		fmt.Fprintf(&out, "item   %s.%s -> %s (%s)\n", binding.Item, binding.Field, binding.Env, req)
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
	completions := make([]string, 0, len(manifest.Bindings))
	for _, binding := range manifest.Bindings {
		key := binding.Key()
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

func (a *App) resolveProjectBindings(configPathFlag, vaultPathFlag string, manifest project.Manifest) ([]project.ResolvedBinding, []project.Diagnostic, error) {
	if len(manifest.Bindings) == 0 {
		return nil, nil, nil
	}
	vault, err := a.loadVaultForRead(configPathFlag, vaultPathFlag)
	if err != nil {
		return nil, nil, err
	}
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	entries, diagnostics := project.ResolveBindings(manifest, resolver)
	return entries, diagnostics, nil
}

func buildProjectBinding(resolver vaultmodel.Resolver, req ProjectAddRequest) (project.Binding, error) {
	selector := req.Selector
	if selector == "" {
		return project.Binding{}, fmt.Errorf("selector is required")
	}
	itemSelector := selector
	fieldID := req.Field
	if before, after, ok := strings.Cut(selector, "."); ok {
		itemSelector = before
		if fieldID == "" {
			fieldID = after
		}
	}
	_, item, err := resolver.ResolveItem(itemSelector)
	if err != nil {
		return project.Binding{}, err
	}
	if fieldID == "" {
		primary, err := resolver.PrimarySecretField(item)
		if err != nil {
			return project.Binding{}, err
		}
		fieldID = primary.ID
	}
	_, field, err := resolver.Field(item, fieldID)
	if err != nil {
		return project.Binding{}, fmt.Errorf("field not found: %s.%s", item.ID, fieldID)
	}
	if field.Sensitivity != vaultmodel.SensitivitySecret {
		return project.Binding{}, fmt.Errorf("field is not secret: %s.%s", item.ID, fieldID)
	}
	envName := req.Env
	if envName == "" {
		var ok bool
		envName, ok, err = resolver.EnvName(item, fieldID)
		if err != nil {
			return project.Binding{}, err
		}
		if !ok {
			envName = ""
		}
	}
	if envName == "" {
		return project.Binding{}, fmt.Errorf("env name is required for %s.%s", item.ID, fieldID)
	}
	binding := project.Binding{Item: item.ID, Field: fieldID, Env: envName}
	if req.Optional {
		binding.Required = new(bool)
	}
	return binding, nil
}
