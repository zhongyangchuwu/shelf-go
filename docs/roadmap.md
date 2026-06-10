# Shelf-Go Roadmap

This roadmap is a working priority list for the Go rewrite. It is not a release promise.

Shelf-Go should keep the useful ideas from the Python Shelf without copying its command surface or tree-shaped store. The Go version stays focused on fast local secret management, direct export, and small Unix-like commands.

## Current baseline

Released as `v0.1.0`:

- Flat JSON store with `version: 1`.
- First-class secret objects stored under canonical `group_path:key` map keys.
- Code-level secret identity split into `group_path` and `key`.
- `secret set/get/list/info/edit/rm`.
- `secret edit` can modify `group_path` and `key`, which renames the secret.
- `export <path-or-prefix> --format shell|env|json`.
- `export` defaults to secrets with explicit `env`; `--all` includes derived env names.
- `init`.
- `completion`.
- `project id`.
- `--version` based on Go build info and git tags.

## Product rules

- Extract concepts from the Python version, not command names.
- Keep one canonical command for each job.
- Do not add aliases for migration comfort.
- Do not ship placeholder project commands.
- Do not restore group objects, group metadata, or generic `meta set/show/rm`.
- Keep `secret edit` as the unified identity + metadata edit path unless repeated scripted usage proves a separate command is needed.
- Keep the store file simple: `map[canonical_path]Secret`.
- Prefer explicit commands and readable JSON over hidden automation.
- Treat writes as manual management operations; scripts should use `get` or `export`, not mutate the store.


## v0.2.0 — Reliability and visibility

### File locking — done

Write-side locking is implemented for mutating commands.

Covered commands:

- `init --force`
- `secret set`
- `secret edit`
- `secret rm`

Behavior:

- Lock path is `<data-path>.lock`.
- Mutating commands lock before loading the store, so they mutate the latest on-disk data.
- Save still uses backup + temp file + rename.
- Concurrent `secret set` operations preserve all writes instead of overwriting stale snapshots.

### `doctor` — done

`shelf doctor` checks local Shelf configuration and data health.

Implemented checks:

- Config resolves.
- Installed binary reports version.
- Data file exists or would be created at the expected path.
- Store JSON loads and validates.
- Data file mode is not too permissive; prefer `0600`.
- Data file appears or does not appear to be tracked by ordinary git in plaintext.
- Completion file exists in the directories listed by `FPATH` / `fpath` and ends with `/_shelf`.

Output uses `ok`, `warn`, and `fail`; `fail` exits non-zero.

The first `doctor` is intentionally not coupled to chezmoi, age, or external secret managers.

## Next candidates

There are no write-path convenience features planned right now.

Rejected ideas:

- `secret set --stdin --raw`: `set` is a low-frequency management command. Scripts should not mutate Shelf as part of normal execution.
- `export --output`: Unix redirection already handles this well: `shelf export ai --format env > .env.local`.
- `.env` import: `.env` keys do not reliably encode Shelf's `group_path:key` identity, so import would need arbitrary naming rules.

The next feature should either improve safety/diagnostics or prove the Git-aware project model without adding write-side convenience APIs.

## Future — Git-aware project flow experiment

The old Go rewrite idea card had a strong product direction: Shelf can autofill Git projects the way password managers autofill websites. This is promising, but should be introduced through read-only explanation first.

### `project explain`

Start with a read-only command:

```bash
shelf project explain
```

Purpose:

- Show current project identity.
- Show whether any project binding exists.
- Show which secret paths would provide which env vars.
- Show missing bindings or missing secrets.

Do not inject anything in `explain`.

### `project bind`

Only after `project explain` proves useful, add explicit binding:

```bash
shelf project bind <path-or-prefix>
shelf project unbind <path-or-prefix>
```

Preferred storage:

```text
~/.config/shelf/projects.json
```

Avoid writing personal secret bindings into `.git/config` unless there is a strong reason later.

### `project run`

After bindings are stable:

```bash
shelf project run -- command args...
```

Rules:

- Resolve current Git project.
- Resolve bindings.
- Inject env into the child process only.
- Do not mutate shell parent environment.
- Provide clear errors for missing secrets or conflicting env names.

Defer shell hooks until project commands are proven. Hooks are expensive to debug and can hurt shell startup performance.

## Later / careful design only

### Built-in encryption

Potentially useful, but not before the file-locking, backup, migration, and doctor story is solid.

If added, prefer whole-file encryption over field-level encryption.

Open questions:

- Age only, or pluggable backend?
- Is plaintext-on-disk allowed for encrypted stores?
- What happens to `.bak` files?
- How are migrations performed?

### Template injection

The old `inject` command was useful, but can turn Shelf into a template engine.

If restored, keep it narrow:

```text
shelf://group/path:key
```

Avoid custom template syntax, conditionals, loops, or project-aware magic in the first version.

### External secret-manager integrations

Possible later integrations:

- 1Password CLI
- Bitwarden CLI
- Vault
- Doppler

Keep out of scope until the local workflow is stable.

## Explicit non-goals for now

- Reintroducing tree-shaped group objects.
- Group metadata.
- Generic `meta` commands.
- `mv` as a separate command; `secret edit` already supports identity changes.
- TUI browser.
- Remote sync service.
- Multi-user permissions.
- Audit log or value history.
- SQLite backend.
- Automatically exporting secrets from shell startup without explicit user configuration.
