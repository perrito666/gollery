# Prompt 31 — Backend album path-based lookup

Add the missing album path query endpoint.

Implement:
- `GET /api/v1/albums?path=/relative/path` — look up an album by its filesystem path
- build a path-to-album index in `Server.SetSnapshot()`
- ACL enforcement on the result
- return 404 if path is not found or not published
- tests

Do not change existing album-by-ID behavior.
