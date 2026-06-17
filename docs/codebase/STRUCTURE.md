# Codebase Structure

**Analysis Date:** 2026-06-16

## Directory Layout

```text
shelf-go/
├── cmd/
│   └── shelf/              # Executable package with `main`
├── internal/
│   ├── cli/                # Cobra command surface and CLI-only orchestration
│   ├── config/             # Runtime config path/editor resolution
│   ├── manifest/           # `.shelf.json` project manifest model, validation, IO
│   ├── render/             # Env/shell/JSON output formatting
│   ├── store/              # Secret data model, path grammar, locking, JSON persistence
│   └── version/            # Version string and VCS revision lookup
├── docs/                   # Product, usage, data, and roadmap docs
│   └── superpowers/plans/  # Planning artifacts
├── .planning/codebase/     # Generated codebase maps for GSD workflows
├── bin/                    # Local build output
├── go.mod                  # Go module and dependency declarations
├── go.sum                  # Go dependency checksums
├── justfile                # Local task shortcuts
└── README.md               # Project overview and command surface
```

## Directory Purposes

**`cmd/shelf`:**
- Purpose: Contains the compiled command's process entry point.
- Contains: `main.go`.
- Key files: `cmd/shelf/main.go`.

**`internal/cli`:**
- Purpose: Owns all Cobra command definitions, command-specific helpers, CLI output behavior, and CLI integration tests.
- Contains: one file per command group or feature area, plus co-located `_test.go` files.
- Key files: `internal/cli/root.go`, `internal/cli/secret.go`, `internal/cli/project.go`, `internal/cli/run.go`, `internal/cli/export.go`, `internal/cli/init.go`, `internal/cli/doctor.go`, `internal/cli/completion.go`, `internal/cli/test_helpers_test.go`.

**`internal/config`:**
- Purpose: Resolves runtime configuration from flags, env vars, YAML config, defaults, and path expansion.
- Contains: config/runtime structs and path resolution helpers.
- Key files: `internal/config/config.go`.

**`internal/store`:**
- Purpose: Owns the local secret store schema and persistence behavior.
- Contains: data model, path grammar, JSON value validation, file locking, load/save, CRUD-style store methods, and store tests.
- Key files: `internal/store/model.go`, `internal/store/path.go`, `internal/store/validate.go`, `internal/store/io.go`, `internal/store/lock.go`, `internal/store/io_test.go`.

**`internal/manifest`:**
- Purpose: Owns the project-level `.shelf.json` manifest schema and persistence behavior.
- Contains: manifest model, entry mutation helpers, strict JSON IO, validation, and manifest tests.
- Key files: `internal/manifest/manifest.go`, `internal/manifest/io.go`, `internal/manifest/validate.go`, `internal/manifest/manifest_test.go`.

**`internal/render`:**
- Purpose: Converts store secrets and resolved bindings into output formats.
- Contains: env-name derivation, JSON raw-value conversion, shell quoting, and env/shell/JSON renderers.
- Key files: `internal/render/export.go`.

**`internal/version`:**
- Purpose: Provides the root Cobra command version string.
- Contains: version constant and VCS build-info lookup.
- Key files: `internal/version/version.go`.

**`docs`:**
- Purpose: Human-readable specs and roadmap that describe command behavior, data model, and product direction.
- Contains: Markdown specs and planning artifacts.
- Key files: `docs/usage-spec.md`, `docs/data-spec.md`, `docs/roadmap.md`, `docs/superpowers/plans/2026-06-10-shelf-go-mvp.md`.

**`.planning/codebase`:**
- Purpose: Generated architecture/stack/testing/convention/concern maps consumed by GSD planning and execution workflows.
- Contains: generated Markdown reference docs.
- Key files: `.planning/codebase/ARCHITECTURE.md`, `.planning/codebase/STRUCTURE.md`.

**`bin`:**
- Purpose: Local build artifact directory.
- Contains: compiled binaries such as `bin/shelf`.
- Key files: `bin/shelf`.

## Key File Locations

**Entry Points:**
- `cmd/shelf/main.go`: Process entry point for the `shelf` binary.
- `internal/cli/root.go`: Root Cobra command, persistent flags, command registration, runtime load helpers.

**Configuration:**
- `go.mod`: Module path, Go version, and dependencies.
- `justfile`: Local task runner commands.
- `internal/config/config.go`: Runtime config resolution and defaults.
- `README.md`: Current supported command surface.

**Core Logic:**
- `internal/cli/secret.go`: Secret CRUD, interactive `secret add`, edit flow, secret completion.
- `internal/cli/project.go`: Project manifest commands, project binding resolution, Git project identity.
- `internal/cli/run.go`: Runtime env injection and child process exit code propagation.
- `internal/store/io.go`: Store load/save and mutation methods.
- `internal/store/model.go`: JSON store data types.
- `internal/store/path.go`: Secret path grammar.
- `internal/store/validate.go`: Secret value/env/tag validation.
- `internal/store/lock.go`: File lock for mutating store commands.
- `internal/manifest/manifest.go`: Project manifest data types and entry mutation.
- `internal/manifest/io.go`: Strict manifest JSON load/save.
- `internal/manifest/validate.go`: Manifest schema validation.
- `internal/render/export.go`: Env/shell/JSON rendering.

