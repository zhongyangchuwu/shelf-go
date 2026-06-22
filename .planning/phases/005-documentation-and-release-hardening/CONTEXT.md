# Context: Phase 5 Documentation and Release Hardening

## Goal

Make the encrypted-vault workflow clear enough for real use and produce final release evidence that every v1 requirement is complete or explicitly deferred.

## Constraints

- Documentation must match implemented code, not earlier MVP wording.
- Explain encrypted vault, age recipients, identity paths, and chezmoi-safe file boundaries.
- Clearly separate Shelf config, `.shelf.json`, encrypted vault data, generated/exported env files, and plaintext migration sources.
- Warn about commands and UI actions that intentionally reveal values: `secret get`, `export`, `project export`, terminal output, browser reveal, and old plaintext stores.
- Keep docs concise; update existing docs instead of creating unnecessary new docs.

## Decisions

- Update `README.md`, `docs/usage-spec.md`, and `docs/data-spec.md` because those are the user-facing docs currently linked from README.
- Treat `docs/roadmap.md` as older project roadmap material unless it contains active user guidance.
- Use Phase 1-4 verification artifacts plus `go test ./...` as final release evidence.

## Open Questions

- None.

## Verification Expectations

- Verify docs mention current command surface, vault defaults/config keys, migration, doctor safety, manager safety controls, and plaintext/reveal warnings.
- Run `go test ./...` after docs/planning updates to ensure code remains verified.
