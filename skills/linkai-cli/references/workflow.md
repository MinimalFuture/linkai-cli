# Workflow

List and run automated workflows on the LinkAI platform.

**Required scopes**: `workflow:read` (list), `workflow:run` (run)

## List workflows

```bash
linkai workflow list
linkai workflow list --json
```

## Run a workflow

```bash
linkai workflow run <app_code> --input "<text>"
```

- `<app_code>` — the workflow's application code (from `workflow list`)
- `--input <text>` — input text to pass to the workflow

### Flags

| Flag | Type | Description |
|---|---|---|
| `--input` | string | Input text (required) |
| `--arg` | string (repeatable) | Extra argument in `key=value` format |
| `--session` | string | Session ID for multi-turn workflow |
| `--json` | bool | JSON output |

### Examples

```bash
linkai workflow list
linkai workflow run wf_abc123 --input "Summarize this report"
linkai workflow run wf_abc123 --input "Process data" --arg format=csv --arg verbose=true
linkai workflow run wf_abc123 --input "Follow up" --session sess-001    # multi-turn
```

## Typical workflow

1. `linkai workflow list` → find the workflow code
2. `linkai workflow run <code> --input "..."` → run it with input
