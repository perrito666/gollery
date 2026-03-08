# Prompt 76 — Search API route

Add a search endpoint and wire the search index into the API server.

## Implement

1. Add search index to the `Deps` struct in `api.go`:
   ```go
   SearchIndex search.Index // optional — nil means search disabled
   ```

2. Add route in `Handler()`:
   ```
   GET /api/v1/search?q={text}&album={path}&sort={field}&order={asc|desc}&offset={n}&limit={n}
   ```

3. Create `backend/internal/api/search.go` with `handleSearch`:
   - If `SearchIndex` is nil, return 404 with message "search not available".
   - Parse query parameters into `search.Query`.
   - Default limit: 50, max limit: 200.
   - Call `SearchIndex.Search()`.
   - Filter results through ACL: only include assets the requesting principal can view.
   - Return JSON `{ results: [...], total: N }`.

4. In `app.go`, create the search index and rebuild it:
   ```go
   idx := memory.New()
   idx.Rebuild(snapshot)
   ```
   Pass into `Deps`. Also rebuild the search index in `doReindex()`.

5. Add tests in `backend/internal/api/search_test.go`:
   - Search returns matching assets.
   - Search respects ACL (restricted assets hidden from anonymous).
   - Search with nil index returns 404.
   - Pagination works.

## Verify

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```

## Do not

- Add the search route to the OpenAPI spec (do separately if spec exists)
- Add full-text search — substring matching is sufficient for now
