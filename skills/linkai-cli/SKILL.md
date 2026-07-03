---
name: linkai-cli
description: LinkAI is an all-in-one agent platform; the `linkai` CLI lets an agent tap its capabilities — AI models (chat, image/video/audio generation) and platform resources (apps, knowledge bases, databases, workflows, plugins). Use when an agent needs these AI capabilities or LinkAI platform resources.
---

# LinkAI CLI (agent edition)

This skill is optimized for agent invocation, not interactive human use. Default to JSON, non-streaming, non-interactive flags.

> If `linkai` is not installed, see the fallback install note at the [end of this file](#install-fallback).

## Agent defaults — always apply

| Concern | Default | Reason |
|---|---|---|
| Output | append `--json` | machine-parseable |
| `chat` | non-streaming is automatic when output is piped (agent case); add `--no-stream` to force it | full reply in one block |
| `score buy` | add `--agent` | returns `qr_base64`, not ASCII QR |
| `knowledge delete` | add `--force` | skips confirmation prompt |
| `auth login` | use the two-step flow (see below) | login needs the user to authorize in a browser |
| Long async tasks | `video gen` polls internally — just wait | don't re-poll |
| Unknown command/flags | run `linkai <command> --help` | `--help` is the authoritative source when a reference is missing |

## Capability map

| Intent | Command | Reference |
|---|---|---|
| Chat with an AI app | `linkai chat` | [chat.md](references/chat.md) |
| Knowledge base (list/files/search/create/delete) | `linkai knowledge ...` | [knowledge.md](references/knowledge.md) |
| Image / Video / Audio generation | `linkai image gen` / `video gen` / `audio speech` | [content-gen.md](references/content-gen.md) |
| Database query | `linkai database ...` | [database.md](references/database.md) |
| Plugin | `linkai plugin ...` | [plugin.md](references/plugin.md) |
| Workflow | `linkai workflow ...` | [workflow.md](references/workflow.md) |
| App / Model / Account / Credits | `linkai app/model/account/score ...` | [admin.md](references/admin.md) |
| Login / auth status | `linkai auth ...` | [auth.md](references/auth.md) |
| Exit codes & scope recovery | — | [errors.md](references/errors.md) |

## Decision flow

1. Pick the command from the capability map.
2. Open the matching reference for required flags and JSON output fields.
3. Resolve missing IDs (`app_code`, `kb_code`, `db_code`, `plugin_code`) by listing first with `--json`. Don't guess codes.
4. Run with `--json`. Parse the result; surface only what the user needs.
5. On non-zero exit, classify via [errors.md](references/errors.md). For scope errors, **stop and ask the user to re-login** — do not retry.

## Pre-flight

Before the first LinkAI call in a session, verify auth: `linkai auth status --json`. If status is not `valid`, run the login flow below.

## Login (agent flow)

Login needs the user to authorize in a browser, so run it in two non-blocking steps instead of the plain `linkai auth login` (which waits for that authorization):

1. Get the URL and send it to the user to authorize:

   ```bash
   linkai auth login --no-wait --json
   ```

2. Poll until the user finishes (each call returns within `--wait` seconds):

   ```bash
   linkai auth login --device-code <code> --wait 60 --json
   ```

   Re-run the same command while `event` is `authorization_pending`; stop on `authorization_complete` or `authorization_failed`.

See [auth.md](references/auth.md) for the JSON fields and event handling.

## Install (fallback) {#install-fallback}

If the `linkai` command is missing, install it (no interaction needed):

```bash
npm i -g linkai-cli   # when Node.js is available
```

If npm is unavailable or fails, see [install.md](references/install.md) for the
install-script and manual-download methods (macOS/Linux/Windows). Verify with
`linkai --version`.
