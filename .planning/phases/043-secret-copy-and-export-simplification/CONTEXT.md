# Context: Secret Copy and Export Simplification

## Goal

Make the daily interactive secret-retrieval flow clipboard-first without breaking the existing stdout-oriented scripting contract.

## Accepted Decisions

- Add `shelf secret copy <selector> [--field FIELD]` and `shelf secret copy --env ENV [--field FIELD]`.
- `secret copy` writes only to the system clipboard and prints a value-free success confirmation.
- Clipboard failures return an error; Shelf must not fall back to printing the secret.
- Keep `secret get` unchanged for scripts and explicit stdout materialization.
- Remove redundant `env` export format from `secret export` and `project export`; retain `shell` (default) and `json`.
- Use `github.com/atotto/clipboard` outside WSL. On WSL, invoke `clip.exe` to target the Windows clipboard.

## Constraints

- Never include secret values in success, error, diagnostic, list, info, completion, or test-failure output.
- Clipboard is an explicit OS side effect, isolated behind `internal/clipboard`.
- Linux systems outside WSL require `xclip` or `xsel` through the selected backend.
- Public release work remains out of scope.
