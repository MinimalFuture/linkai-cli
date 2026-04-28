# plugin

## Scopes

- `plugin:read` — `list` / `detail` (in default scopes)
- `plugin:run` — `exec` (in default scopes)

## Commands

```
linkai plugin list [--category <name>] [--json]
linkai plugin detail <plugin_code> [--json]
linkai plugin exec <plugin_code> --input "<text>" [--arg key=value ...] [--json]
```

## Workflow

1. Have a plugin in mind? Skip to step 3.
2. `linkai plugin list --json` (optionally `--category`) to discover.
3. `linkai plugin detail <code> --json` — read expected `input` and `args` shape.
4. `linkai plugin exec <code> --input "..." --arg k=v --json`.

## Notes

- `--arg` is repeatable; each `key=value` is one argument.
- The shape of the JSON `result` field is plugin-specific — check `detail` first.
