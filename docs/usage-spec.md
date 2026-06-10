# Shelf Go Rewrite Usage Spec

## Purpose

Shelf-Go is not a line-by-line port of the Python CLI. It keeps the useful ideas from the MVP and reshapes them into a smaller, faster local secret manager.

MVP product focus:

```text
Fast local secret management and direct export.
```

Git-aware project workflows remain a possible future direction, but they do not drive the MVP. The first Go version should complete the secret manager mission before designing project setup/sync behavior.

MVP scope proves the new secret object model, direct export workflow, and a small Git project identity utility without shipping incomplete project workflow commands.

## Design rules

- Prefer one canonical command for each purpose.
- Do not add aliases for migration comfort.
- Do not ship placeholder commands.
- A command ships only when it has complete behavior.
- Keep command behavior small, explicit, and scriptable.
- Secret metadata belongs to the secret object, not to a group node.
- Group/path is identity; tags are auxiliary labels.

## MVP command surface

```bash
shelf secret set <path> <value> [--env NAME] [--description TEXT] [--tag TAG ...] [--force]
shelf secret get <path>
shelf secret list [prefix]
shelf secret info <path>
shelf secret edit <path>

shelf export <path-or-prefix> --format shell|env|json

shelf project id
```

Deferred project commands:

```bash
shelf project setup
shelf project status
shelf project sync
shelf project explain
```

Those commands are possible future extensions, but not MVP requirements. They should not appear in the binary until their full behavior is specified and implemented.

## Secret paths

A secret path uses the existing `group:key` shape:

```text
providers/openrouter/accounts/personal:api_key
providers/openai/accounts/work:api_key
github/accounts/personal:token
```

The part before `:` is the path namespace. The part after `:` is the value key.

The path is the secret's canonical ID.

## `shelf secret set`

Create a new secret object.

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
  "value": "sk-xxx",
  "env": "OPENROUTER_API_KEY",
  "description": "Personal OpenRouter key",
  "tags": ["ai", "openrouter", "personal"]
}
```

Rules:

- Opens the user's editor.
- Edits the full secret object.
- The edit buffer is JSON.
- The edited object is validated before writing.
- Invalid edits are not written.
- `edit` is the unified metadata modification path.
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

## Deferred project workflow

MVP is not project-first. Future project workflows should be based on real usage experience, not placeholder commands.

The likely future direction is project-local environment materialization:

```text
Git project identity
→ selected secret paths
→ generated project-local env file such as .env.local
```

Possible future commands:

```bash
shelf project setup
shelf project status
shelf project sync
shelf project explain
```

MVP intentionally defers them. The first Go version should not include placeholder project commands. Runtime injection commands such as `shelf project run` are not the default product story unless future usage proves they are needed.

## Non-goals for MVP

- Prompt-based secret creation.
- Stdin-based secret creation.
- Field-specific metadata mutation commands.
- Group metadata.
- Project binding.
- Project-aware env projection.
- Shell hook.
- Capture from `.env` files.
- Clipboard integration.
- fzf picker.
- Chezmoi integration.
- External secret-manager backends.
- Built-in encryption.
- History/versioning.
- TUI.
