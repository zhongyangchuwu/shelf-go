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
- field input through stdin, including the default primary field and an explicit public field;
- public-field flag writes and public-field clear behavior;
- item selector collision rejection;
- invalid public data leaving the vault byte-for-byte unchanged;
- secret field `--value` rejection without secret values in errors;
- value-free update confirmation and metadata output;
- removal of the obsolete `secret set` command.

## Installed-Binary Smoke Tests

Passed with a temporary vault and fake values:

```text
go install ./cmd/shelf
shelf --vault <temp>/vault.json secret create demo --env DEMO_TOKEN --stdin
shelf --vault <temp>/vault.json secret update demo --title "Demo Personal" --add-label local --add-alias demo-local --field value_env --value DEMO_PERSONAL_TOKEN
shelf --vault <temp>/vault.json secret info demo-local
```

Observed: update confirmation was exactly `Updated secret item.`; title, label, alias, and public env metadata changed; no fake secret appeared in output.

A condition-driven pseudo-terminal smoke test waited for the default `secret update demo` prompt before supplying input. The prompt appeared, the primary field changed, and the supplied fake value was not echoed.
