# Manual vault migrations

Shelf does **not** migrate a vault in place. Make and verify a copy before replacing the configured plaintext vault.

## V2 to V3

### 1. Back up the plaintext vault

```bash
cp -p ~/.local/share/shelf/vault.json ~/.local/share/shelf/vault.v2.json
chmod 600 ~/.local/share/shelf/vault.v2.json
```

If the vault is managed by chezmoi or another encryption tool, back up and edit the encrypted source according to that tool's workflow. Do not commit a plaintext copy.

### 2. Change the document version

Set the top-level `version` to `3`.

V3 item field instances contain only `id` and `value`. Move the old per-item `label`, `type`, `sensitivity`, `purpose`, and `env_for` data out of every item field. Field definitions now belong to the item's template.

### 3. Rename structural fields

Use these direct mappings for built-in templates:

| V2 field | V3 field |
| --- | --- |
| `value_env` | `env_value` |
| `credential_env` | `env_credential` |
| `expires_at` for an API credential | `validity_credential` |

`validity_<field>` must be `unknown`, `permanent`, or a `YYYY-MM-DD` date. Choose `unknown` when the old expiry was absent or uncertain; use `permanent` only when the credential does not expire.

An `env_<field>` value is the default environment name for the exact suffix field. It may map a secret or public `string`, `concealed`, or `url` field. Delete mappings you do not want `shelf env export` to materialize.

### 4. Recreate custom template definitions

For every custom template, move the field shape into its `templates[].fields` entry and set `primary_secret` to one secret field. Keep each item field as only:

```json
{"id":"field_name","value":"field value"}
```

A structural definition must have a real atomic target in the same template:

- `env_<field>`: public `string`, target type `string`, `concealed`, or `url`.
- `validity_<field>`: required public `string`, target a secret field.

Every stored structural field must also have its target stored on that item. This prevents dangling mappings.

### 5. Validate the copy without materializing values

Point Shelf at the migrated copy and run a value-free command:

```bash
shelf --vault ~/.local/share/shelf/vault.v3.json secret list --long
shelf --vault ~/.local/share/shelf/vault.v3.json secret audit
```

Correct all schema errors before replacing the configured vault. Then verify projections and project bindings deliberately:

```bash
shelf --vault ~/.local/share/shelf/vault.v3.json env export --label <label>
shelf --vault ~/.local/share/shelf/vault.v3.json project status
```

Only replace the active vault after these checks pass. Keep the V2 backup until the new vault has been used successfully.

## V3 to V4: aliases become descriptions

Schema version 4 removes item selector aliases. An item is now selected only by its stable `id` or exact `title`; the user-facing replacement is the optional `description` field. Strict V4 decoding rejects the old `aliases` and `notes` keys.

1. Back up the V3 vault using the same `0600` procedure above.
2. Set the top-level `version` to `4`.
3. For every item, remove `aliases`. Copy its text into `description`; when both old `notes` and aliases exist, retain both texts in `description`. Do not alter `id`, `title`, `template`, fields, labels, or field values.
4. Replace any scripts that selected an item through an alias with its stable `id` or title.

For example, only the safe metadata changes:

```json
{"id":"item_openai","title":"OpenAI personal","description":"personal, llm"}
```

Validate the V4 copy without materializing values before promotion:

```bash
shelf --vault ~/.local/share/shelf/vault.v4.json secret list --long
shelf --vault ~/.local/share/shelf/vault.v4.json secret audit
```

Then deliberately verify any label projection and project bindings you use. Keep the V3 backup until the V4 vault has been used successfully.