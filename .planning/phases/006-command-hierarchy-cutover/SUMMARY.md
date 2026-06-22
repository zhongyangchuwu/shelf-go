# Summary: Phase 6 Command Hierarchy Cutover

## Completed Changes

- Replaced ambiguous top-level command registrations with scoped pre-release command paths.
- Added `shelf setup` for global config/vault onboarding.
- Added `shelf vault init`, `shelf vault migrate`, `shelf vault status`/`check`, and `shelf vault open`.
- Moved direct export to `shelf secret export`.
- Moved runtime injection to `shelf project run`.
- Removed root registrations for old top-level `init`, `migrate`, `export`, `run`, and `manager` commands.
- Updated README and usage spec command surfaces.

## Files Changed

- `internal/cli/root.go`
- `internal/cli/init.go`
- `internal/cli/manager.go`
- `internal/cli/secret.go`
- `internal/cli/project.go`
- `internal/cli/doctor.go`
- `internal/cli/*_test.go`
- `README.md`
- `docs/usage-spec.md`
- `docs/data-spec.md`
- `.planning/PROJECT.md`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/phases/006-command-hierarchy-cutover/CONTEXT.md`
- `.planning/phases/006-command-hierarchy-cutover/PLAN.md`

## Deviations

- `shelf vault status` was implemented in this phase because it is a small vault UX improvement and helps validate the new vault namespace immediately.
- `shelf vault check` is an alias for `shelf vault status`.

## Evidence

- `go test ./internal/cli -run 'Test(Setup|Migrate|Export|Secret|Run|Root|Manager|Completion)'` passed.
- `go test ./internal/cli -run 'Test(Vault|Root|Manager|Migrate|Setup)'` passed.
- `go test ./...` passed.
