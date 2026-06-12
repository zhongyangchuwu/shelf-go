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


## Completed since v0.1.0

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


### Project manifest (`project init` / `project explain`) — done

Implemented for v0.2 scope:

- `.shelf.json` manifest in Git root.
- `shelf project init` with `--force` overwrite behavior.
- `shelf project explain` read-only status output with `ok` / `warn` / `fail`.
- Env resolution order: project entry `env` → secret `env` → derived path env.
- Required missing secret → `fail`; optional missing secret → `warn`.
- Env name conflict detection with non-zero exit.

## Next candidates

The next release should focus on v0.4 runtime injection (`shelf run --`) with direct child process env injection. Rejected ideas:
- `export --output`: Unix redirection already handles this well: `shelf export ai --format env > .env.local`.
- `.env` as project config: `.env` keys do not reliably encode Shelf's `group_path:key` identity, and `.env` invites users to paste real secret values into project files. Shelf uses `.shelf.json` instead.
- `.env` import: same identity-encoding problem. Project manifest with `shelf run` is the correct direction.

## v0.2.0 — Project manifest and explain

Goal: let Shelf read what secrets a project needs, without injecting or exporting anything.

### `.shelf.json` project manifest

Project bindings live in `<git-root>/.shelf.json`, a file that declares which Shelf secret paths the project needs. It is a manifest of intent, not a secret store.

Why not `.env`:

- `.env` is a key-value file for environment variables. It cannot reliably encode Shelf's `group_path:key` identity.
- `.env` invites users to paste real secret values into project files, contradicting Shelf's goal of keeping secrets out of projects.
- `.env` cannot express include/exclude/required/collision strategies or profiles.

`.shelf.json` can be committed to Git; `.env.local` (the generated output) must be gitignored.

**v0.2 minimal format:**

```json
{
  "version": 1,
  "secrets": [
    {
      "path": "providers/openai/accounts/personal:api_key",
      "env": "OPENAI_API_KEY",
      "required": true
    },
    {
      "path": "providers/openrouter/accounts/personal:api_key",
      "env": "OPENROUTER_API_KEY",
      "required": true
    }
  ]
}
```

Field rules:

- `version`: required, fixed at `1`.
- `secrets`: required array.
- `path`: exact Shelf secret path (`group_path:key`).
- `env`: optional project-level override for the environment variable name.
- `required`: optional, defaults to `true`.
- `prefix` and `profiles` are deferred to v0.3+.

Fields **prohibited** from `.shelf.json`:

- `value` / resolved values.
- Fallback plaintext secrets.
- Shell commands or template expressions.

### `shelf project init`

Writes a minimal `.shelf.json` in the Git root.

```bash
shelf project init
```

Behavior:

- Must run inside a Git worktree.
- Writes `<git-root>/.shelf.json`.
- Fails if `.shelf.json` already exists unless `--force` is given.
- Writes no secret values.

### `shelf project explain`

Read-only explanation of the current project's secret requirements.

```bash
shelf project explain
```

Example output:

```text
project: github.com/alex/my-app
root:    /Users/alex/code/my-app
config:  .shelf.json

ok   providers/openai/accounts/personal:api_key -> OPENAI_API_KEY
ok   providers/openrouter/accounts/personal:api_key -> OPENROUTER_API_KEY
warn providers/anthropic/accounts/personal:api_key missing optional
fail providers/deepseek/accounts/personal:api_key missing required
```

Rules:

- Does **not** print secret values.
- Does **not** execute commands.
- Does **not** generate or modify files.
- Validates `.shelf.json` format, secret path existence, env name uniqueness, and required/optional coverage.
- `ok`/`warn`/`fail` per entry; `fail` exits non-zero.

Env name resolution order:

1. `.shelf.json` entry `env` (project-level override).
2. Secret object's `env` field.
3. Derived from full secret path (uppercase, non-alphanumeric → `_`).

Conflict: two entries resolving to the same env name → `fail`.

**Acceptance criteria:**

- Not in Git repo → fail.
- No `.shelf.json` → clear prompt.
- Invalid JSON → fail.
- `version` ≠ 1 → fail.
- Invalid `path` format → fail.
- Invalid `env` name → fail.
- Required secret missing → fail.
- Optional secret missing → warn.
- Duplicate env name → fail.
- All secrets present, no conflicts → ok.

## v0.3.0 — Project binding management and export

Goal: let users manage `.shelf.json` without hand-editing JSON, and export environment variables from the project manifest.

### New commands

```bash
shelf project add <path-or-prefix> [--env NAME] [--optional]
shelf project rm <path-or-prefix>
shelf project list
shelf project export --format env|shell|json
```

### `shelf project add`

Add a secret path or prefix to `.shelf.json`.

