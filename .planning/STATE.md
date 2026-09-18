# State

## Current Position

- Version: v0.2.0 final experimental milestone
- Schema: V4
- Status: maintenance / research artifact
- Active Artifact: none
- Active Feature Roadmap: none
- Next Action: use established secret infrastructure for new personal workflows; change Shelf only for a concrete bug, security issue, documentation correction, or proven integration gap.

## Blockers

- None.

## Historical and Current Evidence
- The default CLI version is `v0.2.0`; `shelf --version` reported the local milestone without creating a public tag or release.

- Phase 53 completed the historical V3 breaking schema cutover: vault fields persisted only `id` and `value` while templates owned field shape. V4 later replaced item aliases/legacy notes with public `description`; V4 is the active schema.
- `env_<field>` maps eligible secret and public `string`, `concealed`, or `url` fields for `shelf env export`; `secret export` is removed with no alias.
- `validity_<secret-field>` provides field-level lifecycle audit states: unknown, permanent, invalid, expired, expiring, and healthy. Untracked fields are omitted.
- Project bindings accept the same eligible public and secret field types; manifests stay value-free.
- Secret and project TUI workspaces remain optional, keyboard-first, and value-free. Secret detail groups structural env/validity metadata under its target field.
- `docs/v3-migration.md` remains a historical manual migration record. Current maintained docs describe the V4 command contract.
- Phase 53 verification passed: focused V3 package tests, `bash scripts/test.sh`, `go build ./cmd/shelf`, a temporary-vault CLI projection/audit/project-export smoke, and a 120×32 pseudo-terminal secret TUI smoke with secret redaction and alternate-screen restoration.
- The active local plaintext vault was manually migrated from V2 to V3: 16 generic-secret items changed `value_env` to `env_value`, with atomic fields only. Exact V2 backup: `/home/han/.local/share/shelf/backups/pre-v3-migration-20260714-024329/vault.json` (`0600`). Active `doctor`, `secret list --long`, audit, info, and both label projection groups passed; one group intentionally projected no mapped fields.
- The local V3 vault was semantically consolidated: 16 legacy `item:field` titles became 15 item titles without `:`; the one two-field group uses one custom template with two concealed secret fields. All 16 secret reads and both label projections matched the pre-consolidation V3 vault. No `.shelf.json` manifests exist under `/home/han/project`; selector mapping and exact V3 backup: `/home/han/.local/share/shelf/backups/pre-semantic-item-migration-20260714-025446/` (`0600`).
- The user chose template-first modeling: `generic_secret` remains a single concealed secret; multi-field services use explicit templates. The CLI Proxy item now uses built-in `api_credential`, preserving its credential and marking `base_url` public `url`; lifecycle validity is deliberately `unknown`. Removed the migration-only custom template. Exact pre-change V3 backup and mapping: `/home/han/.local/share/shelf/backups/pre-cli-proxy-template-correction-20260714-031500/` (`0600`). Resolver now treats an absent optional `env_<field>` instance as unmapped, so that public URL is safely skipped by `env export` until mapped.
- The proposed personal workflow sequence is scriptable item lifecycle → wide-screen secret workspace TUI → project binding TUI → credential health audit.
- TUI remains optional and keyboard-first; it must never display secret values and must reuse application workflows.
- Implemented the V4 TUI UX direction: Catalog is title-first with accessible colored `×`/`!`/`●`/`·` action-state symbols; Detail owns fields, lifecycle reasons, and ENV mappings. Create and Edit are continuous scrollable forms; all mutable stored fields are editable atomically through `app`, public values prefill, and secret replacements are masked and never prefilled. Create exposes `Single secret` versus `API credential`, then optional provider presets such as OpenAI. Stable IDs stay generated/readonly, and template conversion remains separate.
- Schema V4 replaces item aliases and legacy `notes` with `description`; item lookup now accepts only ID or title. The active local vault was manually migrated from V3 to V4 with 14 alias texts preserved as descriptions and all item-field payloads unchanged. Exact V3 backup and migration manifest: `/home/han/.local/share/shelf/backups/pre-v4-description-migration-20260715-175049/` (`0600`). Source `doctor`, value-free list/info, and a rebuilt 120×32 pseudo-terminal TUI smoke passed with redaction confirmed; `bash scripts/test.sh` passed.
- Named dev/staging/prod environments are not planned for the personal-use path.
- Default vault path is `~/.local/share/shelf/vault.json`.
- Local config points at `/home/han/.local/share/shelf/vault.json`.
- Local backup before cutover: `/home/han/.local/share/shelf/backups/pre-plain-json-cutover-20260710-162158`.
- Latest Phase 52 verification passed:
  - secret and project model panel/focus/value-free tests
  - 120×32 secret and project pseudo-terminal Mocha panel smokes
  - `bash scripts/test.sh` (`go vet`, `go test`, and `go-arch-lint`)
  - `go build ./cmd/shelf`

## Updated

- 2026-09-18
