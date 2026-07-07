# Spec: v0.2 Template Item Vault Model

## Purpose

Define the v0.2 core vault contract before implementation. This spec replaces the product-level `group:key` mental model with template-backed items and fields while preserving scriptable developer-secret workflows.

## Model

### Vault

```go
type Vault struct {
    Version   int        `json:"version"`
    Items     []Item     `json:"items"`
    Templates []Template `json:"templates,omitempty"`
}
```

Rules:

- `version` for the v2 format is `2`.
- `items` is an ordered array to preserve user-visible ordering.
- Runtime code builds indexes after load; the persisted model is not keyed by path.
- `templates` is optional and reserved for future value-free custom templates. v0.2 ships built-ins in code.

### Item

```go
type Item struct {
    ID        string  `json:"id"`
    Title     string  `json:"title"`
    Template  string  `json:"template"`
    Fields    []Field `json:"fields"`
    Labels    []string `json:"labels,omitempty"`
    Aliases   []string `json:"aliases,omitempty"`
    Notes     string   `json:"notes,omitempty"`
    CreatedAt string   `json:"created_at,omitempty"`
    UpdatedAt string   `json:"updated_at,omitempty"`
}
```

Rules:

- `id` is the canonical stable identity and must be unique inside a vault.
- `title` is the primary human label and must be non-empty.
- `template` must resolve to a built-in or future custom template.
- `fields` are ordered and must satisfy the selected template.
- `aliases` are optional value-free selectors used for migration from v1 paths. They are not project-manifest lookup contracts unless a later phase explicitly accepts that.
- `labels` replace v1 tags.
- `notes` receives v1 descriptions during migration.

### Field

```go
type Field struct {
    ID          string          `json:"id"`
    Label       string          `json:"label"`
    Type        FieldType       `json:"type"`
    Sensitivity Sensitivity     `json:"sensitivity"`
    Purpose     string          `json:"purpose,omitempty"`
    Value       json.RawMessage `json:"value,omitempty"`
    Bindings    []Binding       `json:"bindings,omitempty"`
}
```

Field types:

- `string`
- `concealed`
- `url`
- `date`
- `string_list`
- `json`

Sensitivity:

- `public`: may appear in list/status/info/search metadata output.
- `secret`: must not appear unless a command explicitly materializes secrets, such as `get`, `copy`, export, or project run.

Rules:

- Field IDs must be unique within an item.
- Secret-bearing fields use `sensitivity: secret` and usually `type: concealed`.
- Public fields must not carry env bindings in v0.2. Env bindings resolve values and therefore belong only on secret fields.
- Required template fields must exist before save.
- Unknown fields may be rejected for built-in templates in v0.2; future custom templates can relax this only with explicit validation rules.

### Binding

```go
type Binding struct {
    Type string `json:"type"`
    Name string `json:"name"`
}
```

Rules:

- v0.2 binding type is `env`.
- `name` must be a valid env var name for `env` bindings.
- Duplicate env binding names across secret fields are invalid unless a later phase defines explicit conflict resolution.
- Bindings attach to fields, never to whole items.

### Template

```go
type Template struct {
    ID          string          `json:"id"`
    Name        string          `json:"name"`
    Description string          `json:"description,omitempty"`
    Fields      []FieldTemplate `json:"fields"`
}

type FieldTemplate struct {
    ID          string      `json:"id"`
    Label       string      `json:"label"`
    Type        FieldType   `json:"type"`
    Sensitivity Sensitivity `json:"sensitivity"`
    Purpose     string      `json:"purpose,omitempty"`
    Required    bool        `json:"required,omitempty"`
    Repeatable  bool        `json:"repeatable,omitempty"`
}
```

Rules:

- Built-in templates are compiled into Go for v0.2.
- Custom templates are deferred until the model proves useful.
- A template may define at most one default `primary_secret` field unless later CLI selection rules support multiple defaults.

## Built-In Templates

### `generic_secret`

Fields:

| ID | Type | Sensitivity | Purpose | Required |
| --- | --- | --- | --- | --- |
| `value` | `concealed` | `secret` | `primary_secret` | yes |

