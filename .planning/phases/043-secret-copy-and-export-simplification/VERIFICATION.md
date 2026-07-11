# Verification: Secret Copy and Export Simplification

## Automated

Passed:

```text
go test ./...
go vet ./...
go-arch-lint check
go build ./cmd/shelf
```

Coverage includes:

- `secret copy` copies the resolved value through an injected clipboard writer and prints only `Copied secret to clipboard.`
- Clipboard failures do not print the secret.
- WSL clipboard routing chooses `clip.exe` from `PATH` or the standard Windows-system fallback path.
- `secret export` and `project export` reject removed `env` format.
- Shell and JSON export remain covered.

## Installed-binary smoke test

Passed with a temporary vault and a fake value:

```text
go install ./cmd/shelf
shelf --vault <temp>/vault.json secret create demo --env DEMO_TOKEN --stdin
shelf --vault <temp>/vault.json secret copy --env DEMO_TOKEN
```

Observed output:

```text
Copied secret to clipboard.
```

The WSL fallback bridge `/mnt/c/Windows/System32/clip.exe` was exercised successfully. No real vault value was copied or printed during the smoke test.
