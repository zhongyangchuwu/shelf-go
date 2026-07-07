# Context: Secret CLI v2 Workflows

## Goal

Add common non-interactive secret workflows over the v2 item/field model while preserving value-free output rules and avoiding premature TUI/prompt work.

## Constraints

- Do not evolve the old interactive `secret add` prompt in this phase.
- Keep secret values out of list/info/status/log/planning output.
- `get`, `copy`, export, and run are explicit secret-materializing paths; this phase focuses on secret CLI create/set/get/copy/list/info.
- Use fake secrets only in tests.
- Do not remove v1 command behavior unless a clean compatibility/cutover path is implemented and tested.
- Clipboard behavior may need an explicit dependency/OS decision; if not safe, scope copy behind a small adapter or defer it with evidence.

## Decisions

- Prefer scriptable commands first: create, set from stdin, get, copy, list long, info.
- Use selectors from Phase 34: item ID, exact title, `--env`, and unambiguous query.
- Default secret field is the template primary secret when unambiguous.
- Value-free commands may show item title, template, labels, aliases, public fields, field IDs, env names, and value-set booleans.

## Open Questions

- Whether v2 commands live beside existing v1 commands during transition or replace them behind current command names.
- Whether `copy` is implemented now or deferred until clipboard dependency and platform behavior are designed.
- How the CLI chooses a v2 repository path before full runtime cutover.

## Verification Expectations

- Focused CLI/app tests prove non-interactive create/set/get/list/info behavior without leaking secret values in value-free outputs.
- Existing v1 CLI tests remain passing if compatibility is preserved.
- Any command that materializes secret values is explicitly tested as such.
