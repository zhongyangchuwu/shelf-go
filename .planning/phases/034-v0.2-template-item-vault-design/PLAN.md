# Plan: v0.2.0 Template Item Vault Design

## Objective

Produce the v0.2.0 design plan for replacing Shelf's flat path-keyed secret model with a template-based item/field core model and storage adapter boundary. The output should let a later implementation phase redesign the vault format, CLI contracts, project bindings, and migration path without relying on conversation memory.

## Scope

- Define the v0.2 core model: `Vault`, `Item`, `Field`, `Template`, field sensitivity, field purpose, and field bindings.
- Define built-in templates for at least `generic_secret` and `api_credential`.
- Define repository/codec/adapter boundaries so storage and encryption remain replaceable.
- Compare canonical storage candidates: JSON file, age-wrapped JSON, plaintext SQLite, age-wrapped SQLite, and encrypted SQLite/SQLCipher-like options.
- Define CLI command direction for common non-interactive workflows: create, set, get, copy, list, info.
- Define project manifest v2 direction for value-free item/field env bindings.
- Define migration strategy from v1 flat secrets and `.shelf.json` manifests.
- Define implementation phases and gates for a later v0.2.0 execution track.

## Non-Goals

- No production code changes in this design phase.
- No TUI or prompt redesign in this design phase.
- No browser extension, autofill, or direct website capture implementation.
- No hosted sync, team sharing, multi-user permissions, audit log service, or permanent daemon.
- No immediate decision to use SQLite unless the design evidence beats JSON on Shelf's portability and leakage constraints.
- No deletion of v1 migration support before an explicit migration/export path is designed.

## Inputs Read

- `.planning/PROJECT.md`
- `.planning/ROADMAP.md`
- `.planning/REQUIREMENTS.md`
- `.planning/STATE.md`
- `skill://omp-workflow`
- `skill://omp-planning-artifacts`
- `skill://omp-workflow/references/context-and-plan.md`
- `skill://omp-planning-artifacts/references/phase-artifacts.md`
- `skill://omp-planning-artifacts/references/state.md`
- Current conversation decisions about API credential templates, item/field models, adapters, and storage candidates.

## Proposed Core Model

### Vault

```go
type Vault struct {
    Version   int        `json:"version"`
    Items     []Item     `json:"items"`
    Templates []Template `json:"templates,omitempty"`
}
```

Persist items as an ordered array. Build indexes in memory by ID, title, env binding, label, and template.

### Item

```go
type Item struct {
    ID        string  `json:"id"`
    Title     string  `json:"title"`
    Template  string  `json:"template"`
    Fields    []Field `json:"fields"`
    Labels    []string `json:"labels,omitempty"`
    Notes     string   `json:"notes,omitempty"`
    CreatedAt string   `json:"created_at,omitempty"`
    UpdatedAt string   `json:"updated_at,omitempty"`
}
```

Items are user-visible objects such as `OpenAI personal API key` or `Local app token`. Templates determine expected fields and default CLI/display behavior.

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

Field types should start small:

- `string`
- `concealed`
- `url`
- `date`
- `string_list`
- `json`

Sensitivity values:

- `public`: may appear in metadata/list/info outputs.
- `secret`: never appears unless explicitly revealed by `get`, `copy`, export, project run, or another secret-materializing command.

Common purposes:

- `primary_secret`: default value for `get`/`copy` when a template has exactly one primary secret.
- `env_name`: public field that suggests a default env binding.
- `expiry`: date used for expired/expiring views.
- `website`: source or management URL.

### Binding

```go
type Binding struct {
    Type string `json:"type"`
    Name string `json:"name"`
}
```

MVP binding type is `env`. Bindings attach to fields, not items.

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

Built-in templates can be compiled into Go first. Custom templates are deferred.

## Built-In Templates

### `generic_secret`

Fields:

- `value`: concealed, secret, required, `primary_secret`.

Recommended bindings:

- env bindings attach to `value`.

### `api_credential`

Fields:

- `credential`: concealed, secret, required, `primary_secret`.
- `provider`: string, public.
- `website`: url, public, `website`.
- `expires_at`: date, public, `expiry`.
- `scopes`: string_list, public.

Recommended bindings:

- env bindings attach to `credential`.

## Adapter Boundaries

### Repository

```go
type Repository interface {
    Load(context.Context) (*Vault, error)
    Save(context.Context, *Vault) error
}
```

Repository implementations own storage and optional encryption. Domain/app code depends on this interface only.

### Codec

```go
type Codec interface {
    Decode([]byte) (*Vault, error)
    Encode(*Vault) ([]byte, error)
}
```

JSON should be the initial codec. Version detection and migration live near the codec/repository boundary.

### Template Registry

Responsibilities:

- Return built-in templates.
- Validate item fields against template definitions.
- Identify primary secret field.
- Identify public fields for info/list display.
- Identify expiry fields for future expired/expiring filters.

### Env Adapter

Responsibilities:

- Resolve item/field env bindings into value-bearing env pairs.
- Detect duplicate env names and conflicts.
- Preserve value-free manifest behavior.

### Legacy Adapter

Responsibilities:

- Read v1 flat `secrets` maps.
- Convert old `path -> Secret` records into v2 items.
- Preserve old path as a non-secret alias or label if the design chooses selector aliases.
- Migrate `env` to an env binding on the generated primary secret field.
- Migrate `description` to notes.
- Migrate `tags` to labels.
- Do not infer `api_credential` automatically unless a later explicit migration rule is approved.

## Storage Candidate Comparison

### JSON file

Pros:

