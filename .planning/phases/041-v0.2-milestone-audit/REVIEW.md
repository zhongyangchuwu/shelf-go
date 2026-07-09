# Review: v0.2.0 Milestone Audit and Migration Scope Reset

## Scope Reviewed

- v0.2.0 phases 34-40 planning artifacts and verification records.
- Architecture boundaries: `.go-arch-lint.yml`, `internal/app`, `internal/project`, `internal/vaultmodel`, `internal/vaultjson`, `internal/manager`, `internal/secret`.
- Functional design: v2 secret CLI, label export, v2 project bindings, v2 manager gate, vault path behavior, plaintext v2 JSON cutover risk.
- Documentation management: `README.md`, `SECURITY.md`, `docs/architecture.md`, `docs/contributing.md`, `docs/reference.md`, `docs/portable-vault.md`, `CHANGELOG.md`, and root `.planning/` records.
- Migration automation scope: `internal/vaultmigrate` and references to public or internal v1-to-v2 conversion.

## Findings

### Blocker: automatic v1-to-v2 migration was premature

`internal/vaultmigrate` had no production consumers and encoded a large model conversion without a user-facing dry-run, review, backup, rollback, or no-secret-log contract. One-user local migration evidence is insufficient for release automation that rewrites secret models.

**Resolution:** Removed `internal/vaultmigrate` and its architecture-lint component. `V2-IMPL-04` is now deferred. Future migration automation requires explicit product and safety requirements before implementation.

### Blocker: security documentation was stale

`SECURITY.md` still described removed commands (`shelf vault open`, `shelf secret edit`) as active plaintext boundaries and did not state the v2 plaintext JSON sidecar risk in the security policy.

**Resolution:** Updated `SECURITY.md` to use `shelf manager`, stdin-based v2 writes, manager v2 gating, and explicit `*.v2.json` plaintext warnings.

### Warning: architecture docs omitted v0.2 packages

`docs/architecture.md` and `docs/contributing.md` did not introduce `internal/vaultmodel` and `internal/vaultjson`, and the package map still implied an `internal/exportfmt` package that no longer exists.

**Resolution:** Updated the package map, dependency direction, v2 model/storage sections, and contributor layout. `internal/util` is now documented as the current env/shell/JSON formatter home.

### Warning: v2 storage lacks a provider abstraction

V1 storage has `vault.Provider`/`vault.Repository`; v2 storage currently constructs `vaultjson.Repository` directly in `internal/app`. This is acceptable for the current JSON-only cutover, but storage encryption or alternate backends should introduce a v2 provider/repository interface before widening persistence behavior.

**Disposition:** Follow-up design item, not a release blocker for the current cutover.

### Info: `internal/secret` is legacy v1 editing support

`internal/secret` still contains v1 add/edit/temp helpers used by manager/app flows. The package name could attract future v2 editing code by mistake.

**Resolution:** Documented `internal/secret` as legacy v1 editing support; future v2 editing should rename/split before adding behavior there.

## Release Assessment

The v2 model, v2 JSON boundary, v2 secret CLI, label export, v2 project bindings, and manager gate are functionally coherent after the audit fixes. The main remaining product risk is not correctness of current workflows; it is security positioning: v2 JSON sidecars are plaintext until age wrapping is implemented.

Recommended release posture:

- If v0.2.0 is a cutover/preview release: proceed after verification, with prominent plaintext warnings.
- If v0.2.0 is meant to be the safe portable-vault release: add a v2 age-wrapping phase before release.
