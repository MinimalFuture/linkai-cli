# linkai-cli

LinkAI 平台的官方命令行工具，让 Agent 能够管理应用、知识库、数据库、插件、模型和工作流，生成文本/图片/视频/语音等内容，以及与智能体对话。

## 安装

按你的环境任选一种方式：

### 一键安装脚本（macOS / Linux）

```bash
curl -fsSL https://raw.githubusercontent.com/MinimalFuture/linkai-cli/main/install.sh | sh
```

默认安装到 `$HOME/.local/bin/linkai`，无需 sudo。锁定版本：`INSTALL_VERSION=v1.2.3 sh install.sh`；安装到 `/usr/local/bin`：`INSTALL_PREFIX=/usr/local sh install.sh`。

### Homebrew（macOS / Linux）

```bash
brew install MinimalFuture/tap/linkai
```

### npm（任何装了 Node 16+ 的环境）

```bash
npm i -g @linkai/cli       # 全局安装
npx @linkai/cli --help     # 一次性运行
```

也可以加进 `package.json` 的 `devDependencies`，团队 `npm ci` 后即可 `npx linkai`。

### Go（需要本地装 Go 1.21+）

```bash
go install github.com/MinimalFuture/linkai-cli@latest
```

### Windows

从 [Releases 页面](https://github.com/MinimalFuture/linkai-cli/releases/latest) 下载 `linkai_*_windows_*.zip`，解压后把 `linkai.exe` 放到 PATH 中。或使用上面的 npm 方式。

### 从源码构建

```bash
git clone https://github.com/MinimalFuture/linkai-cli.git
cd linkai-cli
make build    # 带版本号和构建日期注入
make install  # 安装到 $GOPATH/bin
```

## 快速开始

### 1. 登录

```bash
linkai auth login
```

命令会打印一个授权 URL，在浏览器中打开并完成登录，CLI 自动完成授权。

```
Requesting authorization with scope: app:read chat:send user:read workflow:read workflow:run knowledge:read db:read image:gen video:gen audio:gen plugin:read plugin:run score:read score:buy
Open the following URL in your browser to authorize:
  https://app.link-ai.tech/cli-login?code=xxxxxx

Waiting for authorization...
✓ Logged in as 张三 (138xxxx0000)
```

### 2. 查看登录状态

```bash
linkai auth status
linkai auth status --json
```

### 3. 登出

```bash
linkai auth logout
```

登出时会自动吊销服务端 token，确保 token 无法被重用。

---

## 命令参考

### auth — 认证

```bash
linkai auth login                          # 登录（Device Flow）
linkai auth login --scope "db:read db:write knowledge:create knowledge:delete"
linkai auth logout                         # 登出（同时吊销服务端 token）
linkai auth status                         # 查看当前登录状态
```

### account — 账户

```bash
linkai account info                        # 查看用户名、积分、版本
linkai account info --json
```

### app — 应用管理

```bash
linkai app list                            # 列出应用（表格）
linkai app list --key "客服" --page 2      # 关键词搜索 + 翻页
linkai app list --json                     # JSON 输出
linkai app detail <code>                   # 查看应用详情（配置、模型、插件）
```

### chat — 对话（需要 `chat:send` 权限）

```bash
linkai chat "你好" --app <app_code>                       # 单轮对话（流式输出）
linkai chat "继续" --app <app_code> --session <session_id> # 多轮对话
linkai chat "Hello" --app <app_code> --no-stream          # 非流式，等待完整回复
linkai chat "Hello" --app <app_code> --json               # JSON 输出（自动禁用流式）
linkai chat "Hello" --app <app_code> --dry-run            # 只打印请求，不实际执行
```

### knowledge — 知识库

```bash
linkai knowledge list                      # 列出知识库
linkai knowledge create --name "产品文档"  # 创建知识库
linkai knowledge delete <code>             # 删除知识库
linkai knowledge files <code>              # 列出知识库文件
linkai knowledge search <code> <query>     # 向量搜索
```

### database — 数据库（需要 `db:read` 权限；写操作需 `db:write`）

```bash
linkai database list                       # 列出数据库连接
linkai database tables <code>              # 列出指定库的所有表
linkai database describe <code> <table>    # 查看表结构
linkai database exec <code> "SELECT ..."   # 执行 SQL（SELECT 需 db:read，写操作需 db:write）
```

### model — 模型

```bash
linkai model list                          # 列出可用模型
linkai model list --json
```

### image — 图片生成（需要 `image:gen` 权限）

```bash
linkai image gen "a cat on the moon"
linkai image gen "日落风景" --model dall-e-3 --size 1024x1024
linkai image gen "portrait" --aspect-ratio 9:16 --json
```

输出图片 CDN URL，可直接在浏览器中打开。

### video — 视频生成（需要 `video:gen` 权限）

```bash
linkai video gen "ocean waves at sunset"
linkai video gen "城市夜景延时" --duration 10 --aspect-ratio 16:9 --mode pro
linkai video gen "a flying bird" --model jimeng_t2v_v30 --json
```

CLI 自动轮询等待生成完成（约 30s–3min），完成后输出视频 URL。

### audio — 语音合成（需要 `audio:gen` 权限）

```bash
linkai audio speech "Hello, welcome to LinkAI"
linkai audio speech "你好，欢迎使用 LinkAI" --output hello.mp3   # 下载到本地
linkai audio speech "Test" --model tts-1-hd --voice alloy --json
```

### plugin — 插件（需要 `plugin:read` / `plugin:run` 权限）

```bash
linkai plugin list                                # 列出可用插件
linkai plugin list --category "工具"              # 按分类筛选
linkai plugin detail <code>                       # 查看插件详情
linkai plugin exec <code> --input "查询内容"      # 执行插件
linkai plugin exec <code> --arg key1=val1 --arg key2=val2  # 带结构化参数
```

### workflow — 工作流（需要 `workflow:read` / `workflow:run` 权限）

```bash
linkai workflow list                                        # 列出工作流
linkai workflow run <app_code> --input "输入文本"            # 运行工作流
linkai workflow run <app_code> --input "text" --arg k=v     # 带额外参数
linkai workflow run <app_code> --input "text" --session <id> # 多轮会话
```

### score — 积分管理（需要 `score:read` / `score:buy` 权限）

```bash
linkai score list                          # 查看积分套餐
linkai score buy --product <id>            # 购买积分（终端显示付款二维码）
linkai score buy --product <id> --json     # agent 模式，返回付款链接
linkai score orders                        # 查看购买历史
linkai score orders --page 2 --page-size 20
```

---

## 通用 Flag

所有命令均支持以下 flag：

| Flag | 说明 |
|------|------|
| `--json` | JSON 格式输出，适合脚本和 agent 集成 |
| `--dry-run` | 仅打印将要发送的请求，不实际执行（写操作命令） |

---

## 权限

登录时通过 `--scope` 指定权限范围，权限精确到每个资源的操作级别。命令行 flag 沿用 OAuth 标准术语 `--scope`，但权限字符串本身按 `资源:动作` 命名，动作词与命令语义对齐（`chat:send`、`image:gen`、`score:buy`）。

| 权限 | 说明 | 默认授予 |
|------|------|---------|
| `app:read` | 查询应用列表、详情 | ✅ |
| `user:read` | 查询用户信息 | ✅ |
| `chat:send` | 与应用对话 | ✅ |
| `knowledge:read` | 查询知识库 | ✅ |
| `db:read` | 查询数据库/表/执行 SELECT | ✅ |
| `image:gen` | 生成图片 | ✅ |
| `video:gen` | 生成视频 | ✅ |
| `audio:gen` | 语音合成（TTS） | ✅ |
| `plugin:read` | 查询插件列表、详情 | ✅ |
| `plugin:run` | 执行插件 | ✅ |
| `workflow:read` | 查询工作流 | ✅ |
| `workflow:run` | 执行工作流 | ✅ |
| `score:read` | 查看积分套餐、购买历史 | ✅ |
| `score:buy` | 购买积分 | ✅ |
| `knowledge:create` | 创建知识库 | ❌ |
| `knowledge:delete` | 删除知识库 | ❌ |
| `db:write` | 执行写操作（INSERT/UPDATE/DELETE） | ❌ |

需要额外权限时重新授权：

```bash
# 数据库写操作
linkai auth login --scope "db:read db:write"

# 知识库管理
linkai auth login --scope "knowledge:read knowledge:create knowledge:delete"

# 一次性获取所有权限
linkai auth login --scope "app:read user:read chat:send knowledge:read knowledge:create knowledge:delete db:read db:write image:gen video:gen audio:gen plugin:read plugin:run workflow:read workflow:run score:read score:buy"
```

---

## Shell 补全

```bash
# Bash
source <(linkai completion bash)

# Zsh
source <(linkai completion zsh)
# 或永久安装:
linkai completion zsh > "${fpath[1]}/_linkai"

# Fish
linkai completion fish | source
```

---

## 安全设计

- **Opaque Token**：服务端存储，可随时撤销
- **双 Token**：`access_token` 有效期 2 小时，`refresh_token` 有效期 7 天
- **自动刷新**：`access_token` 到期前 5 分钟自动使用 `refresh_token` 换取新 token，无需重新登录
- **服务端吊销**：`linkai auth logout` 同时调用服务端 revoke 接口，确保 token 立即失效
- **设备绑定**：每台机器生成唯一 `device_id`，所有请求携带 `X-Device-ID` 头，服务端将 token 与 device_id 绑定
- **Keychain 存储**：macOS 上 token 存入系统钥匙串（Keychain），避免明文存储；其他平台使用文件存储（`0600` 权限）
- **输入验证**：拒绝控制字符、Bidi 覆盖等危险 Unicode 输入；数据库命令禁止 DDL 操作（DROP/TRUNCATE/ALTER）
- **输出净化**：表格输出自动剥离 ANSI 转义序列和终端控制字符
- **网络容错**：内置重试机制，502/503/504 自动指数退避重试（最多 3 次）
- **结构化错误**：区分退出码（0=成功, 1=一般错误, 2=参数错误, 3=认证错误, 4=网络错误），便于脚本和 agent 集成

## 本地文件

```
~/.linkai/
├── config.json   # device_id、已登录用户信息
└── token.json    # access_token、refresh_token、scope、过期时间（keychain 的文件备份）
```

macOS 上 token 优先存储在系统钥匙串中（服务名: `linkai-cli`），文件作为后备。

## 开发

```bash
make build               # 构建（带版本注入）
make test                # 运行测试（带 race detector）
make lint                # 代码检查（golangci-lint）
make tidy                # 整理依赖
make clean               # 清理构建产物
```

### 添加新命令

参考 `cmd/database/` 或 `cmd/image/` 的结构（`permission` 包来自 `github.com/MinimalFuture/linkai-cli/internal/permission`）：

```go
func NewCmdXxx(f *cmdutil.Factory, runF func(*XxxOptions) error) *cobra.Command {
    cmd := &cobra.Command{
        Use:         "xxx",
        Annotations: map[string]string{permission.RequiredKey: permission.AppRead.String()},
        RunE: func(cmd *cobra.Command, args []string) error {
            client, err := f.APIClient()
            // ...
        },
    }
    return cmd
}
```

在 `cmd/root.go` 中注册后即可使用。

## License

MIT
