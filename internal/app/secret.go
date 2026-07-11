package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/util"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

var itemIDPattern = regexp.MustCompile(`[^a-z0-9_]+`)

type CreateSecretRequest struct {
	Title    string
	ID       string
	Template string
	Profile  string
	Env      string
	Stdin    bool
	In       io.Reader
	Provider string
	Website  string
	Expires  string
	Scopes   []string
}

type SetSecretFieldRequest struct {
	Selector string
	Field    string
	Stdin    bool
	In       io.Reader
}

type GetSecretFieldRequest struct {
	Selector string
	Env      string
	Field    string
}

type ListSecretItemsRequest struct {
	Query string
	Long  bool
}

type ExportSecretItemsRequest struct {
	Labels []string
	Format string
}

type SecretItemSummary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Template string `json:"template"`
	Env      string `json:"env,omitempty"`
	ValueSet bool   `json:"value_set"`
}

type SecretFieldInfo struct {
	ID          string                 `json:"id"`
	Type        vaultmodel.FieldType   `json:"type"`
	Sensitivity vaultmodel.Sensitivity `json:"sensitivity"`
	Value       string                 `json:"value,omitempty"`
	ValueSet    bool                   `json:"value_set"`
}

type SecretInfo struct {
	ID       string            `json:"id"`
	Title    string            `json:"title"`
	Template string            `json:"template"`
	Labels   []string          `json:"labels,omitempty"`
	Aliases  []string          `json:"aliases,omitempty"`
	Fields   []SecretFieldInfo `json:"fields"`
}

func (a *App) CreateSecret(configPathFlag, vaultPathFlag string, req CreateSecretRequest) (string, error) {
	if strings.TrimSpace(req.Title) == "" {
		return "", fmt.Errorf("title is required")
	}
	vault, repo, err := a.loadVaultForWrite(configPathFlag, vaultPathFlag)
	if err != nil {
		return "", err
	}
	registry := vaultmodel.NewRegistryWithDefinitions(vault.Templates, vault.Profiles)
	item, err := buildSecretItem(req, registry)
	if err != nil {
		return "", err
	}
	if _, exists := findItemByID(vault.Items, item.ID); exists {
		return "", fmt.Errorf("item already exists: %s", item.ID)
	}
	for _, existing := range vault.Items {
		if existing.Title == item.Title {
			return "", fmt.Errorf("item title already exists: %s", item.Title)
		}
	}
	vault.Items = append(vault.Items, item)
	if err := repo.Save(vault); err != nil {
		return "", err
	}
	return item.ID, nil
}

func (a *App) SetSecretField(configPathFlag, vaultPathFlag string, req SetSecretFieldRequest) error {
	if !req.Stdin {
		return fmt.Errorf("secret set requires --stdin")
	}
	vault, repo, err := a.loadVaultForWrite(configPathFlag, vaultPathFlag)
	if err != nil {
		return err
	}
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	itemIndex, item, err := resolver.ResolveItem(req.Selector)
	if err != nil {
		return err
	}
	fieldIndex, field, err := resolver.DefaultField(item, req.Field)
	if err != nil {
		return err
	}
	value, err := readFieldValue(req.In, field.Type)
	if err != nil {
		return err
	}
	item.Fields[fieldIndex].Value = value
	vault.Items[itemIndex] = item
	if err := repo.Save(vault); err != nil {
		return err
	}
	return nil
}

func (a *App) GetSecretField(configPathFlag, vaultPathFlag string, req GetSecretFieldRequest) (string, error) {
	vault, err := a.loadVaultForRead(configPathFlag, vaultPathFlag)
	if err != nil {
		return "", err
	}
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	var item vaultmodel.Item
	var field vaultmodel.Field
	if req.Env != "" {
		ref, err := resolver.ResolveEnv(req.Env)
		if err != nil {
			return "", err
		}
		item = ref.Item
		field = ref.Field
		if req.Field != "" {
			_, field, err = resolver.Field(item, req.Field)
			if err != nil {
				return "", err
			}
		}
	} else {
		_, item, err = resolver.ResolveItem(req.Selector)
		if err != nil {
			return "", err
		}
		_, field, err = resolver.DefaultField(item, req.Field)
		if err != nil {
			return "", err
		}
	}
	return valueForOutput(field)
}

