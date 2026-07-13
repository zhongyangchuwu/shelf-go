# Specification: Item-First Environment Projection V3

**Status:** Approved for implementation.

## Goal

Make each vault item a service, integration, or account. Keep fields atomic and express field relationships with stable prefix conventions, not embedded target metadata. Project and label workflows produce explicit environment projections from both secret and public fields.

## V3 Persistence Contract

- Vault schema version is `3`; V2 is intentionally unsupported after cutover. Shelf does not modify a user vault automatically.
- Item field instances persist only `id` and `value`; template field definitions own labels, types, sensitivity, and requiredness.
- A template declares `primary_secret`, naming exactly one secret field.
- `env_<field_id>` is a public string field containing the item's default ENV name for `<field_id>`.
- `validity_<field_id>` is a public string field for a tracked secret field. Its required value is `unknown`, `permanent`, or an ISO date (`YYYY-MM-DD`).
- A template field without a `validity_<field_id>` companion is `not_tracked` for lifecycle audit.
- `env_` and `validity_` are reserved structural prefixes. Their target comes from the suffix; validation must reject dangling, ineligible, or nested targets.
- ENV targets may be `string`, `concealed`, or `url` fields of either sensitivity. JSON/list/date values are not environment-projectable.

## Environment Projection

- `shelf env export --label LABEL` cleanly replaces `shelf secret export` with no alias.
- A matched item contributes every field that has a non-empty `env_<field_id>` mapping.
- Both secret and public eligible fields project. Unmapped fields are metadata only.
- Duplicate ENV names or mapped fields without values fail without rendering values. Output is sorted by ENV.
- Project manifests remain value-free and bind any eligible `item.field` to a project-specific ENV. Project bindings can override item default ENV names.

## Lifecycle Audit

- `unknown` is the default value for template-declared `validity_<secret_field>` fields.
- `permanent` is explicit and has no warning.
- ISO dates produce expired/due/healthy status under the existing UTC horizon semantics.
- `not_tracked` fields are omitted from audit.
- Audit output contains `unknown`, `permanent`, `invalid`, `expired`, `expiring`, or `healthy`; only unknown/invalid/expiring/expired affect health exits.

## Built-in Templates

```text
generic_secret
  primary_secret: value
  value                    concealed secret required
  env_value                string public

api_credential
  primary_secret: credential
  credential               concealed secret required
  env_credential           string public
  validity_credential      string public required, default unknown
  base_url                 url public
  env_base_url             string public
  client_id                string public
  env_client_id            string public
  provider                 string public
  website                  url public
  scopes                   string_list public
```

## UI Contract

- Catalog remains item-first and reports aggregate projectable-field and set-field counts.
- Detail shows atomic fields; structural `env_` and `validity_` fields are grouped with their target rather than presented as unrelated credentials.
- No secret value becomes renderable.

## Non-goals

- No automatic V2 vault migration, old-command alias, named environments, field-level labels, or public values in `.shelf.json`.
- No project or user vault file is modified by this implementation.

## Acceptance

- V3 validation rejects stale V2 structures and malformed convention fields.
- Label export and project export/run materialize all explicitly mapped eligible fields, including public fields, with duplicate/missing safeguards.
- Lifecycle audit distinguishes unknown, permanent, date-based health, invalid, and not-tracked fields.
- CLI, TUI, docs, completions, and tests use `env export` and the V3 field contract.