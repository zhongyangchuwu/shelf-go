# Phase 19 Context: Secret Tag Selection

## Goal

Add compact CLI tag selectors for direct secret workflows: `shelf secret list --tag` and `shelf secret export --tag`.

## Constraints

- Do not add `secret meta` or `secret tag` command groups.
- Keep existing path/prefix export behavior unchanged.
- Preserve `secret export --all` behavior: by default export only secrets with explicit env names; `--all` permits derived env names.
- Multiple `--tag` flags use AND semantics.
- Output must remain deterministic and value-free for `secret list`.
- This phase does not add project manifest tag bindings; Phase 20 owns that.

## Decisions

- Add a reusable store-level tag matcher so Phase 20 can use the same AND semantics.
- Allow `secret export --tag <tag>` without a path argument to export all matching tagged secrets.
- When both path/prefix and tags are provided, apply both filters.
- Keep `--tag` as repeatable `StringArray`, matching existing `secret set --tag` input style.

## Open Questions

- None.

## Verification Expectations

- Tests prove list tag filtering is deterministic and value-free.
- Tests prove export by tag works for env, shell, and JSON formats.
- Tests prove multiple tags use AND semantics.
- Tests prove default env-only filtering and `--all` still apply to tag-selected exports.
- `go test ./internal/cli ./internal/store` and `go test ./...` pass.
