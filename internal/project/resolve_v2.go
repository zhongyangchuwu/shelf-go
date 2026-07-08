package project

import (
	"fmt"

	"github.com/zhongyangchuwu/shelf-go/internal/util"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func ResolveV2Bindings(manifest Manifest, vault vaultmodel.Vault) ([]Binding, []Diagnostic) {
	entries := make([]Binding, 0, len(manifest.Bindings))
	diagnostics := make([]Diagnostic, 0)
	envOwners := map[string]string{}
	items := itemsByID(vault.Items)

	for _, binding := range manifest.Bindings {
		path := binding.Key()
		item, ok := items[binding.Item]
		if !ok {
			appendMissingV2Diagnostic(&diagnostics, binding, path, "missing item")
			continue
		}
		field, ok := fieldByID(item, binding.Field)
		if !ok {
			appendMissingV2Diagnostic(&diagnostics, binding, path, "missing field")
			continue
		}
		if field.Sensitivity != vaultmodel.SensitivitySecret {
			diagnostics = append(diagnostics, Diagnostic{Status: "fail", Path: path, Message: "field is not secret"})
			continue
		}
		if len(field.Value) == 0 {
			appendMissingV2Diagnostic(&diagnostics, binding, path, "missing value")
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
		entries = append(entries, Binding{Path: path, EnvName: binding.Env, Value: value})
	}
	return entries, diagnostics
}

func itemsByID(items []vaultmodel.Item) map[string]vaultmodel.Item {
	out := make(map[string]vaultmodel.Item, len(items))
	for _, item := range items {
		out[item.ID] = item
	}
	return out
}

func fieldByID(item vaultmodel.Item, fieldID string) (vaultmodel.Field, bool) {
	for _, field := range item.Fields {
		if field.ID == fieldID {
			return field, true
		}
	}
	return vaultmodel.Field{}, false
}

func appendMissingV2Diagnostic(diagnostics *[]Diagnostic, binding V2Binding, path, message string) {
	if binding.IsRequired() {
		*diagnostics = append(*diagnostics, Diagnostic{Status: "fail", Path: path, Message: message + " required"})
		return
	}
	*diagnostics = append(*diagnostics, Diagnostic{Status: "warn", Path: path, Message: message + " optional"})
}
