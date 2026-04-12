# Plugin

Browse and execute plugins on the LinkAI platform. Plugins are third-party tools that extend platform capabilities.

**Required scopes**: `plugin:read` (list/detail), `plugin:run` (exec)

## List plugins

```bash
linkai plugin list
linkai plugin list --category "search"    # filter by category
```

Flags:
- `--category <name>` — filter by category
- `--json` — JSON output

## View plugin detail

```bash
linkai plugin detail <code>
linkai plugin detail <code> --json
```

Shows the plugin's description, expected inputs, and parameters.

## Execute a plugin

```bash
linkai plugin exec <code> --input "<text>"
linkai plugin exec <code> --input "<text>" --arg key1=value1 --arg key2=value2
```

- `<code>` — plugin code (from `plugin list`)
- `--input <text>` — input text for the plugin
- `--arg key=value` — structured argument, can be repeated
- `--json` — JSON output

### Example

```bash
linkai plugin exec web-search --input "latest AI news"
linkai plugin exec translator --input "Hello world" --arg target_lang=zh
```

## Typical workflow

1. `linkai plugin list` → browse available plugins
2. `linkai plugin detail <code>` → check what inputs/args it expects
3. `linkai plugin exec <code> --input "..." --arg ...` → run it
