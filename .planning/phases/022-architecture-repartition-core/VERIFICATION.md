# Verification: Phase 22 Architecture Repartition Core

## Claims Checked

1. `shelf manager` is registered as the manager entrypoint.
2. `shelf vault open` is not registered.
3. Vault core and diagnostics live under `internal/vault`.
4. Project manifest schema/IO/validation live under `internal/project`.
5. Version composition lives under `internal/app`.
6. Export env/shell/JSON formatting lives under `internal/exportfmt`.
7. Behavior remains unchanged apart from the intentional manager command rename.
8. Removed internal package paths are gone from active code/docs references.

## Evidence Observed

- `go test ./internal/cli -run 'Test.*Manager|TestVault'`
  - Result: passed.
  - Covers root manager command presence and absence of `vault open`, plus vault status/check command behavior.
- `go test ./internal/vault`
  - Result: passed.
  - Covers encrypted vault core, legacy plaintext IO compatibility, file locking, and tag/path store helpers.
- `go test ./internal/project`
  - Result: passed.
  - Covers project manifest validation and project tag binding resolution after manifest merge.
- `go test ./...`
  - Result: passed.
  - Covers all package compilation and existing behavior tests.
- `go vet ./...`
  - Result: passed.
- LSP workspace diagnostics
  - Result: no Go issues found.
- Active reference search for removed paths
  - Result: no active code/docs references to `internal/store`, `internal/manifest`, `internal/render`, `internal/version`, or `internal/atomicfile` remain outside current phase/planning references.
- `glob internal/*`
  - Result: final package set is `app`, `cli`, `config`, `exportfmt`, `manager`, `project`, `secret`, `vault`.

## Coverage

- Command wiring: covered by focused CLI tests.
- Package movement: covered by full compile/test and package-specific tests.
- Docs consistency: minimal architecture/contributing/reference/getting-started/troubleshooting references updated; full prose pass deferred.
- Regression: full test suite and vet passed.

## Gaps

- Full user documentation rewrite is deferred to Phase 23.
- Release snapshot after package repartition is deferred to Phase 24 release hardening.
- Native Windows test run was not repeated in this phase.

## Result

Passed. Phase 22 acceptance criteria are satisfied.
