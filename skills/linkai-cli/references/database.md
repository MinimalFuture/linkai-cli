# Database

Query and explore databases connected to the LinkAI platform.

**Required scopes**: `db:read` (list/tables/describe/SELECT), `db:write` (INSERT/UPDATE/DELETE — enforced server-side)

## Explore databases

```bash
linkai database list                          # list all database connections
linkai database tables <code>                 # list tables in a database
linkai database describe <code> <table>       # show table structure (columns, types)
```

All three support `--json` for JSON output. `database list` also supports `--page` / `--page-size`.

## Execute SQL

```bash
linkai database exec <code> "<sql>"
```

- `<code>` — database connection code (from `database list`)
- `<sql>` — the SQL statement to execute

### Safety

- Default to SELECT queries unless the user explicitly asks for writes.
- The CLI blocks DDL (DROP, TRUNCATE, ALTER) client-side.
- INSERT/UPDATE/DELETE require `db:write` scope, which the server checks.

### Examples

```bash
linkai database exec mydb "SELECT * FROM users LIMIT 10"
linkai database exec mydb "SELECT count(*) FROM orders WHERE status = 'completed'"
linkai database exec mydb "INSERT INTO logs (msg) VALUES ('test')"    # needs db:write scope
```

### Flags

| Flag | Type | Description |
|---|---|---|
| `--json` | bool | JSON output |
| `--dry-run` | bool | Print request without executing |

## Typical workflow

A common pattern for database exploration:

1. `linkai database list` → find the database code
2. `linkai database tables <code>` → see available tables
3. `linkai database describe <code> <table>` → understand the schema
4. `linkai database exec <code> "SELECT ..."` → run the query
