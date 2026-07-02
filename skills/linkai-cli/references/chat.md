# chat

Talk to a LinkAI application, or call an LLM directly. Required scope: `chat:send` (in default scopes).

## Command

```
linkai chat "<message>" [--app <app_code>] [--model <model_code>] [--session <id>] [--no-stream] [--json] [--dry-run]
```

## Required

- `<message>` — positional, the user's message. Validate it has no control chars before passing.

## App vs model (both optional)

- `--app <app_code>` — chat with a configured application. Get the code from `linkai app list --json`.
- `--model <model_code>` — call an LLM directly. Get the code from `linkai model list --json` (use an entry whose type is `LLM`).
- Both given → `--model` overrides the app's configured model.
- Neither given → the platform default model is used.

## Agent recommendation

Always pass `--no-stream --json` for a single, parseable JSON reply (streaming is only for humans at a terminal).

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
