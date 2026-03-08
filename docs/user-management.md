# User Management

Gollery uses a file-based user store (`users.json`) for authentication. The `gollery-users` CLI tool manages this file.

## Building the tool

```bash
make backend-build
# or directly:
cd backend && go build -o gollery-users ./cmd/gollery-users
```

## Usage

```
gollery-users -file <users.json> <command> [flags]
```

The `-file` flag defaults to `users.json` in the current directory.

## Commands

### List users

```bash
gollery-users list
```

Output:
```
  admin [admin] groups=admins
  alice groups=editors,viewers
```

### Add a user

```bash
# Basic user
gollery-users add -username alice -password secret123

# Admin user with groups
gollery-users add -username alice -password secret123 -admin -groups admins,editors
```

The password is automatically bcrypt-hashed before storing.

### Remove a user

```bash
gollery-users remove -username alice
```

### Change a password

```bash
gollery-users passwd -username alice -password newpassword
```

### Set admin status

```bash
# Grant admin
gollery-users set-admin -username alice -admin

# Revoke admin
gollery-users set-admin -username alice -admin=false
```

### Set groups

```bash
# Set groups
gollery-users set-groups -username alice -groups editors,viewers

# Clear all groups
gollery-users set-groups -username alice -groups ""
```

## users.json format

The file is a JSON array of user objects:

```json
[
  {
    "username": "admin",
    "password": "$2a$10$...",
    "groups": ["admins"],
    "is_admin": true
  }
]
```

| Field | Type | Description |
|-------|------|-------------|
| `username` | string | Unique username |
| `password` | string | bcrypt hash (`$2a$` prefix) |
| `groups` | string[] | Group memberships for ACL evaluation |
| `is_admin` | bool | Full admin access (reindex, diagnostics, all albums) |

## How auth works

- The server loads `users.json` at startup (it looks for it in the working directory first, then in `content_root/`)
- Login via `POST /api/v1/auth/login` with JSON `{"username": "...", "password": "..."}`
- Sessions are HMAC-signed cookies (configured via `auth.session_secret` or `GOLLERY_SESSION_SECRET`)
- Auth endpoints are rate-limited when `auth.rate_limit` is configured in `gollery.json`
- The server does **not** hot-reload `users.json` — restart after changes

## Groups and ACLs

Groups are used in `album.json` access rules:

```json
{
  "access": {
    "view": "restricted",
    "allowed_groups": ["family", "friends"],
    "allowed_users": ["bob"]
  }
}
```

Access modes:
- `"public"` — anyone can view
- `"authenticated"` — any logged-in user
- `"restricted"` — only users/groups listed in `allowed_users`/`allowed_groups`

Admins (`is_admin: true`) bypass all access checks.
