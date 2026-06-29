package gopass

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/source"
)

const defaultBinary = "gopass"

var ErrTagsUnsupported = errors.New("gopass source does not support tag selectors")

type CommandRunner interface {
	Run(name string, args ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return nil, fmt.Errorf("gopass %s: %s", strings.Join(args, " "), message)
	}
	return stdout.Bytes(), nil
}

type Reader struct {
	Binary string
	Runner CommandRunner
}

func NewReader(binary string) Reader {
	if binary == "" {
		binary = defaultBinary
	}
	return Reader{Binary: binary, Runner: ExecRunner{}}
}

func (r Reader) Get(path string) (source.Secret, error) {
	gopassPath, err := toGopassPath(path)
	if err != nil {
		return source.Secret{}, err
	}
	out, err := r.run("show", "--password", gopassPath)
	if err != nil {
		if isNotFound(err) {
			return source.Secret{}, source.ErrNotFound
		}
		return source.Secret{}, err
	}
	value := strings.TrimRight(string(out), "\r\n")
	return source.Secret{Path: path, Value: value}, nil
}

func (r Reader) List(prefix string) ([]string, error) {
	out, err := r.run("list", "--flat")
	if err != nil {
		return nil, err
	}
	paths := fromGopassPaths(parseLines(out))
	if prefix == "" {
		return paths, nil
	}
	filtered := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == prefix || strings.HasPrefix(path, prefix+"/") || strings.HasPrefix(path, prefix+":") {
			filtered = append(filtered, path)
		}
	}
	return filtered, nil
}

func (r Reader) ListByTags(prefix string, tags []string) ([]string, error) {
	return nil, ErrTagsUnsupported
}

func (r Reader) run(args ...string) ([]byte, error) {
	runner := r.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	binary := r.Binary
	if binary == "" {
		binary = defaultBinary
	}
	out, err := runner.Run(binary, args...)
	if err != nil {
		return nil, normalizeError(binary, err)
	}
	return out, nil
}

func toGopassPath(path string) (string, error) {
	id, err := source.ParseSecretID(path)
	if err != nil {
		return "", err
	}
	return id.GroupPath + "/" + id.Key, nil
}

func fromGopassPaths(paths []string) []string {
	mapped := make([]string, 0, len(paths))
	for _, path := range paths {
		idx := strings.LastIndex(path, "/")
		if idx <= 0 || idx == len(path)-1 {
			continue
		}
		mappedPath := path[:idx] + ":" + path[idx+1:]
		if source.ValidatePath(mappedPath) == nil {
			mapped = append(mapped, mappedPath)
		}
	}
	sort.Strings(mapped)
	return mapped
}

func parseLines(out []byte) []string {
	lines := strings.Split(string(out), "\n")
	paths := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}
	return paths
}

func normalizeError(binary string, err error) error {
	var notFound *exec.Error
	if errors.As(err, &notFound) && notFound.Err == exec.ErrNotFound {
		return fmt.Errorf("gopass binary not found: %s", binary)
	}
	return err
}

func isNotFound(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "not found") || strings.Contains(message, "not in the password store") || strings.Contains(message, "entry is not in the password store")
}
