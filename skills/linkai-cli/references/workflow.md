# workflow

## Scopes

- `workflow:read` — `list` (in default scopes)
- `workflow:run` — `run` (in default scopes)

## Commands

```
linkai workflow list [--json]
linkai workflow run <app_code> --input "<text>" [--arg k=v ...] [--session <id>] [--json]
```

## Notes

- `<app_code>` is the workflow's app code; get it from `workflow list --json`.
- `--input` is required.
- `--arg` is repeatable for extra parameters the workflow expects.
- `--session <id>` enables multi-turn workflow context — pass the same id across calls.
