# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o linkai .

# Build and verify all packages compile
go build ./...

# Run
./linkai <command>

# Tidy dependencies
go mod tidy
```

## Architecture

This is a Go CLI for the LinkAI platform, modeled after `reference/cli/` (lark-cli). The `reference/` directory contains read-only reference implementations — do not modify them.

### Package layout

```
main.go                           → calls cmd.Execute(), returns exit code
cmd/root.go                       → root cobra command, PersistentPreRunE scope check, registers subcommands
cmd/auth/                         → auth subcommands (login/logout/status)
cmd/app/                          → app subcommands (list)
internal/auth/device_flow.go      → Device Flow HTTP calls (opaque token, no AppID/AppSecret)
internal/auth/token_store.go      → StoredToken persisted at ~/.linkai-cli/token.json
internal/api/client.go            → unified HTTP client (auth header, X-Device-ID, error unwrapping)
internal/output/print.go          → output helpers (JSON, table, success/error messages)
internal/config/config.go         → config + device_id at ~/.linkai-cli/config.json
internal/cmdutil/factory.go       → dependency injection (Config, HttpClient, IOStreams, APIClient)
internal/cmdutil/iostreams.go     → In/Out/ErrOut/IsTerminal abstraction
internal/cmdutil/scope.go         → fine-grained scope checking (HasScope, CheckScope)
```

### Auth flow (Device Flow)

`linkai auth login` implements Device Flow OAuth:

1. `POST /api/cli/auth/device` → sends `device_id` + `scope`, returns `device_code` + `verification_uri_complete`
2. CLI prints the URL to stderr; user opens it in browser and approves (can select granted scopes)
3. CLI polls `POST /api/cli/auth/token` (with `device_code` + `client_id`) until authorized or timed out
4. On success: opaque `access_token` + `refresh_token` saved to `~/.linkai-cli/token.json`, user info + `device_id` to `~/.linkai-cli/config.json`

Supports `--no-wait` (print URL + device_code, return immediately) and `--device-code` (resume polling from a prior `--no-wait` call). Error strings: `authorization_pending`, `slow_down`, `access_denied`, `expired_token`.

### Token design

- **Opaque tokens** (not JWT): stored in Redis on the server, can be revoked
- **access_token** TTL: 2 hours; **refresh_token** TTL: 7 days
- **client_id**: random secret generated at login, stored in `token.json`, required for refresh (binds refresh to device)
- **device_id**: persistent machine UUID stored in `config.json`, sent as `X-Device-ID` header on every request; server rejects tokens used from a different device
- `TokenStatus()` returns `"valid"` / `"needs_refresh"` (within 5 min of expiry) / `"expired"`

### Fine-grained scope / permission system

Scopes follow the pattern `{resource}:{action}` (e.g. `app:read`, `app:write`, `workflow:delete`).

Default scope on login: `app:read chat:read user:read workflow:read knowledge:read`

Write/delete scopes require explicit authorization via `--scope` flag or re-login.

Commands declare their required scope via a Cobra annotation:
```go
cmd.Annotations = map[string]string{cmdutil.RequiredScopeKey: "app:write"}
```

`PersistentPreRunE` in `cmd/root.go` enforces this automatically before every command runs.

### API client (`internal/api/client.go`)

`api.Client` wraps all backend HTTP calls:
- Attaches `Authorization: Bearer {token}` automatically
- `X-Device-ID` is injected by `deviceIDTransport` in the Factory
- Standard envelope: `{"code":200,"msg":"...","data":...}` — non-200 codes are returned as errors automatically
- Methods: `Get`, `Post`, `Delete`, `Stream` (SSE)

Obtain via `f.APIClient()` in command RunE functions.

### Backend API endpoints

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| POST | `/api/cli/auth/device` | public | initiate device authorization (sends device_id + scope) |
| POST | `/api/cli/auth/token` | public | poll for token (returns opaque access+refresh tokens) |
| POST | `/api/cli/auth/authorize` | user token | called by web page after user approves (sends granted_scope) |
| POST | `/api/cli/auth/refresh` | client_id + X-Device-ID | refresh access token |
| POST | `/api/cli/auth/revoke` | access token | logout / revoke tokens |
| GET | `/api/cli/app/list` | CLI token | list apps for current user; params: `key`, `type[]`, `pageNo`, `pageSize` |

Redis keys: `cli_device_auth:{device_code}`, `cli_access:{token}` (2h TTL), `cli_refresh:{token}` (7d TTL).
Tokens are bound to `device_id` and `client_id` server-side; mismatched device returns 401.

### Adding new commands

Follow the pattern in `cmd/auth/` or `cmd/app/`:
- Define an `Options` struct with a `Factory` field
- `NewCmdXxx(f *cmdutil.Factory, runF func(*Options) error) *cobra.Command` — `runF` allows test injection
- Register in `cmd/root.go` via `rootCmd.AddCommand(...)`
- Declare required scope: `cmd.Annotations = map[string]string{cmdutil.RequiredScopeKey: "resource:action"}`
- Use `f.APIClient()` for authenticated requests, `f.IOStreams` for output
- Use `output.PrintJSON` / `output.PrintTable` for formatted output
- For list commands with pagination: use `--page` / `--page-size` flags (maps to backend `pageNo`/`pageSize`)
- Truncate displayed strings by **rune count**, not byte length: `[]rune(s)[:n]` to avoid corrupting CJK/emoji

### Config & token files

- `~/.linkai-cli/config.json` — API base URL + persistent `device_id` + logged-in user info
- `~/.linkai-cli/token.json` — `access_token`, `refresh_token`, `client_id`, `scope`, expiry timestamps (Unix ms)
