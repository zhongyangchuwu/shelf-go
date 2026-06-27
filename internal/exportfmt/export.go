package exportfmt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

var nonEnvChar = regexp.MustCompile(`[^A-Za-z0-9]+`)

func EnvName(path string, secret vault.Secret) (string, error) {
	if secret.Env != "" {
		return secret.Env, nil
	}
	name := nonEnvChar.ReplaceAllString(path, "_")
	name = strings.Trim(name, "_")
	name = strings.ToUpper(name)
	if !vault.IsEnvName(name) {
		return "", fmt.Errorf("derived env name for %s is invalid: %s", path, name)
	}
	return name, nil
}

func ValueString(raw json.RawMessage) (string, error) {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return "", err
	}
	if compact.String() == "null" {
		return "", nil
	}
	return compact.String(), nil
}

func Env(paths []string, secrets map[string]vault.Secret) (string, error) {
	entries, err := Bindings(paths, secrets)
	if err != nil {
		return "", err
	}
	return EnvBindings(entries)
}

func Shell(paths []string, secrets map[string]vault.Secret) (string, error) {
	entries, err := Bindings(paths, secrets)
	if err != nil {
		return "", err
	}
	return ShellBindings(entries)
}

func JSON(paths []string, secrets map[string]vault.Secret) (string, error) {
	entries, err := Bindings(paths, secrets)
	if err != nil {
		return "", err
	}
	return JSONBindings(entries)
}

type Binding struct {
	EnvName string
	Value   string
}

func Bindings(paths []string, secrets map[string]vault.Secret) ([]Binding, error) {
	entries := make([]Binding, 0, len(paths))
	for _, path := range paths {
		secret := secrets[path]
		envName, err := EnvName(path, secret)
		if err != nil {
			return nil, err
		}
		value, err := ValueString(secret.Value)
		if err != nil {
			return nil, err
		}
		entries = append(entries, Binding{EnvName: envName, Value: value})
	}
	return entries, nil
}

func EnvBindings(entries []Binding) (string, error) {
	var b strings.Builder
	for _, entry := range entries {
		fmt.Fprintf(&b, "%s=%s\n", entry.EnvName, entry.Value)
	}
	return b.String(), nil
}

func ShellBindings(entries []Binding) (string, error) {
	var b strings.Builder
	for _, entry := range entries {
		fmt.Fprintf(&b, "export %s=%s\n", entry.EnvName, ShellQuote(entry.Value))
	}
	return b.String(), nil
}

func JSONBindings(entries []Binding) (string, error) {
	payload := map[string]string{}
	for _, entry := range entries {
		payload[entry.EnvName] = entry.Value
	}
	keys := make([]string, 0, len(payload))
	for key := range payload {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b bytes.Buffer
	b.WriteString("{\n")
	for i, key := range keys {
		value, err := json.Marshal(payload[key])
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "  %q: %s", key, string(value))
		if i < len(keys)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("}\n")
	return b.String(), nil
}

func ShellQuote(value string) string {
	if value == "" {
		return "''"
	}
	if isShellBare(value) {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func isShellBare(value string) bool {
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("_@%+=:,./-", r) {
			continue
		}
		return false
	}
	return true
}
