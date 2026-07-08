# CLI Spec: Secret CLI v2 Workflows

## Intent

Replace the old `group:key` secret CLI with a v2 item/field/template interface. First version stays non-interactive and scriptable; richer discovery, filtering, batching, and editing belong to a later TUI.

## Model Decisions

- Template defines item structure, field types, sensitivity, primary secret field, env-name fields, and validation.
- Profile defines defaults over a template. `openai` is a profile over `api_credential`, not its own template and not a provider subsystem.
- Provider/source remains a normal public field when a template includes it.
- Env name is a normal public field, not a special binding object. By convention and template metadata, the default env-name field for a secret field is `<field>_env`.
- Explicit user flags override profile defaults, and profile defaults override template defaults.
- Secret values must enter through stdin, not argv.
- Duplicate env-name field values are allowed in the vault; ambiguity/conflict is reported by selectors and project/runtime resolution where context exists.

## Built-in Templates, First Pass

### `generic_secret`

- `value`: concealed, secret, primary, required.
- `value_env`: string, public, env name for `value`.

### `api_credential`

- `credential`: concealed, secret, primary, required.
- `credential_env`: string, public, env name for `credential`.
- `provider`: string, public.
- `website`: url, public.
- `expires_at`: date, public.
- `scopes`: string_list, public.

## Built-in Profiles, First Pass

Profiles are registry data, not CLI conditionals.

### `openai`

- template: `api_credential`
- aliases: `openai`
- defaults:
  - `provider`: `OpenAI`
  - `website`: `https://platform.openai.com`
  - `credential_env`: `OPENAI_API_KEY`

### Optional later profiles

Add only when backed by registry definitions and tests, not switch statements:

- `anthropic`
- `github`
- `cloudflare`
- `openrouter`
- `deepseek`

## Commands

### Create

```text
shelf secret create <title> [--template generic|api] [--profile PROFILE] [--id ID] [--env NAME] [--stdin]
```

Rules:

- `--template` defaults to `generic` unless `--profile` selects a template.
- `--profile` applies profile defaults.
- `--env` writes the primary secret field's env-name field, e.g. `value_env` or `credential_env`.
- `--stdin` reads the primary secret field value from stdin.
- No secret value positional argument exists.
- Public template fields may be set by first-version flags where useful: `--provider`, `--website`, `--expires`, repeated `--scope`.

Examples:

```text
printf %s "$TOKEN" | shelf secret create github --env GITHUB_TOKEN --stdin
printf %s "$OPENAI_API_KEY" | shelf secret create openai --profile openai --stdin
printf %s "$TOKEN" | shelf secret create openai --profile openai --env LLM_API_KEY --stdin
```

### Set

```text
shelf secret set <selector> [--field FIELD] --stdin
```

Rules:

- `--field` defaults to the template primary secret field.
- stdin is parsed according to the target field type.
- `string_list` accepts newline-separated values in first version.
- Secret values via argv are rejected by design.

Examples:

```text
printf %s "$TOKEN" | shelf secret set openai --stdin
printf %s OPENAI_API_KEY | shelf secret set openai --field credential_env --stdin
printf "responses\nembeddings\n" | shelf secret set openai --field scopes --stdin
```

### Get

```text
shelf secret get <selector> [--field FIELD]
shelf secret get --env ENV [--field FIELD]
```

Rules:

- `--field` defaults to the template primary secret field.
- `--env` finds fields whose env-name field equals `ENV`, then reads the corresponding secret field unless `--field` overrides it.
- Multiple matches fail with a value-free ambiguity error.
- Scalar values print as raw strings plus newline. Lists and JSON print JSON plus newline.

### List

```text
shelf secret list [query] [--long]
```

Rules:

- Always value-free.
- Short output prints titles.
- Long output prints ID, title, template, primary env name, and primary secret value-set state.
- `query` filters by exact/substring title, alias, template, profile/default metadata if present; no fuzzy selection in first version.

### Info

```text
shelf secret info <selector>
```

Rules:

- Always value-free for secret fields.
- Public field values are visible.
- Secret fields show `value_set=true|false`.
- Env-name fields are shown as normal public fields.

### Remove

```text
shelf secret rm <selector> --yes
```

Rules:

- Requires `--yes` in first version.
- Ambiguous selectors fail before mutation.

## Removed or Deferred Commands

- `secret add`: removed; replaced by `secret create`.
- `secret edit`: deferred to TUI/wizard.
- `secret export`: deferred to project/export v2 design.
- `secret copy`: deferred until clipboard adapter and platform behavior are designed.
- `secret bind` / `secret unbind` / `secret env set`: not needed because env names are normal public fields.

## Selector Rules

First version is precise, not fuzzy.

Resolution order:

1. exact item ID
2. exact title
3. exact alias
4. `--env` env-name field lookup

Ambiguity output is value-free and lists item IDs/titles/field names only.

## Storage Rule for Phase 38

Use the v2 JSON repository boundary. Do not overwrite the existing v1 encrypted vault by default during development/cutover. If a default v2 path is needed before full config migration, use a distinct v2 path and document it in tests and CLI help.

## Acceptance Criteria

- `secret create` can create `generic_secret` and `api_credential` items.
- `secret create --profile openai` applies registry defaults without CLI hardcoding provider behavior.
- `secret create --stdin` and `secret set --stdin` are the only first-version ways to write secret field values from CLI.
- `secret get` materializes only explicit requested values.
- `secret list --long` and `secret info` do not print secret field values.
- Env names are stored and edited as ordinary public fields.
- Ambiguous selector and duplicate env lookup failures are value-free.
- Removed legacy v1 command paths fail clearly or are absent.
