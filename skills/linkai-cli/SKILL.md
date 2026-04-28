---
name: linkai-cli
description: Invoke LinkAI platform services via the `linkai` CLI — chat with AI apps, RAG search over knowledge bases, generate images/video/audio, query connected databases, run workflows, execute plugins, manage account/credits. Use when the user references "linkai", a LinkAI app, knowledge base (知识库), workflow, plugin, or asks to generate AI content / query a managed database. Optimized for non-interactive agent use (JSON output, no streaming, no interactive prompts).
---

# LinkAI CLI (agent edition)

This skill is optimized for agent invocation, not interactive human use. Default to JSON, non-streaming, non-interactive flags.

## Agent defaults — always apply

| Concern | Default | Reason |
|---|---|---|
| Output | append `--json` | machine-parseable |
| `chat` | add `--no-stream` (unless streaming live to the user) | full reply in one block |
| `score buy` | add `--agent` | returns `qr_base64`, not ASCII QR |
| `knowledge delete` | add `--force` | skips confirmation prompt |
| `auth login` | **never run from agent** | needs a browser; ask the user |
| Long async tasks | `video gen` polls internally — just wait | don't re-poll |

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

Before the first LinkAI call in a session, verify auth: `linkai auth status --json`. If status is not `valid`, ask the user to run `linkai auth login` themselves.
