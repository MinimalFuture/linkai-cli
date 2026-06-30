# knowledge

| Command | Scope | In default scopes? |
|---|---|---|
| `linkai knowledge list [--json]` | `knowledge:read` | yes |
| `linkai knowledge files <kb_code> [--name <kw>] [--page <n>] [--page-size <n>] [--json]` | `knowledge:read` | yes |
| `linkai knowledge search <kb_code> "<query>" [--k <n>] [--json] [--dry-run]` | `knowledge:read` | yes |
| `linkai knowledge create --name <name> [--desc <txt>] [--json] [--dry-run]` | `knowledge:create` | **no** |
| `linkai knowledge update <kb_code> [--name <name>] [--desc <txt>] [--json] [--dry-run]` | `knowledge:update` | **no** |
| `linkai knowledge add <kb_code> (--text <txt> \| --question <q> --answer <a>) [--file-id <id>] [--json] [--dry-run]` | `knowledge:create` | **no** |
| `linkai knowledge data delete <kb_code> --file-id <id> --id <data_id> --force [--dry-run]` | `knowledge:delete` | **no** |
| `linkai knowledge delete <kb_code> --force [--dry-run]` | `knowledge:delete` | **no** |

## Agent rules

- **Always use `--force` on delete** — never run interactively.
- `search --k` defaults to 5; cap at ~20.
- `add`: provide either `--text` (raw chunk) or `--question` + `--answer` (QA). `--file-id` is optional — omit it to auto-create a file; the returned `file_id` can be reused on subsequent `add` calls to append to the same file.
- `data delete` removes a single entry (chunk/QA) inside a file; `knowledge delete` removes the whole base.
- `create` / `update` / `add` / `delete` need sensitive scopes; on scope error see [errors.md](errors.md) — do not retry, ask the user to re-login with the missing scope.

## Resolving `kb_code`

```
linkai knowledge list --json
```

Array of `{ code, name, description, ... }`. Match by `name`, use `code`.

## search JSON shape

```json
{
  "results": [
    { "content": "...", "score": 0.83, "file_name": "..." }
  ]
}
```
