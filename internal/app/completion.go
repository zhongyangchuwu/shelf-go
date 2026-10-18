package app

import (
	"fmt"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

type SecretCompletionTarget uint8

const (
	CompleteSecretSelectors SecretCompletionTarget = iota
	CompleteSecretQueries
	CompleteSecretLabels
	CompleteSecretEnvNames
	CompleteSecretFields
	CompleteSecretTemplates
	CompleteSecretProfiles
	CompleteSecretSetEnvNames
	CompleteSecretSetTargets
)

type SecretCompletionRequest struct {
	Target          SecretCompletionTarget
	Selector        string
	Env             string
	Template        string
	Profile         string
	ProjectableOnly bool
}

// SecretCompletions reads one vault snapshot and returns names and public
// metadata only. It never initializes storage or materializes secret values.
func (a *App) SecretCompletions(configPath, vaultPath string, req SecretCompletionRequest) ([]string, error) {
	vault, err := a.loadVaultForRead(configPath, vaultPath)
	if err != nil && req.Target != CompleteSecretTemplates && req.Target != CompleteSecretProfiles {
		return nil, err
	}
	if err != nil {
		vault = vaultmodel.Vault{}
	}
	var candidates []string
	switch req.Target {
	case CompleteSecretTemplates:
		registry := vaultmodel.NewRegistryWithDefinitions(vault.Templates, vault.Profiles)
		compatible := ""
		if req.Profile != "" {
			profile, ok := registry.Profile(req.Profile)
			if !ok {
				return nil, nil
			}
			compatible = profile.Template
		}
		for _, definition := range append(vaultmodel.BuiltInTemplates(), vault.Templates...) {
			template, _ := registry.Template(definition.ID)
			if compatible != "" && template.ID != compatible {
				continue
			}
			candidates = append(candidates, template.ID)
			candidates = append(candidates, template.Aliases...)
		}
	case CompleteSecretProfiles:
		registry := vaultmodel.NewRegistryWithDefinitions(vault.Templates, vault.Profiles)
		compatible := ""
		if req.Template != "" {
			template, ok := registry.Template(req.Template)
			if !ok {
				return nil, nil
			}
			compatible = template.ID
		}
		for _, definition := range append(vaultmodel.BuiltInProfiles(), vault.Profiles...) {
			profile, _ := registry.Profile(definition.ID)
			if compatible != "" && profile.Template != compatible {
				continue
			}
			candidates = append(candidates, profile.ID)
			candidates = append(candidates, profile.Aliases...)
		}
	case CompleteSecretSelectors:
		counts := make(map[string]int, len(vault.Items)*2)
		for _, item := range vault.Items {
			counts[item.ID]++
			if item.Title != item.ID {
				counts[item.Title]++
			}
		}
		for _, item := range vault.Items {
			if counts[item.ID] == 1 {
				candidates = append(candidates, item.ID)
			}
			if item.Title != item.ID && counts[item.Title] == 1 {
				candidates = append(candidates, item.Title)
			}
		}
	case CompleteSecretQueries:
		for _, item := range vault.Items {
			candidates = append(candidates, item.ID, item.Title, item.Template, item.Description)
			candidates = append(candidates, item.Labels...)
		}
	case CompleteSecretLabels:
		if req.Selector != "" {
			resolver := vaultmodel.NewResolverFromValidatedVault(vault)
			_, item, err := resolver.ResolveItem(req.Selector)
			if err != nil {
				return nil, err
			}
			candidates = append(candidates, item.Labels...)
		} else {
			for _, item := range vault.Items {
				candidates = append(candidates, item.Labels...)
			}
		}
	case CompleteSecretSetTargets:
		if _, err := newEnvironmentSecret(req.Env); err != nil {
			return nil, err
		}
		return secretSetTargets(vault, req.Env)
	case CompleteSecretEnvNames, CompleteSecretSetEnvNames:
		resolver := vaultmodel.NewResolverFromValidatedVault(vault)
		counts := make(map[string]int)
		for _, item := range vault.Items {
			for _, field := range item.Fields {
				if !resolver.IsProjectableField(item, field.ID) {
					continue
				}
				name, mapped, err := resolver.EnvName(item, field.ID)
				if err != nil {
					return nil, err
				}
				if mapped {
					candidates = append(candidates, name)
					counts[name]++
				}
			}
		}
		unique := candidates[:0]
		for _, name := range candidates {
			if counts[name] == 1 || req.Target == CompleteSecretSetEnvNames {
				unique = append(unique, name)
			}
		}
		candidates = unique
	case CompleteSecretFields:
		resolver := vaultmodel.NewResolverFromValidatedVault(vault)
		var item vaultmodel.Item
		if req.Env != "" {
			ref, err := resolver.ResolveEnv(req.Env)
			if err != nil {
				return nil, err
			}
			item = ref.Item
		} else {
			_, resolved, err := resolver.ResolveItem(req.Selector)
			if err != nil {
				return nil, err
			}
			item = resolved
		}
		for _, field := range item.Fields {
			if req.ProjectableOnly && !resolver.IsProjectableField(item, field.ID) {
				continue
			}
			candidates = append(candidates, field.ID)
		}
	default:
		return nil, fmt.Errorf("unknown secret completion target: %d", req.Target)
	}
	return candidates, nil
}
