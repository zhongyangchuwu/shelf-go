package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/util"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

var itemIDPattern = regexp.MustCompile(`[^a-z0-9_]+`)

type CreateSecretRequest struct {
	Title       string
	ID          string
	Template    string
	Profile     string
	Env         string
	Stdin       bool
	Labels      []string
	Description string
	In          io.Reader
	Provider    string
	Website     string
	Validity    string
	Scopes      []string
	BaseURL     string
	BaseURLEnv  string
	ClientID    string
	ClientIDEnv string
}

type FieldUpdate struct {
	Field         string
	Values        []string
	FieldValueSet bool
	ClearField    bool
	In            io.Reader
	ReadField     bool
}

type UpdateSecretRequest struct {
	Selector       string
	Title          string
	TitleSet       bool
	Description    string
	DescriptionSet bool
	AddLabels      []string
	RemoveLabels   []string
	FieldUpdates   []FieldUpdate
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

type ExportEnvironmentRequest struct {
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

type SecretCatalogItem struct {
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	Template          string   `json:"template"`
	Env               string   `json:"env,omitempty"`
	EnvNames          []string `json:"env_names,omitempty"`
	Labels            []string `json:"labels,omitempty"`
	Description       string   `json:"description,omitempty"`
	ValueSet          bool     `json:"value_set"`
	ProjectableFields int      `json:"projectable_fields"`
	SetFields         int      `json:"set_fields"`
	EnvMappings       int      `json:"env_mappings"`
}

type SecretFieldInfo struct {
	ID          string                 `json:"id"`
	Type        vaultmodel.FieldType   `json:"type"`
	Sensitivity vaultmodel.Sensitivity `json:"sensitivity"`
	Value       string                 `json:"value,omitempty"`
	ValueSet    bool                   `json:"value_set"`
}

type SecretInfo struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Template    string            `json:"template"`
	Labels      []string          `json:"labels,omitempty"`
	Description string            `json:"description,omitempty"`
	PrimaryEnv  string            `json:"-"`
	Fields      []SecretFieldInfo `json:"fields"`
}

// SecretCreateField describes a value-free field that the TUI can collect.
type SecretCreateField struct {
	ID          string
	Label       string
	Type        string
	Sensitivity string
	Required    bool
}

// SecretCreatePreset supplies safe public defaults for one create option.
type SecretCreatePreset struct {
	ID       string
	Label    string
	Defaults map[string]string
}

// SecretCreateOption describes one TUI-visible item kind and its presets.
type SecretCreateOption struct {
	ID       string
	Label    string
	Fields   []SecretCreateField
	Defaults map[string]string
	Presets  []SecretCreatePreset
}

// SecretCreateOptions returns the built-in item kinds exposed by the TUI.
func (a *App) SecretCreateOptions() []SecretCreateOption {
	return DefaultSecretCreateOptions()
}

// DefaultSecretCreateOptions provides the value-free built-in create schema.
func DefaultSecretCreateOptions() []SecretCreateOption {
	registry := vaultmodel.NewDefaultRegistry()
	profiles := vaultmodel.BuiltInProfiles()
	options := make([]SecretCreateOption, 0, 2)
	for _, id := range []string{vaultmodel.TemplateGenericSecret, vaultmodel.TemplateAPICredential} {
		template, ok := registry.Template(id)
		if !ok {
			continue
		}
		label := template.Name
		if template.ID == vaultmodel.TemplateGenericSecret {
			label = "Single secret"
		}
		option := SecretCreateOption{ID: template.ID, Label: label, Fields: make([]SecretCreateField, 0, len(template.Fields)), Defaults: map[string]string{}}
		for _, field := range template.Fields {
			option.Fields = append(option.Fields, SecretCreateField{ID: field.ID, Label: field.Label, Type: string(field.Type), Sensitivity: string(field.Sensitivity), Required: field.Required})
			if _, validity := vaultmodel.IsValidityFieldID(field.ID); validity {
				option.Defaults[field.ID] = "unknown"
			}
		}
		for _, profile := range profiles {
			if profile.Template != template.ID {
				continue
			}
			preset := SecretCreatePreset{ID: profile.ID, Label: profile.Name, Defaults: make(map[string]string, len(profile.Defaults))}
			for fieldID, raw := range profile.Defaults {
				var value string
				if json.Unmarshal(raw, &value) == nil {
					preset.Defaults[fieldID] = value
				}
			}
			option.Presets = append(option.Presets, preset)
		}
		options = append(options, option)
	}
	return options
}

func (a *App) CreateSecret(configPathFlag, vaultPathFlag string, req CreateSecretRequest) (string, error) {
	if strings.TrimSpace(req.Title) == "" {
		return "", fmt.Errorf("title is required")
	}
	labels, err := normalizedTokens(req.Labels, "label")
	if err != nil {
		return "", err
	}
	req.Labels = labels
	req.Description = strings.TrimSpace(req.Description)
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

func (a *App) UpdateSecret(configPathFlag, vaultPathFlag string, req UpdateSecretRequest) error {
	if err := validateSecretUpdate(req); err != nil {
		return err
	}
	if !hasSecretUpdate(req) {
		return fmt.Errorf("secret update requires at least one change")
	}
	addLabels, err := normalizedTokens(req.AddLabels, "label")
	if err != nil {
		return err
	}
	removeLabels, err := normalizedTokens(req.RemoveLabels, "label")
	if err != nil {
		return err
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
	if req.TitleSet {
		title := strings.TrimSpace(req.Title)
		if title == "" {
			return fmt.Errorf("title is required")
		}
		if err := ensureSelectorAvailable(vault.Items, itemIndex, title); err != nil {
			return err
		}
		item.Title = title
	}
	if req.DescriptionSet {
		item.Description = strings.TrimSpace(req.Description)
	}
	item.Labels = removeTokens(addTokens(item.Labels, addLabels), removeLabels)
	if err := applyFieldUpdates(resolver, &item, req.FieldUpdates); err != nil {
		return err
	}
	vault.Items[itemIndex] = item
	return repo.Save(vault)
}

func applyFieldUpdates(resolver vaultmodel.Resolver, item *vaultmodel.Item, updates []FieldUpdate) error {
	seen := make(map[string]struct{}, len(updates))
	for _, update := range updates {
		fieldIndex, field, err := resolver.DefaultField(*item, update.Field)
		if err != nil {
			return err
		}
		if _, exists := seen[field.ID]; exists {
			return fmt.Errorf("secret update contains duplicate field: %s", field.ID)
		}
		seen[field.ID] = struct{}{}
		if update.ClearField {
			if field.Sensitivity != vaultmodel.SensitivityPublic {
				return fmt.Errorf("only public fields can be cleared: %s", field.ID)
			}
			item.Fields[fieldIndex].Value = nil
			continue
		}
		if update.ReadField {
			value, err := readFieldValue(update.In, field.Type)
			if err != nil {
				return err
			}
			item.Fields[fieldIndex].Value = value
			continue
		}
		if field.Sensitivity != vaultmodel.SensitivityPublic {
			return fmt.Errorf("secret field requires stdin input: %s", field.ID)
		}
		value, err := publicFieldValue(field.Type, update.Values)
		if err != nil {
			return err
		}
		item.Fields[fieldIndex].Value = value
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
	var field vaultmodel.ResolvedField
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

func (a *App) SecretCatalog(configPathFlag, vaultPathFlag string) ([]SecretCatalogItem, error) {
	vault, err := a.loadVaultForRead(configPathFlag, vaultPathFlag)
	if err != nil {
		return nil, err
	}
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	items := make([]SecretCatalogItem, 0, len(vault.Items))
	for _, item := range vault.Items {
		primary, err := resolver.PrimarySecretField(item)
		if err != nil {
			return nil, fmt.Errorf("catalog %s: %w", item.Title, err)
		}
		env, _, err := resolver.EnvName(item, primary.ID)
		if err != nil {
			return nil, fmt.Errorf("catalog %s env name: %w", item.Title, err)
		}
		_, field, err := resolver.Field(item, primary.ID)
		if err != nil {
			return nil, fmt.Errorf("catalog %s: %w", item.Title, err)
		}
		projectableFields, setFields, envMappings := 0, 0, 0
		var envNames []string
		for _, storedField := range item.Fields {
			if !resolver.IsProjectableField(item, storedField.ID) {
				continue
			}
			projectableFields++
			_, projected, err := resolver.Field(item, storedField.ID)
			if err != nil {
				return nil, fmt.Errorf("catalog %s field %s: %w", item.Title, storedField.ID, err)
			}
			if len(projected.Value) > 0 {
				setFields++
			}
			if name, mapped, err := resolver.EnvName(item, storedField.ID); err != nil {
				return nil, fmt.Errorf("catalog %s env name: %w", item.Title, err)
			} else if mapped {
				envMappings++
				if !hasString(envNames, name) {
					envNames = append(envNames, name)
				}
			}
		}
		sort.Strings(envNames)
		items = append(items, SecretCatalogItem{
			ID: item.ID, Title: item.Title, Template: item.Template, Env: env, EnvNames: envNames,
			Labels: append([]string(nil), item.Labels...), Description: item.Description, ValueSet: len(field.Value) > 0,
			ProjectableFields: projectableFields, SetFields: setFields, EnvMappings: envMappings,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Title < items[j].Title })
	return items, nil
}

func (a *App) ExportEnvironment(configPathFlag, vaultPathFlag string, req ExportEnvironmentRequest) (string, error) {
	labels := normalizeLabels(req.Labels)
	if len(labels) == 0 {
		return "", fmt.Errorf("env export requires --label")
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
		return "", fmt.Errorf("unsupported env export format: %s", req.Format)
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
	primary, err := resolver.PrimarySecretField(item)
	if err != nil {
		return SecretInfo{}, err
	}
	primaryEnv, _, err := resolver.EnvName(item, primary.ID)
	if err != nil {
		return SecretInfo{}, err
	}
	info := SecretInfo{ID: item.ID, Title: item.Title, Template: item.Template, Labels: append([]string(nil), item.Labels...), Description: item.Description, PrimaryEnv: primaryEnv}
	for _, storedField := range item.Fields {
		_, field, err := resolver.Field(item, storedField.ID)
		if err != nil {
			return SecretInfo{}, err
		}
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
	item := vaultmodel.Item{ID: req.ID, Title: req.Title, Template: tmpl.ID, Labels: append([]string(nil), req.Labels...), Description: req.Description}
	if item.ID == "" {
		item.ID = itemIDFromTitle(req.Title)
	}
	for _, templateField := range tmpl.Fields {
		field := vaultmodel.Field{ID: templateField.ID}
		if value, ok := defaults[templateField.ID]; ok {
			field.Value = append([]byte(nil), value...)
		} else if _, validity := vaultmodel.IsValidityFieldID(templateField.ID); validity {
			field.Value = rawString("unknown")
		}
		item.Fields = append(item.Fields, field)
	}
	if req.Env != "" {
		primary, err := registry.PrimarySecretField(tmpl.ID)
		if err != nil {
			return vaultmodel.Item{}, err
		}
		if _, ok, err := registry.FieldTemplate(tmpl.ID, vaultmodel.EnvFieldID(primary.ID)); err != nil {
			return vaultmodel.Item{}, err
		} else if !ok {
			return vaultmodel.Item{}, fmt.Errorf("template field has no env-name field: %s", primary.ID)
		}
		if err := setFieldValueIfPresent(&item, vaultmodel.EnvFieldID(primary.ID), rawString(req.Env)); err != nil {
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
	for _, update := range []struct {
		field string
		value string
		label string
	}{
		{field: "base_url", value: req.BaseURL, label: "base URL"},
		{field: vaultmodel.EnvFieldID("base_url"), value: req.BaseURLEnv, label: "base URL environment variable"},
		{field: "client_id", value: req.ClientID, label: "client ID"},
		{field: vaultmodel.EnvFieldID("client_id"), value: req.ClientIDEnv, label: "client ID environment variable"},
	} {
		if update.value == "" {
			continue
		}
		if err := setFieldValueIfPresent(&item, update.field, rawString(update.value)); err != nil {
			return vaultmodel.Item{}, fmt.Errorf("%s is not supported by template %s: %w", update.label, tmpl.ID, err)
		}
	}
	if req.Validity != "" {
		primary, err := registry.PrimarySecretField(tmpl.ID)
		if err != nil {
			return vaultmodel.Item{}, err
		}
		if err := setFieldValueIfPresent(&item, vaultmodel.ValidityFieldID(primary.ID), rawString(req.Validity)); err != nil {
			return vaultmodel.Item{}, fmt.Errorf("--validity is not supported by template %s: %w", tmpl.ID, err)
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
		primary, err := registry.PrimarySecretField(tmpl.ID)
		if err != nil {
			return vaultmodel.Item{}, err
		}
		value, err := readFieldValue(req.In, primary.Type)
		if err != nil {
			return vaultmodel.Item{}, err
		}
		if err := setFieldValueIfPresent(&item, primary.ID, value); err != nil {
			return vaultmodel.Item{}, err
		}
	}
	return item, nil
}

func rawString(value string) json.RawMessage {
	return json.RawMessage(strconv.AppendQuote(nil, value))
}

func hasSecretUpdate(req UpdateSecretRequest) bool {
	return req.TitleSet || req.DescriptionSet || len(req.AddLabels) > 0 || len(req.RemoveLabels) > 0 || len(req.FieldUpdates) > 0
}

func validateSecretUpdate(req UpdateSecretRequest) error {
	for index, update := range req.FieldUpdates {
		if err := validateFieldUpdate(update); err != nil {
			return fmt.Errorf("secret update field %d: %w", index+1, err)
		}
		if update.Field == "" && len(req.FieldUpdates) > 1 {
			return fmt.Errorf("secret update field %d requires a field id", index+1)
		}
	}
	return nil
}

func validateFieldUpdate(update FieldUpdate) error {
	if update.FieldValueSet && update.ReadField {
		return fmt.Errorf("cannot use --value with stdin input")
	}
	if update.FieldValueSet && update.ClearField {
		return fmt.Errorf("cannot use --value with --clear-field")
	}
	if update.ReadField && update.ClearField {
		return fmt.Errorf("cannot read stdin with --clear-field")
	}
	if update.ReadField && update.In == nil {
		return fmt.Errorf("stdin input is required")
	}
	if update.FieldValueSet || update.ClearField {
		if update.Field == "" {
			return fmt.Errorf("requires --field")
		}
		return nil
	}
	if !update.ReadField && (update.Field != "" || len(update.Values) > 0) {
		return fmt.Errorf("requires --value, --clear-field, or stdin input")
	}
	return nil
}

func normalizedTokens(values []string, kind string) ([]string, error) {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, fmt.Errorf("%s is required", kind)
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out, nil
}

func addTokens(existing, additions []string) []string {
	if len(additions) == 0 {
		return existing
	}
	seen := make(map[string]struct{}, len(existing)+len(additions))
	out := make([]string, 0, len(existing)+len(additions))
	for _, value := range existing {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	for _, value := range additions {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func removeTokens(existing, removals []string) []string {
	if len(removals) == 0 {
		return existing
	}
	remove := make(map[string]struct{}, len(removals))
	for _, value := range removals {
		remove[value] = struct{}{}
	}
	out := make([]string, 0, len(existing))
	for _, value := range existing {
		if _, exists := remove[value]; !exists {
			out = append(out, value)
		}
	}
	return out
}

func ensureSelectorAvailable(items []vaultmodel.Item, itemIndex int, selector string) error {
	for i, item := range items {
		if i == itemIndex {
			continue
		}
		if item.ID == selector || item.Title == selector {
			return fmt.Errorf("item selector already exists: %s", selector)
		}
	}
	return nil
}

func publicFieldValue(fieldType vaultmodel.FieldType, values []string) (json.RawMessage, error) {
	switch fieldType {
	case vaultmodel.FieldTypeString, vaultmodel.FieldTypeURL, vaultmodel.FieldTypeDate:
		if len(values) != 1 {
			return nil, fmt.Errorf("public field requires exactly one --value")
		}
		return json.Marshal(values[0])
	case vaultmodel.FieldTypeStringList:
		return json.Marshal(values)
	case vaultmodel.FieldTypeJSON:
		if len(values) != 1 {
			return nil, fmt.Errorf("json public field requires exactly one --value")
		}
		if !json.Valid([]byte(values[0])) {
			return nil, fmt.Errorf("public field value must be valid json")
		}
		return json.RawMessage(values[0]), nil
	default:
		return nil, fmt.Errorf("unsupported public field type: %s", fieldType)
	}
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

func valueForOutput(field vaultmodel.ResolvedField) (string, error) {
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

func setFieldValueIfPresent(item *vaultmodel.Item, fieldID string, value json.RawMessage) error {
	for i := range item.Fields {
		if item.Fields[i].ID == fieldID {
			item.Fields[i].Value = value
			return nil
		}
	}
	return fmt.Errorf("field not found: %s", fieldID)
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
		for _, storedField := range item.Fields {
			if !resolver.IsProjectableField(item, storedField.ID) {
				continue
			}
			envName, ok, err := resolver.EnvName(item, storedField.ID)
			if err != nil {
				return nil, fmt.Errorf("read env name for %s.%s: %w", item.Title, storedField.ID, err)
			}
			if !ok {
				continue
			}
			_, field, err := resolver.Field(item, storedField.ID)
			if err != nil {
				return nil, fmt.Errorf("missing mapped field for %s.%s", item.Title, storedField.ID)
			}
			if len(field.Value) == 0 {
				return nil, fmt.Errorf("missing value for %s.%s", item.Title, storedField.ID)
			}
			path := item.ID + "." + storedField.ID
			if previous, exists := seenEnv[envName]; exists {
				return nil, fmt.Errorf("duplicate env name %s: %s, %s", envName, previous, path)
			}
			value, err := valueForOutput(field)
			if err != nil {
				return nil, err
			}
			seenEnv[envName] = path
			bindings = append(bindings, util.Binding{EnvName: envName, Value: value})
		}
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
	return strings.Contains(item.ID, query) || strings.Contains(item.Title, query) || strings.Contains(item.Template, query) || strings.Contains(item.Description, query)
}

func hasString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
