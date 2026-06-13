# Shelf Go Rewrite Usage Spec

## Purpose

Shelf-Go is not a line-by-line port of the Python CLI. It keeps the useful ideas from the MVP and reshapes them into a smaller, faster local secret manager.

MVP product focus:

```text
Fast local secret management and direct export.
```

Git-aware project workflows are the next priority after the secret manager foundation is stable. The release progression is:

- v0.2: `.shelf.json` project manifest, `shelf project init`, and `shelf project explain` (implemented).
- v0.3: `shelf project add/rm/list/export` (management and materialization).
- v0.4: `shelf run` (runtime injection into child process).

The command surface below is implemented through v0.4. Later / profiles remains future work.

## Design rules

- Prefer one canonical command for each purpose.
- Do not add aliases for migration comfort.
- Do not ship placeholder commands.
- A command ships only when it has complete behavior.
- Keep command behavior small, explicit, and scriptable.
- Secret metadata belongs to the secret object, not to a group node.
- Group/path is identity; tags are auxiliary labels.

## Current command surface

```bash
shelf secret add [path-or-group]
shelf secret set <path> <value> [--env NAME] [--description TEXT] [--tag TAG ...] [--force]
shelf secret get <path>
shelf secret list [prefix]
shelf secret info <path>
shelf secret edit <path>

shelf export <path-or-prefix> --format shell|env|json

shelf doctor

shelf project id
shelf project init
shelf project explain
shelf project add <path-or-prefix> [--env NAME] [--optional]
shelf project rm <path-or-prefix>
shelf project list
shelf project export --format env|shell|json

shelf run -- command args...
shelf run --dry-run -- command args...
```

## Secret paths

Secret paths have the form `group_path:key`.

`group_path` is the namespace part before the single colon; `key` is the leaf name after it.

`group_path` may contain `/` separators. `key` must stay a single leaf token.


```bash
shelf secret set providers/openrouter/accounts/personal:api_key sk-xxx \
  --env OPENROUTER_API_KEY \
  --description "Personal OpenRouter key" \
  --tag ai \
  --tag openrouter \
  --tag personal
```

Rules:

- `<value>` is a required command-line argument in MVP.
- No prompt input in MVP.
- No stdin input in MVP.
- The value is JSON-parsed when possible.
- If the secret already exists, `set` fails by default.
- `--force` replaces the existing secret object.
- `--env`, `--description`, and `--tag` set fields on the same secret object.
- There is no separate generic metadata system.

Examples:

```bash
shelf secret set app:port 34222
shelf secret set flags:enabled true
shelf secret set app:options '{"debug":false}'
shelf secret set providers/openrouter/accounts/personal:api_key sk-xxx --env OPENROUTER_API_KEY
```


### `shelf secret add`

```bash
shelf secret add
shelf secret add providers/openai/accounts/personal:api_key
shelf secret add providers/openai/accounts/personal
```

Interactively creates a secret for human use. `secret set` remains the scriptable non-interactive write path.

Behavior:

- Requires a terminal; non-interactive scripts should use `secret set`.
- Shows existing group paths as lightweight hints before prompting.
- With no argument, prompts for full secret path.
- With a full `group:key` path, prompts only for secret fields.
- With a group path argument, prompts for `key` and stores `group:key`.
- Prompts for value using hidden input.
- Prompts for optional env, description, and comma-separated tags.
- Existing paths are not overwritten unless the user confirms.
- Does not print secret values.

This does not add group objects or group metadata; group remains the path prefix used to organize keys.

## `shelf secret get`

Print a single secret value.

```bash
shelf secret get providers/openrouter/accounts/personal:api_key
```

Output:

```text
sk-xxx
```

Rules:

- `get` reveals the value by design.
- It does not mask output.
- It does not print metadata.
- It is the script-friendly value lookup command.

## `shelf secret list`

List secret paths only.

```bash
shelf secret list
```

Output:

```text
providers/deepseek/accounts/personal:api_key
providers/openrouter/accounts/personal:api_key
```

Prefix filter:

```bash
shelf secret list providers/openrouter
```

Output:

```text
providers/openrouter/accounts/personal:api_key
```

Rules:

- Does not print values.
- Does not print metadata.
- Sorts paths lexicographically.
- `[prefix]` is a path-prefix filter.
- There is no separate group object in MVP.

## `shelf secret info`

Print non-secret information about one secret as JSON.