Use cases:

- One-off app token.
- Local credential where extra metadata is not worth modeling.
- Migration target for all v1 flat secret records unless the user explicitly chooses a richer template.

### `api_credential`

Fields:

| ID | Type | Sensitivity | Purpose | Required |
| --- | --- | --- | --- | --- |
| `credential` | `concealed` | `secret` | `primary_secret` | yes |
| `provider` | `string` | `public` | provider/source | no |
| `website` | `url` | `public` | `website` | no |
| `expires_at` | `date` | `public` | `expiry` | no |
| `scopes` | `string_list` | `public` | scopes | no |

Use cases:

- API keys and bearer tokens where provider, website, expiry, and scopes help review without revealing the credential.

## Example JSON

```json
{
  "version": 2,
  "items": [
    {
      "id": "item_01JEXAMPLE0000000000000001",
      "title": "OpenAI personal API key",
      "template": "api_credential",
      "fields": [
        {
          "id": "credential",
          "label": "Credential",
          "type": "concealed",
          "sensitivity": "secret",
          "purpose": "primary_secret",
          "bindings": [{ "type": "env", "name": "OPENAI_API_KEY" }],
          "value": "example-redacted"
        },
        {
          "id": "provider",
          "label": "Provider",
          "type": "string",
          "sensitivity": "public",
          "value": "openai"
        },
        {
          "id": "website",
          "label": "Website",
          "type": "url",
          "sensitivity": "public",
          "value": "https://platform.openai.com/api-keys"
        },
        {
          "id": "expires_at",
          "label": "Expires",
          "type": "date",
          "sensitivity": "public",
          "value": "2026-12-31"
        },
        {
          "id": "scopes",
          "label": "Scopes",
          "type": "string_list",
          "sensitivity": "public",
          "value": ["responses", "embeddings"]
        }
      ],
      "labels": ["ai", "personal"]
    }
  ]
}
```

The example value is intentionally fake. Planning artifacts must not store real secret values.

## Validation Rules

Vault-level:

- Version must be supported.
- Item IDs must be non-empty and unique.
- Titles must be non-empty.
- Template IDs must be known.
- Env binding names must be unique across secret fields.

Item-level:

- Required template fields must exist.
- Field IDs must be unique inside the item.
- Field type and sensitivity must match the built-in template for built-in fields.
- Labels and aliases are value-free strings and must not be used to store secret values.

Field-level:

- Secret fields may carry env bindings.
- Public fields must not carry env bindings in v0.2.
- `date` fields use `YYYY-MM-DD`.
- `url` fields must parse as URLs with a scheme.
- `string_list` values must be arrays of strings.

## Value Exposure Rules

Allowed value-free surfaces:

- `secret list --long`
- `secret info`
- `project status`
- manager list/search responses
- project manifests
- logs and diagnostics

Secret-materializing surfaces:

- `secret get`
- `secret copy`
- `secret export`
- `project export`
- `project run`

Rules:

- Secret fields are never printed by value-free surfaces.
- Public fields may be printed by value-free surfaces.
- Diagnostics name item title, item ID, field ID, env name, labels, and template; they do not print secret field values.

## Selector Rules

Initial selector precedence:

1. Exact item ID.
2. Exact title.
3. Exact env binding when the user passes `--env NAME`.
4. Query match only if exactly one result.
5. Ambiguous matches fail with value-free candidates and guidance.

Aliases from v1 paths are migration aids. They may be accepted by CLI selectors during the migration window, but project manifests should store canonical item ID + field ID.

## Project Manifest v2

Candidate shape:

```json
{
  "version": 2,
  "bindings": [
    {
      "item": "item_01JEXAMPLE0000000000000001",
      "field": "credential",
      "env": "OPENAI_API_KEY",
      "required": true,
      "title": "OpenAI personal API key"
    }
  ]
}
```

Rules:

- `item` and `field` are canonical lookup values.
- `title` is a value-free review snapshot only.
- Manifests never store field values.
- Export/run resolves only explicit field bindings.
- Missing item/field and duplicate env names produce value-free diagnostics.
