# workflow

## Scopes

- `workflow:read` — `list` (in default scopes)
- `workflow:run` — `run` (in default scopes)
- `workflow:create` — `create` (**not** in default scopes)
- `workflow:update` — `update` (**not** in default scopes)
- `workflow:delete` — `delete` (**not** in default scopes)

## Commands

```
linkai workflow list [--json]
linkai workflow run <app_code> --input "<text>" [--arg k=v ...] [--session <id>] [--json]
linkai workflow create --name <name> [--desc <txt>] [--json] [--dry-run]
linkai workflow update <code> [--name <name>] [--desc <txt>] [--json] [--dry-run]
linkai workflow delete <code> --force [--dry-run]
```

## Notes

- `<app_code>` is the workflow's app code; get it from `workflow list --json`.
- `--input` is required for `run`.
- `--arg` is repeatable for extra parameters the workflow expects.
- `--session <id>` enables multi-turn workflow context — pass the same id across calls.
- `create` makes a **blank** workflow shell — the node orchestration is done in the console. The create response includes a `links.console` URL that opens the workflow editor; surface it to the user.
- `create` / `update` / `delete` need sensitive scopes; on scope error do not retry, ask the user to re-login with the missing scope.

## Clickable links

`create` (and `app create` / `app detail` / `knowledge create`) return a `links`
object built by the server (never hard-coded in the CLI). The CLI prints these
under the success line. Forward them to the user so they can click straight into
the console.
