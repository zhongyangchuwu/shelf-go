# Codebase Concerns

**Analysis Date:** 2026-06-16

## Tech Debt

**Plaintext Store Boundary:**
- Issue: Secrets are persisted as plaintext JSON through `json.MarshalIndent` and `os.Rename`; this is an intentional MVP policy, but it is still the primary security debt for a secret manager.
- Files: `internal/store/io.go:57`, `internal/store/io.go:72`, `docs/data-spec.md`
- Impact: Anyone with filesystem access, backup access, synced-folder access, or ordinary Git access to `secrets.json` or `secrets.json.bak` can read every secret value.
- Fix approach: Add an encryption boundary inside `internal/store/io.go` so command code continues to operate on `store.Data` after decrypt/load and before encrypt/save. Keep file permissions and locking as defense-in-depth, not as the only protection.

**Store Write Responsibilities Are Concentrated:**
- Issue: `Store.Save` validates data, creates backups, writes temp files, fsyncs, closes, and renames in one function.
- Files: `internal/store/io.go:57`, `internal/store/io.go:77`, `internal/store/io.go:80`, `internal/store/io.go:87`, `internal/store/io.go:101`, `internal/store/io.go:108`
- Impact: Adding encryption, migrations, backup rotation, or platform-specific durability will make one already central method harder to reason about.
- Fix approach: Split `Store.Save` into small helpers for validation, serialization, backup, atomic write, and optional encryption. Keep `Save` as the orchestration point used by CLI write commands.

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

**Store Files With Multiple JSON Values Are Accepted:**
- Symptoms: `store.Load` decodes one JSON object with `dec.Decode(&data)` and does not verify EOF after the first value.
- Files: `internal/store/io.go:31`, `internal/store/io.go:34`
- Trigger: A `secrets.json` containing a valid store object followed by another JSON token can load successfully while trailing data is ignored.
- Workaround: Do not manually edit `secrets.json`; use `shelf doctor` and CLI commands for normal operations.

## Security Considerations

**Plaintext Secrets at Rest:**
- Risk: The application protects file mode on writes but does not encrypt secret values.
- Files: `internal/store/io.go:72`, `internal/store/io.go:93`, `internal/store/io.go:193`, `docs/data-spec.md`
- Current mitigation: Data directory creation uses `0700`, store temp files and backups use `0600`, and `doctor` warns when the data file mode is broader than user-only.
- Recommendations: Add encryption before the first non-local or sync-heavy use case. Include backup encryption because `secrets.json.bak` is created next to the primary store.

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

**Whole-File Store Load and Save:**
- Problem: Every read command loads the whole JSON store, and every write command rewrites the whole store.
- Files: `internal/store/io.go:20`, `internal/store/io.go:57`, `internal/store/io.go:72`, `internal/store/io.go:108`
- Cause: The MVP store is a single JSON file with an in-memory map.
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

**Single Global Write Lock Per Data File:**
- Current capacity: One writer at a time per data path.
- Limit: Long-running write flows like `secret edit` hold the write lock while the editor is open.
- Scaling path: For edit flows, consider loading under lock, releasing while editing a copy, then reacquiring and applying an optimistic update with conflict detection before save.

## Dependencies at Risk

**`syscall.Flock`:**
- Risk: Platform-specific locking behavior is embedded directly in `internal/store/lock.go`.
- Impact: Portability issues for non-Unix platforms and ambiguous behavior on some shared/network filesystems.
- Migration plan: Use build-tagged lock implementations or a maintained cross-platform file lock package with explicit behavior.

**`gopkg.in/yaml.v3`:**
- Risk: Config parsing currently uses permissive YAML unmarshalling without strict unknown-field rejection.
- Impact: Typos in `config.yaml` can be silently ignored, leading to unexpected default data paths or editor behavior.
- Migration plan: Decode config with `KnownFields(true)` or equivalent strict parsing in `internal/config/config.go`.

## Missing Critical Features

**Encrypted Backend:**
- Problem: The project is a secret manager whose MVP store is plaintext.
- Blocks: Safer use on synced directories, shared workstations, unmanaged backups, and repositories with accidental data-file tracking.

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
- What's not tested: Multiple processes or goroutines contending for the same `secrets.json.lock`.
- Files: `internal/store/lock.go`, `internal/cli/root.go`
- Risk: Concurrent write regressions can lose secrets or corrupt backups.
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
