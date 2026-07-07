# Migration Plan: v1 Flat Secrets to v2 Items

## Goal

Define a safe migration path from v1 flat secret records and project manifests to v2 template-based items and field bindings.

## Inputs

Current v1 vault records:

```go
type Secret struct {
    Value       json.RawMessage `json:"value"`
    Env         string          `json:"env,omitempty"`
    Description string          `json:"description,omitempty"`
    Tags        []string        `json:"tags,omitempty"`
}
```

Current v1 project manifest entries:

```go
type Entry struct {
    Path     string   `json:"path,omitempty"`
    Prefix   string   `json:"prefix,omitempty"`
    Tags     []string `json:"tags,omitempty"`
    Env      string   `json:"env,omitempty"`
    Required *bool    `json:"required,omitempty"`
}
```

## v1 Secret Conversion

For every v1 `path -> Secret` record:

- Create one v2 item.
- Use template `generic_secret`.
- Generate a stable item ID.
- Set item title from the old path until a better naming rule is explicitly accepted.
- Preserve old path as an alias for migration-period lookup.
- Map `Secret.Value` to field `value`.
- Map `Secret.Env` to an env binding on field `value` when set.
- Map `Secret.Description` to item notes.
- Map `Secret.Tags` to item labels.

Do not infer `api_credential` from old path/key names in the default migration. Heuristic template inference risks incorrect semantics and must be a separately reviewed rule.

## Item ID Strategy

Implementation phase must choose one deterministic strategy before writing migrations:

1. Generate new IDs and write a migration map from old path to new ID.
2. Derive stable IDs from old paths using a collision-resistant deterministic function.

Recommendation: generate IDs for new user-created items; use deterministic migration IDs for v1 imports so rerunning migration can be idempotent. Preserve old paths as aliases either way.

## Project Manifest Conversion

Path entries:

- Resolve old `path` through the migration map.
- Bind `item` to the migrated item ID.
- Bind `field` to `value`.
- Preserve explicit `env`; otherwise resolve from the v2 field binding.
- Preserve `required`.
- Store `title` as a value-free review snapshot.

Prefix entries:

- Expand against the v1 vault during migration.
- Create one v2 binding per matched migrated item/field.
- Mark derived entries with value-free title snapshots.
- If prefix expansion has no matches, preserve a migration warning rather than inventing values.

Tag entries:

- Expand against v1 tags during migration.
- Create one v2 binding per matched migrated item/field.
- Preserve required/optional semantics.
- Do not store tag selectors as executable v2 manifest selectors unless a later project-manifest design explicitly supports label selectors.

## Safety Rules

- Never delete or overwrite the v1 vault as part of initial migration.
- Write v2 output to a new path or backup-protected target.
- Validate the v2 vault before reporting migration success.
- Validate the converted manifest before replacing `.shelf.json`.
- Report value-free migration diagnostics: counts, paths/titles, item IDs, env names, labels, and conflicts. Do not print secret values.
- Conflicting env bindings fail before writing unless the command supports an explicit reviewed resolution mode.

## Required Diagnostics

Migration should report:

- Number of v1 secrets read.
- Number of v2 items written.
- Number of project bindings converted.
- Duplicate env binding conflicts.
- Unknown or unparseable v1 records.
- Prefix/tag selectors that expand differently than expected.
- Any skipped item with value-free reason.

## Rollback

Rollback is ordinary file restore:

- Keep the v1 vault unchanged.
- Keep a backup of any replaced v2 vault or project manifest.
- If validation fails, leave existing files untouched and report the failed validation rule.

## Open Implementation Questions

- Exact deterministic ID format for migrated records.
- Whether aliases stay permanently or only for a migration window.
- Whether project manifests may ever bind by alias/title, or only canonical item ID and field ID.
- Whether migrated `title` should default to old path or a cleaned last path segment plus key.
