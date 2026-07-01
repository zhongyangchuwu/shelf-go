package project

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

type EnvFileStatus struct {
	Name    string
	Vars    int
	Empty   int
	Bound   int
	Unbound int
	Error   string
}

func EnvFileStatuses(root string, boundEnvNames map[string]struct{}) []EnvFileStatus {
	entries, err := os.ReadDir(root)
	if err != nil {
		return []EnvFileStatus{{Name: ".", Error: err.Error()}}
	}
	statuses := make([]EnvFileStatus, 0)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !isEnvFileName(name) {
			continue
		}
		statuses = append(statuses, readEnvFileStatus(filepath.Join(root, name), name, boundEnvNames))
	}
	sort.Slice(statuses, func(i, j int) bool { return statuses[i].Name < statuses[j].Name })
	return statuses
}

func RenderEnvFileStatuses(w io.Writer, statuses []EnvFileStatus) {
	fmt.Fprintln(w, "env files:")
	if len(statuses) == 0 {
		fmt.Fprintln(w, "  none found")
		return
	}
	for _, status := range statuses {
		if status.Error != "" {
			fmt.Fprintf(w, "  %s warn %s\n", status.Name, status.Error)
			continue
		}
		fmt.Fprintf(w, "  %s vars %d, empty %d, bound %d, unbound %d\n", status.Name, status.Vars, status.Empty, status.Bound, status.Unbound)
	}
}

func isEnvFileName(name string) bool {
	return name == ".env" || strings.HasPrefix(name, ".env.")
}

func readEnvFileStatus(path, name string, boundEnvNames map[string]struct{}) EnvFileStatus {
	file, err := os.Open(path)
	if err != nil {
		return EnvFileStatus{Name: name, Error: err.Error()}
	}
	defer file.Close()

	status := EnvFileStatus{Name: name}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, hasValue, ok := parseEnvLine(scanner.Text())
		if !ok {
			continue
		}
		status.Vars++
		if !hasValue || isEmptyEnvValue(value) {
			status.Empty++
		}
		if _, bound := boundEnvNames[key]; bound {
			status.Bound++
		} else {
			status.Unbound++
		}
	}
	if err := scanner.Err(); err != nil {
		status.Error = err.Error()
	}
	return status
}

func parseEnvLine(line string) (string, string, bool, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false, false
	}
	line = strings.TrimPrefix(line, "export ")
	key, value, hasValue := strings.Cut(line, "=")
	key = strings.TrimSpace(key)
	if key == "" || !vault.IsEnvName(key) {
		return "", "", false, false
	}
	if !hasValue {
		return key, "", false, true
	}
	return key, strings.TrimSpace(value), true, true
}

func isEmptyEnvValue(value string) bool {
	value = strings.TrimSpace(value)
	return value == "" || value == `""` || value == "''"
}