```bash
shelf project add providers/openai/accounts/personal:api_key --env OPENAI_API_KEY
shelf project add providers/openrouter/accounts/personal --optional
```

Behavior:

- If `.shelf.json` does not exist, prompts to run `shelf project init` first.
- `path`: validates the secret exists in the store. Fails if not found.
- `prefix`: validates at least one secret matches. Fails if zero matches.
- Rejects duplicate entries (same path or prefix already present).
- Appends entry and preserves JSON formatting.

### `shelf project rm`

Remove an entry from `.shelf.json`.

```bash
shelf project rm providers/openai/accounts/personal:api_key
```

### `shelf project list`

List entries in `.shelf.json` without resolving secrets.

```text
path   providers/openai/accounts/personal:api_key -> OPENAI_API_KEY (required)
prefix providers/openrouter/accounts/personal (optional)
```

Does not print secret values.

### `.shelf.json` v0.3 format

Adds `prefix` entries:

```json
{
  "version": 1,
  "secrets": [
    {
      "path": "providers/openai/accounts/personal:api_key",
      "env": "OPENAI_API_KEY",
      "required": true
    },
    {
      "prefix": "providers/openrouter/accounts/personal",
      "required": false
    }
  ]
}
```

Prefix rules:

- `prefix` matches all secrets whose canonical path starts with the given string.
- `path` and `prefix` are mutually exclusive within one entry.
- Prefix entries should not carry `env` because a prefix may expand to multiple secrets with different env names.
- `profiles` are deferred to v0.5.

### `shelf project export`

Export environment variables from the project manifest.

```bash
shelf project export --format env
shelf project export --format shell
shelf project export --format json
```

Distinction from `shelf export`:

| Command | Source | Awareness |
| --- | --- | --- |
| `shelf export <path-or-prefix>` | explicit path/prefix argument | direct |
| `shelf project export` | `.shelf.json` in Git root | project-aware |

Reuses existing value conversion and shell quoting from `shelf export`.

Behavior:

- Resolves all `path` and `prefix` entries from `.shelf.json`.
- Expands prefix entries into matching secrets (stable sort).
- Resolves env names (project override → secret `env` → derived).
- On env name conflict: fail with clear message.
- On required secret missing: fail.
- On optional secret missing: skip with warning to stderr.

Typical usage:

```bash
shelf project export --format env > .env.local
```

`.env.local` is a generated artifact, not a configuration source. It is never committed to Git.

**Acceptance criteria:**

- `project add` appends `path` entries.
- `project add` appends `prefix` entries.
- `project add` rejects duplicates.
- `project rm` removes entries.
- `project list` shows entries without values.
- `project export env` outputs `KEY=value`.
- `project export shell` is safe for `eval`.
- `project export json` outputs a `{"KEY":"value"}` object.
- Prefix expansion produces stable sorted output.
- Env name conflicts fail export.
- Required missing fails export.
- Optional missing skips with warning.

## v0.4.0 — Runtime injection: `shelf run`

Goal: run commands with secrets injected directly into the child process, without writing files.

```bash
shelf run -- command args...
shelf run --dry-run -- command args...
```

`--dry-run` prints which env vars would be injected (without values) and does not execute the command.

Examples:

```bash
shelf run -- npm run dev
shelf run -- python app.py
shelf run -- go test ./...
```

### Runtime flow

1. Find Git root.
2. Load `<git-root>/.shelf.json`.
3. Resolve `path`/`prefix` entries.
4. Read secret values from store.
5. Compute env names.
6. Check required missing → fail.
7. Check env name conflicts → fail.
8. Construct child process environment (Shelf env overrides parent env).
9. Execute command.
10. Return child process exit code.

### Key principles

- Only injects into the child process.
- Does **not** mutate the parent shell.
- Does **not** write `.env.local` (use `shelf project export` for that).
- Does **not** print secret values.
- If resolution fails, the command never executes.
- If the command fails, `shelf run` returns its exit code.

### Env override strategy

Shelf-resolved env vars override existing parent environment variables. Users run `shelf run` to explicitly activate the project manifest; stale env vars in the parent shell should not silently win.

`explain` and `--dry-run` warn when Shelf would override an existing env var:

```text
warn OPENAI_API_KEY overrides existing environment variable
```

A future `--preserve-existing` flag is deferred.

### What v0.4 does not do

- Shell hooks (`eval "$(shelf hook zsh)"`).
- Auto-injection on `cd`.
- Profiles.
- `.env` import.
- Automatic `.env.local` generation.

## Later / profiles

After the single-profile workflow is stable, add `profiles` to `.shelf.json`:

```json
{
  "version": 1,
  "profiles": {
    "default": {
      "secrets": [...]
    },
    "work": {
      "secrets": [...]
    }
  }
}
```

Commands: `shelf project add --profile work ...`, `shelf run --profile work -- ...`.

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
