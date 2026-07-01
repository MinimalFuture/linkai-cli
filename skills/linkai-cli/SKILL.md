---
name: linkai-cli
description: Invoke LinkAI platform services via the `linkai` CLI — chat with AI apps, RAG search over knowledge bases, generate images/video/audio, query connected databases, run workflows, execute plugins, manage account/credits. Use when the user references "linkai", a LinkAI app, knowledge base, workflow, plugin, or asks to generate AI content / query a managed database. Optimized for non-interactive agent use (JSON output, no streaming, no interactive prompts).
---

# LinkAI CLI (agent edition)

This skill is optimized for agent invocation, not interactive human use. Default to JSON, non-streaming, non-interactive flags.

## Install (if `linkai` is not found)

If the `linkai` command is missing, install it first (no interaction needed):

```bash
# Preferred when Node.js is available (npm's global bin is already on PATH):
npm i -g linkai-cli
# Zero-dependency fallback (macOS/Linux):
curl -fsSL https://cdn.link-ai.tech/cli/install.sh | sh
```

Then verify: `linkai --version`. On Windows use `irm https://cdn.link-ai.tech/cli/install.ps1 | iex`.

## Agent defaults — always apply

| Concern | Default | Reason |
|---|---|---|
| Output | append `--json` | machine-parseable |
| `chat` | non-streaming is automatic when output is piped (agent case); add `--no-stream` to force it | full reply in one block |
| `score buy` | add `--agent` | returns `qr_base64`, not ASCII QR |
| `knowledge delete` | add `--force` | skips confirmation prompt |
| `auth login` | use the **non-blocking flow** (see below) — never the plain blocking `auth login` | needs a browser; must not block a tool call |
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
| Exit codes & scope recovery | — | [errors.md](references/errors.md) |

## Decision flow

1. Pick the command from the capability map.
2. Open the matching reference for required flags and JSON output fields.
3. Resolve missing IDs (`app_code`, `kb_code`, `db_code`, `plugin_code`) by listing first with `--json`. Don't guess codes.
4. Run with `--json`. Parse the result; surface only what the user needs.
5. On non-zero exit, classify via [errors.md](references/errors.md). For scope errors, **stop and ask the user to re-login** — do not retry.

## Pre-flight

Before the first LinkAI call in a session, verify auth: `linkai auth status --json`. If status is not `valid`, run the non-blocking login flow below.

## Login (non-blocking, agent flow)

Never run the plain `linkai auth login` — it blocks up to several minutes waiting for the browser. Instead:

1. **Get the URL** (returns instantly):

   ```bash
   linkai auth login --no-wait --json
   # → {"verification_url":"...","device_code":"xxx","next_action":{"command":"...","instruction":"..."}}
   ```

   Send `verification_url` to the user and ask them to open it and authorize. The `next_action` field tells you the exact next command.

2. **Poll with a bounded wait** (each call blocks at most ~60s, then returns):

   ```bash
   linkai auth login --device-code xxx --wait 60 --json
   ```

   - `event=authorization_pending` → user hasn't finished; re-run the SAME command.
   - `event=authorization_complete` → done, proceed.
   - `event=authorization_failed` → stop (expired or denied); restart from step 1.

   Keep polling on `pending` until complete or the code expires (~5 min). Do not block indefinitely.
