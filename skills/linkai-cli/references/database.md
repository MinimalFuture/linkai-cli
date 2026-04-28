# database

## Scopes

- `db:read` — `list` / `tables` / `describe` / `SELECT` via `exec` (in default scopes)
- `db:write` — `INSERT/UPDATE/DELETE` via `exec` (**not** in default scopes; classified server-side from the SQL)

## Commands

| Command | Purpose |
|---|---|
| `linkai database list [--page <n>] [--page-size <n>] [--json]` | list connections |
| `linkai database tables <db_code> [--json]` | list tables in a connection |
| `linkai database describe <db_code> <table> [--json]` | columns + types |
| `linkai database exec <db_code> "<sql>" [--json] [--dry-run]` | run SQL |

## Safety rules — agent must follow

- **Default to `SELECT`.** Only run mutating SQL when the user explicitly asks.
- **Always include `LIMIT`** on exploratory selects (e.g. `LIMIT 100`).
- DDL (`DROP`, `TRUNCATE`, `ALTER`) is blocked client-side. Do not attempt; surface the rejection to the user.
- Mutating SQL needs `db:write` (not in defaults). On scope failure see [errors.md](errors.md).

## Discovery flow

1. `linkai database list --json` → pick `code`.
2. `linkai database tables <code> --json` → pick table.
3. `linkai database describe <code> <table> --json` → understand columns.
4. `linkai database exec <code> "SELECT ... LIMIT N" --json`.
