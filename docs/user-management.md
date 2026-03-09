# User Management

Gollery uses a file-based user store (`users.json`) for authentication. The `gollery-users` CLI tool manages users, creates album configs, and provides validated editing.

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

## User commands

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

### Set groups (replace all)

```bash
# Replace all groups
gollery-users set-groups -username alice -groups editors,viewers

# Clear all groups
gollery-users set-groups -username alice -groups ""
```

### Add groups (incremental)

```bash
# Add groups without removing existing ones
gollery-users add-groups -username alice -groups newgroup1,newgroup2
```

Groups that the user already belongs to are silently skipped.

### Remove groups (incremental)

```bash
# Remove specific groups without affecting others
gollery-users remove-groups -username alice -groups oldgroup
```

Groups not present are silently skipped.

## Album commands

### Create album.json

```bash
# Minimal public album
gollery-users init-album -dir /path/to/album -title "My Album"

# Restricted album with allowed users
gollery-users init-album -dir /path/to/album -title "Private" \
  -access restricted -allowed-users alice,bob

# Authenticated-only album
gollery-users init-album -dir /path/to/album -title "Members Only" \
  -access authenticated

# Full example with all flags
gollery-users init-album -dir /path/to/album -title "Family Photos" \
  -description "Photos from the trip" \
  -access restricted \
  -allowed-users alice,bob \
  -allowed-groups family \
  -admins alice \
  -no-inherit
```

Flags:

| Flag | Required | Description |
|------|----------|-------------|
| `-dir` | yes | Directory to create album.json in (created if missing) |
| `-title` | yes | Album title |
| `-description` | no | Album description |
| `-access` | no | Access mode: `public`, `authenticated`, or `restricted` |
| `-allowed-users` | no | Comma-separated allowed users (for restricted mode) |
| `-allowed-groups` | no | Comma-separated allowed groups (for restricted mode) |
| `-admins` | no | Comma-separated album admins |
| `-no-inherit` | no | Disable config inheritance from parent album |

The command refuses to overwrite an existing album.json. Use `edit album` to modify.

## Validated editing (visudo-style)

Edit files in `$EDITOR` with validation on save. If validation fails, you're prompted to re-edit or abort (Ctrl+C). The original file is only overwritten once validation passes.

### Edit users.json

```bash
gollery-users edit users
# or with explicit path:
gollery-users edit users -file /etc/gollery/users.json
```

Validates:
- Valid JSON array
- No empty usernames
- No duplicate usernames
- No empty password hashes

### Edit album.json

```bash
gollery-users edit album -file /path/to/album.json
```

Validates:
- Valid JSON object
- Access mode (if set) is one of `public`, `authenticated`, `restricted`

### Requirements

The `$EDITOR` environment variable must be set. If unset, the command fails with an error rather than guessing.

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

- The server loads `users.json` at startup (searches: next to config file, working directory, then `content_root/`)
- Explicit path via `auth.users_file` in `gollery.json`
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
