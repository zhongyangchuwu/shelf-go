package project

import (
	"fmt"
	"strings"

	"errors"

	"github.com/zhongyangchuwu/shelf-go/internal/source"
)

type AddEntryRequest struct {
	Selector string
	Env      string
	Optional bool
	Tags     []string
}

func BuildEntry(reader source.Reader, req AddEntryRequest) (Entry, error) {
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
		matches, err := reader.ListByTags("", req.Tags)
		if err != nil {
			return Entry{}, err
		}
		if len(matches) == 0 {
			return Entry{}, fmt.Errorf("no secrets match tags: %s", strings.Join(req.Tags, ","))
		}
		entry.Tags = req.Tags
	} else {
		isPrefix := !strings.Contains(req.Selector, ":")
		if isPrefix && req.Env != "" {
			return Entry{}, fmt.Errorf("--env is only valid for path entries")
		}
		if isPrefix {
			matches, err := reader.List(req.Selector)
			if err != nil {
				return Entry{}, err
			}
			if len(matches) == 0 {
				return Entry{}, fmt.Errorf("no secrets match prefix: %s", req.Selector)
			}
			entry.Prefix = req.Selector
		} else {
			if _, err := reader.Get(req.Selector); errors.Is(err, source.ErrNotFound) {
				return Entry{}, fmt.Errorf("secret not found: %s", req.Selector)
			} else if err != nil {
				return Entry{}, err
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

func AddEntry(m Manifest, reader source.Reader, req AddEntryRequest) (Manifest, Entry, error) {
	entry, err := BuildEntry(reader, req)
	if err != nil {
		return m, Entry{}, err
	}
	if err := m.AddEntry(entry); err != nil {
		return m, Entry{}, err
	}
	return m, entry, nil
}
