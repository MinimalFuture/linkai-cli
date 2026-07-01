<h1 align="center">LinkAI CLI</h1>

<p align="center">LinkAI 智能体平台的命令行工具</p>

<p align="center">
  <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="MIT License">
  <img src="https://img.shields.io/badge/Go-1.21+-blue.svg" alt="Go 1.21+">
  <a href="https://www.npmjs.com/package/linkai-cli"><img src="https://img.shields.io/npm/v/linkai-cli.svg" alt="npm"></a>
</p>

<p align="center">
  <a href="./README.md">English</a> · 中文
</p>

**LinkAI CLI** 是 LinkAI 平台的命令行工具，让 Agent 和人类都可以在终端中调用 LinkAI 的全部能力（模型、应用、知识库、数据库、工作流、插件、账号等），为 Agent 提供丰富的基础设施资源，扩展 Agent 的能力边界。

## 能力总览

LinkAI CLI 核心提供两类能力：**平台资源**（管理和使用在 LinkAI 上配置的资源）和 **模型能力**（调用不同模型生成内容）。

**📦 平台资源**

| 资源 | 命令 | 说明 |
|------|------|------|
| 应用 | `app` | 查询 AI 应用列表与配置 |
| 知识库 | `knowledge` | 向量检索私有知识库，管理文件与知识库 |
| 数据库 | `database` | 查询业务数据库、表结构，执行 SQL |
| 工作流 | `workflow` | 运行 LinkAI 上编排好的工作流 |
| 插件 | `plugin` | 调用平台插件能力 |
| 账户 | `account` | 用户账号信息、积分管理 |

**🧠 AI 模型能力**

| 能力 | 命令 | 说明 |
|------|------|------|
| 对话 | `chat` | 与 AI 应用或语言模型对话，支持多轮会话 |
| 图片生成 | `image` | 文生图，输出图片链接 |
| 视频生成 | `video` | 文生视频，内置轮询等待完成 |
| 语音合成 | `audio` | 文本转语音（TTS），可下载到本地 |
| 模型列表 | `model` | 查询可用模型（LLM / IMAGE / VIDEO） |

> 运行 `linkai <命令> --help` 查看任意命令的完整参数。

## 安装

### 方式一：npm（推荐）

需要 Node.js 16 或更高版本。npm 会将命令安装到全局目录，无需手动配置 PATH：

```bash
npm i -g linkai-cli
```

### 方式二：一键安装脚本（零依赖）

```bash
# macOS / Linux
curl -fsSL https://cdn.link-ai.tech/cli/install.sh | sh
# Windows (PowerShell)
irm https://cdn.link-ai.tech/cli/install.ps1 | iex
```

脚本会下载预编译二进制、自动配置 PATH，并把 Agent Skill 安装到常见 AI 工具目录（Claude Code / Cursor / Codex / CowAgent 等）。

<details>
<summary>脚本环境变量 & 其他安装方式</summary>

| 变量 | 说明 | 默认 |
|------|------|------|
| `LINKAI_VERSION` | 指定版本 | `latest` |
| `LINKAI_INSTALL_DIR` | 二进制安装目录 | 自动选择（优先已在 PATH 的目录） |
| `LINKAI_SOURCE` | 下载源 `cdn` / `github` | `cdn`（不可达自动回退 GitHub） |
| `LINKAI_NO_SKILL` | 设为 `1` 跳过 Skill 安装 | — |

