# Prompt 57 — Create unified migration package

Extract migration running into a shared package so all subsystems use one migration runner.

## Current state

`analytics/postgres` embeds its own migrations and runs them via `store.Migrate(ctx)`. If another subsystem needs PostgreSQL tables, it would duplicate this pattern. There is also a stale duplicate at `backend/migrations/001_create_analytics.sql` that should be removed.

## Implement

1. Create `backend/internal/migrate/migrate.go` with:

```go
package migrate

// Runner manages tern v2 migrations against a pgx pool.
type Runner struct {
    pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Runner

// RunDir runs all migrations from an embed.FS.
// The FS should contain files like 001_name.sql at its root.
func (r *Runner) RunDir(ctx context.Context, migrations fs.FS) error

// Version returns the current migration version.
func (r *Runner) Version(ctx context.Context) (int32, error)
```

2. Implement `RunDir` using the same tern v2 logic currently in `analytics/postgres/postgres.go` (`Migrate` method). Use the same schema version table name (`public.schema_version`).

3. Add tests for `Runner`:
   - `TestRunDir` — use `//go:build integration` tag, run a simple migration that creates a table, verify it exists.
   - `TestRunDirIdempotent` — run the same migrations twice, verify no error.

## Verify

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```

## Do not

- Change `analytics/postgres` yet (next prompt)
- Remove the stale `backend/migrations/` file yet (next prompt)
- Change the migration file format
