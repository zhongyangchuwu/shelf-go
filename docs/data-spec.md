# Shelf Go Data Spec

## Purpose

This document defines the encrypted vault data model and the plaintext in-memory model used after decrypt/load.

The model intentionally does not copy the Python store architecture. It keeps the useful `group:key` path idea, but makes each secret a first-class object with its metadata stored beside the value.

Core principles:

```text
Usability beats storage efficiency. Shelf data is expected to be small.
A secret is a first-class object.
Metadata lives with the secret.
Namespaces are derived from secret paths, not stored as nested group nodes.
```

## Plaintext model inside the encrypted vault

After decrypting the vault, Shelf validates this JSON shape:

```json
{
  "version": 1,
  "secrets": {
    "providers/openrouter/accounts/personal:api_key": {
      "value": "sk-xxx",
      "env": "OPENROUTER_API_KEY",
      "description": "Personal OpenRouter key",
      "tags": ["ai", "openrouter", "personal"]
    }
  }
}
```

## Storage and encryption policy

Shelf persists secrets as an age-encrypted vault file, not a database.

Reasons:

- The expected data size is small.
- The access pattern is simple path lookup, prefix listing, object editing, and export.
- A database does not solve plaintext-at-rest by itself; an unencrypted SQLite file is still plaintext storage.
- A portable encrypted file keeps backup, sync, migration, and chezmoi workflows straightforward.

Security stance:

- `vault.age` is the durable encrypted source of truth.
- `config.yaml`, if present, is non-sensitive runtime configuration and must not contain secret values or private age identities.
- Config may contain public age recipients and identity file paths.
- Commands operate on the same plaintext data model only after decrypt/load and before validate/encrypt/save.
- Users should not manually edit the encrypted vault; `secret edit` is the supported edit path.
- `shelf vault migrate --from <plaintext.json>` is the supported plaintext-to-vault conversion path. It preserves the plaintext source, writes an encrypted target, then decrypts and validates the target before reporting success.

## Runtime config boundary

Shelf config is YAML and must stay value-free:

```yaml
version: 1
vault_path: ~/.local/share/shelf/vault.age
recipients:
  - age1...
identity_paths:
  - ~/.config/shelf/identity.txt
editor: vim
```

Rules:

- `recipients` are public age recipients.
- `identity_paths` are filesystem paths to private identity files; the private identity material is not embedded in config.
- `vault_path` points at the encrypted durable store.
- Config can be reviewed or committed only if it contains no private key material or secret values.

## Value materialization boundaries

These operations intentionally materialize plaintext values:

- decrypting the vault into memory during command execution;
- `secret get`;
- `secret edit` editor buffer;
- `export` and `project export` output;
- `run` child process environment;
- `manager` explicit reveal endpoint and browser reveal action.

Generated env files and copied terminal/browser output are plaintext artifacts outside the vault and must not be committed.

Recommended filesystem behavior:

- Create the vault file with user-only permissions where the platform supports it.
- Warn when plaintext legacy stores are still present or tracked by ordinary Git.
- Keep atomic writes and backups in the storage layer, and encrypt backup bytes before durable persistence.
- When replacing an encrypted vault, backups are encrypted vault bytes; plaintext migration sources are never rewritten or backed up by Shelf.
- Mutating commands must take an exclusive lock at `<vault-path>.lock` before loading the store for modification.
- Write flow is: lock, decrypt latest vault, mutate in memory, validate, encrypt to temp file, rename, unlock.

## Why flat secrets

The Python MVP stored values and metadata under group nodes:

```text
$values
$meta
$value_meta
```

The Go model avoids that split.

Reasons:

- A secret's value and metadata should be readable in one place.
- Rename and edit operations should update one object, not two parallel maps.
- Path hierarchy should not require reserved keys.
- Group metadata is not part of the MVP product loop.
- Listing, lookup, and export can all operate directly on stable secret IDs.

## Secret path grammar

A secret path has this shape:

```text
<namespace>:<key>
```

Examples:

```text
providers/openrouter/accounts/personal:api_key
providers/openai/accounts/work:api_key
github/accounts/personal:token
```

Definitions:

```text
group_path  the part before the single colon
key         the part after the single colon
path        group_path + ':' + key
```

Rules:

- A path must contain exactly one `:` separator.
- Group path must not be empty.
- Key must not be empty.
- Group path segments are separated by `/`.
- Group path segments must not be empty.
- Group path segments must not contain `:`.
- Key must not contain `/` or `:`.
- Code may model `group_path` and `key` separately, but the store file uses the full path string as the map key.
- The full path is the secret's unique ID.

