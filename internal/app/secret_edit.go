package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

// ErrSecretEditInvalidDocument identifies an invalid whole-item edit. Diagnostics
// deliberately omit document contents, including public field values.
var ErrSecretEditInvalidDocument = errors.New("invalid secret edit document")

// ErrSecretEditConflict means the selected item or its schema changed since the
// edit began. It does not imply transactional locking of other workflows.
var ErrSecretEditConflict = errors.New("secret edit conflict; item or schema changed")

type SecretEditRequest struct {
	Selector string
	Env      string
	Target   string
}

// SecretEditSession exposes only the plaintext draft. The baseline is private and
// independent of Document so invalid edits and failed saves can be retried.
type SecretEditSession struct {
	Document []byte
	path     string
	original vaultmodel.Item
	template vaultmodel.Template
}

// BeginSecretEdit resolves an existing item before exposing its editable values.
// The resolved absolute vault path is pinned for the lifetime of the session.
func (a *App) BeginSecretEdit(configPathFlag, vaultPathFlag string, req SecretEditRequest) (*SecretEditSession, error) {
	if (req.Selector == "") == (req.Env == "") || (req.Target != "" && req.Env == "") {
		return nil, errors.New("secret edit requires either an item selector or an environment mapping")
	}
	path, err := vaultPath(configPathFlag, vaultPathFlag)
	if err != nil {
		return nil, errors.New("resolve vault for secret edit failed")
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, errors.New("resolve vault for secret edit failed")
	}
	vault, err := vaultjson.NewRepository(path).Load()
	if err != nil {
		return nil, errors.New("load vault for secret edit failed")
	}
	var item vaultmodel.Item
	if req.Env == "" {
		_, item, err = vaultmodel.NewResolverFromValidatedVault(vault).ResolveItem(req.Selector)
		if err != nil {
			return nil, errors.New("secret edit selector must uniquely identify an existing item")
		}
	} else {
		if _, err := newEnvironmentSecret(req.Env); err != nil {
			return nil, errors.New("invalid environment mapping for secret edit")
		}
		targets, err := secretSetTargets(vault, req.Env)
		if err != nil {
			return nil, errors.New("resolve environment mapping for secret edit failed")
		}
		target := req.Target
		if target == "" {
			if len(targets) != 1 {
				return nil, errors.New("secret edit requires one existing environment mapping; choose --target ITEM.FIELD when ambiguous")
			}
			target = targets[0]
		}
		_, item, _, err = secretSetItem(vault, req.Env, target)
		if err != nil {
			return nil, errors.New("secret edit target must identify an existing item-ID.field mapping")
		}
	}
	template, ok := vaultmodel.NewRegistryWithDefinitions(vault.Templates, vault.Profiles).Template(item.Template)
	if !ok {
		return nil, errors.New("resolve schema for secret edit failed")
	}
	fields := make(map[string]json.RawMessage, len(item.Fields))
	for _, field := range item.Fields {
		if len(field.Value) != 0 {
			fields[field.ID] = field.Value
		}
	}
	labels := item.Labels
	if labels == nil {
		labels = []string{}
	}
	document, err := json.MarshalIndent(struct {
		ID          string                     `json:"id"`
		Template    string                     `json:"template"`
		Title       string                     `json:"title"`
		Description string                     `json:"description"`
		Labels      []string                   `json:"labels"`
		Fields      map[string]json.RawMessage `json:"fields"`
		ClearFields []string                   `json:"clear_fields"`
	}{item.ID, item.Template, item.Title, item.Description, labels, fields, []string{}}, "", "  ")
	if err != nil {
		return nil, errors.New("prepare secret edit document failed")
	}
	template.Aliases = slices.Clone(template.Aliases)
	template.Fields = slices.Clone(template.Fields)
	return &SecretEditSession{Document: append(document, '\n'), path: path, original: cloneSecretEditItem(item), template: template}, nil
}

