# Prompt 47 — Backend API pagination

Add pagination to album asset listings.

Implement:
- `GET /api/v1/albums/{id}` accepts optional `?offset=N&limit=N` query parameters
- default limit: 100, max limit: 500
- return `total_assets` count in the response alongside the paginated `assets` array
- `GET /api/v1/albums/root` gets the same treatment
- tests for default pagination, custom offset/limit, and boundary cases

Do not change the album children listing (children are typically few).