Recommended allowed characters for MVP:

```text
group path segment: A-Z a-z 0-9 _ - .
key:                A-Z a-z 0-9 _ - .
```

Implementations may reject other characters to keep shell, JSON, and path-like usage predictable.

## Secret object

A secret object contains one required value and optional metadata.

```json
{
  "value": "sk-xxx",
  "env": "OPENROUTER_API_KEY",
  "description": "Personal OpenRouter key",
  "tags": ["ai", "openrouter", "personal"]
}
```

Fields:

```text
value        required JSON-compatible value
env          optional environment variable name
description  optional string
tags         optional list of strings
```

No generic `meta`, `attrs`, or arbitrary metadata map exists in MVP.

## Value field

`value` may be any JSON-compatible value:

```text
string
number
boolean
null
object
array
```

`secret set` JSON-parses the command-line value when possible.

Examples:

```bash
shelf secret set app:port 34222
shelf secret set flags:enabled true
shelf secret set app:options '{"debug":false}'
shelf secret set providers/openrouter/accounts/personal:api_key sk-xxx
```

Stored values:

```json
{
  "secrets": {
    "app:port": { "value": 34222 },
    "flags:enabled": { "value": true },
    "app:options": { "value": { "debug": false } },
    "providers/openrouter/accounts/personal:api_key": { "value": "sk-xxx" }
  }
}
```

## Env field

`env` stores the preferred environment variable name for direct export.

Example:

```json
{
  "secrets": {
    "providers/openrouter/accounts/personal:api_key": {
      "value": "sk-xxx",
      "env": "OPENROUTER_API_KEY"
    }
  }
}
```

Validation:

```regex
^[A-Za-z_][A-Za-z0-9_]*$
```

Rules:

- Invalid env names must be rejected on write.
- `export --format env` and `export --format shell` use `env` when present.
- If `env` is absent, export derives a name from the secret path.
- The env field is not part of secret identity.

## Description field

`description` is optional human-readable text.

Example:

```json
{
  "description": "Personal OpenRouter key"
}
```

Rules:

- Must be a string when present.
- Empty description is equivalent to absent description for display purposes.
- It does not participate in lookup or export.

## Tags field

`tags` is an optional list of cross-cutting labels.

Example:

```json
{
  "tags": ["ai", "openrouter", "personal"]
}
```

Rules:

- Tags are strings.
- Tags should be unique within a secret object.
- Tags do not participate in secret identity.
- Tags do not participate in lookup or command routing.
- Tags are for future filtering and explanation.
- Tags must not replace path hierarchy.

Recommended allowed characters for MVP:

```text
A-Z a-z 0-9 _ - .
```

## Group/path vs tags

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

Example:

```json
{
  "secrets": {
    "providers/openrouter/accounts/personal:api_key": {
      "value": "sk-xxx",
      "env": "OPENROUTER_API_KEY",
      "tags": ["ai", "personal"]
    }
  }
}
```

Identity:

```text
providers/openrouter/accounts/personal:api_key
```

Namespace:

```text
providers/openrouter/accounts/personal
```

Tags:

```text
ai, personal
```

Rules:

- The path is required and unique.
- Tags are optional.
- Tags are unordered labels.
- A tag cannot identify a secret by itself.
- `secret get`, `secret info`, and `export` route by path or path prefix, not by tags.

## Prefix matching

Commands that accept a prefix operate on path strings.

Example store:

```json
{
  "secrets": {
    "providers/openrouter/accounts/personal:api_key": { "value": "sk-xxx" },
    "providers/openrouter/accounts/work:api_key": { "value": "sk-yyy" },
    "providers/openai/accounts/personal:api_key": { "value": "sk-zzz" }
  }
}
```

Prefix:

```text
providers/openrouter
```

Matches:

```text
providers/openrouter/accounts/personal:api_key
providers/openrouter/accounts/work:api_key
```

Does not match:

```text
providers/openai/accounts/personal:api_key
```

## Info output model

`secret info <path>` emits non-secret JSON.

Example:

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

- `value` is not included.
- `value_set` is true for a valid MVP secret object because `value` is required.
- Optional string fields are omitted when absent or empty.
- `tags` is always emitted; absent tags are emitted as an empty array.

When optional fields are absent:

```json
{
  "path": "providers/openrouter/accounts/personal:api_key",
  "group_path": "providers/openrouter/accounts/personal",
  "key": "api_key",
  "value_set": true,
  "tags": []
}
```

## Export value conversion

Env output converts JSON values to strings.

