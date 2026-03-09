# Attributions

This file lists all external libraries and tools used by gollery, organized by component.

## Backend (Go)

### Direct Dependencies

| Module | Version | License | Purpose |
|--------|---------|---------|---------|
| [github.com/jackc/pgx/v5](https://github.com/jackc/pgx) | v5.8.0 | MIT | PostgreSQL driver and connection pool |
| [github.com/jackc/tern/v2](https://github.com/jackc/tern) | v2.3.5 | MIT | PostgreSQL schema migrations |
| [github.com/rwcarlsen/goexif](https://github.com/rwcarlsen/goexif) | v0.0.0-20190401172101 | BSD 2-Clause | EXIF metadata extraction from JPEG images |
| [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) | v0.48.0 | BSD 3-Clause | bcrypt password hashing |
| [golang.org/x/image](https://pkg.go.dev/golang.org/x/image) | v0.36.0 | BSD 3-Clause | CatmullRom image scaling for derivatives |
| [golang.org/x/time](https://pkg.go.dev/golang.org/x/time) | v0.15.0 | BSD 3-Clause | Token-bucket rate limiting |

### Indirect Dependencies

These are transitive dependencies pulled in by the direct dependencies above.

| Module | Version | License | Pulled in by |
|--------|---------|---------|--------------|
| [dario.cat/mergo](https://github.com/darccio/mergo) | v1.0.1 | BSD 3-Clause | tern (via sprig) |
| [github.com/Masterminds/goutils](https://github.com/Masterminds/goutils) | v1.1.1 | Apache 2.0 | tern (via sprig) |
| [github.com/Masterminds/semver/v3](https://github.com/Masterminds/semver) | v3.3.0 | MIT | tern (via sprig) |
| [github.com/Masterminds/sprig/v3](https://github.com/Masterminds/sprig) | v3.3.0 | MIT | tern |
| [github.com/google/uuid](https://github.com/google/uuid) | v1.6.0 | BSD 3-Clause | tern (via sprig) |
| [github.com/huandu/xstrings](https://github.com/huandu/xstrings) | v1.5.0 | MIT | tern (via sprig) |
| [github.com/jackc/pgpassfile](https://github.com/jackc/pgpassfile) | v1.0.0 | MIT | pgx |
| [github.com/jackc/pgservicefile](https://github.com/jackc/pgservicefile) | v0.0.0-20240606120523 | MIT | pgx |
| [github.com/jackc/puddle/v2](https://github.com/jackc/puddle) | v2.2.2 | MIT | pgx |
| [github.com/mitchellh/copystructure](https://github.com/mitchellh/copystructure) | v1.2.0 | MIT | tern (via sprig) |
| [github.com/mitchellh/reflectwalk](https://github.com/mitchellh/reflectwalk) | v1.0.2 | MIT | tern (via sprig) |
| [github.com/shopspring/decimal](https://github.com/shopspring/decimal) | v1.4.0 | MIT | tern (via sprig) |
| [github.com/spf13/cast](https://github.com/spf13/cast) | v1.7.0 | MIT | tern (via sprig) |
| [golang.org/x/sync](https://pkg.go.dev/golang.org/x/sync) | v0.19.0 | BSD 3-Clause | pgx |
| [golang.org/x/text](https://pkg.go.dev/golang.org/x/text) | v0.34.0 | BSD 3-Clause | pgx |

## Frontend (JavaScript)

The frontend has **zero runtime dependencies**. It is vanilla JavaScript with no frameworks.

### Development Dependencies

| Package | Version | License | Purpose |
|---------|---------|---------|---------|
| [esbuild](https://github.com/evanw/esbuild) | ^0.24.0 | MIT | JavaScript bundler and minifier |

Tests use only built-in Node.js modules (`node:test`, `node:assert/strict`).

## Docker Images

The deployment configuration uses the following base images:

| Image | Purpose |
|-------|---------|
| `golang:1.25-bookworm` | Backend build stage |
| `node:22-bookworm-slim` | Frontend build stage |
| `debian:bookworm-slim` | Runtime container |
| `nginx` | Frontend static file serving and API reverse proxy |
| `postgres` | Optional analytics database |

## Go Standard Library

The backend makes extensive use of Go standard library packages including `net/http`, `crypto/hmac`, `crypto/sha256`, `crypto/rand`, `encoding/json`, `log/slog`, `image/jpeg`, `image/png`, `sync`, `os`, `path/filepath`, and others. These are part of the Go distribution and not listed individually.