func (a *App) ListSecretItems(configPathFlag, vaultPathFlag string, req ListSecretItemsRequest) ([]SecretItemSummary, error) {
	vault, err := a.loadVaultForRead(configPathFlag, vaultPathFlag)
	if err != nil {
		return nil, err
	}
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	out := make([]SecretItemSummary, 0, len(vault.Items))
	for _, item := range vault.Items {
		if req.Query != "" && !matchesItemQuery(item, req.Query) {
			continue
		}
		primary, err := resolver.PrimarySecretField(item)
		if err != nil {
			return nil, fmt.Errorf("list %s: %w", item.Title, err)
		}
		env, _, err := resolver.EnvName(item, primary.ID)
		if err != nil {
			return nil, fmt.Errorf("list %s env name: %w", item.Title, err)
		}
		_, field, err := resolver.Field(item, primary.ID)
		if err != nil {
			return nil, fmt.Errorf("list %s: %w", item.Title, err)
		}
		out = append(out, SecretItemSummary{ID: item.ID, Title: item.Title, Template: item.Template, Env: env, ValueSet: len(field.Value) > 0})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Title < out[j].Title })
	return out, nil
}

func (a *App) ExportSecretItems(configPathFlag, vaultPathFlag string, req ExportSecretItemsRequest) (string, error) {
	labels := normalizeLabels(req.Labels)
	if len(labels) == 0 {
		return "", fmt.Errorf("secret export requires --label")
	}
	vault, err := a.loadVaultForRead(configPathFlag, vaultPathFlag)
	if err != nil {
		return "", err
	}
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	bindings, err := exportBindingsForLabels(vault, resolver, labels)
	if err != nil {
		return "", err
	}
	switch req.Format {
	case "", "shell":
		return util.ShellBindings(bindings)
	case "json":
		return util.JSONBindings(bindings)
	default:
		return "", fmt.Errorf("unsupported export format: %s", req.Format)
	}
}

func (a *App) SecretInfo(configPathFlag, vaultPathFlag, selector string) (SecretInfo, error) {
	vault, err := a.loadVaultForRead(configPathFlag, vaultPathFlag)
	if err != nil {
		return SecretInfo{}, err
	}
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	_, item, err := resolver.ResolveItem(selector)
	if err != nil {
		return SecretInfo{}, err
	}
	info := SecretInfo{ID: item.ID, Title: item.Title, Template: item.Template, Labels: append([]string(nil), item.Labels...), Aliases: append([]string(nil), item.Aliases...)}
	for _, field := range item.Fields {
		fieldInfo := SecretFieldInfo{ID: field.ID, Type: field.Type, Sensitivity: field.Sensitivity, ValueSet: len(field.Value) > 0}
		if field.Sensitivity == vaultmodel.SensitivityPublic && len(field.Value) > 0 {
			fieldInfo.Value, _ = valueForOutput(field)
		}
		info.Fields = append(info.Fields, fieldInfo)
	}
	return info, nil
}

func (a *App) RemoveSecret(configPathFlag, vaultPathFlag, selector string) error {
	vault, repo, err := a.loadVaultForWrite(configPathFlag, vaultPathFlag)
	if err != nil {
		return err
	}
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	idx, _, err := resolver.ResolveItem(selector)
	if err != nil {
		return err
	}
	vault.Items = append(vault.Items[:idx], vault.Items[idx+1:]...)
	return repo.Save(vault)
}

func (a *App) loadVaultForRead(configPathFlag, vaultPathFlag string) (vaultmodel.Vault, error) {
	path, err := vaultPath(configPathFlag, vaultPathFlag)
	if err != nil {
		return vaultmodel.Vault{}, err
	}
	repo := vaultjson.NewRepository(path)
	vault, err := repo.Load()
	if errors.Is(err, os.ErrNotExist) {
		return vaultmodel.Vault{}, fmt.Errorf("item vault not found: %s", path)
	}
	return vault, err
}

func (a *App) loadVaultForWrite(configPathFlag, vaultPathFlag string) (vaultmodel.Vault, vaultjson.Repository, error) {
	path, err := vaultPath(configPathFlag, vaultPathFlag)
	if err != nil {
		return vaultmodel.Vault{}, vaultjson.Repository{}, err
	}
	repo := vaultjson.NewRepository(path)
	vault, err := repo.Load()
	if errors.Is(err, os.ErrNotExist) {
		return vaultmodel.Vault{Version: vaultmodel.CurrentVersion}, repo, nil
	}
	return vault, repo, err
}

func vaultPath(configPathFlag, vaultPathFlag string) (string, error) {
	runtime, err := config.Resolve(configPathFlag, vaultPathFlag)
	if err != nil {
		return "", err
	}
	return runtime.VaultPath, nil
}

