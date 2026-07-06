# knowledge

| Command | Scope | In default scopes? |
|---|---|---|
| `linkai knowledge list [--json]` | `knowledge:read` | yes |
| `linkai knowledge files <kb_code> [--name <kw>] [--page <n>] [--page-size <n>] [--json]` | `knowledge:read` | yes |
| `linkai knowledge search <kb_code> "<query>" [--k <n>] [--json] [--dry-run]` | `knowledge:read` | yes |
| `linkai knowledge create --name <name> [--desc <txt>] [--json] [--dry-run]` | `knowledge:create` | **no** |
| `linkai knowledge update <kb_code> [--name <name>] [--desc <txt>] [--json] [--dry-run]` | `knowledge:update` | **no** |
| `linkai knowledge add <kb_code> (--text <txt> \| --question <q> --answer <a>) [--file-id <id>] [--json] [--dry-run]` | `knowledge:create` | **no** |
| `linkai knowledge import <kb_code> --file <path> [--type doc\|qa\|table] [--json] [--dry-run]` | `knowledge:create` | **no** |
| `linkai knowledge file delete <kb_code> --file-id <id> --force [--dry-run]` | `knowledge:delete` | **no** |
| `linkai knowledge data delete <kb_code> --file-id <id> --id <data_id> --force [--dry-run]` | `knowledge:delete` | **no** |
| `linkai knowledge delete <kb_code> --force [--dry-run]` | `knowledge:delete` | **no** |

## Agent rules

- **Always use `--force` on delete** — never run interactively.
- `search --k` defaults to 5; cap at ~20.
- `add`: provide either `--text` (raw chunk) or `--question` + `--answer` (QA). `--file-id` is optional — omit it to auto-create a file; the returned `file_id` can be reused on subsequent `add` calls to append to the same file.
- `import`: upload a whole local file (same as the platform's file import). `--type doc` (default) for unstructured docs (pdf/txt/word/md/...), `--type qa` for a two-column question/answer csv/excel, `--type table` for multi-column tabular excel/csv. Use `import` for files; use `add` for a single text chunk or QA pair typed inline. Embedding is **async** — the command returns once accepted; poll `knowledge files <kb_code>` to confirm it appears.
- **Deleting** — three levels:
  - `file delete --file-id <id>` removes a **whole file** and all its entries. Get `file-id` from `knowledge files <kb_code>` (the `fileId` field). This is what you want to drop one uploaded/imported file.
  - `data delete --file-id <id> --id <data_id>` removes a **single entry** (chunk/QA) inside a file — only use when you truly need to drop one entry.
  - `knowledge delete <kb_code>` removes the **whole base**.
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
