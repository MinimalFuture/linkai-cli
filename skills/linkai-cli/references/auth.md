# Auth (agent flow)

Login uses OAuth Device Flow: the user opens a URL and authorizes in a browser. The plain `linkai auth login` blocks until that finishes, so agents use the two-step flow below.

## Check status first

```bash
linkai auth status --json
```

`status` is `valid`, `needs_refresh`, or `expired` (or the user is not logged in). Auto-refresh is handled by the CLI on `needs_refresh`, so a login is only needed on `expired` / not logged in.

## Step 1 — get the authorization URL

```bash
linkai auth login --no-wait --json
```

Returns immediately:

```json
{
  "verification_url": "https://...",
  "device_code": "xxx",
  "next_action": { "command": "...", "instruction": "..." }
}
```

Send `verification_url` to the user and ask them to open it and authorize. `next_action.command` is the exact command to run in step 2.

## Step 2 — poll with a bounded wait

```bash
linkai auth login --device-code <code> --wait 60 --json
```

`--wait N` blocks for at most N seconds, then returns one of:

| `event` | Meaning | Action |
|---|---|---|
| `authorization_pending` | User hasn't finished yet | Re-run the same command |
| `authorization_complete` | Logged in | Proceed |
| `authorization_failed` | Code expired or denied | Restart from step 1 |

Keep polling on `pending` until complete. The device code expires after ~5 minutes.

## Requesting extra scopes

Default login scopes exclude `db:write`, `knowledge:create`, `knowledge:delete`. To grant them, add `--scope`:

```bash
linkai auth login --scope "<existing scopes> <missing scope>" --no-wait --json
```

Then run the same two-step flow. Only do this when a command has failed with a scope error (exit 3) and the user has agreed.
