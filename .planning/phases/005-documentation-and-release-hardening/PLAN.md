# Plan: Phase 5 Documentation and Release Hardening

## Objective

Bring user-facing docs in sync with the implemented encrypted vault, migration, CLI compatibility, and localhost manager behavior, then record final v1 requirement verification.

## Scope

In scope:

- Update README command/status overview.
- Update usage spec for encrypted vault config, migration, doctor, manager, export/run value exposure, and generated env file warnings.
- Update data spec for vault/config/manifest/export boundaries and localhost manager data-flow warnings.
- Final verification matrix for all v1 requirements.
- Root planning completion updates.

Out of scope:

- Published release archive or version tagging.
- New feature work.
- Rich UI documentation beyond current manager MVP.

## Tasks

1. Update `README.md`.
   - Describe Shelf as an encrypted local secret manager.
   - Include `migrate` and `manager` commands in current command surface.
   - Update status to reflect encrypted vault, project workflows, run, and manager completion.

2. Update `docs/usage-spec.md`.
   - Replace stale MVP text with implemented encrypted vault status.
   - Add configuration section for `--config`, `--vault`, config YAML, recipients, identity paths, and defaults.
   - Document `shelf init`, `migrate`, `doctor`, `manager`, export/project/run behavior.
   - Add safety warnings for plaintext exports, generated `.env.local`, terminal output, browser reveal, token URL, and old plaintext store cleanup.

3. Update `docs/data-spec.md`.
   - Clarify durable encrypted vault format vs decrypted in-memory JSON model.
   - Document config and `.shelf.json` as value-free boundaries.
   - Document manager read/reveal/write boundaries and reuse of encrypted persistence.

4. Verify release requirements.
   - Confirm DOCS-01, DOCS-02, DOCS-03 are addressed in docs.
   - Confirm every v1 requirement is Complete in roadmap/requirements or explicitly deferred.
   - Run `go test ./...`.

5. Close the phase.
   - Write SUMMARY, REVIEW, VERIFICATION, CAPTURE.
   - Update roadmap, requirements, project, and state.

## Acceptance Criteria

- DOCS-01: Docs explain encrypted vault model, age recipient and identity configuration, and chezmoi-friendly workflow.
- DOCS-02: Docs separate Shelf config, `.shelf.json`, encrypted vault data, and generated/exported env files.
- DOCS-03: Docs warn about plaintext export, terminal output, browser reveal/copy actions, and old plaintext store cleanup.
- Release verification confirms all 33 v1 requirements are complete or explicitly deferred.

## Verification

```bash
go test ./...
```

Docs checks:

- README lists current commands including `migrate` and `manager`.
- Usage spec documents vault config, migration cleanup, doctor git safety, project manifest value-free behavior, manager token/reveal, and export/run value-printing rules.
- Data spec documents encrypted source of truth and non-secret config/manifest boundaries.

## Risks

- Docs can overpromise UI polish. Keep manager description to implemented search/metadata/reveal/create/update/delete behavior.
- Docs can imply generated `.env` is safe. State that generated/exported env files contain plaintext values and must not be committed.
- Docs can imply config identity paths are private material. State identity paths are paths only; private identity files themselves remain sensitive and outside config.
