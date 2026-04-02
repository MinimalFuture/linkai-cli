# linkai-cli

LinkAI 平台的官方命令行工具，让你在终端中管理应用、发起对话、操作工作流。

## 安装

### 环境要求

- Go 1.20+

### 从源码构建

```bash
git clone https://github.com/MinimalFuture/linkai-cli.git
cd linkai-cli
go build -o linkai .
```

将 `linkai` 移动到 PATH 中的任意目录：

```bash
mv linkai /usr/local/bin/
```

## 快速开始

### 1. 登录

```bash
linkai auth login
```

命令会打印一个授权 URL，在浏览器中打开并完成登录，CLI 自动完成授权。

```
Requesting authorization with scope: app:read chat:read user:read workflow:read knowledge:read
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

## 权限（Scope）

登录时可通过 `--scope` 指定请求的权限范围。权限精确到每个资源的操作级别。

| Scope | 说明 | 默认授予 |
|-------|------|---------|
| `app:read` | 查询应用列表、详情 | ✅ |
| `app:write` | 创建/更新应用 | ❌ |
| `app:delete` | 删除应用 | ❌ |
| `chat:read` | 查询对话记录 | ✅ |
| `chat:write` | 发起对话 | ❌ |
| `user:read` | 查询用户信息 | ✅ |
| `workflow:read` | 查询工作流 | ✅ |
| `workflow:write` | 创建/更新/执行工作流 | ❌ |
| `workflow:delete` | 删除工作流 | ❌ |
| `knowledge:read` | 查询知识库数据 | ✅ |
| `knowledge:write` | 添加/编辑知识库数据 | ❌ |
| `knowledge:delete` | 删除知识库数据 | ❌ |

默认登录只获取所有只读权限。需要写权限时重新授权：

```bash
linkai auth login --scope "app:read app:write"
```

授权页面会展示所请求的权限，用户可在页面上选择实际授予的权限范围。

## 安全设计

- **Opaque Token**：服务端存储，可随时撤销，有别于 JWT 自包含令牌
- **双 Token**：access_token 有效期 2 小时，refresh_token 有效期 7 天
- **设备绑定**：每台机器生成唯一 `device_id` 存于 `~/.linkai-cli/config.json`，所有请求携带 `X-Device-ID` 头。服务端将 token 与 device_id 绑定，token 被盗后在其他机器上无法使用
- **client_id**：每次登录生成的随机密钥，用于绑定 refresh token，存于 `~/.linkai-cli/token.json`

## 本地文件

```
~/.linkai-cli/
├── config.json   # API 地址、device_id、已登录用户信息
└── token.json    # access_token、refresh_token、client_id、scope、过期时间
```

## 开发

```bash
# 构建
go build -o linkai .

# 验证所有包可编译
go build ./...

# 整理依赖
go mod tidy
```

### 添加新命令

参考 `cmd/auth/` 的结构：

```go
func NewCmdXxx(f *cmdutil.Factory, runF func(*XxxOptions) error) *cobra.Command {
    cmd := &cobra.Command{
        Use:         "xxx",
        Annotations: map[string]string{cmdutil.RequiredScopeKey: "resource:action"},
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
