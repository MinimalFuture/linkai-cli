# Admin (Auth / Account / Credits / Models)

Authentication, account info, credits management, and model listing.

## Authentication

### Login

```bash
linkai auth login
linkai auth login --scope "app:read chat:write knowledge:read db:read"
```

Opens a browser URL for authorization. The CLI waits until the user approves.

Default scope: `app:read chat:read user:read workflow:read knowledge:read`

To request additional permissions (e.g., write/delete), specify `--scope` explicitly.

| Flag | Type | Description |
|---|---|---|
| `--scope` | string | Space-separated scopes to request |
| `--no-wait` | bool | Print URL and device code, return immediately |
| `--device-code` | string | Resume polling from a prior `--no-wait` call |
| `--json` | bool | Structured JSON output |

If the user is already logged in with a valid token and the requested scopes are all covered, login is skipped. To upgrade scopes, pass a `--scope` that includes new ones.

### Logout

```bash
linkai auth logout
```

Revokes the token server-side and clears local storage.

### Status

```bash
linkai auth status
linkai auth status --json
```

Shows login state, username, token expiry, and granted scopes.

## Account info

```bash
linkai account info
linkai account info --json
```

**Required scope**: `user:read`

Shows the user's name, remaining credits, and plan version.

## Models

```bash
linkai model list
linkai model list --type LLM
linkai model list --supplier openai
```

**Required scope**: (uses default auth)

| Flag | Type | Description |
|---|---|---|
| `--type` | string | Filter by model type (e.g., `LLM`, `EMBEDDING`) |
| `--supplier` | string | Filter by supplier (e.g., `openai`, `claude`) |
| `--json` | bool | JSON output |

## Credits / Score

### List credit packages

```bash
linkai score list
linkai score list --json
```

**Required scope**: `score:read`

### Purchase credits

```bash
linkai score buy
linkai score buy --product <id> --pay wechat
linkai score buy --product <id> --pay alipay
```

**Required scope**: `score:write`

| Flag | Type | Default | Description |
|---|---|---|---|
| `--product` | string | | Product ID (skip interactive selection) |
| `--pay` | string | `wechat` | Payment channel: `wechat` or `alipay` |
| `--agent` | bool | | Agent mode: return QR URL instead of ASCII QR |
| `--json` | bool | | JSON output (agent mode) |

### Purchase history

```bash
linkai score orders
linkai score orders --page 2 --page-size 20
```

**Required scope**: `score:read`

| Flag | Type | Default | Description |
|---|---|---|---|
| `--page` | int | 1 | Page number |
| `--page-size` | int | 10 | Items per page |
| `--json` | bool | | JSON output |

## Scope reference

| Scope | Commands |
|---|---|
| `app:read` | `app list`, `app detail` |
| `user:read` | `account info` |
| `chat:write` | `chat` |
| `knowledge:read` | `knowledge list/files/search` |
| `knowledge:write` | `knowledge create` |
| `knowledge:delete` | `knowledge delete` |
| `db:read` | `database list/tables/describe/exec` (SELECT) |
| `db:write` | `database exec` (INSERT/UPDATE/DELETE) |
| `image:write` | `image gen` |
| `video:write` | `video gen` |
| `audio:write` | `audio speech` |
| `plugin:read` | `plugin list/detail` |
| `plugin:run` | `plugin exec` |
| `workflow:read` | `workflow list` |
| `workflow:run` | `workflow run` |
| `score:read` | `score list/orders` |
| `score:write` | `score buy` |