- **Homebrew：** `brew install MinimalFuture/tap/linkai`
- **Go：** `go install github.com/MinimalFuture/linkai-cli@latest`
- **二进制：** 从 [GitHub Releases](https://github.com/MinimalFuture/linkai-cli/releases/latest) 下载解压后放入 PATH。
- **源码：** `git clone` 后 `make build && make install`。

</details>

## 人类使用

三步开始：

```bash
linkai auth login                     # 1. 登录（浏览器完成授权）
linkai app list                       # 2. 查看应用，获取 app_code
linkai chat "你好" --app <app_code>    # 3. 与应用对话
```

`auth status` 查看登录状态，`auth logout` 登出。

## Agent 使用

安装后，Agent 通过 Skill 了解如何调用 CLI，无需额外配置。Skill 内容见 [`skills/linkai-cli/`](./skills/linkai-cli/SKILL.md)。

调用时的几点说明：

- 加 `--json` 获取结构化输出，便于解析。
- 写操作可先加 `--dry-run` 预览请求；`knowledge delete` 等需确认的命令加 `--force` 跳过交互。
- 运行 `linkai <命令> --help` 查看命令的完整参数。

## 命令参考

| 命令 | 说明 |
|------|------|
| `auth login` / `logout` / `status` | 登录 / 登出 / 查看登录状态 |
| `app list` / `app detail <code>` | 查看应用列表与详情 |
| `chat "<消息>" --app <code>` | 与应用对话，`--session` 支持多轮 |
| `knowledge list` / `files` / `search` / `create` / `delete` | 知识库查询与管理 |
| `database list` / `tables` / `describe` / `exec` | 数据库查询与 SQL 执行 |
| `workflow list` / `run <code> --input "<文本>"` | 查看与运行工作流 |
| `plugin list` / `detail` / `exec <code>` | 查看与执行插件 |
| `image gen "<描述>"` | 文生图 |
| `video gen "<描述>"` | 文生视频（内置轮询等待） |
| `audio speech "<文本>" [--output a.mp3]` | 语音合成，可下载 |
| `model list [--type LLM\|IMAGE\|VIDEO]` | 查看可用模型 |
| `account info` | 查看账号信息与积分余额 |
| `account credits` / `recharge` / `orders` | 积分套餐、充值购买与订单 |

> 每个命令的完整参数见 `linkai <命令> --help`。

## 权限

登录时通过 `--scope` 申请权限，格式为「资源:动作」。默认授予只读与内容生成权限；写操作权限需显式申请。

| 权限 | 说明 | 默认授予 |
|------|------|:---:|
| `app:read` | 查询应用列表、详情 | ✅ |
| `app:create` | 创建应用 | ✅ |
| `user:read` | 查询用户信息 | ✅ |
| `chat:send` | 与应用对话 | ✅ |
| `knowledge:read` | 查询知识库 | ✅ |
| `knowledge:create` | 创建知识库 / 添加文件 | ✅ |
| `db:read` | 查询数据库 / 执行 SELECT | ✅ |
| `image:gen` / `video:gen` / `audio:gen` | 生成图片 / 视频 / 语音 | ✅ |
| `plugin:read` / `plugin:run` | 查询 / 执行插件 | ✅ |
| `workflow:read` / `workflow:run` | 查询 / 运行工作流 | ✅ |
| `workflow:create` | 创建工作流 | ✅ |
| `score:read` / `score:buy` | 查看 / 购买积分 | ✅ |
| `app:update` / `app:delete` | 更新 / 删除应用 | ❌ |
| `knowledge:update` / `knowledge:delete` | 更新 / 删除知识库 | ❌ |
| `workflow:update` / `workflow:delete` | 更新 / 删除工作流 | ❌ |
| `db:write` | 数据库写操作（INSERT/UPDATE/DELETE） | ❌ |

申请额外权限时重新登录，例如：

```bash
linkai auth login --scope "db:read db:write knowledge:update knowledge:delete"
```

## 更多

<details>
<summary><strong>安全设计</strong></summary>

- **Token**：Opaque token 服务端存储、可撤销；`access` 2h / `refresh` 7d，到期前自动刷新；`logout` 同步吊销服务端 token。
- **设备绑定**：请求携带 `X-Device-ID`，服务端绑定 token 与设备。
- **本地存储**：macOS 存入系统钥匙串（服务名 `linkai-cli`），其他平台文件存储（`0600`）；配置在 `~/.linkai/`。
- **输入/输出防护**：拒绝危险 Unicode，数据库禁 DDL，表格输出剥离 ANSI。
- **容错**：5xx 自动指数退避重试；结构化退出码（0 成功 / 1 一般 / 2 参数 / 3 认证 / 4 网络）。

</details>

<details>
<summary><strong>Shell 补全</strong></summary>

```bash
source <(linkai completion bash)                # Bash
source <(linkai completion zsh)                 # Zsh
linkai completion fish | source                 # Fish
```

</details>

<details>
<summary><strong>开发</strong></summary>

```bash
make build   # 构建（带版本注入）
make test    # 测试
make lint    # golangci-lint
```

添加新命令参考 `cmd/database/`，在 `cmd/root.go` 注册，通过 `permission.RequiredKey` 声明权限。

</details>

## License

[MIT](./LICENSE)
