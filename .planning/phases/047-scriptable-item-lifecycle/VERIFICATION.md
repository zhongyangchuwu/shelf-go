# Verification: Scriptable Item Lifecycle

## Automated

Passed:

```text
go test ./...
go vet ./...
go-arch-lint check
go build ./cmd/shelf
```

Coverage includes:

- atomic title, label, alias, public string, and public string-list updates;
- public-field clear behavior;
- item selector collision rejection;
- invalid public data leaving the vault byte-for-byte unchanged;
- secret field update rejection without secret values in errors;
- Cobra flag wiring, value-free update confirmation, and value-free metadata output.

## Installed-Binary Smoke Test

Passed with a temporary vault and fake value:

```text
go install ./cmd/shelf
shelf --vault <temp>/vault.json secret create demo --env DEMO_TOKEN --stdin
shelf --vault <temp>/vault.json secret update demo --title "Demo Personal" --add-label local --add-alias demo-local --field value_env --value DEMO_PERSONAL_TOKEN
shelf --vault <temp>/vault.json secret info demo-local
```

Observed: update confirmation was exactly `Updated secret item.`; title, label, alias, and public env metadata changed; the fake secret value did not appear in output.
