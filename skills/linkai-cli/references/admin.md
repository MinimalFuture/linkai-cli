# Apps / Models / Account / Credits / Auth

## Apps — scope `app:read`

| Command | Notes |
|---|---|
| `linkai app list [--key <kw>] [--page <n>] [--page-size <n>] [--json]` | search + paginate |
| `linkai app detail <app_code> [--json]` | full info for one app |

## Models — uses default auth

```
linkai model list [--type LLM|EMBEDDING] [--supplier <name>] [--json]
```

## Account — scope `user:read`

```
linkai account info [--json]
```

Returns `{ name, credits, plan_version, ... }`.

## Credits

| Command | Scope | Agent notes |
|---|---|---|
| `linkai account credits [--json]` | `score:read` | available credit packages |
| `linkai account recharge --product <id> --pay wechat\|alipay --agent [--json]` | `score:buy` | **always pass `--agent`** — returns `qr_file` (a local PNG path), not an ASCII QR |
| `linkai account order <order_no> [--json]` | `score:read` | poll order status by order number |
| `linkai account orders [--page <n>] [--page-size <n>] [--json]` | `score:read` | purchase history |

`recharge --agent` writes the payment QR to a local PNG and returns its path as
`qr_file` (plus `code_url`, the raw pay URL). Show the image to the user; once
they pay, poll `account order <order_no>` until `status` becomes `PAID`.

## Auth — agent rules

| Command | Agent action |
|---|---|
| `linkai auth status [--json]` | run before the first call; expect `valid` |
| `linkai auth login` | **never run from agent** — needs a browser; tell the user to run it themselves with the right `--scope` |
| `linkai auth logout` | only on explicit user request |

## Default scopes granted at login

```
app:read chat:send user:read workflow:read workflow:run knowledge:read db:read image:gen video:gen audio:gen plugin:read plugin:run score:read score:buy
```

Sensitive scopes **not** in defaults — require explicit user re-login: `db:write`, `knowledge:create`, `knowledge:delete`.