func buildSecretItem(req CreateSecretRequest, registry vaultmodel.Registry) (vaultmodel.Item, error) {
	templateID := req.Template
	profileID := req.Profile
	defaults := map[string]json.RawMessage{}
	if profileID != "" {
		profile, ok := registry.Profile(profileID)
		if !ok {
			return vaultmodel.Item{}, fmt.Errorf("profile not found: %s", profileID)
		}
		templateID = profile.Template
		for id, value := range profile.Defaults {
			defaults[id] = append([]byte(nil), value...)
		}
	}
	if templateID == "" {
		templateID = vaultmodel.TemplateGenericSecret
	}
	tmpl, ok := registry.Template(templateID)
	if !ok {
		return vaultmodel.Item{}, fmt.Errorf("template not found: %s", templateID)
	}
	item := vaultmodel.Item{ID: req.ID, Title: req.Title, Template: tmpl.ID}
	if item.ID == "" {
		item.ID = itemIDFromTitle(req.Title)
	}
	for _, templateField := range tmpl.Fields {
		field := vaultmodel.Field{ID: templateField.ID, Label: templateField.Label, Type: templateField.Type, Sensitivity: templateField.Sensitivity, Purpose: templateField.Purpose}
		if value, ok := defaults[templateField.ID]; ok {
			field.Value = append([]byte(nil), value...)
		}
		item.Fields = append(item.Fields, field)
	}
	if req.Env != "" {
		primary, err := primaryFieldTemplateForTemplate(tmpl)
		if err != nil {
			return vaultmodel.Item{}, err
		}
		if err := setEnvNameField(&item, tmpl, primary.ID, rawString(req.Env)); err != nil {
			return vaultmodel.Item{}, err
		}
	}
	if req.Provider != "" {
		if err := setFieldValueIfPresent(&item, "provider", rawString(req.Provider)); err != nil {
			return vaultmodel.Item{}, fmt.Errorf("--provider is not supported by template %s: %w", tmpl.ID, err)
		}
	}
	if req.Website != "" {
		if err := setFieldValueIfPresent(&item, "website", rawString(req.Website)); err != nil {
			return vaultmodel.Item{}, fmt.Errorf("--website is not supported by template %s: %w", tmpl.ID, err)
		}
	}
	if req.Expires != "" {
		if err := setFieldValueIfPresent(&item, "expires_at", rawString(req.Expires)); err != nil {
			return vaultmodel.Item{}, fmt.Errorf("--expires is not supported by template %s: %w", tmpl.ID, err)
		}
	}
	if len(req.Scopes) > 0 {
		value, err := json.Marshal(req.Scopes)
		if err != nil {
			return vaultmodel.Item{}, err
		}
		if err := setFieldValueIfPresent(&item, "scopes", value); err != nil {
			return vaultmodel.Item{}, fmt.Errorf("--scope is not supported by template %s: %w", tmpl.ID, err)
		}
	}
	if req.Stdin {
		primary, err := primaryFieldTemplateForTemplate(tmpl)
		if err != nil {
			return vaultmodel.Item{}, err
		}
		idx, field, err := findField(item, primary.ID)
		if err != nil {
			return vaultmodel.Item{}, err
		}
		value, err := readFieldValue(req.In, field.Type)
		if err != nil {
			return vaultmodel.Item{}, err
		}
		item.Fields[idx].Value = value
	}
	if err := vaultmodel.Validate(vaultmodel.Vault{Version: vaultmodel.CurrentVersion, Items: []vaultmodel.Item{item}}); err != nil {
		return vaultmodel.Item{}, err
	}
	return item, nil
}

func rawString(value string) json.RawMessage {
	content, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return content
}

func itemIDFromTitle(title string) string {
	id := strings.ToLower(strings.TrimSpace(title))
	id = itemIDPattern.ReplaceAllString(id, "_")
	id = strings.Trim(id, "_")
	if id == "" {
		id = "secret"
	}
	return "item_" + id
}

func readFieldValue(r io.Reader, fieldType vaultmodel.FieldType) (json.RawMessage, error) {
	if r == nil {
		r = strings.NewReader("")
	}
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	text := stripFinalNewline(string(content))
	switch fieldType {
	case vaultmodel.FieldTypeString, vaultmodel.FieldTypeConcealed, vaultmodel.FieldTypeURL, vaultmodel.FieldTypeDate:
		return json.Marshal(text)
	case vaultmodel.FieldTypeStringList:
		lines := strings.Split(text, "\n")
		values := make([]string, 0, len(lines))
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				values = append(values, line)
			}
		}
		return json.Marshal(values)
	case vaultmodel.FieldTypeJSON:
		if !json.Valid([]byte(text)) {
			return nil, fmt.Errorf("stdin value must be valid json")
		}
		return json.RawMessage(text), nil
	default:
		return nil, fmt.Errorf("unsupported field type: %s", fieldType)
	}
}

func stripFinalNewline(value string) string {
	if strings.HasSuffix(value, "\r\n") {
		return strings.TrimSuffix(value, "\r\n")
	}
	return strings.TrimSuffix(value, "\n")
}

