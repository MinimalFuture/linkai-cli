# Auth (agent flow)

Login uses OAuth Device Flow: the user opens a URL and authorizes in a browser. After you send the link, do NOT stop and wait for the user to reply in chat — go straight to your next tool call and poll, repeating until authorization completes.

## Check status first

```bash
linkai auth status --json
```

`status` is `valid`, `needs_refresh`, or `expired` (or the user is not logged in). Auto-refresh is handled by the CLI on `needs_refresh`, so a login is only needed on `expired` / not logged in.

## Get the authorization URL

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

Tell the user to open `verification_url` and authorize. Then **make your next tool call the poll below** — do NOT stop to wait for the user to say "done".

## Poll until complete (next tool call, keep going)

Run `next_action.command` (a bounded poll that blocks up to 60s):

```bash
linkai auth login --device-code <code> --wait 60 --json
```

It returns one of:

| `event` | Meaning | Action |
|---|---|---|
| `authorization_pending` | User hasn't finished yet | **Re-run the same command as your next tool call** (loop) |
| `authorization_complete` | Logged in | Proceed to the user's original request |
| `authorization_failed` | Code expired or denied | Restart from "Get the authorization URL" |

Keep re-running the poll on `pending` (each call blocks ~60s, so looping is not busy-spinning) until you get `complete` or `failed`. The device code expires after ~5 minutes; if the user is away that long, it is fine to stop and let them re-trigger.

## Requesting extra scopes

Default login scopes exclude `db:write`, `knowledge:create`, `knowledge:delete`. To grant them, add `--scope`:

```bash
linkai auth login --scope "<existing scopes> <missing scope>" --no-wait --json
```

Then run the same two-step flow. Only do this when a command has failed with a scope error (exit 3) and the user has agreed.
