# linkai-cli

LinkAI 平台的官方命令行工具，让你在终端中管理应用、知识库、数据库，生成图片/视频/语音。

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

---

## 命令参考

### auth — 认证

```bash
linkai auth login                          # 登录（Device Flow）
linkai auth login --scope "db:read db:write image:write video:write audio:write"
linkai auth logout                         # 登出
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
```

### knowledge — 知识库

```bash
linkai knowledge list                      # 列出知识库
linkai knowledge create --name "产品文档"  # 创建知识库
linkai knowledge delete <code>             # 删除知识库
linkai knowledge files <code>             # 列出知识库文件
linkai knowledge search <code> <query>    # 向量搜索
```

### database — 数据库（需要 `db:read` scope）

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

### image — 图片生成（需要 `image:write` scope）

```bash
linkai image gen "a cat on the moon"
linkai image gen "日落风景" --model dall-e-3 --size 1024x1024
linkai image gen "portrait" --aspect-ratio 9:16 --json
```

输出图片 CDN URL，可直接在浏览器中打开。

### video — 视频生成（需要 `video:write` scope）

```bash
linkai video gen "ocean waves at sunset"
linkai video gen "城市夜景延时" --duration 10 --aspect-ratio 16:9 --mode pro
linkai video gen "a flying bird" --model jimeng_t2v_v30 --json
```

CLI 自动轮询等待生成完成（约 30s–3min），完成后输出视频 URL。

### audio — 语音合成（需要 `audio:write` scope）

```bash
linkai audio speech "Hello, welcome to LinkAI"
linkai audio speech "你好，欢迎使用 LinkAI" --output hello.mp3   # 下载到本地
linkai audio speech "Test" --model tts-1-hd --voice alloy --json
```

---

## 权限（Scope）

登录时通过 `--scope` 指定权限范围，权限精确到每个资源的操作级别。

| Scope | 说明 | 默认授予 |
|-------|------|---------|
| `app:read` | 查询应用列表、详情 | ✅ |
| `chat:read` | 查询对话记录 | ✅ |
| `user:read` | 查询用户信息 | ✅ |
| `workflow:read` | 查询工作流 | ✅ |
| `knowledge:read` | 查询知识库 | ✅ |
| `knowledge:write` | 创建/编辑知识库 | ❌ |
| `knowledge:delete` | 删除知识库 | ❌ |
| `db:read` | 查询数据库/表/执行 SELECT | ❌ |
| `db:write` | 执行写操作（INSERT/UPDATE/DELETE） | ❌ |
| `image:write` | 生成图片 | ❌ |
| `video:write` | 生成视频 | ❌ |
| `audio:write` | 语音合成（TTS） | ❌ |

需要额外权限时重新授权：

```bash
# 数据库读写
linkai auth login --scope "db:read db:write"

# 内容生成
linkai auth login --scope "image:write video:write audio:write"

# 一次性获取所有权限
linkai auth login --scope "app:read chat:read user:read workflow:read knowledge:read knowledge:write knowledge:delete db:read db:write image:write video:write audio:write"
```

---

## 安全设计

- **Opaque Token**：服务端存储，可随时撤销
- **双 Token**：`access_token` 有效期 2 小时，`refresh_token` 有效期 7 天
- **设备绑定**：每台机器生成唯一 `device_id`，所有请求携带 `X-Device-ID` 头，服务端将 token 与 device_id 绑定

## 本地文件

```
~/.linkai/
├── config.json   # device_id、已登录用户信息
└── token.json    # access_token、refresh_token、scope、过期时间
```

## 开发

```bash
go build -o linkai .     # 构建
go build ./...           # 验证所有包可编译
go mod tidy              # 整理依赖
```

### 添加新命令

参考 `cmd/database/` 或 `cmd/image/` 的结构：

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
