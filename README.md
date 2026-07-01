<h1 align="center">LinkAI CLI</h1>

<p align="center">The command-line interface for the LinkAI agent platform.</p>

<p align="center">
  <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="MIT License">
  <img src="https://img.shields.io/badge/Go-1.21+-blue.svg" alt="Go 1.21+">
  <a href="https://www.npmjs.com/package/linkai-cli"><img src="https://img.shields.io/npm/v/linkai-cli.svg" alt="npm"></a>
</p>

<p align="center">
  English · <a href="./README.zh.md">中文</a>
</p>

**LinkAI CLI** brings the full LinkAI platform to the terminal for both agents and people — models, apps, knowledge bases, databases, workflows, plugins, and accounts. It gives AI agents a rich set of platform capabilities as infrastructure, extending what they can do.

## Capabilities

The CLI exposes two kinds of capabilities: **platform resources** (manage and use the resources you've configured on LinkAI) and **model capabilities** (call models to generate content).

**📦 Platform resources**

| Resource | Command | Description |
|----------|---------|-------------|
| Apps | `app` | List AI apps and inspect their configuration |
| Knowledge | `knowledge` | Vector-search private knowledge bases; manage files and bases |
| Database | `database` | Query business databases and schemas, run SQL |
| Workflow | `workflow` | Run workflows orchestrated on LinkAI |
| Plugin | `plugin` | Invoke platform plugins |
| Account | `account` | Account profile, credits, and recharge |

**🧠 Model capabilities**

| Capability | Command | Description |
|------------|---------|-------------|
| Chat | `chat` | Talk to an AI app or language model, with multi-turn sessions |
| Image | `image` | Text-to-image; returns an image URL |
| Video | `video` | Text-to-video, with built-in polling until the result is ready |
| Speech | `audio` | Text-to-speech (TTS), optionally saved to a file |
| Models | `model` | List available models (LLM / IMAGE / VIDEO) |

> Run `linkai <command> --help` for the full flags of any command.

## Installation

### Option 1 — npm (recommended)

Requires Node.js 16 or later. npm installs the command to its global bin directory, which is already on your `PATH`:

```bash
npm i -g linkai-cli
```

### Option 2 — install script (no dependencies)

```bash
# macOS / Linux
curl -fsSL https://cdn.link-ai.tech/cli/install.sh | sh
# Windows (PowerShell)
irm https://cdn.link-ai.tech/cli/install.ps1 | iex
```

The script downloads the pre-built binary, sets up your `PATH`, and installs the agent skill into common AI tool directories (Claude Code, Cursor, Codex, CowAgent, etc.).

<details>
<summary>Other install methods (Homebrew / Go / Binary / Source)</summary>

- **Homebrew:** `brew install MinimalFuture/tap/linkai`
- **Go:** `go install github.com/MinimalFuture/linkai-cli@latest`
- **Binary:** download from [GitHub Releases](https://github.com/MinimalFuture/linkai-cli/releases/latest), extract, and put it on your `PATH`.
- **Source:** `git clone`, then `make build && make install`.

</details>

<details>
<summary>Install script environment variables</summary>

| Variable | Description | Default |
|----------|-------------|---------|
| `LINKAI_VERSION` | Version to install | `latest` |
| `LINKAI_INSTALL_DIR` | Binary install directory | Auto (prefers a directory already on `PATH`) |
| `LINKAI_SOURCE` | Download source: `cdn` / `github` | `cdn` (falls back to GitHub if unreachable) |
| `LINKAI_NO_SKILL` | Set to `1` to skip installing the skill | — |

</details>

## Getting started

Three steps to your first call:

```bash
linkai auth login                     # 1. Log in (authorize in the browser)
linkai app list                       # 2. List apps and grab an app_code
linkai chat "Hello" --app <app_code>   # 3. Chat with the app
```

Use `auth status` to check your session and `auth logout` to sign out.

## Using with agents

Once installed, agents learn how to drive the CLI through its bundled skill — no extra setup required. The skill lives in [`skills/linkai-cli/`](./skills/linkai-cli/SKILL.md).

A few notes when calling from an agent:

- Add `--json` for structured, easy-to-parse output.
- Preview write operations with `--dry-run`; add `--force` to skip confirmations (e.g. `knowledge delete`).
- Run `linkai <command> --help` for a command's full flags.

## Command reference

| Command | Description |
|---------|-------------|
| `auth login` / `logout` / `status` | Log in / out / show login status |
| `app list` / `app detail <code>` | List apps and view details |
| `chat "<message>" --app <code>` | Chat with an app; `--session` for multi-turn |
| `knowledge list` / `files` / `search` / `create` / `delete` | Query and manage knowledge bases |
| `database list` / `tables` / `describe` / `exec` | Query databases and run SQL |
| `workflow list` / `run <code> --input "<text>"` | List and run workflows |
| `plugin list` / `detail` / `exec <code>` | List and run plugins |
| `image gen "<prompt>"` | Text-to-image |
| `video gen "<prompt>"` | Text-to-video (waits for completion) |
| `audio speech "<text>" [--output a.mp3]` | Text-to-speech, optionally downloaded |
| `model list [--type LLM\|IMAGE\|VIDEO]` | List available models |
| `account info` | Account profile and credit balance |
| `account credits` / `recharge` / `orders` | Credit packages, recharge, and orders |

> See `linkai <command> --help` for each command's full flags.

## Permissions

Permissions are requested at login via `--scope`, in `resource:action` form. Read and content-generation scopes are granted by default; write scopes must be requested explicitly.

| Scope | Description | Default |
|-------|-------------|:---:|
| `app:read` | List and inspect apps | ✅ |
| `app:create` | Create apps | ✅ |
| `user:read` | Read user info | ✅ |
| `chat:send` | Chat with apps | ✅ |
| `knowledge:read` | Query knowledge bases | ✅ |
| `knowledge:create` | Create knowledge bases / add files | ✅ |
| `db:read` | Query databases / run SELECT | ✅ |
| `image:gen` / `video:gen` / `audio:gen` | Generate images / video / speech | ✅ |
| `plugin:read` / `plugin:run` | List / run plugins | ✅ |
| `workflow:read` / `workflow:run` | List / run workflows | ✅ |
| `workflow:create` | Create workflows | ✅ |
| `score:read` / `score:buy` | View / purchase credits | ✅ |
| `app:update` / `app:delete` | Update / delete apps | ❌ |
| `knowledge:update` / `knowledge:delete` | Update / delete knowledge bases | ❌ |
| `workflow:update` / `workflow:delete` | Update / delete workflows | ❌ |
| `db:write` | Database writes (INSERT/UPDATE/DELETE) | ❌ |

Request extra scopes by logging in again, for example:

```bash
linkai auth login --scope "db:read db:write knowledge:update knowledge:delete"
```

## More

<details>
<summary><strong>Security</strong></summary>

- **Tokens:** opaque tokens stored server-side and revocable; `access` lasts 2h, `refresh` lasts 7d, refreshed automatically before expiry; `logout` revokes them server-side.
- **Device binding:** every request carries `X-Device-ID`, and the server binds tokens to the device.
- **Local storage:** tokens go into the macOS Keychain (service `linkai-cli`), with a `0600` file fallback on other platforms; config lives in `~/.linkai/`.
- **Input / output hardening:** dangerous Unicode is rejected, DDL is blocked in the database commands, and ANSI is stripped from table output.
- **Resilience:** 5xx responses are retried with exponential backoff; exit codes are structured (0 ok / 1 general / 2 validation / 3 auth / 4 network).

</details>

<details>
<summary><strong>Shell completion</strong></summary>

```bash
source <(linkai completion bash)                # Bash
source <(linkai completion zsh)                 # Zsh
linkai completion fish | source                 # Fish
```

</details>

<details>
<summary><strong>Development</strong></summary>

```bash
make build   # build (with version injection)
make test    # run tests
make lint    # golangci-lint
```

To add a command, follow the pattern in `cmd/database/`, register it in `cmd/root.go`, and declare its scope via `permission.RequiredKey`.

</details>

## License

[MIT](./LICENSE)
