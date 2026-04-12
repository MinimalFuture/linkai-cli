# Chat

Chat with a LinkAI application. Supports streaming, multi-turn conversation, and JSON output.

**Required scope**: `chat:write`

## Basic usage

```bash
linkai chat "<message>" --app <app_code>
```

The message is the first positional argument. `--app` specifies which application to talk to.

## Finding the right app

If the user doesn't know the app code, list available apps first:

```bash
linkai app list                    # list all apps
linkai app list --key "keyword"    # search by keyword
linkai app detail <code>           # view app detail
```

`app list` flags:
- `--key <keyword>` — filter by name
- `--page <n>` / `--page-size <n>` — pagination (default: page 1, 20 per page)
- `--json` — JSON output

## Streaming (default)

By default, `linkai chat` streams the response to stdout as it arrives (SSE). This gives the best user experience for interactive use.

## Disable streaming

```bash
linkai chat "<message>" --app <code> --no-stream
```

Waits for the full response and prints it at once. Useful when piping output to another command.

## Multi-turn conversation

```bash
# First turn
linkai chat "hello" --app <code> --session my-session-1

# Subsequent turns — same session ID keeps context
linkai chat "tell me more" --app <code> --session my-session-1
```

The `--session` flag ties messages together into a conversation. Use any string as the session ID — it just needs to be consistent across turns.

## JSON output

```bash
linkai chat "<message>" --app <code> --json
```

Returns structured JSON (disables streaming). Useful for programmatic consumption.

## Dry run

```bash
linkai chat "<message>" --app <code> --dry-run
```

Prints the HTTP request that would be sent, without executing it.

## Complete flags

| Flag | Type | Description |
|---|---|---|
| `--app` | string | Application code (required) |
| `--session` | string | Session ID for multi-turn conversation |
| `--no-stream` | bool | Disable streaming, wait for full reply |
| `--json` | bool | Output JSON (disables streaming) |
| `--dry-run` | bool | Print request without executing |
