# Review: Secret CLI v2 Workflows

## Scope Reviewed

- `internal/vaultmodel/model.go`
- `internal/vaultmodel/templates.go`
- `internal/vaultmodel/validate.go`
- `internal/vaultmodel/index.go`
- `internal/vaultjson/codec_test.go`
- `internal/vaultmigrate/vault.go`
- `internal/app/secret_v2.go`
- `internal/cli/secret.go`
- v2 app/CLI/model/migration tests
- User-facing docs and planning records

## Findings

- No blocking findings.
- The old env binding subsystem is removed from the v2 model; env names are modeled as public fields and validated through template metadata.
- The `openai` provider defaults are registry profile data, not CLI switch logic.
- Secret writes through v2 CLI require stdin for `create --stdin` and `set --stdin`; no v2 command accepts secret values as positional argv.
- `secret list --long`, `secret info`, and duplicate env ambiguity errors are value-free for secret fields.
- Selector resolution is intentionally exact: item ID, title, alias, or `--env` lookup.
- Removed v1 command paths are tested as absent instead of carried through compatibility shims.
- Architecture lint now explicitly allows `internal/app` to depend on v2 model/repository packages because the v2 app service owns command orchestration over those boundaries.

## Fixes Applied

- Added a distinct default v2 path rule: v2 secret commands use `<resolved-vault-path>.v2.json` unless `--vault` is explicitly supplied.
- Documented that Phase 38 v2 JSON vault files are plaintext cutover artifacts and must not be committed or synced.
- Added direct app-layer v2 tests after CLI tests so reusable behavior is defended outside Cobra.
- Removed stale v1 direct-export/tag-selection CLI tests that conflicted with the clean breaking cutover.
- Updated docs to stop teaching old `group:key` value argv workflows as current secret CLI usage.

## Waivers

- v2 storage encryption is not implemented in this phase. The plaintext v2 repository boundary is inherited from Phase 36 and is documented as a cutover risk.
- Project export/run still use v1 secret resolution. Phase 39 owns project manifest v2 bindings and runtime integration.
- Manager adaptation remains deferred to Phase 40.
- `secret copy` remains deferred until clipboard/platform behavior is designed.

## Remaining Risks

- Users must not treat `*.v2.json` as git-safe until v2 age wrapping lands.
- The v2 secret CLI is useful for create/set/get/list/info/rm, but project workflows are incomplete until Phase 39.
- The repository now contains both v1 and v2 app surfaces during the transition; future phases should remove or gate v1 surfaces only after project/manager cutover is complete.