```bash
shelf secret info providers/openrouter/accounts/personal:api_key
```

Output:

```json
{
  "path": "providers/openrouter/accounts/personal:api_key",
  "group_path": "providers/openrouter/accounts/personal",
  "key": "api_key",
  "value_set": true,
  "env": "OPENROUTER_API_KEY",
  "description": "Personal OpenRouter key",
  "tags": ["ai", "openrouter", "personal"]
}
```

Rules:

- Default output is JSON.
- The real value is not printed.
- `value_set` reports whether a value exists.
- `info` is the metadata read command.
- There is no `secret show` command in MVP.

## `shelf secret edit`

Edit a complete secret object in `$EDITOR`.

```bash
shelf secret edit providers/openrouter/accounts/personal:api_key
```


Editor buffer:

```json
{
  "group_path": "providers/openrouter/accounts/personal",
  "key": "api_key",
  "value": "sk-xxx",
  "env": "OPENROUTER_API_KEY",
  "description": "Personal OpenRouter key",
  "tags": ["ai", "openrouter", "personal"]
}
```

Rules:

- Opens the user's editor.
- Edits the full secret record.
- The edit buffer is JSON.
- `group_path` and `key` are the editable identity fields; together they form the canonical path `group_path:key`.
- Changing `group_path` or `key` renames the secret.
- Rename fails if the destination path already exists.
- The edited object is validated before writing.
- Invalid edits are not written.
- `edit` is the unified metadata and identity modification path.
- MVP does not include field-specific commands like `secret tag add`, `secret tag rm`, or `secret env set`.

## Group/path and tags

Core rule:

```text
group/path = canonical secret identity
tags       = optional cross-cutting labels
```

Or:

```text
The path tells where the secret lives.
The tags tell what the secret is about.
```

Example path:

```text
providers/openrouter/accounts/personal:api_key
```

The namespace portion is:

```text
providers/openrouter/accounts/personal
```

Example tags:

```json
{
  "tags": ["ai", "personal"]
}
```

Rules:

- The path is the unique ID.
- Tags do not participate in uniqueness.
- Tags do not participate in lookup or command routing.
- Tags are optional labels for future filtering and explanation.
- Tags must not replace path hierarchy.
- MVP does not include group metadata.

## `shelf export`

Directly export secrets from the secret namespace.

```bash
shelf export providers/openrouter/accounts/personal:api_key --format shell
shelf export providers/openrouter/accounts/personal --format env
shelf export providers/openrouter/accounts/personal --format json
```

Formats:

```text
shell  export KEY=value lines
env    KEY=value lines
json   JSON object
```

Rules:

- `export` is not project-aware.
- `export` operates on explicit secret paths or path prefixes.
- `shell` output must be safe for `eval "$(shelf export ... --format shell)"`.
- `env` and `shell` use a secret's `env` field when present.
- If `env` is not present, the env name is derived from the secret path.

Future distinction:

```text
shelf export        = direct path/prefix export
shelf project setup = possible future project-local env materialization
```

`shelf project env` is not an MVP command. If it returns later, it should be a project-aware projection helper distinct from direct export.

## `shelf project id`

Identify the current Git project.

```bash
shelf project id
```

Output:

```text
github.com/owner/repo
```

Rules:

- Read-only.
- Uses the current working directory.
- Finds the Git worktree root.
- Reads the default remote, expected to be `origin` in MVP.
- Normalizes common Git remote URL forms to `host/owner/repo`.
- Fails if the current directory is not inside a Git worktree.
- Fails if no usable remote URL exists.

## `shelf doctor`

Check local Shelf configuration and data health.

```bash
shelf doctor
```

Output example:

```text
ok   config resolves (/home/han/.config/shelf/config.yaml)
ok   version (v0.1.0 go1.26 linux/amd64)
ok   data file exists (/home/han/.local/share/shelf/secrets.json)
ok   data file mode (-rw-------)
ok   store loads (/home/han/.local/share/shelf/secrets.json)
ok   git tracking (data file is not inside a Git worktree)
ok   completion installed (/home/han/.zfunc/_shelf)
```

Rules:

- `ok` means the check passed.
- `warn` means usable but needs attention.
- `fail` means doctor exits non-zero.
- Initial checks are local only: config resolution, version, data file existence and mode, store validation, ordinary Git tracking, and zsh completion paths discovered from `FPATH` / `fpath`.
- It does not check chezmoi, age, or external secret-manager integrations.

