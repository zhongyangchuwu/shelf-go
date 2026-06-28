package project

import "fmt"

func ChildEnv(parent []string, entries []Binding) []string {
	values := make(map[string]string, len(entries))
	for _, entry := range entries {
		values[entry.EnvName] = entry.Value
	}

	out := make([]string, 0, len(parent)+len(entries))
	seen := make(map[string]struct{}, len(parent)+len(entries))
	for _, item := range parent {
		key, _, ok := splitEnv(item)
		if !ok {
			if _, exists := values[item]; exists {
				continue
			}
			out = append(out, item)
			continue
		}
		if value, exists := values[key]; exists {
			out = append(out, key+"="+value)
			seen[key] = struct{}{}
			continue
		}
		out = append(out, item)
		seen[key] = struct{}{}
	}
	for _, entry := range entries {
		if _, exists := seen[entry.EnvName]; exists {
			continue
		}
		out = append(out, entry.EnvName+"="+entry.Value)
	}
	return out
}

func EnvOverrideWarnings(entries []Binding, parent []string) []string {
	parentNames := make(map[string]struct{}, len(parent))
	for _, item := range parent {
		key, _, ok := splitEnv(item)
		if ok {
			parentNames[key] = struct{}{}
		}
	}
	warnings := make([]string, 0)
	for _, entry := range entries {
		if _, exists := parentNames[entry.EnvName]; exists {
			warnings = append(warnings, fmt.Sprintf("warn %s overrides existing environment variable", entry.EnvName))
		}
	}
	return warnings
}

func splitEnv(item string) (string, string, bool) {
	for i := 0; i < len(item); i++ {
		if item[i] == '=' {
			return item[:i], item[i+1:], true
		}
	}
	return "", "", false
}
