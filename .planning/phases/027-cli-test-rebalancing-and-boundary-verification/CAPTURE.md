# Capture: Phase 27 CLI Test Rebalancing and Boundary Verification

## Final Package Boundary

- `internal/cli`: Cobra command tree, flags, args, completions, terminal prompts, stdout/stderr routing, error wording, child process execution, HTTP server lifecycle, signal handling, and representative command smoke coverage.
- `internal/app`: reusable command orchestration across config, vault, export formatting, setup, migration, and manager helper primitives. App services return strings/results/errors and do not import Cobra.
- `internal/project`: project manifest model/IO/validation, project selector entry construction, project binding resolution, diagnostics, child env merge, and parent env override warnings.
- `internal/secret`: secret edit workflow, editable JSON representation, and temp file lifecycle. Interactive prompt handling remains CLI-owned.
- `internal/vault`: encrypted vault file lifecycle, age encryption, store mutation primitives, path/env/value validation, locking, atomic writes, status, and doctor checks.
- `internal/exportfmt`: env/shell/JSON formatting and value/env-name conversion.
- `internal/manager`: local Web manager server/API/UI behavior.

## Final Test Ownership Model

- CLI tests should cover command wiring, flags, completions, stdout/stderr contracts, user-facing error wording, no-leak command output, interactive prompts, child process/server lifecycle, and a small number of end-to-end smoke workflows.
- App tests should cover cross-package orchestration rules directly without Cobra.
- Project tests should cover manifest, selector, resolution, diagnostics, env conflict, and project session environment rules directly.
- Vault tests should cover persistence, encryption, locking, validation, status, and store primitives.
- Secret tests should cover edit/temp workflows directly; CLI prompt tests stay in CLI.

## Follow-On Work

- If developer docs need non-planning architecture documentation, copy this package/test ownership model into the developer architecture docs during a docs phase.
- Keep future CLI tests short and contract-focused; add behavior-rule coverage beside the owning package first.

## Documentation Impact

- No user-facing documentation update is required; command behavior did not change.
- This capture is the durable planning reference for the final boundary model.
