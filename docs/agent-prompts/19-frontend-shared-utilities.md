# Prompt 19 — Frontend shared utilities

Extract duplicated code in the frontend `ui-default` layer.

Implement:
- `ui-default/util/html.js` with shared `esc()` function for XSS-safe escaping
- update all views and components to import from the shared utility
- remove the per-file `esc()` definitions

Do not change view behavior or styling.
Verify the bundle still compiles.
