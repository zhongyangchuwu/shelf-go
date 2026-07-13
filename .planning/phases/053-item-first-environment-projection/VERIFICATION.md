# Verification: Item-First Environment Projection V3

## Claims Checked

- Vault schema version 3 stores atomic `id`/`value` fields; templates own type, sensitivity, and requiredness.
- `env_<field>` maps eligible secret and public fields for `shelf env export`; duplicate and missing-value errors remain value-free.
- Project bindings accept public and secret `string`, `concealed`, and `url` fields even without an item default env mapping.
- `validity_<secret-field>` classifies unknown, permanent, invalid, expired, expiring, and healthy field lifecycle without rendering secret values.
- TUI catalog summaries are item-first and detail groups structural mappings with their target without revealing secret values.
- No V2 command alias or automatic vault migration exists; `docs/v3-migration.md` provides a manual checklist.

## Evidence Observed

- Focused regression suite passed: `go test ./internal/vaultmodel ./internal/vaultjson ./internal/project ./internal/app ./internal/cli ./internal/tui`.
- Full validation passed: `bash scripts/test.sh` (`go vet ./...`, `go test ./...`, and `go-arch-lint check`) and `go build ./cmd/shelf`.
- Temporary-vault CLI smoke created an API item, mapped secret and public URL fields, and verified JSON environment projection, permanent audit status, and public project export.
- 120×32 pseudo-terminal secret TUI smoke rendered the catalog, exited cleanly, restored the alternate screen, and did not render the test secret.

## Result

passed