# Running Gollery Locally

This guide explains how to run gollery locally using Docker Compose.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) (v2+)
- A directory of images to use as gallery content

## Architecture

The Docker Compose setup runs three services:

| Service     | Purpose                              | Port |
|-------------|--------------------------------------|------|
| `galleryd`  | Go backend API server                | 8080 |
| `nginx`     | Serves frontend static files, proxies API requests | 8090 |
| `postgres`  | PostgreSQL for optional analytics    | 5432 (internal only) |

The **nginx** container serves the frontend SPA and proxies `/api/*` and `/healthz` requests to the backend.

## Quick Start

### 1. Build the frontend

```bash
make frontend-build
```

This produces `frontend/dist/` with the bundled JavaScript, HTML, and CSS.

### 2. Prepare content

Create a content directory with your images. Each directory that should appear as an album needs an `album.json` file:

```bash
mkdir -p sample-content/MyAlbum
cp /path/to/your/images/* sample-content/MyAlbum/
```

Create `sample-content/album.json` (root album):

```json
{
  "title": "My Gallery",
  "description": "A personal image gallery",
  "access": {
    "view": "public"
  },
  "derivatives": {
    "thumbnail_sizes": [200, 400],
    "preview_sizes": [800, 1600]
  }
}
```

Create `sample-content/MyAlbum/album.json`:

```json
{
  "title": "My Album",
  "description": "Album description"
}
```

Supported image formats: `.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`.

### 3. Create a users file

The easiest way is with the `gollery-users` tool:

```bash
# Build the tool first
make backend-build

# Create an admin user (creates users.json if it doesn't exist)
./backend/gollery-users add -username admin -password admin -admin -groups admins
```

This automatically bcrypt-hashes the password and writes `users.json`. See [docs/user-management.md](user-management.md) for the full command reference.

Alternatively, create `users.json` manually with pre-hashed passwords:

```json
[
  {
    "username": "admin",
    "password": "$2a$10$YOUR_BCRYPT_HASH_HERE",
    "groups": ["admins"],
    "is_admin": true
  }
]
```

### 4. Start with Docker Compose

```bash
docker-compose up --build
```

The gallery will be available at **http://localhost:8090**.

The API is also directly accessible at `http://localhost:8080/api/v1/...`.

### 5. Stop

```bash
docker-compose down
```

To also remove volumes (database data, cache):

```bash
docker-compose down -v
```

## Configuration

### gollery.json

The server configuration file. Key fields:

| Field | Description | Default |
|-------|-------------|---------|
| `content_root` | Path to content directory (inside container) | `/data/content` |
| `cache_dir` | Path to derivative cache (inside container) | `/data/cache` |
| `listen_addr` | Backend listen address | `:8080` |
| `auth.provider` | Auth provider (`"static"` for file-based) | — |
| `auth.session_secret` | HMAC session signing key | — |
| `analytics.enabled` | Enable PostgreSQL analytics | `false` |

### Environment variable overrides

| Variable | Overrides |
|----------|-----------|
| `GOLLERY_LISTEN_ADDR` | `listen_addr` |
| `GOLLERY_POSTGRES_DSN` | `analytics.postgres_dsn_env` |
| `GOLLERY_SESSION_SECRET` | `auth.session_secret` |

### Volume mounts

The docker-compose.yml mounts:

- `./sample-content` → `/data/content` (your images)
- `./gollery.json` → `/etc/gollery/gollery.json` (server config)
- `./users.json` → `/etc/gollery/users.json` (user credentials)
- `./frontend/dist` → nginx html root (frontend files)
- Named volume `gallery-cache` → `/data/cache` (derivative images)
- Named volume `pgdata` → PostgreSQL data

## Content Structure

```
sample-content/
├── album.json              ← root album config
├── Vacation/
│   ├── album.json          ← album config (inherits from root)
│   ├── beach.jpg
│   └── sunset.png
└── Family/
    ├── album.json
    └── portrait.jpg
```

Albums are directories. Any directory with an `album.json` is published. Child directories inherit their parent's config unless `"inherit": false` is set.

## Troubleshooting

- **Frontend shows blank page**: Ensure `make frontend-build` was run and `frontend/dist/` exists.
- **API returns 404**: The backend only serves `/api/v1/*` and `/healthz`. Frontend is served by nginx.
- **Auth not working**: Ensure `users.json` exists and passwords are bcrypt-hashed. The server looks for `users.json` in the working directory first, then in the content root.
- **No albums appear**: Each album directory needs a valid `album.json` file.
- **PostgreSQL connection errors**: Wait for the health check — the backend depends on `service_healthy` condition.
