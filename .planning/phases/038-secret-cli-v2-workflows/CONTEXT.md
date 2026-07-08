# Context: Secret CLI v2 Workflows

## Goal

Add common non-interactive secret workflows over the v2 item/field model while preserving value-free output rules and avoiding premature TUI/prompt work.

## Constraints

- Do not evolve the old interactive `secret add` prompt in this phase.
- Keep secret values out of list/info/status/log/planning output.
- `get`, `copy`, export, and run are explicit secret-materializing paths; this phase focuses on secret CLI create/set/get/copy/list/info.
- Use fake secrets only in tests.
- Treat the v2 secret CLI as a clean breaking redesign instead of preserving old v1 command behavior; rely on git rollback if the cutover proves wrong.
- Clipboard behavior may need an explicit dependency/OS decision; if not safe, scope copy behind a small adapter or defer it with evidence.

## Decisions

- Prefer scriptable commands first: create, set from stdin, get, list long, info, and rm.
- Use selectors from Phase 34: item ID, exact title, exact alias, `--env`, and unambiguous matches only.
- Default secret field is the template primary secret field when unambiguous.
- Value-free commands may show item title, template, labels, aliases, public fields, field IDs, env-name fields, and value-set booleans.
- The old `group:key` CLI shape is not a compatibility burden to preserve; the v2 interface should expose item/template/field concepts directly and delete obsolete command paths once replaced.
- Env names are ordinary public fields, defaulting by template convention to `<secret_field>_env`; they are not special binding objects.
- Profiles define defaults over templates. `openai` is a profile over `api_credential`, not a dedicated template and not a provider subsystem.
- Provider/source is just a public field inside templates that need it.

## Open Questions

- How much model refactor is needed before implementing the CLI: replacing field bindings with env-name fields, adding template env metadata, and adding profile registry support.
- How the CLI chooses a v2 repository path during the cutover.
- Which public docs should be updated in the same phase once the command behavior is implemented.

## Verification Expectations

- Focused CLI/app tests prove non-interactive create/set/get/list/info behavior without leaking secret values in value-free outputs.
- Existing v1 CLI tests remain passing if compatibility is preserved.
- Any command that materializes secret values is explicitly tested as such.
