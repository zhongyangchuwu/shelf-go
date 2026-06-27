# Review: Phase 5 Documentation and Release Hardening

## Scope Reviewed

- `README.md`
- `docs/usage-spec.md`
- `docs/data-spec.md`
- v1 requirement traceability in `.planning/REQUIREMENTS.md` and `.planning/ROADMAP.md`

## Findings

- README no longer describes Shelf as only a Python rewrite/local export MVP; it now names encrypted vault storage and current implemented workflows.
- Usage spec no longer treats project/run flows as future v0.2-v0.4 work; it documents implemented v1 behavior.
- Usage/data docs now separate config, encrypted vault, `.shelf.json`, generated env files, and explicit reveal/output boundaries.
- Docs warn about plaintext migration sources, generated env files, terminal value output, editor buffers, browser reveal, and tokenized manager URLs.

## Fixes Applied

- Replaced stale usage spec content with current v1 behavior.
- Added config/vault and manager boundary documentation.
- Added storage and value materialization warnings to README and data spec.

## Waivers

- No external docs generator/subagent was used; the relevant docs were small and directly verified against code and phase artifacts.
- No release tag/archive was created because the user asked to continue development phases, not publish a release.

## Remaining Risks

- Docs should be revisited before public release if CLI flags change after this phase.
