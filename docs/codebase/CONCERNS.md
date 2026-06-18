# Codebase Concerns

**Analysis Date:** 2026-06-16

## Tech Debt

**Plaintext Store Boundary:**
- Status: Phase 1 encrypted vault core now routes CLI runtime through `store.Vault`; encrypted vault writes and `.bak` backups are covered by tests.
- Remaining risk: legacy plaintext helpers still exist for internal/plaintext boundaries and migration work, and `secret edit` still writes a plaintext editor temp file.
- Files: `internal/store/vault.go`, `internal/store/io.go`, `internal/cli/root.go`, `internal/cli/secret.go`, `docs/data-spec.md`
- Follow-up: Phase 2 should provide plaintext migration and git/chezmoi safety classification instead of treating any encrypted vault path as fully safe.

**Store Write Responsibilities Are Split Across Helpers:**
- Status: Phase 1 split encrypted vault persistence into `store.Vault`, `encodeStore`, `writeStoreFile`, and plaintext model helpers.
- Remaining risk: backup retention, restore commands, platform-specific durability, and migration still converge around the same persistence helpers.
- Files: `internal/store/io.go`, `internal/store/vault.go`
- Follow-up: Keep future migration and restore behavior at the store/vault boundary instead of adding command-local file writes.

**Large CLI Command Files:**
- Issue: Command construction, command behavior, project resolution, completion, and rendering glue are concentrated in large files.
- Files: `internal/cli/secret.go`, `internal/cli/project.go`
- Impact: Small command changes can accidentally affect completions, editor behavior, project export, or diagnostics because command wiring and business logic live together.
- Fix approach: Move pure behavior into smaller helpers that can be unit-tested without Cobra command setup. Keep command files responsible for flags, args, I/O streams, and invoking helpers.

**String Prefix Matching for Secret Groups:**
- Issue: Secret listing and project prefix expansion use raw `strings.HasPrefix`.
- Files: `internal/store/io.go:154`, `internal/store/io.go:157`, `internal/cli/project.go:351`, `internal/cli/project.go:353`
- Impact: A prefix such as `app` also matches `app2:token`, `apple:key`, and `app-prod:key`; project manifests can export broader sets than a user expects.
- Fix approach: Add a store-level group-prefix matcher that treats prefixes as group path boundaries. Match exact group `app` and descendants like `app/api:token`, but not sibling names with the same byte prefix.

## Known Bugs

**Project Prefixes Can Over-Match Unrelated Secret Groups:**
- Symptoms: `shelf project add app` or a manifest entry with `"prefix":"app"` includes any path where the full secret path starts with `app`, including `app2:token`.
- Files: `internal/store/io.go:154`, `internal/store/io.go:157`, `internal/cli/project.go:353`, `internal/cli/project.go:363`
- Trigger: Store both `app:token` and `app2:token`, then add/export prefix `app`.
- Workaround: Use explicit path entries in `.shelf.json` instead of broad prefixes when group names share leading characters.

**Plaintext Store Files With Multiple JSON Values Are Accepted:**
- Symptoms: plaintext `store.Load` decodes one JSON object with `dec.Decode(&data)` and does not verify EOF after the first value.
- Files: `internal/store/io.go:31`, `internal/store/io.go:34`
- Trigger: A legacy plaintext store containing a valid store object followed by another JSON token can load successfully while trailing data is ignored.
- Workaround: Use encrypted vault mode for CLI workflows; migration hardening should strict-decode any plaintext source before encrypting it.

## Security Considerations

**Plaintext Secrets at Rest:**
- Risk: legacy plaintext stores and editor temp files can still contain secret values outside the encrypted vault boundary.
- Files: `internal/store/io.go`, `internal/store/vault.go`, `internal/cli/secret.go`, `docs/data-spec.md`
- Current mitigation: CLI runtime writes the active vault through `store.Vault`, encrypted temp files, encrypted replacement backups, and user-only file modes.
- Recommendations: Add the migration flow and git safety checks before declaring synced repositories fully safe; harden `secret edit` temp-file handling separately.

**Editor Temp File Contains Secret Values:**
- Risk: `secret edit` writes the complete secret object, including `value`, to an OS temp file outside the private Shelf data directory.
- Files: `internal/cli/secret.go:375`, `internal/cli/secret.go:379`, `internal/cli/secret.go:384`, `internal/cli/secret.go:401`
- Current mitigation: The temp file is removed after editing, and `os.CreateTemp` creates a private file on normal Unix platforms.
- Recommendations: Create edit temp files in a `0700` Shelf-owned temp directory, explicitly chmod to `0600`, document editor swap/backup risks, and consider an edit mode that excludes `value` unless requested.

