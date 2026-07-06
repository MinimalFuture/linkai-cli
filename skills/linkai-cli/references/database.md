# database

## Scopes

- `db:read` — `list` / `tables` / `describe` / `SELECT` via `exec` (in default scopes)
- `db:write` — `create`, and `INSERT/UPDATE/DELETE` via `exec` (**not** in default scopes; `exec` is classified server-side from the SQL)

## Commands

| Command | Purpose |
|---|---|
| `linkai database list [--page <n>] [--page-size <n>] [--json]` | list databases |
| `linkai database create --name <name> [--description <text>] [--json] [--dry-run]` | create a builtin (platform-hosted) database |
| `linkai database create-table <db_code> --name <table> --field name:type[:comment] ... [--json] [--dry-run]` | create a table in a builtin database |
| `linkai database tables <db_code> [--json]` | list tables in a database |
| `linkai database describe <db_code> <table> [--json]` | columns + types |
| `linkai database exec <db_code> "<sql>" [--json] [--dry-run]` | run SQL |

`create` / `create-table` only work on **builtin** (platform-hosted) databases — no external connection. Both need `db:write` (not in defaults); on scope failure see [errors.md](errors.md).

Full builtin flow from the CLI: `create` a database → `create-table` to define columns → `exec` INSERT/SELECT to read/write rows. (Table structure cannot be created via `exec` — `CREATE TABLE` and other DDL are blocked; use `create-table`.)

`create-table` column types: `text`, `text1024`, `longtext`, `number`, `decimal`, `datetime`. An auto-increment `id` primary key is added automatically. Define columns with repeated `--field name:type[:comment]`, or (agent-preferred) `--fields-json '[{"name":"...","type":"..."}]'`. Table name must not contain spaces.

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
