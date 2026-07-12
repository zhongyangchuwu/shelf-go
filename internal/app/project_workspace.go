package app

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/zhongyangchuwu/shelf-go/internal/project"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

type ProjectWorkspace struct {
	ProjectID       string
	Root            string
	ManifestPresent bool
	Bindings        []ProjectBindingStatus
}

type ProjectBindingStatus struct {
	Key      string
	Item     string
	Title    string
	Field    string
	Env      string
	Required bool
	Status   string
	Message  string
	Health   CredentialHealthStatus
	Expires  string
}

type ProjectBindingCandidate struct {
	Item  string
	Title string
	Field string
	Env   string
}

func (a *App) ProjectWorkspace(configPathFlag, vaultPathFlag string) (ProjectWorkspace, error) {
	root, err := project.Root()
	if err != nil {
		return ProjectWorkspace{}, err
	}
	workspace := ProjectWorkspace{ProjectID: project.IDBestEffort(root), Root: root}
	manifest, err := project.Load(filepath.Join(root, project.FileName))
	if errors.Is(err, os.ErrNotExist) {
		return workspace, nil
	}
	if err != nil {
		return ProjectWorkspace{}, err
	}
	workspace.ManifestPresent = true
	workspace.Bindings = make([]ProjectBindingStatus, 0, len(manifest.Bindings))
	if len(manifest.Bindings) == 0 {
		return workspace, nil
	}

	vault, err := a.loadVaultForRead(configPathFlag, vaultPathFlag)
	if err != nil {
		return ProjectWorkspace{}, err
	}
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	healthByItem := credentialHealthByItem(credentialAuditForVault(vault, DefaultCredentialAuditWithinDays, time.Now()))
	statuses := project.ResolveBindingStatuses(manifest, resolver)
	statusByPath := make(map[string]project.BindingStatus, len(statuses))
	for _, bindingStatus := range statuses {
		statusByPath[bindingStatus.Path] = bindingStatus
	}

	for _, binding := range manifest.Bindings {
		status := ProjectBindingStatus{
			Key:      binding.Key(),
			Item:     binding.Item,
			Title:    binding.Title,
			Field:    binding.Field,
			Env:      binding.Env,
			Required: binding.IsRequired(),
			Status:   "ready",
		}
		if item, ok := resolver.ItemByID(binding.Item); ok {
			status.Title = item.Title
		}
		if bindingStatus, ok := statusByPath[status.Key]; ok {
			status.Status = bindingStatus.Status
			status.Message = bindingStatus.Message
		}
		if health, ok := healthByItem[binding.Item]; ok && health.Field == binding.Field {
			status.Health = health.Status
			status.Expires = health.Expires
		}
		workspace.Bindings = append(workspace.Bindings, status)
	}
	return workspace, nil
}

func (a *App) ProjectBindingCandidates(configPathFlag, vaultPathFlag string) ([]ProjectBindingCandidate, error) {
	vault, err := a.loadVaultForRead(configPathFlag, vaultPathFlag)
	if err != nil {
		return nil, err
	}
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	candidates := make([]ProjectBindingCandidate, 0)
	for _, item := range vault.Items {
		for _, field := range item.Fields {
			if field.Sensitivity != vaultmodel.SensitivitySecret {
				continue
			}
			env, ok, err := resolver.EnvName(item, field.ID)
			if err != nil {
				return nil, err
			}
			if !ok {
				env = ""
			}
			candidates = append(candidates, ProjectBindingCandidate{Item: item.ID, Title: item.Title, Field: field.ID, Env: env})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Title != candidates[j].Title {
			return candidates[i].Title < candidates[j].Title
		}
		if candidates[i].Item != candidates[j].Item {
			return candidates[i].Item < candidates[j].Item
		}
		return candidates[i].Field < candidates[j].Field
	})
	return candidates, nil
}

func (a *App) InitializeProject(force bool) (string, error) {
	return ProjectInit(force)
}

func (a *App) RemoveProjectBinding(key string) (string, error) {
	return ProjectRm(key)
}
