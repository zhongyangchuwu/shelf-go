# Verification: Phase 3 Project Workflow Compatibility

## Claims Checked

- CLI-02: `shelf export` exact-path and prefix flows render from encrypted storage in shell/env/JSON formats.
- CLI-03: `shelf project` commands resolve encrypted-vault paths/prefixes without storing secret values in `.shelf.json`.
- CLI-04: `shelf run -- ...` and `shelf run --dry-run -- ...` work against encrypted storage and preserve value-printing rules.
- CLI-05 / TEST-01: Regression coverage proves storage encryption did not change export/project/run semantics.

## Evidence Observed

- `go test ./internal/cli -run 'Test(Export|Project|Run)'` passed.
- `go test ./internal/cli ./internal/store ./internal/render ./internal/manifest` passed.
- `go test ./...` passed.
- `TestExportFormatsReadEncryptedVault` checks shell/env/JSON exact-path output and asserts vault bytes do not contain known value, path, or env name.
- `TestExportPrefixReadsEncryptedVault` checks prefix export default filtering and `--all` derived env behavior.
- `TestProjectAddKeepsManifestValueFree` asserts `.shelf.json` contains path/env metadata but not the known secret value.
- `TestProjectExportUsesEncryptedVaultWithoutPlaintextSideData` asserts project export renders the value and encrypted vault bytes do not contain known plaintext project data.
- `TestRunDryRunUsesEncryptedVaultWithoutPlaintextSideData` asserts dry-run prints injection/override diagnostics without secret or parent env values and vault bytes remain value-free.

## Coverage

- Direct export: exact path, prefix, env filtering, `--all`, shell/env/JSON.
- Project workflow: add, list, explain, export, required/optional missing, prefix expansion, conflict paths through existing and new tests.
- Runtime injection: exact path, prefix-derived env names, override behavior, dry-run no-leak behavior, failed resolution preventing execution, child exit propagation through existing and new tests.
- Encryption path: tests pass `--vault`, which triggers the age-backed test config helper and real encrypted vault persistence.

## Gaps

- Localhost vault manager is not covered; it belongs to Phase 4.
- End-user documentation is not updated here; Phase 5 owns docs and release hardening.

## Result

Phase 3 verification passed. All Phase 3 success criteria have automated evidence.
