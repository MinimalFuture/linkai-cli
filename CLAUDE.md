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

# Run tests
go test ./...

# Tidy dependencies
go mod tidy
```

## Architecture

This is a Go CLI for the LinkAI platform.

### Package layout

```
main.go                           → calls cmd.Execute(), returns exit code
cmd/root.go                       → root cobra command, PersistentPreRunE scope check, registers subcommands
cmd/auth/                         → auth subcommands (login/logout/status)
cmd/app/                          → app subcommands (list/detail)
cmd/account/                      → account subcommands (info)
cmd/chat/                         → chat command (streaming SSE, multi-turn with --session)
cmd/knowledge/                    → knowledge subcommands (list/create/delete/files/search)
cmd/model/                        → model subcommands (list)
cmd/database/                     → database subcommands (list/tables/describe/exec)
cmd/image/                        → image subcommands (gen)
cmd/video/                        → video subcommands (gen — with built-in polling)
cmd/audio/                        → audio subcommands (speech — TTS with optional --output download)
cmd/plugin/                       → plugin subcommands (list/detail/exec)
cmd/workflow/                     → workflow subcommands (list/run)
cmd/score/                        → score subcommands (list/buy/orders — with QR code payment polling)
internal/auth/device_flow.go      → Device Flow HTTP calls + token refresh + server-side revoke
internal/auth/token_store.go      → StoredToken persisted at ~/.linkai/token.json
internal/api/client.go            → unified HTTP client (auth header, X-Device-ID, error unwrapping, non-JSON detection)
internal/output/print.go          → output helpers (JSON, table, success/error messages)
internal/output/dryrun.go         → dry-run output helper (prints request without executing)
internal/config/config.go         → config + device_id at ~/.linkai/config.json
internal/cmdutil/factory.go       → dependency injection (Config, HttpClient, IOStreams, APIClient, auto token refresh)
internal/cmdutil/iostreams.go     → In/Out/ErrOut/IsTerminal abstraction
internal/cmdutil/scope.go         → fine-grained scope checking (HasScope, CheckScope)
internal/cmdutil/transport.go     → retry transport (exponential backoff on 502/503/504)
```

### Auth flow (Device Flow)

`linkai auth login` implements Device Flow OAuth:

1. `POST /api/cli/auth/device` → sends `device_id` + `scope`, returns `device_code` + `verification_uri_complete`
2. CLI prints the URL to stderr; user opens it in browser and approves (can select granted scopes)
3. CLI polls `POST /api/cli/auth/token` (with `device_code`) until authorized or timed out
4. On success: opaque `access_token` + `refresh_token` saved to `~/.linkai/token.json`, user info + `device_id` to `~/.linkai/config.json`

Supports `--no-wait` (print URL + device_code, return immediately) and `--device-code` (resume polling from a prior `--no-wait` call). Error strings: `authorization_pending`, `slow_down`, `access_denied`, `expired_token`.

### Token lifecycle

- **Opaque tokens** (not JWT): stored in Redis on the server, can be revoked
- **access_token** TTL: 2 hours; **refresh_token** TTL: 7 days
- **device_id**: persistent machine UUID stored in `config.json`, sent as `X-Device-ID` header on every request; server binds tokens to device_id and rejects mismatched requests
- `TokenStatus()` returns `"valid"` / `"needs_refresh"` (within 5 min of expiry) / `"expired"`
- **Auto-refresh**: `factory.go` calls `RefreshAccessToken()` when status is `needs_refresh`, updates the stored token on disk, and falls back to the current (still valid) token if refresh fails
- **Server-side revoke**: `linkai auth logout` calls `POST /api/cli/auth/revoke` before deleting local token files, ensuring leaked tokens cannot be reused

### Fine-grained scope / permission system

Scopes follow the pattern `{resource}:{action}` (e.g. `app:read`, `app:write`, `workflow:delete`).

Default scope on login: `app:read chat:read user:read workflow:read knowledge:read`

Write/delete scopes require explicit authorization via `--scope` flag or re-login.

Full scope list:

| Scope | Commands |
|-------|---------|
| `app:read` | `app list/detail` |
| `user:read` | `account info` |
| `chat:write` | `chat` |
| `knowledge:read` | `knowledge list/files/search` |
| `knowledge:write` | `knowledge create` |
| `knowledge:delete` | `knowledge delete` |
| `db:read` | `database list/tables/describe/exec` (SELECT) |
| `db:write` | `database exec` (INSERT/UPDATE/DELETE) — checked server-side |
| `image:write` | `image gen` |
| `video:write` | `video gen` |
| `audio:write` | `audio speech` |
| `plugin:read` | `plugin list/detail` |
| `plugin:run` | `plugin exec` |
| `workflow:read` | `workflow list` |
| `workflow:run` | `workflow run` |
| `score:read` | `score list/orders` |
| `score:write` | `score buy` |

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
- Non-JSON error responses (e.g. HTML gateway pages) are detected and reported with a body snippet for diagnosis
- Stream error responses read the response body to extract error detail
- Methods: `Get`, `Post`, `Delete`, `Stream` (SSE)

### Transport layer (`internal/cmdutil/transport.go`)

HTTP requests pass through a transport chain: `retryTransport → deviceIDTransport → http.DefaultTransport`.

- `retryTransport` retries on 502/503/504 with exponential backoff (max 3 retries), respects context cancellation
- `deviceIDTransport` injects `X-Device-ID` header on every request

Obtain the API client via `f.APIClient()` in command RunE functions.

### Backend API endpoints

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| POST | `/api/cli/auth/device` | public | initiate device authorization (sends device_id + scope) |
| POST | `/api/cli/auth/token` | public | poll for token (returns opaque access+refresh tokens) |
| POST | `/api/cli/auth/authorize` | user token | called by web page after user approves (sends granted_scope) |
| POST | `/api/cli/auth/refresh` | X-Device-ID | refresh access token |
| POST | `/api/cli/auth/revoke` | access token | logout / revoke tokens |
| GET | `/api/cli/app/list` | CLI token | list apps; params: `key`, `type[]`, `pageNo`, `pageSize` |
| GET | `/api/cli/app/detail` | CLI token | app detail; param: `code` |
| GET | `/api/cli/account/info` | CLI token | current user's name, credits, plan version |
| POST | `/api/cli/chat/completions` | CLI token | chat; body: `{app_code, message, stream?, session_id?}` |
| GET | `/api/cli/knowledge/list` | CLI token | list knowledge bases |
| POST | `/api/cli/knowledge/create` | CLI token | create knowledge base |
| POST | `/api/cli/knowledge/delete` | CLI token | delete knowledge base |
| GET | `/api/cli/knowledge/files` | CLI token | list files in a knowledge base |
| POST | `/api/cli/knowledge/search` | CLI token | vector search |
| GET | `/api/cli/database/list` | CLI token | list database connections; params: `page`, `page_size` |
| GET | `/api/cli/database/tables` | CLI token | list tables; param: `code` |
| GET | `/api/cli/database/describe` | CLI token | table structure; params: `code`, `table` |
| POST | `/api/cli/database/exec` | CLI token | execute SQL; body: `{code, sql}` |
| POST | `/api/cli/image/gen` | CLI token | generate image; body: `{prompt, model?, size?, aspect_ratio?}` |
| POST | `/api/cli/video/gen` | CLI token | create video task; body: `{prompt, model?, duration?, aspect_ratio?, mode?}` |
| POST | `/api/cli/video/status` | CLI token | query video task; body: `{task_id, model?, duration?, mode?}` |
| POST | `/api/cli/audio/speech` | CLI token | TTS; body: `{text, model?, voice?}` → returns `{url}` |
| GET | `/api/cli/plugin/list` | CLI token | list plugins; param: `category?` |
| GET | `/api/cli/plugin/detail` | CLI token | plugin detail; param: `code` |
| POST | `/api/cli/plugin/exec` | CLI token | execute plugin; body: `{code, input?, args?}` |
| GET | `/api/cli/workflow/list` | CLI token | list workflows |
| POST | `/api/cli/workflow/run` | CLI token | run workflow; body: `{app_code, input_text, args?, session_id?}` |
| GET | `/api/cli/score/products` | CLI token | list credit packages |
| POST | `/api/cli/score/buy` | CLI token | create purchase order; body: `{product_id, pay_channel}` |
| POST | `/api/cli/score/order/status` | CLI token | query order status; body: `{order_no}` |
| GET | `/api/cli/score/orders` | CLI token | list purchase history; params: `pageNum`, `pageSize` |

Redis keys: `cli_device_auth:{device_code}`, `cli_access:{token}` (2h TTL), `cli_refresh:{token}` (7d TTL).
Tokens are bound to `device_id` server-side; mismatched device returns 401.

### Adding new commands

Follow the pattern in `cmd/database/` or `cmd/image/`:
- Define an `Options` struct with a `Factory` field
- `NewCmdXxx(f *cmdutil.Factory, runF func(*Options) error) *cobra.Command` — `runF` allows test injection
- Register in `cmd/root.go` via `rootCmd.AddCommand(...)`
- Declare required scope: `cmd.Annotations = map[string]string{cmdutil.RequiredScopeKey: "resource:action"}`
- Use `f.APIClient()` for authenticated requests, `f.IOStreams` for output
- Use `output.PrintJSON` / `output.PrintTable` for formatted output
- For write commands: support `--dry-run` with `output.PrintDryRun()` to show the request without sending
- For list commands with pagination: use `--page` / `--page-size` flags (maps to backend `pageNo`/`pageSize`)
- Truncate displayed strings by **rune count**, not byte length: `[]rune(s)[:n]` to avoid corrupting CJK/emoji
- For async tasks (e.g. video): poll with `time.Sleep` + context check; print progress to `f.IOStreams.ErrOut`
- For binary downloads (e.g. audio `--output`): use `net/http` GET directly on the CDN URL, no API client needed
- For streaming (e.g. chat): use `client.Stream()` + `bufio.Scanner` with expanded buffer; always log parse errors to `ErrOut`

### Testing

Tests use the `runF` injection seam for command-level tests and `httptest.NewServer` for API client tests.

```bash
go test ./...
```

Key test files:
- `internal/auth/token_store_test.go` — TokenStatus states + MaskToken
- `internal/cmdutil/scope_test.go` — HasScope + CheckScope
- `internal/api/client_test.go` — JSON envelope, non-JSON errors, API error codes, stream errors
- `internal/cmdutil/transport_test.go` — retry behavior, backoff, exhaustion

### Config & token files

- `~/.linkai/config.json` — persistent `device_id` + logged-in user info (api_base is runtime-only via `LINKAI_API_BASE` env var)
- `~/.linkai/token.json` — `access_token`, `refresh_token`, `scope`, expiry timestamps (Unix ms)