// SaveSecretEdit validates a private candidate, then merges it into a fresh vault
// only if the selected item and owning template still match the baseline. Failed
// saves leave the baseline untouched. The check is optimistic, not a file lock.
func (a *App) SaveSecretEdit(session *SecretEditSession, document []byte) (bool, error) {
	if session == nil || session.path == "" || session.original.ID == "" {
		return false, fmt.Errorf("%w: missing edit session", ErrSecretEditInvalidDocument)
	}
	item, err := secretEditCandidate(session, document)
	if err != nil {
		return false, err
	}
	if equalSecretEditItems(item, session.original) {
		return false, nil
	}
	repo := vaultjson.NewRepository(session.path)
	vault, err := repo.Load()
	if errors.Is(err, os.ErrNotExist) {
		return false, ErrSecretEditConflict
	}
	if err != nil {
		return false, errors.New("reload vault before saving secret edit failed")
	}
	template, ok := vaultmodel.NewRegistryWithDefinitions(vault.Templates, vault.Profiles).Template(session.original.Template)
	if !ok || !reflect.DeepEqual(template, session.template) {
		return false, ErrSecretEditConflict
	}
	index := -1
	for i, current := range vault.Items {
		if current.ID == session.original.ID {
			if !equalSecretEditItems(current, session.original) {
				return false, ErrSecretEditConflict
			}
			index = i
			break
		}
	}
	if index < 0 {
		return false, ErrSecretEditConflict
	}
	if item.Title != session.original.Title {
		if err := ensureSelectorAvailable(vault.Items, index, item.Title); err != nil {
			return false, fmt.Errorf("%w: title collides with an existing item", ErrSecretEditInvalidDocument)
		}
	}
	vault.Items[index] = item
	if err := vaultmodel.Validate(vault); err != nil {
		return false, fmt.Errorf("%w: item does not satisfy the vault schema", ErrSecretEditInvalidDocument)
	}
	if err := repo.Save(vault); err != nil {
		return false, errors.New("save secret edit failed")
	}
	return true, nil
}

func secretEditCandidate(session *SecretEditSession, document []byte) (vaultmodel.Item, error) {
	invalid := ErrSecretEditInvalidDocument
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	if err := checkSecretEditJSON(decoder); err != nil {
		return vaultmodel.Item{}, invalid
	}
	if _, err := decoder.Token(); err != io.EOF {
		return vaultmodel.Item{}, invalid
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(document, &members); err != nil || members == nil {
		return vaultmodel.Item{}, invalid
	}
	item := cloneSecretEditItem(session.original)
	var fields map[string]json.RawMessage
	var clearFields []string
	for key, raw := range members {
		switch key {
		case "id", "template", "title", "description":
			var text string
			if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &text) != nil {
				return vaultmodel.Item{}, invalid
			}
			switch key {
			case "id":
				if text != item.ID {
					return vaultmodel.Item{}, invalid
				}
			case "template":
				if text != item.Template {
					return vaultmodel.Item{}, invalid
				}
			case "title":
				if text != session.original.Title {
					item.Title = strings.TrimSpace(text)
				}
			case "description":
				if text != session.original.Description {
					item.Description = strings.TrimSpace(text)
				}
			}
		case "labels":
			var entries []json.RawMessage
			if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &entries) != nil {
				return vaultmodel.Item{}, invalid
			}
			labels := make([]string, len(entries))
			for i, entry := range entries {
				if bytes.Equal(bytes.TrimSpace(entry), []byte("null")) || json.Unmarshal(entry, &labels[i]) != nil {
					return vaultmodel.Item{}, invalid
				}
			}
			if !slices.Equal(labels, session.original.Labels) {
				var err error
				item.Labels, err = normalizedTokens(labels, "label")
				if err != nil {
					return vaultmodel.Item{}, invalid
				}
			}
		case "fields":
			if json.Unmarshal(raw, &fields) != nil || fields == nil {
				return vaultmodel.Item{}, invalid
			}
		case "clear_fields":
			if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &clearFields) != nil {
				return vaultmodel.Item{}, invalid
			}
		default:
			return vaultmodel.Item{}, invalid
		}
	}
	definitions := make(map[string]vaultmodel.FieldTemplate, len(session.template.Fields))
	for _, definition := range session.template.Fields {
		definitions[definition.ID] = definition
	}
	for id := range fields {
		if _, ok := definitions[id]; !ok {
			return vaultmodel.Item{}, invalid
		}
	}
	cleared := make(map[string]bool, len(clearFields))
	for _, id := range clearFields {
		definition, ok := definitions[id]
		if !ok || definition.Sensitivity != vaultmodel.SensitivityPublic || cleared[id] {
			return vaultmodel.Item{}, invalid
		}
		if _, set := fields[id]; set {
			return vaultmodel.Item{}, invalid
		}
		cleared[id] = true
	}
	for i := range item.Fields {
		id := item.Fields[i].ID
		if value, set := fields[id]; set {
			item.Fields[i].Value = value
			delete(fields, id)
		} else if cleared[id] {
			item.Fields[i].Value = nil
		}
	}
	// New optional fields follow canonical schema order, not map iteration order.
	for _, definition := range session.template.Fields {
		if value, set := fields[definition.ID]; set {
			item.Fields = append(item.Fields, vaultmodel.Field{ID: definition.ID, Value: value})
		}
	}
	if err := vaultmodel.Validate(vaultmodel.Vault{Version: vaultmodel.CurrentVersion, Templates: []vaultmodel.Template{session.template}, Items: []vaultmodel.Item{item}}); err != nil {
		return vaultmodel.Item{}, invalid
	}
	return item, nil
}

