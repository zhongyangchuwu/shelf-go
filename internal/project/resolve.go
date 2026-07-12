package project

import (
	"fmt"

	"github.com/zhongyangchuwu/shelf-go/internal/util"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func ResolveBindingStatuses(manifest Manifest, resolver vaultmodel.Resolver) []BindingStatus {
	statuses := make([]BindingStatus, 0, len(manifest.Bindings))
	envOwners := map[string]string{}
	for _, binding := range manifest.Bindings {
		path := binding.Key()
		status := BindingStatus{Binding: binding, Path: path, Status: "ready"}
		item, ok := resolver.ItemByID(binding.Item)
		if !ok {
			status.Status, status.Message = missingBindingStatus(binding, "missing item")
			statuses = append(statuses, status)
			continue
		}
		_, field, err := resolver.Field(item, binding.Field)
		if err != nil {
			status.Status, status.Message = missingBindingStatus(binding, "missing field")
			statuses = append(statuses, status)
			continue
		}
		if field.Sensitivity != vaultmodel.SensitivitySecret {
			status.Status, status.Message = "fail", "field is not secret"
			statuses = append(statuses, status)
			continue
		}
		if len(field.Value) == 0 {
			status.Status, status.Message = missingBindingStatus(binding, "missing value")
			statuses = append(statuses, status)
			continue
		}
		if owner, exists := envOwners[binding.Env]; exists {
			status.Status, status.Message = "fail", fmt.Sprintf("env name %s conflicts with %s", binding.Env, owner)
			statuses = append(statuses, status)
			continue
		}
		envOwners[binding.Env] = path
		statuses = append(statuses, status)
	}
	return statuses
}

func ResolveBindings(manifest Manifest, resolver vaultmodel.Resolver) ([]ResolvedBinding, []Diagnostic) {
	statuses := ResolveBindingStatuses(manifest, resolver)
	entries := make([]ResolvedBinding, 0, len(statuses))
	diagnostics := make([]Diagnostic, 0)
	for _, status := range statuses {
		if status.Status != "ready" {
			diagnostics = append(diagnostics, Diagnostic{Status: status.Status, Path: status.Path, Message: status.Message})
			continue
		}
		item, _ := resolver.ItemByID(status.Binding.Item)
		_, field, err := resolver.Field(item, status.Binding.Field)
		if err != nil {
			diagnostics = append(diagnostics, Diagnostic{Status: "fail", Path: status.Path, Message: "missing field"})
			continue
		}
		value, err := util.ValueString(field.Value)
		if err != nil {
			diagnostics = append(diagnostics, Diagnostic{Status: "fail", Path: status.Path, Message: err.Error()})
			continue
		}
		entries = append(entries, ResolvedBinding{Path: status.Path, EnvName: status.Binding.Env, Value: value})
	}
	return entries, diagnostics
}

func missingBindingStatus(binding Binding, message string) (string, string) {
	if binding.IsRequired() {
		return "fail", message + " required"
	}
	return "warn", message + " optional"
}
