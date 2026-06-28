# Phase 26 Context: App Service Extraction

## Goal

Move reusable command orchestration out of `internal/cli` and into `internal/app` services while keeping CLI responsible for Cobra wiring, prompts, completions, stdout/stderr routing, and process/server lifecycle.

## Constraints

- Preserve all user-visible command behavior, flags, output strings, error strings, vault format, and config format.
- Do not change project/session domain boundaries completed in Phase 25.
- Keep interactive prompting in CLI; app services must accept explicit request structs or values.
- Keep Cobra, `*cobra.Command`, shell completion callbacks, and terminal I/O out of `internal/app`.
- Keep vault encryption, store mutation primitives, and status/doctor primitives in `internal/vault`.
- Keep export formatting in `internal/exportfmt`; app services may orchestrate formatter selection.

## Decisions

- Extract `secret export` selection/filter/format orchestration to `internal/app` first because it currently mixes selector rules, vault reads, and formatter dispatch inside CLI.
- Extract plaintext migration implementation to `internal/app` because it composes source read, vault format detection, encrypted save, verification, and source preservation checks.
- Extract setup file helpers and setup file orchestration to `internal/app` while leaving prompt collection in CLI.
- Extract manager loopback listener and token helpers to `internal/app`; CLI still owns HTTP server lifecycle and signal handling.

## Rejected Options

- Do not move Cobra completions into app services.
- Do not introduce a new storage abstraction or change the vault file format.
- Do not move Web manager server implementation; it already lives in `internal/manager`.
- Do not move CLI prompt code for setup/secret add; prompts are adapter behavior.

## Open Questions

- None blocking. Function names may be adjusted during implementation, but services should remain Cobra-free.

## Canonical References

- `internal/cli/export.go`
- `internal/cli/init.go`
- `internal/cli/migrate.go`
- `internal/cli/manager.go`
- `internal/app/runtime.go`
- `internal/vault/*`
- `internal/exportfmt/*`

## Verification Expectations

- Direct `internal/app` tests cover export selection/filter/format, migration preservation checks, setup helper behavior, loopback listener validation, and manager token shape.
- Focused CLI tests confirm export, setup/init, migrate, manager, vault, and doctor command behavior remains unchanged.
- Full `go test ./...` passes.
