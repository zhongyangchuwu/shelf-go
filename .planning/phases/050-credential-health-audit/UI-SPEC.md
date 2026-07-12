# Specification: Credential Health Audit

**Status:** Implemented and verified.

## Product Target

`Shelf` reports local credential expiry metadata before a developer workflow fails. It remains a value-free maintenance signal: no remote provider calls, rotation, enforcement, value output, background work, or leak claims.

## Audit Semantics

- `shelf secret audit [--within DAYS] [--format text|json]` audits template/custom fields with expiry purpose metadata. The default warning horizon is 30 calendar days; `--within` must be zero or positive.
- Dates use the UTC calendar. A credential is `expired` only when its expiry date is before today; a date today through today + horizon is `expiring`; later dates are `healthy`.
- An expiry field that exists but is unset is `missing`. An expiry-purpose field that is not a valid date is `malformed`.
- Items with no expiry-purpose field are not applicable and are omitted. This keeps generic, non-expiring secrets out of the audit while making tracked-expiry absence explicit.
- Records contain item ID/title, primary secret field ID, expiry date when present, and health status only.

## CLI Contract

- Text output is a value-free table. JSON output is a value-free report for scripts.
- Exit `0`: all tracked credentials are healthy or no expiry metadata exists.
- Exit `2`: at least one expiring, missing, or malformed credential and no expired credential.
- Exit `3`: at least one expired credential (takes precedence over warning states).
- Vault/config failures retain the ordinary command failure exit behavior. `project run` and `project export` semantics remain unchanged.

## TUI Contract

- The secret catalog reloads the same 30-day audit alongside catalog metadata. It adds a text health indicator (`expired`, `due`, `missing`, `invalid`) to tracked items and a value-free health line in detail.
- The project binding workspace reloads the same audit alongside project status. A binding whose item has expiry metadata shows its health status and expiry in row/detail context.
- Health loading failure does not block catalog/binding management; the UI reports only a fixed value-free `health unavailable` state.
- Health status has text labels and never relies on color alone.

## Safety and Verification

- The app audit reads only public expiry metadata and never calls value conversion for secret fields.
- Reports, CLI output, TUI state, diagnostics, and test snapshots must exclude secret values.
- App tests cover healthy, expired, expiring, missing, malformed, absent/non-applicable, sorting, and report value-freedom.
- CLI tests cover formats, validation, and exit-code selection. TUI tests cover catalog and project health indicators and unavailable-health fallback.