**Testing:**
- `internal/cli/*_test.go`: Command-level behavior tests using Cobra command instances.
- `internal/cli/test_helpers_test.go`: Shared CLI test helpers.
- `internal/store/io_test.go`: Store loading/validation tests.
- `internal/manifest/manifest_test.go`: Manifest model/IO/validation tests.

**Documentation:**
- `docs/usage-spec.md`: Command behavior and UX rules.
- `docs/data-spec.md`: Store format, path grammar, and persistence policy.
- `docs/roadmap.md`: Product roadmap.
- `README.md`: Short overview and command list.

## Naming Conventions

**Files:**
- Use short lowercase package-oriented names for implementation files: `internal/store/io.go`, `internal/store/path.go`, `internal/config/config.go`.
- Use command or feature names for CLI files: `internal/cli/secret.go`, `internal/cli/project.go`, `internal/cli/run.go`, `internal/cli/doctor.go`.
- Use co-located Go test files with `_test.go`: `internal/cli/project_test.go`, `internal/store/io_test.go`.

**Directories:**
- Use Go package directories under `internal/` for architectural boundaries: `internal/store`, `internal/manifest`, `internal/render`.
- Use `cmd/<binary>` for executable packages: `cmd/shelf`.
- Use `docs/` for product/reference docs and `.planning/` for generated workflow docs.

**Functions and Types:**
- Command constructors are unexported and named `new<Name>Cmd`: `newProjectCmd`, `newRunCmd`, `newExportCmd`.
- Shared test helpers are unexported and named for their behavior: `runShelf`, `runShelfWithInput`, `setupProjectTest`.
- Exported domain types use Go PascalCase: `store.Store`, `store.Secret`, `manifest.Manifest`, `config.Runtime`.
- Package-private CLI structs use lower camel case: `resolved`, `projectDiagnostic`, `editableSecret`.

## Where to Add New Code

**New CLI Command:**
- Primary code: add a new command file or extend the related command file in `internal/cli`.
- Registration: add the command to `NewRootCmd` in `internal/cli/root.go`, or to a command group such as `newProjectCmd` in `internal/cli/project.go`.
- Tests: add co-located command tests in `internal/cli/<command>_test.go`.
- Use pattern: define `new<Name>Cmd() *cobra.Command`, set `Args`, use `RunE`, and write output through Cobra output handles.

**New Secret Store Behavior:**
- Primary code: add model or persistence behavior under `internal/store`.
- Tests: add or extend `internal/store/*_test.go`.
- Use pattern: validate inputs with `ValidatePath`, `ValidateSecretID`, or `ValidateSecret`; keep load/save semantics centralized in `internal/store/io.go`.

**New Project Manifest Behavior:**
- Primary code: add manifest schema or validation behavior under `internal/manifest`.
- CLI orchestration: wire user-facing command behavior in `internal/cli/project.go`.
- Tests: add manifest-only tests in `internal/manifest/manifest_test.go` and command behavior tests in `internal/cli/project_test.go`.
- Use pattern: exact path and prefix entries remain mutually exclusive; keep duplicate detection in manifest methods or validation.

**New Project Runtime Resolution Feature:**
- Primary code: prefer `internal/cli/project.go` while behavior is CLI-only.
- Refactor path: if non-CLI packages need project resolution, extract `resolved`, `projectDiagnostic`, and `resolveProjectEntries` out of `internal/cli/project.go` into a new internal package.
- Tests: extend `internal/cli/project_test.go` and `internal/cli/run_test.go` where behavior affects runtime injection.

**New Output Format:**
- Primary code: add renderer functions in `internal/render/export.go`.
- CLI wiring: add format selection to `internal/cli/export.go` and `internal/cli/project.go`.
- Tests: add command tests in `internal/cli/secret_test.go` or `internal/cli/project_test.go`; add render package tests if renderer logic becomes branchy.

**New Configuration Option:**
- Primary code: add fields and resolution behavior in `internal/config/config.go`.
- CLI wiring: add flags in `internal/cli/root.go` only for options that should be global.
- Tests: add config package tests if introduced; add command tests for user-visible behavior.

**New Health Check:**
- Primary code: add a `check*` helper in `internal/cli/doctor.go`.
- Tests: extend `internal/cli/doctor_test.go`.
- Use pattern: call `report.ok`, `report.warn`, or `report.fail`; return a failing command only when `doctorReport.failed` is true.

**Utilities:**
- Shared CLI-only helpers: place in `internal/cli` near the command that owns the behavior, or in a small file with a clear feature name.
- Shared domain helpers: place in the package that owns the data or invariant, such as `internal/store` for secret path/value invariants or `internal/render` for output formatting.

## Special Directories

**`internal`:**
- Purpose: Go internal packages that cannot be imported by external modules.
- Generated: No.
- Committed: Yes.

**`cmd`:**
- Purpose: Executable package roots.
- Generated: No.
- Committed: Yes.

**`docs`:**
- Purpose: Product and technical documentation.
- Generated: No.
- Committed: Yes.

**`docs/superpowers/plans`:**
- Purpose: Planning artifacts for previous or active roadmap work.
- Generated: Partly workflow-generated.
- Committed: Yes.

**`.planning`:**
- Purpose: GSD workflow state and generated codebase maps.
- Generated: Yes.
- Committed: Project-dependent; current files are present in the workspace.

**`bin`:**
- Purpose: Local build output.
- Generated: Yes.
- Committed: Should be treated as build output; do not add source code here.

---

*Structure analysis: 2026-06-16*
