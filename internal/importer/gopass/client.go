package gopass

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

const defaultBinary = "gopass"

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

type Client struct {
	Binary string
	Runner CommandRunner
}

func NewClient(binary string) Client {
	if binary == "" {
		binary = defaultBinary
	}
	return Client{Binary: binary, Runner: ExecRunner{}}
}

func (c Client) ListFlat(prefix string) ([]string, error) {
	out, err := c.run("list", "--flat")
	if err != nil {
		return nil, err
	}
	paths := parseLines(out)
	if prefix == "" {
		return paths, nil
	}
	filtered := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == prefix || strings.HasPrefix(path, strings.TrimSuffix(prefix, "/")+"/") {
			filtered = append(filtered, path)
		}
	}
	return filtered, nil
}

func (c Client) ShowPassword(path string) (string, error) {
	out, err := c.run("show", "--password", path)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}

func (c Client) run(args ...string) ([]byte, error) {
	runner := c.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	binary := c.Binary
	if binary == "" {
		binary = defaultBinary
	}
	out, err := runner.Run(binary, args...)
	if err != nil {
		return nil, normalizeError(binary, err)
	}
	return out, nil
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
	sort.Strings(paths)
	return paths
}

func normalizeError(binary string, err error) error {
	var notFound *exec.Error
	if errors.As(err, &notFound) && notFound.Err == exec.ErrNotFound {
		return fmt.Errorf("gopass binary not found: %s", binary)
	}
	return err
}
