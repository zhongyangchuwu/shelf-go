package project

import (
	"fmt"

	"github.com/zhongyangchuwu/shelf-go/internal/util"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func ResolveBindings(manifest Manifest, resolver vaultmodel.Resolver) ([]ResolvedBinding, []Diagnostic) {
	entries := make([]ResolvedBinding, 0, len(manifest.Bindings))
	diagnostics := make([]Diagnostic, 0)
	envOwners := map[string]string{}

	for _, binding := range manifest.Bindings {
		path := binding.Key()
		item, ok := resolver.ItemByID(binding.Item)
		if !ok {
			appendMissingDiagnostic(&diagnostics, binding, path, "missing item")
			continue
		}
		_, field, err := resolver.Field(item, binding.Field)
		if err != nil {
			appendMissingDiagnostic(&diagnostics, binding, path, "missing field")
			continue
		}
		if field.Sensitivity != vaultmodel.SensitivitySecret {
			diagnostics = append(diagnostics, Diagnostic{Status: "fail", Path: path, Message: "field is not secret"})
			continue
		}
		if len(field.Value) == 0 {
			appendMissingDiagnostic(&diagnostics, binding, path, "missing value")
			continue
		}
		value, err := util.ValueString(field.Value)
		if err != nil {
			diagnostics = append(diagnostics, Diagnostic{Status: "fail", Path: path, Message: err.Error()})
			continue
		}
		if owner, exists := envOwners[binding.Env]; exists {
			diagnostics = append(diagnostics, Diagnostic{Status: "fail", Path: path, Message: fmt.Sprintf("env name %s conflicts with %s", binding.Env, owner)})
			continue
		}
		envOwners[binding.Env] = path
		entries = append(entries, ResolvedBinding{Path: path, EnvName: binding.Env, Value: value})
	}
	return entries, diagnostics
}

func appendMissingDiagnostic(diagnostics *[]Diagnostic, binding Binding, path, message string) {
	if binding.IsRequired() {
		*diagnostics = append(*diagnostics, Diagnostic{Status: "fail", Path: path, Message: message + " required"})
		return
	}
	*diagnostics = append(*diagnostics, Diagnostic{Status: "warn", Path: path, Message: message + " optional"})
}
