# Prompt 20 — Frontend shared controller error handling

Extract duplicated error handling from controllers.

Implement:
- `core/controllers/errors.js` with a shared `handleApiError(store, err)` function
- update `AlbumController` and `AssetController` to use it
- remove the per-controller `_handleError` methods

Do not change error handling behavior.
Verify the bundle still compiles.
