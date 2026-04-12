# Knowledge Base

Manage and search knowledge bases on the LinkAI platform. Knowledge bases store documents for vector search (RAG).

**Required scopes**: `knowledge:read` (list/files/search), `knowledge:write` (create), `knowledge:delete` (delete)

## List knowledge bases

```bash
linkai knowledge list
linkai knowledge list --json
```

## Search a knowledge base

```bash
linkai knowledge search <code> "<query>"
linkai knowledge search <code> "<query>" --k 10    # return top 10 results
```

- `<code>` — the knowledge base code (from `knowledge list`)
- `<query>` — the search text
- `--k <n>` — number of results (default: 5)
- `--json` — JSON output
- `--dry-run` — print request without executing

This performs vector similarity search across all documents in the knowledge base.

## List files in a knowledge base

```bash
linkai knowledge files <code>
linkai knowledge files <code> --name "keyword"    # filter by file name
```

Flags:
- `--name <keyword>` — filter by file name
- `--page <n>` / `--page-size <n>` — pagination (default: page 1, 20 per page)
- `--json` — JSON output

## Create a knowledge base

```bash
linkai knowledge create --name "My KB"
linkai knowledge create --name "My KB" --desc "Description"
```

Flags:
- `--name <name>` — name (required)
- `--desc <description>` — description
- `--json` — JSON output
- `--dry-run` — print request without executing

## Delete a knowledge base

```bash
linkai knowledge delete <code>
linkai knowledge delete <code> --force    # skip confirmation
```

Flags:
- `--force` — skip the confirmation prompt
- `--dry-run` — print request without executing

## Complete flag reference

| Command | Flag | Type | Description |
|---|---|---|---|
| `list` | `--json` | bool | JSON output |
| `search` | `--k` | int | Number of results (default 5) |
| `search` | `--json` | bool | JSON output |
| `search` | `--dry-run` | bool | Print request only |
| `files` | `--name` | string | Filter by file name |
| `files` | `--page` | int | Page number (default 1) |
| `files` | `--page-size` | int | Items per page (default 20) |
| `files` | `--json` | bool | JSON output |
| `create` | `--name` | string | KB name (required) |
| `create` | `--desc` | string | Description |
| `create` | `--json` | bool | JSON output |
| `create` | `--dry-run` | bool | Print request only |
| `delete` | `--force` | bool | Skip confirmation |
| `delete` | `--dry-run` | bool | Print request only |
