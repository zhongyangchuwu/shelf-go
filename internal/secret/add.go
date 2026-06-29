package secret

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/adapters/shelfvault"
)

type AddRequest struct {
	Args         []string
	In           io.Reader
	Out          io.Writer
	ReadPassword func() ([]byte, error)
}

func Add(st *shelfvault.Store, req AddRequest) (string, error) {
	prompt := newAddPrompt(req.In, req.Out, st, req.ReadPassword)
	path, secret, force, err := prompt.collect(req.Args)
	if err != nil {
		return "", err
	}
	if err := st.Set(path, secret, force); err != nil {
		return "", err
	}
	return path, nil
}

type addPrompt struct {
	in           *bufio.Reader
	out          io.Writer
	st           *shelfvault.Store
	readPassword func() ([]byte, error)
}

func newAddPrompt(in io.Reader, out io.Writer, st *shelfvault.Store, readPassword func() ([]byte, error)) addPrompt {
	return addPrompt{in: bufio.NewReader(in), out: out, st: st, readPassword: readPassword}
}

func (p addPrompt) collect(args []string) (string, shelfvault.Secret, bool, error) {
	p.printGroupHints()
	path, err := p.collectPath(args)
	if err != nil {
		return "", shelfvault.Secret{}, false, err
	}
	force := false
	if _, exists := p.st.Get(path); exists {
		overwrite, err := p.confirm("secret exists; overwrite? [y/N]: ")
		if err != nil {
			return "", shelfvault.Secret{}, false, err
		}
		if !overwrite {
			return "", shelfvault.Secret{}, false, fmt.Errorf("secret already exists: %s", path)
		}
		force = true
	}
	value, err := p.password("value: ")
	if err != nil {
		return "", shelfvault.Secret{}, false, err
	}
	if value == "" {
		return "", shelfvault.Secret{}, false, fmt.Errorf("secret value is required")
	}
	envName, err := p.line("env (optional): ")
	if err != nil {
		return "", shelfvault.Secret{}, false, err
	}
	description, err := p.line("description (optional): ")
	if err != nil {
		return "", shelfvault.Secret{}, false, err
	}
	tagText, err := p.line("tags comma-separated (optional): ")
	if err != nil {
		return "", shelfvault.Secret{}, false, err
	}
	raw, err := shelfvault.ParseValue(value)
	if err != nil {
		return "", shelfvault.Secret{}, false, err
	}
	secret := shelfvault.Secret{Value: raw, Env: strings.TrimSpace(envName), Description: strings.TrimSpace(description), Tags: parsePromptTags(tagText)}
	return path, secret, force, nil
}

func (p addPrompt) collectPath(args []string) (string, error) {
	if len(args) == 0 {
		path, err := p.line("path (group/key as group:path): ")
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(path), nil
	}
	input := strings.TrimSpace(args[0])
	if strings.Contains(input, ":") {
		return input, nil
	}
	key, err := p.line("key: ")
	if err != nil {
		return "", err
	}
	return input + ":" + strings.TrimSpace(key), nil
}

func (p addPrompt) printGroupHints() {
	groups := existingGroups(p.st.List(""))
	if len(groups) == 0 {
		return
	}
	fmt.Fprintln(p.out, "existing groups:")
	limit := len(groups)
	if limit > 8 {
		limit = 8
	}
	for _, group := range groups[:limit] {
		fmt.Fprintf(p.out, "  %s\n", group)
	}
	if len(groups) > limit {
		fmt.Fprintf(p.out, "  ... %d more\n", len(groups)-limit)
	}
}

func (p addPrompt) line(label string) (string, error) {
	fmt.Fprint(p.out, label)
	text, err := p.in.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimRight(text, "\r\n"), nil
}

func (p addPrompt) password(label string) (string, error) {
	fmt.Fprint(p.out, label)
	bytes, err := p.readPassword()
	fmt.Fprintln(p.out)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (p addPrompt) confirm(label string) (bool, error) {
	answer, err := p.line(label)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func parsePromptTags(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

func existingGroups(paths []string) []string {
	seen := map[string]struct{}{}
	groups := make([]string, 0, len(paths))
	for _, path := range paths {
		group, _, ok := strings.Cut(path, ":")
		if !ok || group == "" {
			continue
		}
		if _, exists := seen[group]; exists {
			continue
		}
		seen[group] = struct{}{}
		groups = append(groups, group)
	}
	sort.Strings(groups)
	return groups
}
