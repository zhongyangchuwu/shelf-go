# UI Spec: Shelf Web Manager

## Subject

Shelf Web Manager is a local vault workbench for solo developers editing project secrets on their own machine. Its single job is to make secret records understandable and editable without making values ambiently visible.

## Design Thesis

The interface should feel like opening a compact encrypted field notebook: deliberate, quiet, local, and precise. The memorable element is a **vault rail**: a narrow left column that groups secrets as path rows with tag chips and value state, while the right side behaves like an inspection bench where the selected record can be edited, revealed, copied, or deleted.

This avoids generic SaaS admin language. The user is not managing a team dashboard; they are maintaining their own local working vault.

## Visual System

### Color Tokens

| Token | Hex | Use |
|-------|-----|-----|
| `ink` | `#18211f` | Primary text and high-emphasis icons |
| `slate` | `#40504c` | Secondary text, quiet metadata |
| `paper` | `#f6f2e8` | App background; resembles notebook paper without using the generic cream/terracotta look |
| `panel` | `#fffaf0` | Cards, forms, editor surfaces |
| `seal` | `#2f5d50` | Primary actions, active selection, safe state |
| `wax` | `#b45f3a` | Destructive/warning emphasis used sparingly |
| `cipher` | `#d7c7a4` | Rules, inactive chips, disabled borders |
| `glow` | `#dbe8d2` | Success and copied feedback |

### Type

No external fonts in v0.1.1. Use system faces for single-binary/offline distribution:

- Display: `ui-serif, Georgia, Cambria, "Times New Roman", serif` for page title and selected secret path. Used sparingly.
- Body: `ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif` for labels, controls, forms, body copy.
- Utility/data: `ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace` for paths, env names, JSON values, and status pills.

### Layout

Desktop:

```text
┌──────────────────────────────────────────────────────────────────────────┐
│ Vault header: local vault, safety state, search                          │
├──────────────────────────────┬───────────────────────────────────────────┤
│ Vault rail                   │ Inspection bench                          │
│ - Add secret                 │ ┌ Path / env / tags / description ┐       │
│ - Search/filter              │ │ Value panel: locked/revealed     │       │
│ - Tag chips                  │ │ Editor actions                   │       │
│ - Secret rows                │ └──────────────────────────────────┘       │
└──────────────────────────────┴───────────────────────────────────────────┘
```

Mobile:

```text
┌──────────────────────┐
│ Header + search      │
├──────────────────────┤
│ Secret list          │
├──────────────────────┤
│ Selected editor      │
└──────────────────────┘
```

### Signature Element

The selected secret path is rendered as a **split seal**:

```text
providers/openai  :  token
```

The group path and key are visually separated, matching Shelf's `group_path:key` grammar. This makes the data model visible without explaining implementation internals.

## UX Requirements

### List and Search

- The list fetches `GET /api/secrets?q=<query>`.
- Rows show path, env, description excerpt, tags, and `value set` state.
- Rows never contain secret values.
- Search matches path, env, description, and tags.
- Empty state copy: `No secrets match this view. Add a secret or clear the search.`

### Add Secret

- Primary action: `Add secret`.
- Opens the editor in create mode with empty fields:
  - Path
  - Value
  - Env name
  - Description
  - Tags
- Primary button: `Save secret`.
- Save uses `POST /api/secrets`.
- On success, refresh list and select the saved record.
- On duplicate, show exact server error and suggest using edit mode.

### Edit Secret

- Selecting a row loads metadata already present in the list.
- Editing path, value, env, description, and tags happens in the inspection bench.
- Existing values are not auto-revealed. Value field starts locked/blank with copy: `Leave value unchanged, or reveal before replacing.`
- For v0.1.1 implementation, editing metadata without replacing the value requires either:
  - a detail endpoint that returns metadata plus `value_set`, or
  - explicit reveal before saving the full record.
- Safer Phase 18 target: add a detail/read endpoint for metadata and keep value changes explicit.
- Rename uses an explicit old-path field or endpoint support so it cannot accidentally overwrite another secret.

### Delete Secret

- Destructive action label: `Delete secret`.
- Requires a confirmation dialog with path repeated.
- Uses `DELETE /api/secrets?path=<path>`.
- On success, refresh list and clear selection.

### Reveal, Hide, Copy

