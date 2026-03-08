# Prompt 58 — Wire unified migrations into analytics and app

Migrate `analytics/postgres` to use the shared `migrate.Runner` from prompt 57, and wire the runner into `app.Run()`.

## Implement

1. In `analytics/postgres/postgres.go`:
   - Remove the `Migrate(ctx)` method.
   - Keep the `//go:embed migrations/*.sql` var — it's still the source of migration files.
   - Export the embedded FS: `var MigrationFS = migrationFS` (or make the existing var exported).

2. In `app.go`, update `setupAnalytics()`:
   - Create a `migrate.Runner` from the pool.
   - Call `runner.RunDir(ctx, pganalytics.MigrationFS)`.
   - Remove the `store.Migrate(ctx)` call.

3. Delete the stale duplicate: `backend/migrations/001_create_analytics.sql`.

4. Update `analytics/postgres` integration tests if they call `store.Migrate()` — replace with `migrate.New(pool).RunDir(ctx, MigrationFS)`.

## Verify

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```

## Do not

- Change the analytics SQL schema
- Change the analytics Store interface
- Add new migration files
