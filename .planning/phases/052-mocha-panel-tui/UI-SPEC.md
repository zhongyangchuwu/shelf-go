# UI Specification: Catppuccin Mocha Panel Frame

**Status:** Implemented and verified.

## Goal

Make the persistent left catalog/binding list and right detail/editor visually distinct in the style of Lazygit, while applying a single Catppuccin Mocha visual language to both `shelf tui` and `shelf project tui`.

## Scope

- Render the persistent workspace as two adjacent bordered panels.
- Keep the catalog/binding panel visible while the right panel shows detail, filters, forms, or confirmations.
- Apply the same panel frame and Mocha colors to secret and project workspaces.
- Preserve every current keyboard interaction, resize threshold, alternate-screen lifecycle, and value-safety rule.

## Layout Contract

```text
Shelf · Catalog · metadata

╭ Catalog ──────────────────────╮ ╭ Detail ─────────────────────────────────────╮
│ › OpenAI personal  api  set   │ │ OpenAI personal                              │
│   GitHub work      generic    │ │ id: item_openai                              │
│                                │ │ fields…                                      │
╰───────────────────────────────╯ ╰──────────────────────────────────────────────╯
Tab pane  a add  e edit  c copy  d delete  / search  f filters  r reload  ? help  q quit
```

- Left outer panel is approximately 36% of terminal width, minimum 34 columns.
- A one-column gutter separates panels; right panel takes the remaining width, minimum 50 columns.
- Panels have rounded borders and a persistent title row inside the frame.
- The active panel has a blue border and `›` title marker. In catalog search, the left panel is active. During a right-pane form, filter, picker, or confirmation, the right panel is active.
- Focus is never conveyed by color alone: title marker and current row cursor remain visible.
- Header and footer form a calm full-width frame around the panels; panels have a fixed shared height for the current terminal size.

## Catppuccin Mocha Palette

Use the fixed dark Mocha flavor, rather than adapting to terminal background color:

| Role | Mocha color |
| --- | --- |
| Workspace/panel base | `#1e1e2e` base |
| Header/footer | `#181825` mantle |
| Inactive panel border | `#45475a` surface1 |
| Active panel border and focus | `#89b4fa` blue |
| Panel title/accent | `#cba6f7` mauve |
| Primary text | `#cdd6f4` text |
| Muted text | `#bac2de` subtext1 |
| Success/status | `#a6e3a1` green |
| Warning | `#f9e2af` yellow |
| Destructive/error | `#f38ba8` red |

## Safety and Non-goals

- No secret reveal, secret-colored indicators, or changes to the renderable-data boundary.
- No custom theme selector, light flavor, narrow layout, mouse support, or action/keymap changes.
- Catppuccin is visual styling only; it must not add a persistence/configuration surface.

## Acceptance

- At 120×32, secret and project workspace renders contain two distinct rounded panel frames with title markers and one gutter.
- Switching secret focus between catalog and detail changes only the active panel frame/title treatment; list/detail selection behavior remains unchanged.
- Secret forms, filters, delete confirmation, and project picker/form/confirmation use the right panel while leaving left context visible.
- Styles use the declared Mocha colors and no longer branch on terminal background color.
- Existing value-free rendering tests still pass; regression tests assert the two-panel structure and active-panel marker.