- Reveal and copy are explicit actions.
- Reveal uses `POST /api/reveal` instead of GET so secret values are not coupled to browser prefetch/history/cache behavior.
- Revealed value is displayed in a monospace field with `Hide value` and `Copy value` controls.
- Copy may call reveal internally only after user activates `Copy value`.
- Values are held only in in-memory JavaScript variables and DOM text while revealed; never `localStorage` or `sessionStorage`.
- Hide removes the value from DOM text and in-memory selection state.
- Auto-hide after 60 seconds is preferred; respect `prefers-reduced-motion` for any visual countdown.

### Tags

- Tags are edited as a comma/space-separated chip input in WebUI.
- Tag names follow existing store token rules.
- List rows show chips.
- Clicking a tag filters the search to that tag text for v0.1.1; richer tag query syntax can wait.

## API Contract for Phase 18

Existing endpoints may be extended, but behavior must remain compatible with existing tests unless the test is intentionally updated for safer behavior.

### `GET /`

- Serves the app shell.
- If `token` query parameter is present and valid, set strict HttpOnly cookie and redirect to `/` without token in URL.
- Response headers:
  - `Content-Type: text/html; charset=utf-8`
  - `Cache-Control: no-store`

### `GET /api/secrets?q=<query>`

- Returns `{"secrets":[...]}`.
- Each item: `path`, `env`, `description`, `tags`, `value_set`.
- Must not include `value`.
- Response headers include `Cache-Control: no-store`.

### `GET /api/secret?path=<path>`

- Returns metadata for one secret without value.
- Same no-value contract as list item.
- Response headers include `Cache-Control: no-store`.

### `POST /api/reveal`

- Accepts JSON body `{"path":"group:key"}`.
- Returns `{"path":"group:key","value":"..."}` only for explicit reveal/copy.
- Requires valid Origin like all unsafe methods.
- Response headers include `Cache-Control: no-store`.
- `GET /api/reveal` should be removed or return `405` after tests are updated.

### `POST /api/secrets`

- Creates a secret.
- Rejects duplicates unless explicitly forced.
- Accepts `path`, `value`, `env`, `description`, `tags`, and optional `force`.

### `PUT /api/secrets`

- Updates an existing secret.
- Supports rename with `old_path` + `path` to avoid delete/create ambiguity.
- If `value` is omitted, preserve the existing value.
- If `value` is present, validate and replace it.

### `DELETE /api/secrets?path=<path>`

- Deletes one secret after UI confirmation.

## Security Contract

- Preserve loopback Host validation.
- Preserve token validation through query, header, or strict HttpOnly cookie.
- Query-token access redirects to remove token from visible URL.
- Preserve Origin validation for unsafe methods.
- Treat reveal as unsafe and use POST.
- All app/API responses use `Cache-Control: no-store`.
- No secret values in list/search/detail responses.
- No secret values in URLs.
- No secret values in persistent browser storage.
- No CDN or externally fetched assets.

## Copy and Writing

- Use active labels: `Add secret`, `Save secret`, `Reveal value`, `Copy value`, `Hide value`, `Delete secret`.
- Errors state what failed and how to fix it: `Secret path is required`, `Env name must look like APP_TOKEN`, `Secret already exists`.
- Avoid apology and marketing tone.
- Avoid SaaS/team language such as workspace, organization, invite, member, role.

## Accessibility and Responsiveness

- Keyboard focus visible on all buttons, inputs, list rows, and dialogs.
- Forms use labels, not placeholder-only fields.
- Destructive action uses a real confirmation dialog/panel with focus management where practical.
- `prefers-reduced-motion` disables non-essential transitions.
- Mobile layout stacks list above editor and keeps primary actions reachable.

## Build and Asset Strategy

- Phase 18 should prefer first-party CSS and JS embedded in Go templates or `embed.FS`.
- No npm toolchain is required for v0.1.1 unless explicitly approved later.
- TailAdmin/daisyUI are visual references, not required runtime dependencies.
- The implementation should remain a single Go binary and pass GoReleaser snapshot builds.

## Acceptance

- The spec covers all Phase 17 success criteria.
- Phase 18 implementation can proceed without selecting a separate framework.
- Open questions are resolved as:
  - CSS: first-party embedded CSS using TailAdmin/daisyUI as visual references.
  - JS: small first-party JavaScript, no htmx for v0.1.1.
  - Layout: left vault rail plus right inspection bench.
  - Reveal/copy: separate explicit actions; copy can reveal internally only after click.
  - Auto-hide: 60 seconds preferred.
  - Value editing: metadata can be edited without revealing; replacing value is explicit.
