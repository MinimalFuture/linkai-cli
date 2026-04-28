# chat

Talk to a LinkAI application. Required scope: `chat:send` (in default scopes).

## Command

```
linkai chat "<message>" --app <app_code> [--session <id>] [--no-stream] [--json] [--dry-run]
```

## Required

- `<message>` — positional, the user's message. Validate it has no control chars before passing.
- `--app <app_code>` — get from `linkai app list --json`.

## Agent recommendation

Always pass `--no-stream --json` unless the user explicitly wants live streaming output.

## Multi-turn

Pass the same `--session <id>` across calls; any non-empty string works. Server keeps context keyed by session id.

## JSON output

```json
{
  "session_id": "...",
  "answer": "...",
  "usage": { "...": "..." }
}
```

## Resolving the app code

```
linkai app list --json --key "<keyword>"
```

Returns `[{ code, name, ... }]`. Match by `name`, pass `code` as `--app`.
