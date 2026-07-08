# Verification: Manager Adaptation and Cutover

## Claims Checked

- `shelf manager` fails before opening the v1 manager service when the resolved valid v2 vault exists.
- Explicit `--vault <v2-json-path>` uses that path as the v2 gate path.
- Default/config vault paths use `<resolved-vault-path>.v2.json` as the v2 gate path.
- The gate error is value-free and directs users to CLI v2 workflows.
- Existing v1 manager behavior still opens when no valid v2 vault is present.
- Existing manager server security and value-free semantics remain intact.
- Documentation describes the temporary v2 manager gate.

## Evidence Observed

Focused verification:

```text
go test ./internal/app ./internal/cli ./internal/manager
```

Full verification:

```text
go test ./...
go vet ./...
go-arch-lint check
```

All commands passed.

Tester agent independently reported:

```text
go test ./internal/app/ PASS
go test ./internal/cli/ PASS
go test ./internal/manager/ PASS
```

## Test Coverage

- `internal/app/manager_test.go`
  - `TestOpenManagerV2VaultSidecarGate`
  - `TestOpenManagerExplicitV2VaultPathGate`
  - `TestOpenManagerSucceedsWithoutV2Vault`
- `internal/cli/manager_test.go`
  - `TestManagerGatesExplicitV2Vault`
  - `TestManagerRejectsV2VaultSidecar`
- Existing `internal/manager` tests continue to cover:
  - token and host checks;
  - origin checks;
  - no-store headers;
  - metadata-only list/detail responses;
  - explicit reveal endpoint behavior.

## Gaps

- No v2 manager UI exists yet.
- V2 repository age wrapping remains deferred.
- The gate intentionally does not treat arbitrary invalid files as v2 vaults.

## Result

Pass. Phase 40 satisfies `V2-IMPL-07` through explicit gating while preserving v1 manager behavior.
