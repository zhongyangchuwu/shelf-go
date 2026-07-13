# State

## Current Position

- Version: v0.2.0 local milestone
- Phase: Item-First Environment Projection V3
- Status: complete
- Active Artifact: `.planning/phases/053-item-first-environment-projection/VERIFICATION.md`
- Next Action: Continue using the active V3 vault; if an external encrypted source is authoritative, apply the same manual migration there before its next sync.

## Blockers

- None.

## Recent Evidence
- The default CLI version is `v0.2.0`; `shelf --version` reported the local milestone without creating a public tag or release.

- Phase 53 completed the V3 breaking schema cutover: vault fields persist only `id` and `value`; templates own field shape; schema version is `3`.
- `env_<field>` maps eligible secret and public `string`, `concealed`, or `url` fields for `shelf env export`; `secret export` is removed with no alias.
- `validity_<secret-field>` provides field-level lifecycle audit states: unknown, permanent, invalid, expired, expiring, and healthy. Untracked fields are omitted.
- Project bindings accept the same eligible public and secret field types; manifests stay value-free.
- Secret and project TUI workspaces remain optional, keyboard-first, and value-free. Secret detail groups structural env/validity metadata under its target field.
- User-facing docs describe the V3 command contract and `docs/v3-migration.md` provides a value-free manual migration checklist.
- Phase 53 verification passed: focused V3 package tests, `bash scripts/test.sh`, `go build ./cmd/shelf`, a temporary-vault CLI projection/audit/project-export smoke, and a 120×32 pseudo-terminal secret TUI smoke with secret redaction and alternate-screen restoration.
- The active local plaintext vault was manually migrated from V2 to V3: 16 generic-secret items changed `value_env` to `env_value`, with atomic fields only. Exact V2 backup: `/home/han/.local/share/shelf/backups/pre-v3-migration-20260714-024329/vault.json` (`0600`). Active `doctor`, `secret list --long`, audit, info, and both label projection groups passed; one group intentionally projected no mapped fields.
- The local V3 vault was semantically consolidated: 16 legacy `item:field` titles became 15 item titles without `:`; the one two-field group uses one custom template with two concealed secret fields. All 16 secret reads and both label projections matched the pre-consolidation V3 vault. No `.shelf.json` manifests exist under `/home/han/project`; selector mapping and exact V3 backup: `/home/han/.local/share/shelf/backups/pre-semantic-item-migration-20260714-025446/` (`0600`).
- The proposed personal workflow sequence is scriptable item lifecycle → wide-screen secret workspace TUI → project binding TUI → credential health audit.
- TUI remains optional and keyboard-first; it must never display secret values and must reuse application workflows.
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

- 2026-07-13
