---
name: linkai-cli
description: Use the `linkai` CLI to interact with the LinkAI platform. Invoke this skill whenever the user wants to chat with an AI application, search or manage knowledge bases, generate images/video/audio, query databases, run workflows, execute plugins, check account info or credits, or do anything related to LinkAI. If the user's request matches any capability below, use this skill proactively — don't wait for them to say "use linkai".
---

# LinkAI CLI

`linkai` is a command-line tool for the LinkAI platform. It lets you chat with AI apps, manage knowledge bases, generate multimedia content, query databases, run workflows, call plugins, and manage your account — all from the terminal.

## Before you start

Check login status before running any command:

```bash
linkai auth status
```

If not logged in or token expired, have the user run `linkai auth login`. Some commands need extra scopes — if you see a scope error, re-login with the required scope:

```bash
linkai auth login --scope "app:read chat:write knowledge:read ..."
```

## Capability map

Use this table to decide which command to use. Once you've identified the right domain, read the corresponding reference file for full flag details and usage patterns.

| User intent | Command | Reference |
|---|---|---|
| Talk to an AI app, ask questions, get completions | `linkai chat` | [chat.md](references/chat.md) |
| Search documents, manage knowledge bases | `linkai knowledge ...` | [knowledge.md](references/knowledge.md) |
| Generate images | `linkai image gen` | [content-gen.md](references/content-gen.md) |
| Generate videos | `linkai video gen` | [content-gen.md](references/content-gen.md) |
| Text-to-speech / generate audio | `linkai audio speech` | [content-gen.md](references/content-gen.md) |
| Query or explore databases | `linkai database ...` | [database.md](references/database.md) |
| Execute a third-party plugin | `linkai plugin ...` | [plugin.md](references/plugin.md) |
| Run an automated workflow / pipeline | `linkai workflow ...` | [workflow.md](references/workflow.md) |
| Check account, credits, models, login/logout | `linkai account/auth/score/model ...` | [admin.md](references/admin.md) |

## Workflow: how to fulfill a request

1. **Identify the domain** from the capability map above.
2. **Read the reference file** for that domain — it has the exact syntax, all flags, and common patterns.
3. **Resolve missing arguments**: if a command needs an app code, knowledge base code, or database code that the user didn't provide, list the available resources first (e.g., `linkai app list`) and let them pick.
4. **Run the command** and present the result.
5. **Handle errors**: show the error and suggest a fix. Common issues:
   - `scope` error → re-login with the right scope
   - `not logged in` → run `linkai auth login`
   - network/5xx → retry once, then report

## Output guidelines

- **Lists**: summarize key fields (name, code, status). Don't dump raw JSON unless the user asks.
- **Generated content** (image/video/audio URL): display the URL prominently. If saved to a local file, tell the user the path.
- **Chat / streaming**: let the output stream naturally.
- **Errors**: show the message and suggest the fix.

## Common flags

Most commands support these:
- `--json` — output raw JSON (useful for piping or debugging)
- `--dry-run` — print the HTTP request without executing (write commands only)
- `--page` / `--page-size` — pagination for list commands
