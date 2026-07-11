# Plan: Secret Copy and Export Simplification

## Scope

1. Add a clipboard adapter and inject it at the CLI boundary.
2. Add `shelf secret copy` without changing `secret get`.
3. Remove `env` export support from CLI flags, app switches, utility code, documentation, and tests.
4. Document platform behavior and preserve architecture boundaries.

## Acceptance Criteria

- `shelf secret copy <selector>` copies the resolved primary secret field and prints exactly a value-free confirmation.
- `shelf secret copy --env ENV` resolves by public env name.
- Clipboard failures return a contextual error without outputting the secret.
- `secret get` retains its existing stdout output.
- `secret export` and `project export` accept only `shell` and `json`; `env` fails without secret leakage.
- Clipboard adapter routes WSL to `clip.exe` and other platforms to the external library.
- Focused tests, full tests, vet, architecture lint, and a temporary-vault clipboard smoke test pass.