- Best fit for portable single-file vault.
- Easy to age-encrypt as a whole file later.
- Easy to manage with chezmoi/git after encryption.
- Simple atomic write and `.bak` recovery.
- Flexible during schema iteration.
- Data scale is expected to be small enough for full load/save.

Cons:

- Whole-file rewrite on every change.
- No native query engine.
- Manual index construction required after load.

### Age-wrapped JSON

Pros:

- Matches current product security and portability model.
- Keeps plaintext off disk if encode/decode happen in memory.
- Compatible with current age direction.

Cons:

- Whole-file conflicts do not merge meaningfully.
- Requires careful migration and backup handling.

### Plaintext SQLite

Pros:

- Strong query/index model.
- Better for history/audit/large datasets.
- Mature migration tooling.

Cons:

- Plaintext at rest violates Shelf's security constraint unless wrapped or encrypted.
- Multiple files/WAL/journal/temp artifacts complicate leakage story.
- Worse fit for chezmoi/git-safe portable encrypted file.

### Age-wrapped SQLite

Pros:

- Could preserve age as outer encryption.
- Allows using SQLite during in-memory or temp-session operation.

Cons:

- SQLite wants random access; age is streaming whole-file encryption.
- Likely requires decrypting to temp files or in-memory DB then re-encrypting.
- WAL/journal/temp leakage risk is high unless extremely constrained.
- More complex than JSON without clear current scale benefit.

### Encrypted SQLite / SQLCipher-like storage

Pros:

- Better fit if real database semantics become necessary.
- Can avoid whole-file decrypt/re-encrypt flows.

Cons:

- New native dependency/build/distribution concerns.
- Diverges from current age/chezmoi mental model.
- More operational and security review burden.

### Recommendation for design

Use JSON as the initial canonical storage and keep SQLite behind a future repository adapter. Reconsider SQLite only if v0.2 requirements add history/audit, large datasets, daemon concurrency, or complex query needs that justify SQLCipher-level dependency and leakage review.

## CLI Direction

MVP commands should prioritize non-interactive common workflows and avoid evolving the current prompt-based `secret add`.

Candidate commands:

```bash
shelf secret create generic --title "Local app token" --env APP_TOKEN --stdin
shelf secret create api-credential --title "OpenAI personal API key" --provider openai --website https://platform.openai.com/api-keys --expires-at 2026-12-31 --scope responses --env OPENAI_API_KEY --stdin
shelf secret set <selector> --field credential --stdin
shelf secret get <selector> [--field credential]
shelf secret copy <selector> [--field credential]
shelf secret list --long
shelf secret info <selector>
```

Selector rules to design:

1. Exact item ID.
2. Exact title.
3. Exact env binding when using `--env NAME`.
4. Query match only if exactly one result.
5. Multiple matches fail with value-free candidates and guidance to use ID/title/`--env`.

## Project Manifest v2 Direction

Project manifests should bind item fields, not old secret paths.

Candidate shape:

```json
{
  "version": 2,
  "bindings": [
    {
      "item": "item_01J...",
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
- Manifest must not store field values.
- Project run/export resolves only bound fields with secret sensitivity and env names.
- Missing item/field, duplicate env, and required/optional semantics must have clear diagnostics.

## Tasks

1. Finalize v0.2 item/field/template terminology.
2. Write v0.2 requirements into `.planning/REQUIREMENTS.md` with traceability IDs.
3. Update `.planning/ROADMAP.md` to introduce v0.2.0 phases and explicitly supersede v0.1.1 storage deferral.
4. Create a design/spec artifact for the v2 model, including JSON examples and validation rules.
5. Create a storage decision record comparing JSON and SQLite with leakage and portability criteria.
6. Create a migration plan for v1 vaults and project manifests.
7. Create implementation-phase plans after the design is approved.
8. Keep `STATE.md` pointed at this design phase until approval or implementation planning begins.

## Acceptance Criteria

- `.planning/phases/034-v0.2-template-item-vault-design/CONTEXT.md` records accepted decisions, constraints, open questions, and verification expectations.
- `.planning/phases/034-v0.2-template-item-vault-design/PLAN.md` defines the model, adapters, storage tradeoffs, project-manifest impact, migration direction, risks, and next planning tasks.
- Root planning state points to this active v0.2 design phase.
- The plan does not promise implementation before approval of the design model.
- Another agent can continue by writing the v2 model spec and storage decision record without re-reading this conversation.

## Verification

- Check phase artifacts include required sections from planning contracts:
  - CONTEXT: Goal, Constraints, Decisions, Open Questions, Verification Expectations.
  - PLAN: Objective, Scope, Tasks, Acceptance Criteria, Verification, Risks.
- Check `.planning/STATE.md` points to this phase and next action.
- Check no source code was modified during design planning.

## Risks

- Replacing path-keyed secrets with item IDs may make CLI selectors less convenient unless exact-title, env, and value-free candidate output are designed well.
- Project manifests become less human-readable if they use only item IDs; title snapshots may be necessary but must not become lookup contracts.
- Template flexibility can become over-engineering if v0.2 ships too many templates. MVP should stay at `generic_secret` and `api_credential` unless a specific workflow requires more.
- SQLite may look attractive for querying but can undermine the portable encrypted-file model unless encrypted SQLite is explicitly justified.
- Migration from old `prefix:name` paths to title/item IDs can lose user intent. Preserve old paths as aliases or labels until a better selector story is accepted.
- Pausing or rewriting the Web manager during v0.2 may temporarily reduce editing UX; the plan should state whether manager adaptation is in or out for each implementation phase.
