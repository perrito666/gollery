# Prompt 75 — Search index: types and interface

Define the interface for a read-only search index that mirrors filesystem content.

## Design constraints

- The filesystem remains the source of truth.
- The search index is a **read-only mirror** rebuilt from snapshots.
- The index is **optional** — the gallery works without it.
- Start with an in-memory implementation; PostgreSQL or SQLite can come later.

## Implement

1. Create `backend/internal/search/search.go`:

```go
package search

type Query struct {
    Text       string   // free-text search
    AlbumPath  string   // filter to album subtree (empty = all)
    SortBy     string   // "filename", "date", "relevance"
    SortOrder  string   // "asc", "desc"
    Offset     int
    Limit      int
}

type Result struct {
    AssetID   string
    AlbumID   string
    AlbumPath string
    Filename  string
    Title     string
    Score     float64  // relevance score (1.0 = exact match)
}

type ResultSet struct {
    Results []Result
    Total   int        // total matching count (before pagination)
}

// Index is the search interface. Implementations must be safe for concurrent reads.
type Index interface {
    // Rebuild replaces the index contents from a snapshot.
    Rebuild(snap *domain.Snapshot) error

    // Search returns matching assets.
    Search(ctx context.Context, q Query) (*ResultSet, error)

    // Close releases resources.
    Close() error
}
```

2. Create `backend/internal/search/memory.go` — an in-memory implementation:
   - `Rebuild` stores all assets in a slice.
   - `Search` does case-insensitive substring matching on filename and album title.
   - Supports pagination via `Offset`/`Limit`.
   - Supports sorting by filename.

3. Create `backend/internal/search/memory_test.go`:
   - Test basic text search matches filename.
   - Test album path filtering.
   - Test pagination.
   - Test empty query returns all.
   - Test rebuild replaces previous data.

## Verify

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```

## Do not

- Add API routes (next prompt)
- Add PostgreSQL/SQLite implementation
- Wire into app.go