**Shell-Based Editor Launch Is More Complex Than Needed:**
- Risk: `secret edit` runs `sh -c "$SHELF_EDITOR \"$SHELF_EDIT_FILE\""`.
- Files: `internal/cli/secret.go:392`, `internal/cli/secret.go:393`, `internal/cli/secret.go:394`
- Current mitigation: Editor and edit file are passed through environment variables, which avoids simple command interpolation of the temp path.
- Recommendations: Prefer a small parser for editor commands or direct `exec.Command` when `runtime.Editor` is a binary path. Preserve support for editor arguments without making shell evaluation the default execution path.

**Export Formats Print Secret Values by Design:**
- Risk: `export`, `project export`, and `secret get` write plaintext secrets to stdout, which can be captured in shell history wrappers, logs, terminal scrollback, or process pipelines.
- Files: `internal/render/export.go:92`, `internal/render/export.go:100`, `internal/render/export.go:108`, `internal/cli/secret.go:283`, `internal/cli/project.go:411`
- Current mitigation: Metadata-oriented commands avoid printing values, and `run --dry-run` prints only env names.
- Recommendations: Keep value-printing commands explicit. Add documentation warnings and tests that non-value commands never regress into printing values.

## Performance Bottlenecks

**Whole-Vault Load and Save:**
- Problem: Every read command decrypts and decodes the whole vault, and every write command rewrites the whole encrypted vault.
- Files: `internal/store/vault.go`, `internal/store/io.go`
- Cause: The vault wraps a single JSON model with an in-memory map.
- Improvement path: Keep this design while data remains small. If large stores become common, add indexing or a backend abstraction under `internal/store` before optimizing CLI code.

**Full Prefix Scans:**
- Problem: Prefix listing scans every secret path and sorts the result.
- Files: `internal/store/io.go:154`, `internal/store/io.go:161`, `internal/cli/project.go:353`
- Cause: Secret paths are stored in a flat map with no prefix index.
- Improvement path: Keep the scan for MVP scale. If project manifests expand many prefixes over large stores, add a grouped index built during `Load`.

**Project Resolution Re-Reads Store and Resolves All Entries:**
- Problem: Project commands load the full store and resolve all manifest entries before rendering output.
- Files: `internal/cli/project.go:303`, `internal/cli/project.go:346`, `internal/cli/project.go:351`
- Cause: Resolution is intentionally simple and validates conflicts across the full manifest.
- Improvement path: Keep all-entry resolution for correctness. Optimize only after benchmarks show large manifests are slow.

## Fragile Areas

**Atomic Save Durability:**
- Files: `internal/store/io.go:87`, `internal/store/io.go:101`, `internal/store/io.go:105`, `internal/store/io.go:108`
- Why fragile: The temp file is synced, but the parent directory is not synced after rename. On some filesystems, a crash after rename can still risk directory-entry durability.
- Safe modification: Add a platform-aware helper that syncs the parent directory after `os.Rename`. Keep existing temp-file and `0600` behavior.
- Test coverage: `internal/store/io_test.go` covers load/save basics, but crash durability and directory fsync behavior are not test-covered.

**Advisory Locking Is Unix-Specific:**
- Files: `internal/store/lock.go:6`, `internal/store/lock.go:13`, `internal/store/lock.go:21`, `internal/cli/root.go:44`
- Why fragile: Locking uses `syscall.Flock`, which is Unix-oriented and advisory. Cross-platform behavior and network filesystem behavior depend on OS and filesystem support.
- Safe modification: Hide locking behind a platform-specific implementation with build tags and clear semantics per OS.
- Test coverage: There are no direct lock contention tests for `internal/store/lock.go`.

**Project Resolution Combines Validation and Materialization:**
- Files: `internal/cli/project.go:346`, `internal/cli/project.go:388`
- Why fragile: `resolveProjectEntries` emits both diagnostics and resolved values, while `appendResolvedEntry` owns env-name derivation, duplicate detection, and value conversion.
- Safe modification: Keep a pure resolver, but return a typed result that separates failures, warnings, and exportable bindings. Add focused tests around duplicate env names, optional entries, prefix entries, and value conversion.
- Test coverage: `internal/cli/project_test.go` covers many CLI scenarios, but resolver behavior is mostly tested through command-level flows.