## Project workflow (v0.2–v0.4)

v0.2 foundation commands (`shelf project init` and `shelf project explain`) are implemented. The remaining commands in this section are specifications for future implementation.

### `.shelf.json` location and format

Project manifest lives at `<git-root>/.shelf.json`. It declares which Shelf secret paths the project needs. It is a manifest of intent, not a secret store.

Reasons for `.shelf.json` over `.env`:

- `.env` is a key-value file for environment variables. It cannot reliably encode Shelf's `group_path:key` identity.
- `.env` invites users to paste real secret values into project files.
- `.env` cannot express include/exclude/required/collision/profiles.

`.shelf.json` can be committed to Git; `.env.local` (generated output) must be gitignored.

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

Field rules:

- `version`: required, fixed at `1`.
- `secrets`: required array of entries.
- `path`: exact Shelf secret path (mutually exclusive with `prefix`).
- `prefix`: matches all secrets whose canonical path starts with this string. Deferred to v0.3; v0.2 supports `path` only.
- `env`: optional project-level override for the environment variable name.
- `required`: optional, defaults to `true`.
- Fields prohibited: `value`, fallback plaintext, shell commands, template expressions.

### `shelf project init`

```bash
shelf project init
```

Writes `<git-root>/.shelf.json` with a minimal scaffold. Must run inside a Git worktree. Fails if `.shelf.json` already exists unless `--force` is given. Writes no secret values.

### `shelf project explain`

```bash
shelf project explain
```

Read-only explanation. Shows project identity, manifest path, resolved env names, and missing/conflict status.

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

- Does not print secret values, execute commands, or modify files.
- Validates `.shelf.json` format, path existence, env name uniqueness, and required/optional coverage.
- `ok`/`warn`/`fail` per entry; `fail` exits non-zero.
- Env name resolution: project override → secret `env` field → derived from path.
- Duplicate env name → `fail`.

### `shelf project add`

```bash
shelf project add <path-or-prefix> [--env NAME] [--optional]
```

Appends an entry to `.shelf.json`. If the file does not exist, prompts to run `shelf project init` first. For `path`: validates secret exists in store; fails if not found. For `prefix` (v0.3): validates at least one match; fails if zero. Rejects duplicates.

### `shelf project rm`

```bash
shelf project rm <path-or-prefix>
```

Removes the matching entry from `.shelf.json`.

### `shelf project list`

```bash
shelf project list
```

Lists `.shelf.json` entries without resolving secret values.

### `shelf project export`

```bash
shelf project export --format env|shell|json
```

Exports environment variables from the project manifest. Reads `.shelf.json`, resolves all `path`/`prefix` entries, expands prefixes to matching secrets (stable sort), computes env names, and outputs in the requested format. Reuses value conversion and shell quoting from `shelf export`.

Distinction from `shelf export`:

| Command | Source | Awareness |
| --- | --- | --- |
| `shelf export <path-or-prefix>` | explicit argument | direct |
| `shelf project export` | `.shelf.json` in Git root | project-aware |

Behavior:

- Env name conflict → fail.
- Required secret missing → fail.
- Optional secret missing → skip with warning to stderr.

Typical usage:

```bash
shelf project export --format env > .env.local
```

### `shelf run`

```bash
shelf run -- command args...
shelf run --dry-run -- command args...
```

Runs a command with secrets from `.shelf.json` injected into the child process environment. `--dry-run` prints which env vars would be injected (no values) and skips execution.

Runtime flow:

1. Find Git root.
2. Load `<git-root>/.shelf.json`.
3. Resolve `path`/`prefix` entries.
4. Read secret values from store.
5. Compute env names.
6. Check required missing → fail.
7. Check env conflict → fail.
8. Construct child env (Shelf overrides parent).
9. Execute command.
10. Return child exit code.

Key principles:

- Only injects into child process; does not mutate parent shell.
- Does not write `.env.local`.
- Does not print secret values.
- Shelf-resolved env vars override parent env vars by default.
- `explain` and `--dry-run` warn about overrides.

## Non-goals for MVP

- Stdin-based secret creation.
- Field-specific metadata mutation commands.
- Group metadata.
- Shell hook.
- Capture from `.env` files.
- Clipboard integration.
- fzf picker.
- Chezmoi integration.
- External secret-manager backends.
- Built-in encryption.
- History/versioning.
- TUI.
