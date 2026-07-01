# Errors and recovery

## Exit codes

| Code | Meaning | Agent action |
|---|---|---|
| 0 | success | continue |
| 1 | general | report to user |
| 2 | validation (bad input/flags) | inspect args, fix and retry once |
| 3 | auth (not logged in / expired / scope denied) | see below — don't retry |
| 4 | network (5xx, connection) | the CLI already retried internally; report |

## Scope errors (exit 3, message mentions scope)

The default-login scopes do **not** include: `db:write`, `knowledge:create`, `knowledge:delete`.

When a command fails because of a missing scope:

1. **Stop. Do not retry the same command.**
2. `linkai auth status --json` to read the current `scope` field.
3. Tell the user verbatim:

   > This action needs scope `<missing>`, which isn't in your current grant. Please run:
   > `linkai auth login --scope "<existing scopes> <missing>"`
   > then re-run the original request.

4. Only run `auth login` after the user agrees — it needs them to authorize in a browser. Use the two-step flow in [auth.md](auth.md).

## Not logged in / expired

`auth status` returns `valid` / `needs_refresh` / `expired` (or "not logged in"). On `expired` or absent token, run the two-step login in [auth.md](auth.md). Auto-refresh is handled by the CLI when status is `needs_refresh`, so you usually won't see that case.

## Resource not found

If a command fails with "not found" on `app_code` / `kb_code` / `db_code` / `plugin_code`:

- List the resource (`linkai app list --json`, etc.) and either pick by name or surface the list to the user.
- Don't guess codes from names.

## Dangerous SQL blocked

`database exec` rejects `DROP`, `TRUNCATE`, `ALTER` client-side. Surface the rejection to the user; don't attempt workarounds.

## Non-JSON gateway errors

Occasionally the upstream returns HTML (e.g. CDN/gateway maintenance page). The CLI detects this and exits with code 4 and a body snippet. Treat as a transient network issue and report.