**Rendering Handles Shell, Env, and JSON Export Semantics Centrally:**
- Files: `internal/render/export.go:31`, `internal/render/export.go:92`, `internal/render/export.go:100`, `internal/render/export.go:108`, `internal/render/export.go:135`
- Why fragile: A quoting or value-conversion regression can leak into `export`, `project export`, and `run` environment construction.
- Safe modification: Add direct unit tests in `internal/render` for shell quoting, newline handling, empty strings, JSON values, null conversion, and duplicate JSON env names.
- Test coverage: No `internal/render/*_test.go` file is present.

## Scaling Limits

**Single JSON Store File:**
- Current capacity: Best suited for small local stores, matching `docs/data-spec.md` guidance.
- Limit: Large stores make every command pay full JSON decode cost; every write rewrites and backs up the whole file.
- Scaling path: Introduce a backend interface below `internal/store` if secret counts grow beyond the expected small local usage.

**Single Global Write Lock Per Vault File:**
- Current capacity: One writer at a time per vault path.
- Limit: Long-running write flows like `secret edit` hold the write lock while the editor is open.
- Scaling path: For edit flows, consider loading under lock, releasing while editing a copy, then reacquiring and applying an optimistic update with conflict detection before save.

## Dependencies at Risk

**`syscall.Flock`:**
- Risk: Platform-specific locking behavior is embedded directly in `internal/store/lock.go`.
- Impact: Portability issues for non-Unix platforms and ambiguous behavior on some shared/network filesystems.
- Migration plan: Use build-tagged lock implementations or a maintained cross-platform file lock package with explicit behavior.

**`gopkg.in/yaml.v3`:**
- Risk: Config parsing currently uses permissive YAML unmarshalling without strict unknown-field rejection.
- Impact: Typos in `config.yaml` can be silently ignored, leading to unexpected default vault paths or editor behavior.
- Migration plan: Decode config with `KnownFields(true)` or equivalent strict parsing in `internal/config/config.go`.

## Missing Critical Features

**Migration and Git-Safety Flow:**
- Problem: encrypted vault mode exists, but there is no command that migrates a legacy plaintext store or proves a git/chezmoi path is safe.
- Blocks: Safe onboarding from existing legacy plaintext store files and reliable guidance for synced repositories.

**Store Migration Framework:**
- Problem: `store.Load` rejects unsupported versions but there is no migration path.
- Blocks: Changing the on-disk data model without a manual migration command.

**Backup Retention and Recovery Commands:**
- Problem: Each save overwrites one `.bak` file, and there is no command to inspect or restore backups.
- Blocks: Recovering from accidental deletion or malformed edits beyond the most recent backup.

## Test Coverage Gaps

**Render Package Unit Tests:**
- What's not tested: Shell quoting, env rendering, JSON rendering, env-name derivation, null handling, and structured value conversion.
- Files: `internal/render/export.go`
- Risk: Export format regressions can expose malformed shell snippets or incorrect runtime environment values.
- Priority: High

**Locking and Concurrent Writes:**
- What's tested: goroutine-level `secret set` contention against one `--vault` path keeps all writes.
- What's not tested: separate OS processes contending for the same `<vault-path>.lock`.
- Risk: cross-process locking regressions can lose secrets or corrupt encrypted backups.
- Priority: High

**Prefix Boundary Matching:**
- What's not tested: Distinguishing `app` from `app2`, `app-prod`, and nested `app/api` groups.
- Files: `internal/store/io.go`, `internal/cli/project.go`, `internal/cli/project_test.go`
- Risk: Project export and run can inject unintended secrets.
- Priority: High

**Strict Config Parsing:**
- What's not tested: Unknown YAML keys and misspelled config fields.
- Files: `internal/config/config.go`
- Risk: Users can believe a config setting is active while the CLI silently ignores it.
- Priority: Medium

**Editor Temp File Permissions and Cleanup Failures:**
- What's not tested: Explicit temp file mode, editor failure cleanup, and behavior when removal fails.
- Files: `internal/cli/secret.go`
- Risk: Secret values can remain in temp locations or editor-created side files.
- Priority: Medium

**Store Decode Trailing Data:**
- What's not tested: Rejecting trailing non-whitespace JSON after the store object.
- Files: `internal/store/io.go`, `internal/store/io_test.go`
- Risk: Corrupt or hand-edited files can be accepted partially.
- Priority: Medium

---

*Concerns audit: 2026-06-16*
