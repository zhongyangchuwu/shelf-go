# Storage Decision: v0.2 Initial Vault Storage

## Decision

Use JSON as the initial canonical v0.2 storage format behind repository and codec interfaces. Keep SQLite as a future repository adapter, not the v0.2 default.

## Context

Shelf is a local-first secret manager for a single developer. The vault must remain portable, git-safe after encryption, and compatible with the existing age/chezmoi mental model. v0.2 changes the data model, but it should not add database distribution or plaintext leakage risk before there is a proven need.

## Candidates Compared

### Plain JSON

Pros:

- Single portable file.
- Easy to wrap with age encryption later.
- Easy to atomically save with backup semantics.
- Simple to inspect in tests and migration fixtures.
- Good fit for expected small local vaults.

Cons:

- Whole-file load/save.
- No native query engine.
- Requires runtime indexes for ID/title/env/template lookup.

### Age-wrapped JSON

Pros:

- Preserves the current encrypted portable-file model.
- Keeps encrypted file suitable for chezmoi/git sync.
- Can avoid plaintext temp files if encryption/decryption stays in memory.

Cons:

- Whole-file conflicts remain coarse.
- Requires careful backup/migration handling.
- Needs explicit recipient/identity config integration.

### Plaintext SQLite

Pros:

- Strong query/index semantics.
- Better for history, audit, large datasets, and daemon-style concurrency.
- Mature migration tooling.

Cons:

- Plaintext at rest violates Shelf security requirements.
- WAL, journal, temp, and backup files complicate leakage boundaries.
- Worse fit for a single encrypted file managed by chezmoi/git.

### Age-wrapped SQLite

Pros:

- Keeps age as outer encryption in theory.
- Could use SQLite during a session.

Cons:

- SQLite expects random access; age is whole-file streaming encryption.
- Likely requires decrypting to temp files or in-memory DB and re-encrypting.
- WAL/journal/temp leakage risk is high unless SQLite is tightly constrained.
- Complexity is not justified by current scale.

### SQLCipher-like encrypted SQLite

Pros:

- Database semantics with encryption at rest.
- Better if Shelf later needs history, audit, complex queries, or concurrent daemon writes.

Cons:

- Native dependency and distribution burden.
- Different key management model from age.
- Requires a larger security review.
- Less aligned with the user's existing age/chezmoi workflow.

## Repository Boundary

```go
type Repository interface {
    Load(context.Context) (*Vault, error)
    Save(context.Context, *Vault) error
}

type Codec interface {
    Decode([]byte) (*Vault, error)
    Encode(*Vault) ([]byte, error)
}
```

Repository implementations own file I/O, locking, backup, and future encryption. Codec implementations own versioned serialization and migration entrypoints.

## Consequences

- v0.2 implementation should start with a JSON codec and file repository.
- Runtime code should build indexes after load instead of persisting map-shaped item storage.
- Encryption returns as an age JSON repository adapter after the plaintext v2 model and migration are validated.
- SQLite remains explicitly deferred until requirements add history/audit, large datasets, daemon concurrency, complex queries, or SQLCipher-level encrypted storage justification.

## Reconsideration Triggers

Reopen this decision if one of these becomes a v0.2 requirement:

- Thousands of items with query latency that matters.
- Durable history/audit/event log.
- Long-running daemon with concurrent writers.
- Complex query language that cannot be implemented with in-memory indexes.
- A strong reason to adopt SQLCipher or another encrypted database despite age/chezmoi tradeoffs.