// checkSecretEditJSON rejects duplicate keys at every depth, including JSON-valued
// fields. Decoder errors are consumed internally and never include edited data.
func checkSecretEditJSON(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, compound := token.(json.Delim)
	if !compound {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]bool)
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return ErrSecretEditInvalidDocument
			}
			seen[name] = true
			if err := checkSecretEditJSON(decoder); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := checkSecretEditJSON(decoder); err != nil {
				return err
			}
		}
	default:
		return ErrSecretEditInvalidDocument
	}
	_, err = decoder.Token()
	return err
}

func cloneSecretEditItem(item vaultmodel.Item) vaultmodel.Item {
	item.Labels = slices.Clone(item.Labels)
	item.Fields = slices.Clone(item.Fields)
	for i := range item.Fields {
		item.Fields[i].Value = slices.Clone(item.Fields[i].Value)
	}
	return item
}

func equalSecretEditItems(left, right vaultmodel.Item) bool {
	if left.ID != right.ID || left.Template != right.Template || left.Title != right.Title || left.Description != right.Description || left.CreatedAt != right.CreatedAt || left.UpdatedAt != right.UpdatedAt || !slices.Equal(left.Labels, right.Labels) || len(left.Fields) != len(right.Fields) {
		return false
	}
	for i, field := range left.Fields {
		other := right.Fields[i]
		if field.ID != other.ID {
			return false
		}
		if bytes.Equal(field.Value, other.Value) {
			continue
		}
		if len(field.Value) == 0 || len(other.Value) == 0 {
			return false
		}
		var values [2]any
		for j, raw := range []json.RawMessage{field.Value, other.Value} {
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			if decoder.Decode(&values[j]) != nil {
				return false
			}
		}
		if !equalSecretEditJSON(values[0], values[1]) {
			return false
		}
	}
	return true
}

func equalSecretEditJSON(left, right any) bool {
	switch value := left.(type) {
	case json.Number:
		other, ok := right.(json.Number)
		if !ok {
			return false
		}
		if value == other {
			return true
		}
		leftCoefficient, leftExponent := normalizeSecretEditNumber(value)
		rightCoefficient, rightExponent := normalizeSecretEditNumber(other)
		return leftCoefficient == rightCoefficient && leftExponent.Cmp(&rightExponent) == 0
	case map[string]any:
		other, ok := right.(map[string]any)
		if !ok || len(value) != len(other) {
			return false
		}
		for key, entry := range value {
			otherEntry, exists := other[key]
			if !exists || !equalSecretEditJSON(entry, otherEntry) {
				return false
			}
		}
		return true
	case []any:
		other, ok := right.([]any)
		if !ok || len(value) != len(other) {
			return false
		}
		for i, entry := range value {
			if !equalSecretEditJSON(entry, other[i]) {
				return false
			}
		}
		return true
	default:
		return left == right
	}
}

// normalizeSecretEditNumber represents a valid JSON number as a signed integer
// coefficient without leading or trailing zeros, times ten to an exact exponent.
// Only exponent digits are stored in big.Int; no power of ten is expanded, so
// work and storage stay bounded by the input even for enormous exponents.
func normalizeSecretEditNumber(number json.Number) (coefficient string, exponent big.Int) {
	text := string(number)
	negative := strings.HasPrefix(text, "-")
	if negative {
		text = text[1:]
	}
	mantissa := text
	var exponentText string
	if index := strings.IndexAny(text, "eE"); index >= 0 {
		mantissa, exponentText = text[:index], text[index+1:]
	}
	coefficient = mantissa
	fractionDigits := 0
	if point := strings.IndexByte(mantissa, '.'); point >= 0 {
		fractionDigits = len(mantissa) - point - 1
		coefficient = mantissa[:point] + mantissa[point+1:]
	}
	coefficient = strings.TrimLeft(coefficient, "0")
	if coefficient == "" {
		return "0", exponent
	}
	trimmed := strings.TrimRight(coefficient, "0")
	trailingZeros := len(coefficient) - len(trimmed)
	coefficient = trimmed
	if negative {
		coefficient = "-" + coefficient
	}
	if exponentText != "" {
		// The decoder already validated JSON number syntax, including the exponent.
		exponent.SetString(exponentText, 10)
	}
	var adjustment big.Int
	adjustment.SetInt64(int64(trailingZeros) - int64(fractionDigits))
	exponent.Add(&exponent, &adjustment)
	return coefficient, exponent
}
