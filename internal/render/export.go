package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/store"
)

var nonEnvChar = regexp.MustCompile(`[^A-Za-z0-9]+`)

func EnvName(path string, secret store.Secret) string {
	if secret.Env != "" {
		return secret.Env
	}
	name := nonEnvChar.ReplaceAllString(path, "_")
	name = strings.Trim(name, "_")
	return strings.ToUpper(name)
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

func Env(paths []string, secrets map[string]store.Secret) (string, error) {
	var b strings.Builder
	for _, path := range paths {
		secret := secrets[path]
		value, err := ValueString(secret.Value)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "%s=%s\n", EnvName(path, secret), value)
	}
	return b.String(), nil
}

func Shell(paths []string, secrets map[string]store.Secret) (string, error) {
	var b strings.Builder
	for _, path := range paths {
		secret := secrets[path]
		value, err := ValueString(secret.Value)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "export %s=%s\n", EnvName(path, secret), ShellQuote(value))
	}
	return b.String(), nil
}

func JSON(paths []string, secrets map[string]store.Secret) (string, error) {
	payload := map[string]json.RawMessage{}
	for _, path := range paths {
		payload[path] = secrets[path].Value
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