func valueForOutput(field vaultmodel.Field) (string, error) {
	if len(field.Value) == 0 {
		return "", nil
	}
	switch field.Type {
	case vaultmodel.FieldTypeString, vaultmodel.FieldTypeConcealed, vaultmodel.FieldTypeURL, vaultmodel.FieldTypeDate:
		return stringValue(field.Value)
	case vaultmodel.FieldTypeStringList, vaultmodel.FieldTypeJSON:
		return string(field.Value), nil
	default:
		return "", fmt.Errorf("unsupported field type: %s", field.Type)
	}
}

func stringValue(value json.RawMessage) (string, error) {
	var out string
	if err := json.Unmarshal(value, &out); err != nil {
		return "", err
	}
	return out, nil
}

func primaryFieldTemplateForTemplate(tmpl vaultmodel.Template) (vaultmodel.FieldTemplate, error) {
	for _, field := range tmpl.Fields {
		if field.Purpose == vaultmodel.PurposePrimarySecret {
			return field, nil
		}
	}
	return vaultmodel.FieldTemplate{}, fmt.Errorf("template has no primary secret field: %s", tmpl.ID)
}

func setEnvNameField(item *vaultmodel.Item, tmpl vaultmodel.Template, fieldID string, value json.RawMessage) error {
	for _, templateField := range tmpl.Fields {
		if templateField.Purpose != vaultmodel.PurposeEnvName || templateField.EnvFor != fieldID {
			continue
		}
		return setFieldValueIfPresent(item, templateField.ID, value)
	}
	return fmt.Errorf("template field has no env-name field: %s", fieldID)
}

func setFieldValueIfPresent(item *vaultmodel.Item, fieldID string, value json.RawMessage) error {
	for i := range item.Fields {
		if item.Fields[i].ID == fieldID {
			item.Fields[i].Value = value
			return nil
		}
	}
	return fmt.Errorf("field not found: %s", fieldID)
}

func findField(item vaultmodel.Item, fieldID string) (int, vaultmodel.Field, error) {
	for i, field := range item.Fields {
		if field.ID == fieldID {
			return i, field, nil
		}
	}
	return 0, vaultmodel.Field{}, fmt.Errorf("field not found: %s", fieldID)
}

func findItemByID(items []vaultmodel.Item, id string) (vaultmodel.Item, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return vaultmodel.Item{}, false
}

func exportBindingsForLabels(vault vaultmodel.Vault, resolver vaultmodel.Resolver, labels []string) ([]util.Binding, error) {
	bindings := make([]util.Binding, 0)
	seenEnv := map[string]string{}
	matched := 0
	for _, item := range vault.Items {
		if !itemHasLabels(item, labels) {
			continue
		}
		matched++
		primary, err := resolver.PrimarySecretField(item)
		if err != nil {
			return nil, err
		}
		envName, ok, err := resolver.EnvName(item, primary.ID)
		if err != nil {
			return nil, fmt.Errorf("read env name for %s: %w", item.Title, err)
		}
		if !ok {
			continue
		}
		_, field, err := resolver.Field(item, primary.ID)
		if err != nil {
			return nil, fmt.Errorf("missing secret field for %s: %s", item.Title, primary.ID)
		}
		if len(field.Value) == 0 {
			return nil, fmt.Errorf("missing value for %s", item.Title)
		}
		if previous, exists := seenEnv[envName]; exists {
			return nil, fmt.Errorf("duplicate env name %s: %s, %s", envName, previous, item.Title)
		}
		value, err := valueForOutput(field)
		if err != nil {
			return nil, err
		}
		seenEnv[envName] = item.Title
		bindings = append(bindings, util.Binding{EnvName: envName, Value: value})
	}
	if matched == 0 {
		return nil, fmt.Errorf("no items matched labels: %s", strings.Join(labels, ","))
	}
	sort.Slice(bindings, func(i, j int) bool { return bindings[i].EnvName < bindings[j].EnvName })
	return bindings, nil
}

func normalizeLabels(labels []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(labels))
	for _, label := range labels {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		if _, exists := seen[label]; exists {
			continue
		}
		seen[label] = struct{}{}
		out = append(out, label)
	}
	return out
}

func itemHasLabels(item vaultmodel.Item, labels []string) bool {
	for _, label := range labels {
		if !hasString(item.Labels, label) {
			return false
		}
	}
	return true
}

func matchesItemQuery(item vaultmodel.Item, query string) bool {
	if strings.Contains(item.ID, query) || strings.Contains(item.Title, query) || strings.Contains(item.Template, query) {
		return true
	}
	for _, alias := range item.Aliases {
		if strings.Contains(alias, query) {
			return true
		}
	}
	return false
}

func hasString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
