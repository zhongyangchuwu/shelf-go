# Capture: Secret Copy and Export Simplification

## Durable Decisions

- `secret copy` owns interactive clipboard retrieval; `secret get` remains the stdout/script command.
- Clipboard failure is terminal and value-free. Never fall back to terminal output.
- WSL tries `clip.exe` first, then `wl-copy` if the Windows bridge is unavailable or fails; neither failure can cause terminal output of the secret.
- `env` export was removed because `shell` already covers sourceable environment assignment and `json` covers structured automation.

## Code Shape

- `internal/clipboard` is the sole system-clipboard boundary.
- `internal/cli` injects the clipboard writer into secret commands, allowing CLI behavior tests without accessing the real clipboard.
- `github.com/atotto/clipboard` is the minimal non-WSL backend; outside WSL on Linux it requires `xclip` or `xsel`.

## Next Work

- Continue `shelf secret` UX improvements or consolidate duplicated item/field/env resolution from the v0.2.0 review.
- Keep public release work deferred until a future v1.0.0 milestone.
