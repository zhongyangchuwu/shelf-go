package project

import (
	"fmt"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

type AddEntryRequest struct {
	Selector string
	Env      string
	Optional bool
	Tags     []string
}

func BuildEntry(st *vault.Store, req AddEntryRequest) (Entry, error) {
	isTag := len(req.Tags) > 0
	if isTag && req.Selector != "" {
		return Entry{}, fmt.Errorf("path-or-prefix must not be set with --tag")
	}
	if !isTag && req.Selector == "" {
		return Entry{}, fmt.Errorf("path-or-prefix or --tag is required")
	}
	if isTag && req.Env != "" {
		return Entry{}, fmt.Errorf("--env is only valid for path entries")
	}

	entry := Entry{}
	if isTag {
		if len(st.ListByTags("", req.Tags)) == 0 {
			return Entry{}, fmt.Errorf("no secrets match tags: %s", strings.Join(req.Tags, ","))
		}
		entry.Tags = req.Tags
	} else {
		isPrefix := !strings.Contains(req.Selector, ":")
		if isPrefix && req.Env != "" {
			return Entry{}, fmt.Errorf("--env is only valid for path entries")
		}
		if isPrefix {
			matches := st.List(req.Selector)
			if len(matches) == 0 {
				return Entry{}, fmt.Errorf("no secrets match prefix: %s", req.Selector)
			}
			entry.Prefix = req.Selector
		} else {
			if _, ok := st.Get(req.Selector); !ok {
				return Entry{}, fmt.Errorf("secret not found: %s", req.Selector)
			}
			entry.Path = req.Selector
			if req.Env != "" {
				entry.Env = req.Env
			}
		}
	}
	if req.Optional {
		entry.Required = new(bool)
	}
	return entry, nil
}

func AddEntry(m Manifest, st *vault.Store, req AddEntryRequest) (Manifest, Entry, error) {
	entry, err := BuildEntry(st, req)
	if err != nil {
		return m, Entry{}, err
	}
	if err := m.AddEntry(entry); err != nil {
		return m, Entry{}, err
	}
	return m, entry, nil
}