| JSON value | Env value |
| --- | --- |
| string | unchanged string |
| number | decimal string |
| true | `true` |
| false | `false` |
| null | empty string |
| object | compact JSON |
| array | compact JSON |

## Export env name resolution

For each exported secret:

1. Use the secret object's `env` field if present.
2. Otherwise derive a name from the full secret path.

Derived env name rules:

- Replace non-alphanumeric characters with `_`.
- Collapse repeated `_`.
- Trim leading/trailing `_`.
- Uppercase the result.

Example:

```text
providers/openrouter/accounts/personal:api_key
```

Derived env:

```text
PROVIDERS_OPENROUTER_ACCOUNTS_PERSONAL_API_KEY
```

## Shell quoting

`shelf secret export --format shell` must be safe for:

```bash
eval "$(shelf secret export providers/openrouter/accounts/personal:api_key --format shell)"
```

Format:

```text
export NAME=<quoted-value>
```

Examples:

```text
hello        -> export X=hello
hello world  -> export X='hello world'
a'b          -> export X='a'\''b'
empty string -> export X=''
null         -> export X=''
{"a":1}      -> export X='{"a":1}'
```

## Edit object format

`secret edit <path>` opens the full secret object as JSON.

Example buffer:

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

Validation after edit:

- JSON must parse.
- `group_path` and `key` must form a valid secret path.
- `value` must exist.
- `env`, when present, must match env-name validation.
- `description`, when present, must be a string.
- `tags`, when present, must be a list of strings.
- If `group_path` or `key` changes, the edit is a rename and must fail when the destination path already exists.

## `.shelf.json` project manifest

Project manifest lives at `<git-root>/.shelf.json`. It declares which Shelf secret paths a project needs. It is a manifest of intent, not a secret store.

Rationale: `.env` is a key-value file for environment variables, not a project binding format. It cannot reliably encode Shelf's `group_path:key` identity, and it invites users to paste real secret values into project files.

`.shelf.json` can be committed to Git when reviewed as value-free. `.env.local` and other generated exports contain plaintext values and must be gitignored.

### Format

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

### Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `version` | number | yes | Schema version; fixed at `1`. |
| `secrets` | array | yes | Ordered list of secret references. |
| `secrets[].path` | string | one of `path`/`prefix` | Exact Shelf secret path (`group_path:key`). |
| `secrets[].prefix` | string | one of `path`/`prefix` | Prefix matching all secrets whose canonical path starts with this string. |
| `secrets[].env` | string | no | Project-level env name override. Must match `^[A-Za-z_][A-Za-z0-9_]*$`. |
| `secrets[].required` | boolean | no | Whether a missing secret should fail. Defaults to `true`. |

### Entry rules

- `path` and `prefix` are mutually exclusive; exactly one must be present per entry.
- `prefix` entries should not carry `env` because a prefix may expand to multiple secrets with different env names.
- `env` provides a project-level override; it does not modify the secret object's `env` field.
- Duplicate `path` or `prefix` values within the same `.shelf.json` are rejected.

### Prohibited fields

`.shelf.json` MUST NOT contain:

- `value` / resolved secret values.
- Fallback plaintext secrets.
- Shell commands or template expressions.
- Secret object fields other than the reference (`path`/`prefix`) and projection hints (`env`, `required`).

### Env name resolution

When resolving an env name for a secret referenced from `.shelf.json`:

1. Use the entry's `env` if present (project-level override).
2. Otherwise use the secret object's `env` field.
3. Otherwise derive from the full secret path (uppercase, non-alphanumeric → `_`).

Two entries resolving to the same env name is a conflict and must be reported as an error.

### Version evolution

- v0.2: `path` entries only; no `prefix`, no `profiles`.
- v0.3: adds `prefix` entries.
- v0.5 (future): adds `profiles` dictionary.

## Localhost manager API boundary

The localhost vault manager exposes the same vault data through a local HTTP server.

Rules:

- List/search responses include path, env, description, tags, and `value_set`; they do not include `value`.
- Reveal is a separate explicit route that returns plaintext value only after token/Host checks.
- Create/update/delete routes reuse `Vault.Update` and `Store.Set`/`Store.Delete`, so validation, locking, encrypted save, and encrypted backups remain centralized.
- Unsafe write routes require token and Origin/Host controls because malicious browser pages can target loopback.
- The manager is not a durable store. The encrypted vault remains the only source of truth.

## Unsupported in v1 data model

- Group objects.
- Group metadata.
- Parallel value/meta maps.
- Arbitrary metadata map.
- Provenance/source field.
- Secret history.
- External secret-manager references.

These may be added later, but are not part of the MVP store contract.

