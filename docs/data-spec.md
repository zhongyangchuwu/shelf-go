# Shelf Go Rewrite Data Spec

## Purpose

This document defines the MVP data model for the Go rewrite.

The model intentionally does not copy the Python store architecture. It keeps the useful `group:key` path idea, but makes each secret a first-class object with its metadata stored beside the value.

Core principles:

```text
Usability beats storage efficiency. Shelf data is expected to be small.
A secret is a first-class object.
Metadata lives with the secret.
Namespaces are derived from secret paths, not stored as nested group nodes.
```

## Store format

MVP store shape:

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

Top-level fields:

```text
version  required integer
secrets  required object mapping secret path to secret object
```

Default MVP data path:

```text
~/.local/share/shelf/secrets.json
```

The store is JSON because it is a managed data file, not a hand-edited user configuration file. Users should inspect and modify secrets through `shelf secret info`, `shelf secret set`, and `shelf secret edit`, not by editing the store file directly.

JSON also keeps the future encryption boundary simple: the plaintext data model is one deterministic JSON document, and a later encrypted backend can wrap the load/save layer without changing command semantics.

## Storage and encryption policy

MVP uses a managed JSON file, not a database.

Reasons:

- The expected data size is small.
- The MVP access pattern is simple path lookup, prefix listing, object editing, and export.
- A database does not solve plaintext-at-rest by itself; an unencrypted SQLite file is still plaintext storage.
- A file store keeps backup, sync, migration, and future encryption boundaries straightforward.

Security stance:

- `secrets.json` is sensitive.
- `config.yaml`, if present, is non-sensitive runtime configuration and must not contain secret values.
- The MVP may store plaintext JSON, but the data model is designed so a later encrypted backend can wrap the load/save layer.
- Future encryption should preserve command semantics: commands operate on the same plaintext data model after decryption.
- Users should not manually edit `secrets.json`; `secret edit` is the supported edit path.

Recommended filesystem behavior:

- Create the data file with user-only permissions where the platform supports it.
- Warn when the data file is tracked by ordinary Git without an encryption workflow.
- Keep atomic writes and backups in the storage layer.

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
namespace  the part before the single colon
key        the part after the single colon
path       namespace + ':' + key
```

Rules:

- A path must contain exactly one `:` separator.
- Namespace must not be empty.
- Key must not be empty.
- Namespace segments are separated by `/`.
- Namespace segments must not be empty.
- Namespace segments must not contain `:`.
- Key must not contain `/` or `:`.
- The full path is the secret's unique ID.

Recommended allowed characters for MVP:

```text
namespace segment: A-Z a-z 0-9 _ - .
key:               A-Z a-z 0-9 _ - .
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

`shelf export --format shell` must be safe for:

```bash
eval "$(shelf export providers/openrouter/accounts/personal:api_key --format shell)"
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
  "value": "sk-xxx",
  "env": "OPENROUTER_API_KEY",
  "description": "Personal OpenRouter key",
  "tags": ["ai", "openrouter", "personal"]
}
```

Validation after edit:

- JSON must parse.
- `value` must exist.
- `env`, when present, must match env-name validation.
- `description`, when present, must be a string.
- `tags`, when present, must be a list of strings.
- Secret path is not edited inside the object; rename/move is out of MVP scope.

## Unsupported in MVP data model

- Group objects.
- Group metadata.
- Parallel value/meta maps.
- Arbitrary metadata map.
- Provenance/source field.
- Project bindings.
- Views/projections.
- Secret history.
- Built-in encryption.
- External secret-manager references.

These may be added later, but are not part of the MVP store contract.
