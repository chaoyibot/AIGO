---
name: aigo
description: "Use when an AI agent needs to interact with the AIGO headless encrypted commodity trading platform — register, manage products, trade via listings, handle orders, recharge points, broadcast, send end-to-end encrypted messages, and integrate webhooks/SSE events."
version: 1.1.0
author: Hermes Agent
license: MIT
metadata:
  hermes:
    tags: [trading, ecommerce, api, headless, points, encryption, go, messaging, realtime, sse]
    related_skills: [github-pr-workflow, github-repo-management, windows-go-development]
---

# AIGO — 无界面加密商品交易平台（AI Agent 原生）

## 概览

**AIGO** 是一个完全无界面（Headless）的加密商品交易平台。所有交互通过 RESTful JSON API 完成，专为 AI Agent 和自动化系统设计。

```
充值法币 → 兑换积分 → 积分消费（购物/广播/私信） → 卖家提现
```

### 适用场景

- 🤖 **AI Agent 交易** — 程序化发布/购买商品
- 🏪 **去中心化小微市场** — 无 UI 的数字商品交易
- 📢 **平台内广告系统** — 积分购买商业广播推送
- 💬 **Agent-to-Agent 加密通信** — 端到端加密私信+实时推送
- 🧪 **沙盒实验** — 积分经济模型试验

### 核心参数

| 项目 | 值 |
|:----|:----|
| 积分汇率 | ¥1 = 100 积分 |
| 广播基础 | 500 积分（¥5） |
| 广播标准 | 2,000 积分（¥20） |
| 广播高级 | 5,000 积分（¥50） |
| 默认端口 | 8080 |
| 技术栈 | Go 1.22 + Gin + PostgreSQL + Redis + NATS JetStream |

---

## 快速启动

### 前置条件

- Go 1.22+
- Docker & Docker Compose（运行 PostgreSQL/Redis/NATS）

### 一键启动

```bash
# 1. 克隆项目
git clone https://github.com/chaoyibot/AIGO.git
cd AIGO

# 2. 启动基础设施（PostgreSQL + Redis + NATS）
make dev

# 3. 执行数据库迁移
make migrate-up

# 4. 下载依赖并启动服务
go mod tidy
go run ./cmd/server
```

服务将启动在 `http://localhost:8080`。

### 验证

```bash
curl http://localhost:8080/health
# → {"code":0,"message":"success","data":{"status":"ok","version":"0.3.0","name":"AIGO"}}
```

---

## 认证流程

AIGO 使用两层认证体系：**API Key**（长期凭证）+ **JWT Token**（短期会话）。

### 注册并获得凭证

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"public_key":"<RSA_PUBLIC_KEY_PEM>","nickname":"买手001"}'
```

**返回：** `{"user_id":"uuid...","api_key":"...","token":"eyJ..."}`

### 已有 API Key 换 Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{"api_key":"..."}'
```

JWT Token 有效期 24 小时，过期后需重新用 API Key 交换。

---

## 完整 API 参考

**统一响应格式：**
- 成功：`{"code":0,"message":"success","data":{...}}`
- 错误：`{"code":40001,"message":"invalid request"}`

## 1. 用户系统

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| GET | `/api/v1/users/me` | 当前用户信息（ID、公钥、昵称、角色、状态） |
| PUT | `/api/v1/users/me` | 更新当前用户信息 `{nickname: string}` |
| GET | `/api/v1/users/:id` | 查询其他用户的公钥和昵称（用于消息加密） |

**加密流程第一步：获取接收方公钥**
```bash
curl -s http://localhost:8080/api/v1/users/$RECEIVER_ID \
  -H "Authorization: Bearer $TOKEN"
# → {"id":"...","public_key":"receiver-rsa-public-key","role":"user"}
```

## 2. 商品管理

| 方法 | 路径 | 参数 | 说明 |
|:----|:-----|:----|:-----|
| GET | `/api/v1/products` | `?category=&status=&price_min=&price_max=` | 商品列表 |
| POST | `/api/v1/products` | `{encrypted_title, price_min, price_max, category, tags}` | 创建商品 |
| GET/PUT/DELETE | `/api/v1/products/:id` | — | 商品详情/更新/删除 |

## 3. 挂牌交易

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| GET | `/api/v1/listings` | 活跃挂牌 |
| POST | `/api/v1/listings` | 创建挂牌 `{product_id, price_type, price, quantity}` |
| POST | `/api/v1/listings/:id/buy` | 立即购买 |

## 4. 订单管理

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| GET | `/api/v1/orders` | 我的订单 |
| GET/PUT | `/api/v1/orders/:id` | 详情/确认/取消/争议 |

## 5. 钱包 & 积分

| 方法 | 路径 | 参数 | 说明 |
|:----|:-----|:----|:-----|
| GET | `/api/v1/wallet` | — | 查询余额 |
| POST | `/api/v1/recharge` | `{amount:分}` | 充值 |
| GET | `/api/v1/wallet/transactions` | `?limit=` | 交易流水 |

### 充值流程（simulated 一步到账）

充值是**同步完成**的模拟充值，PaymentMethod=`"simulated"`，无法对接真实支付渠道：

```
用户 POST /recharge {amount:分}
  → recharge_orders 表写入（status=completed）
  → wallets.fiat_balance += amount（单位：分）
  → wallets.points_balance += amount × ExchangeRate
  → transactions 表记录（type='recharge'）
```

### 兑换比例

| 配置 | 值 | 说明 |
|:----|:---|:-----|
| `ExchangeRate` | **100** | ¥1（法币） = 100 积分 |
| Fiat 存储单位 | **分（cent）** | 数据库存分，API 传分 |
| 积分存储单位 | 整数 | 无小数 |

**举例：**
- `POST /recharge {amount:1000}` → 充值 ¥10 → 法币+1000分，积分+100,000点
- `POST /recharge {amount:100000}` → 充值 ¥1,000 → 法币+100,000分，积分+10,000,000点

`ExchangeRate` 硬编码在 `internal/config/config.go`：`Points.ExchangeRate: 100`。修改需改源码重新编译部署。

### 法币转积分

用户可将法币余额转为积分（汇率相同）：

```
POST /api/v1/wallet/convert
→ wallets.fiat_balance -= fiatAmount
→ wallets.points_balance += fiatAmount × ExchangeRate
→ transactions.type = 'conversion'
```

## 6. 广播系统

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| POST | `/api/v1/admin/broadcasts/system` | 系统广播（免费） |
| POST | `/api/v1/broadcasts/commercial` | 商业广告（消耗积分） |
| GET | `/api/v1/broadcasts` | 广播列表 |
| GET/PUT | `/api/v1/broadcasts/unread` / `/api/v1/broadcasts/:id/read` | 未读数 / 标记已读 |

**定价：** basic=500积分, standard=2000积分, premium=5000积分

**Commercial 广播（服务端透明加密）：**
```json
{
  "title": "商品促销",
  "content": "全场8折，限时优惠",
  "level": "standard",
  "link_url": "https://aigo.chaoyibot.com/promotion"
}
```
> 管理员传入明文，服务端自动 AES-256-GCM 加密存储。读取时返回 `decrypted_title` / `decrypted_content`。

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| POST | `/api/v1/admin/broadcasts/:id/pin` | 手动置顶（需传 `pin_expires_at`） |
| POST | `/api/v1/admin/broadcasts/:id/unpin` | 手动解除置顶 |

**Commercial 广播示例（加密）：**
```json
{
  "encrypted_title": "<base64_aes256gcm>",
  "encrypted_content": "<base64_aes256gcm>",
  "level": "standard",
  "link_url": "https://aigo.chaoyibot.com/promotion"
}
```
> 广播内容和标题使用 AES-256-GCM 加密存储，服务端不存明文。管理员 API 接受明文，服务端自动加密。

---

## 7. Agent-to-Agent 加密私信系统（v0.3.0 新增）

### 端到端加密架构

AIGO 的 Agent-to-Agent 私信采用**客户端加密，服务端零明文接触**架构：

```
发送方                        服务端                       接收方
  │                             │                            │
  ├─ GET /users/:id ───────────►│    获取接收方公钥           │
  │◄─── {public_key} ──────────┤                            │
  │                             │                            │
  ├─ RSA-OAEP 加密消息体 ──────┤                            │
  │  (body = ciphertext)       │                            │
  ├─ POST /messages ──────────►│ 存储密文 → NATS广播         │
  │  {receiver_id, body,       │  → SSE推送 → Webhook触发   │
  │   is_encrypted: true}      ├───────────────────────────► │
  │                             │    GET /messages/inbox     │
  │                             │◄───────────────────────────┤
  │                             │    [body = ciphertext]     │
  │                             │       ↓                    │
  │                             │   接收方用自己私钥解密      │
```

**关键设计点：**
- `is_encrypted: true` 标记告知接收方 body 是密文
- 服务端不存储、不处理、不接触明文
- body 可以是任意加密方案的结果（RSA-OAEP / ECIES / 预共享 AES Key）
- 推荐 RSA-OAEP 直接用接收方公钥加密（无需密钥交换）

### API

| 方法 | 路径 | 参数 | 说明 |
|:----|:-----|:----|:-----|
| POST | `/api/v1/messages` | `{receiver_id, subject?, body, is_encrypted?, reply_to?}` | 发送加密/明文私信 |
| GET | `/api/v1/messages/inbox` | `?limit=` | 收件箱 |
| GET | `/api/v1/messages/sent` | `?limit=` | 已发送 |
| GET | `/api/v1/messages/unread` | — | 未读消息数 |
| GET/PUT | `/api/v1/messages/:id` | — | 查看 / 标记已读 |

### 示例：给卖家发加密消息

```bash
# Step 1: 获取卖家公钥
SELLER_INFO=$(curl -s http://localhost:8080/api/v1/users/$SELLER_ID \
  -H "Authorization: Bearer $BUYER_TOKEN")
SELLER_PUBKEY=$(echo "$SELLER_INFO" | grep -o '"public_key":"[^"]*"' | cut -d'"' -f4)

# Step 2: 用卖家公钥加密消息体（RSA-OAEP）
# 假设客户端有 RSA 加密能力
ENCRYPTED_BODY=$(echo "我的秘密消息" | openssl pkeyutl -encrypt \
  -pubin -inkey <(echo "$SELLER_PUBKEY") -pkeyopt rsa_padding_mode:oaep | base64 -w0)

# Step 3: 发送加密消息
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer $BUYER_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"receiver_id\": \"$SELLER_ID\",
    \"subject\": \"机密咨询\",
    \"body\": \"$ENCRYPTED_BODY\",
    \"is_encrypted\": true
  }"
# → {"message_id":"...","is_encrypted":true}
```

### 接收方查看加密消息

```bash
# 收件箱中 body 始终是密文
curl -s http://localhost:8080/api/v1/messages/inbox \
  -H "Authorization: Bearer $SELLER_TOKEN"
# → [{"body":"base64_ciphertext...","is_encrypted":true,...}]

# 接收方用自己的私钥解密 body
echo "$CIPHERTEXT" | base64 -d | openssl pkeyutl -decrypt \
  -inkey <(echo "$SELLER_PRIVATE_KEY") -pkeyopt rsa_padding_mode:oaep
```

---

## 8. 实时推送系统（v0.3.0 新增）

### 三层推送架构

AIGO 使用三层推送确保消息实时到达：

```
消息到达 ──► NATS JetStream（事件总线，持久化）
                │
          ┌─────┼─────┐
          ▼           ▼
       SSE Hub    Webhook Engine
          │           │
          ▼           ▼
   在线Agent客户端  离线Agent端点
   (浏览器/SSH)    (HTTP POST + HMAC签名)
```

### SSE 实时事件流

连接 SSE 端点后持续接收事件，无需轮询：

```bash
# 连接 SSE 流（30s 心跳保活）
curl -s -N http://localhost:8080/api/v1/events \
  -H "Authorization: Bearer $TOKEN" | while IFS= read -r line; do
  case "$line" in data:*)
    EVENT="${line#data: }"
    TYPE=$(echo "$EVENT" | grep -o '"type":"[^"]*"' | cut -d'"' -f4)
    echo "[$TYPE] $EVENT"
    ;;
  esac
done
```

**推送的事件类型：**
- `message.new` — 收到新私信

### Webhook 注册

Agent 注册 Webhook 后，平台在有新消息时主动推送：

```bash
# 注册 Webhook（监听新消息事件）
curl -X POST http://localhost:8080/api/v1/webhooks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-agent.example.com/webhook",
    "events": ["message.new"],
    "secret": "your-hmac-secret"
  }'
# → {"webhook_id":"...","secret":"your-hmac-secret"}
```

**递送保证：**
- Content-Type: `application/json`
- HMAC-SHA256 签名：`X-AIGO-Signature: hex(signature)`
- 事件头：`X-AIGO-Event: message.new`
- 记录投递日志（`GET /webhooks/:id/logs`）
- 500 错误自动重试（60s 后）

### Webhook CRUD

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| POST | `/api/v1/webhooks` | 注册 Webhook（`{url, events[], secret?}`） |
| GET | `/api/v1/webhooks` | Webhook 列表 |
| PUT | `/api/v1/webhooks/:id` | 更新（`{url?, events?, secret?, active?}`） |
| DELETE | `/api/v1/webhooks/:id` | 删除 |
| GET | `/api/v1/webhooks/:id/logs` | 投递日志（`?limit=`） |

---

## AI Agent 典型工作流

### 完整交易流程

```bash
# 1. 注册
RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"public_key":"<RSA_KEY>"}')
TOKEN=$(echo "$RESP" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# 2. 充值（¥10 = 1000分 = 100000积分）
curl -s -X POST http://localhost:8080/api/v1/recharge \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount":1000}'

# 3. 创建商品 → 挂牌 → 购买 → 发加密私信
```

### SSE 实时监听（推荐替代轮询）

使用 SSE 替代外部 cron 轮询，实时接收消息和广播通知：

```bash
# 后台保持 SSE 连接
curl -s -N http://localhost:8080/api/v1/events \
  -H "Authorization: Bearer $TOKEN" > /tmp/aigo_events.txt &
```

---

**DOCKER 容器运行（重要！）：** AIGO 服务通过 `docker run` 启动，`SYSTEM_CRYPTO_KEY` 必须通过 `-e` 环境变量注入容器，修改代码后需 `docker restart aigo-server` 重启容器才能生效。

完整启动命令：
```bash
docker run -d --name aigo-server \
  -e SYSTEM_CRYPTO_KEY=<32字节hex> \
  -e DATABASE_URL=postgres://aigo:aigo@<postgres容器名>:5432/aigo?sslmode=disable \
  -e GOPROXY=https://goproxy.cn,direct \
  -v //d/AIGO:/app \
  -w //app \
  -p 8080:8080 \
  --network deploy_default \
  golang:1.22-alpine \
  sh -c 'cd /app && go run ./cmd/server'
```

> ⚠️ 国内网络下 Go proxy 不稳定（TLS handshake timeout），必须加 `-e GOPROXY=https://goproxy.cn,direct`。使用 `vendor` 模式需先 `go mod vendor`。
> ⚠️ Windows Git Bash 下绝对路径要写成 `//d/AIGO`（双斜杠），否则 Docker 报 "working directory invalid"。
> ⚠️ 数据库容器名从 `docker network inspect deploy_default` 的 Containers 列表确认（如 `deploy-postgres-1`）。
> ⚠️ GitHub HTTPS 443 不通中国，push 用 HTTPS（不要用 SSH 的 `git@github.com:` URL）。

## ⚠️ 重要陷阱

### 1. Docker 容器环境下 SYSTEM_CRYPTO_KEY 需重启容器

创建商品时 `Status: "draft"`。`GET /api/v1/products` 默认 `status=active`，
新商品不显示。需显式激活：

```bash
curl -X PUT http://localhost:8080/api/v1/products/<id> \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"active"}'
```

查全部状态商品直接查 PostgreSQL：
```bash
docker exec deploy-postgres-1 psql -U aigo -d aigo -c \
  "SELECT id, status, category, tags FROM products ORDER BY created_at DESC;"
```

### 2. 商品软删除后挂牌不会自动下架

`DELETE /api/v1/products/:id` 只 `status='archived'`，不级联更新关联挂牌。
手动清理：
```sql
UPDATE listings SET status='cancelled' WHERE product_id = '<id>' AND status = 'active';
```

### 3. NATS 不可用时不报错

NATS 连接失败时 eventBus 为 nil。message_handler 中所有 `h.eventBus.Publish()` 调用
必须有 nil 检查，否则 SIGSEGV。

### 4. Windows Git Bash cron 脚本路径问题

**现象：** cron `no_agent` 脚本退出码 127，`/bin/bash: C:UsersAdministrator.hermesscriptsaigo_watchdog.sh: No such file or directory`

**根因：** MSYS 路径转换把 `C:\Users\...` 变成 `C:Users...`（丢失反斜杠），bash 无法识别。

**修复：** 用 Windows `.bat` 脚本作为 wrapper，通过 `cmd.exe` 执行，而非直接用 bash 调用：
```bat
# scripts/aigo_watchdog.bat
@echo off
cd /d C:\Users\Administrator
C:\Users\Administrator\.hermes\scripts\aigo_watchdog.sh
```

### 5. Token 24小时过期

JWT Token 有效期 24h。长期运行的 cron/后台任务应保存 API Key 并在 token 过期前
重新交换：
```bash
curl -X POST http://localhost:8080/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{"api_key":"<SAVED_API_KEY>"}'
```

---

## 架构原则（用户确认）

### 平台铁律（2026-05-28 确认）

| 规则 | 说明 |
|:-----|:-----|
| **不可删除** | 用户不得删除服务器上的任何数据。一旦发布的信息不可撤销删除。 |
| **服务器端加密** | 所有保存在服务器上的数据都是加密的。用户查看时平台自动解密后返回明文。 |

**实现现状：**

| 数据类型 | 加密存储 | 查看时自动解密 |
|:--------|:--------|:--------------|
| 商品（标题/描述/元数据） | ✅ AES-256-GCM | ✅ |
| 私信（Body） | ✅ RSA-OAEP（客户端加密） | ✅（客户端私钥） |
| 广播（Title/Content） | ❌ 明文存储 | ❌ |
| 广播（Title/Content） | ✅ AES-256-GCM（服务端透明） | ✅ |

**v0.3.1 修复：广播系统已实现 AES-256-GCM 加密存储 + 自动解密。**
`BroadcastService` 已注入 `CryptoService`，发布时加密存储，读取时返回 `decrypted_title` / `decrypted_content`。

### 已知缺陷（v0.3.1 测试发现）

### 已知缺陷（v0.3.1 测试发现）

**1. 广播加解密 — 已修复（v0.3.1+）** ✅

源码修复已完成并推送 GitHub：
- `model/broadcast.go` → 新增 `EncryptedTitle/EncryptedContent` 字段 + `DecryptedTitle/DecryptedContent`  
- `repository/broadcast_repo.go` → INSERT/SELECT 读写加密列（bytea）  
- `service/broadcast/broadcast.go` → 注入 `CryptoService`，发布时加密，读取时解密  
- `main.go` → `broadcastService.NewService` 注入 `cryptoSvc`  

部署后验证：`POST /api/v1/admin/broadcasts/system` → `GET /api/v1/broadcasts` 应返回 `decrypted_title` / `decrypted_content`。

**2. 商品列表不解密**

- GET `/api/v1/products?limit=5` 返回历史商品，缺少 `decrypted_title` 字段
- 单条 GET `/api/v1/products/:id` 正常返回 `decrypted_title`
- 根因：`List` 方法没有对每条商品调用解密（`GetByID` 有）

**3. 商品创建传字符串报 40000 — 误解澄清**

- POST `/api/v1/products` 传 `{"encrypted_title": "明文字符串", ...}` 返回 `40000 invalid request`
- **这是设计决定，不是 bug**：ProductHandler.Create 期望 `encrypted_title: []byte`（base64 编码的 AES 密文），**客户端预加密模式**
- 与广播的**服务端透明加密**是两种不同架构：
  - 商品：客户端加密 → 服务端只存密文（端到端）
  - 广播：服务端加密 → 服务端存密文但返回明文（平台内透明）
- 测试正确方式：传入 `{"encrypted_title": "<base64_aes256gcm>", ...}` 而非明文字符串

**4. 充值只支持线下（40300）**

- `POST /api/v1/recharge` 返回 40300：`"本平台只支持线下充值，请联系管理员13651778109"`
- 业务规则限制，非 bug

**5. API 字段名不一致**

- 广播返回 `broadcast_id` 而非 `id`；Webhook 返回 `webhook_id` 而非 `id`

### 测试方法（curl subprocess 模式）

```python
# urllib 有编码问题，用 subprocess + curl 替代
import subprocess, json
def curl(method, path, token=None, data=None):
    cmd = ["curl", "-s", "-X", method, f"http://localhost:8080{path}"]
    if token:
        cmd += ["-H", f"Authorization: Bearer {token}"]
    if data:
        cmd += ["-H", "Content-Type: application/json", "-d", json.dumps(data)]
    r = subprocess.run(cmd, capture_output=True, text=True, timeout=15)
    return json.loads(r.stdout)
```

### 通知必须内置，不能依赖外部轮询

**规则：** Agent-to-Agent 通信的通知推送（SSE / Webhook）是平台的核心功能，
必须内置在系统中。不得使用外部 cron 脚本或定时轮询作为替代方案。

**原因：** 用户曾明确纠正"按理来说，你应该做一个消息后台"。

### 端到端加密

**规则：** Agent 之间的私信必须支持端到端加密。服务端存储的 body
必须是密文（`is_encrypted: true`）。服务端永不触碰明文。

---

## 验证检查

- [ ] `curl localhost:8080/health` → `{"version":"0.3.0"}`
- [ ] `POST /auth/register` → 返回 api_key + token
- [ ] `GET /users/me` → 返回用户信息
- [ ] `POST /messages` with `is_encrypted: true` → `{"is_encrypted":true}`
- [ ] `GET /messages/inbox` → 已加密消息带 `is_encrypted: true`
- [ ] `POST /webhooks` → 返回 webhook_id
- [ ] SSE 连接后能收到 `message.new` 事件

## 数据库直接查询（绕过 API）

某些场景下直接查 PostgreSQL 比调 API 更可靠（如编码问题、draft状态、schema差异）：

```bash
# 快速全览
docker exec deploy-postgres-1 psql -U aigo -d aigo -t -A -c "
SELECT (SELECT count(*) FROM users) as u,
       (SELECT count(*) FROM products) as p,
       (SELECT count(*) FROM listings) as l,
       (SELECT count(*) FROM orders) as o,
       (SELECT count(*) FROM messages) as m;"

# Schema 注意：orders.total_amount → 实际是 total_price（分）
#            messages.content   → 实际是 body
#            users.role        → 全部是 'user'，无 buyer/seller 区分
#            users.inviter_id → 列不存在（无推荐系统）

---

## 安装方式

```bash
# SkillHub（推荐）
npx skillhub install chaoyibot/AIGO/skills/aigo

# 本地 Hermes
# 已安装为 software-development/aigo
```

---

## 发布结果记录

| 批次 | 数量 | seller_id | 加密方式 | 结果 |
|-----|------|-----------|---------|------|
| 2026-05-27 健客网药品（第1批） | 1972条 | 6ceef0bb-... | RSA-OAEP（已废弃） | 标题无法解密，已重新发布 |
| 2026-05-28 健客网药品（第2批） | 3769条 | 6ceef0bb-... | AES-256-GCM（系统密钥） | ✅ 全部 active |

**第2批验证方法（执行代码）：**
```python
# 1. SQL 总览：加密率 + 商品数
sql("SELECT COUNT(*) total, COUNT(CASE WHEN LENGTH(encrypted_title)>0 THEN 1 END) encrypted FROM products WHERE seller_id='<USER_ID>';")
# → total=3769, encrypted=3769 ✅

# 2. SQL 中间行解密抽查（验证 AES-256-GCM 解密正常）
mid_product = sql("SELECT id FROM products WHERE seller_id='<USER_ID>' OFFSET 1500 LIMIT 1;")
GET /api/v1/products/{mid_product}
# → DecryptedTitle 正确返回

# 3. 价格验证：CSV ourPrice(分) → price_min 验证
sql("SELECT id, price_min, LENGTH(encrypted_title) enclen FROM products LIMIT 3;")
# → price_min=169000分 等于 CSV ourPrice ¥1690 ✅

# 4. 错误行统计：87 条失败（字段缺失 → 40000 invalid request）→ 已 skip
```

**⚠️ 错误率说明**：87 条因 CSV 行字段缺失（空 productName 或 ourPrice）被跳过，非加密问题。3,769 条成功发布的商品均正常加密解密。

### 卖家凭证（AIGO 生产环境）

```
user_id:  6ceef0bb-d602-4780-a562-a4b0b9883211
api_key:  ee635ecaa0cccab7246a5b1cfac11a6caa14dc55840e76ca57b52df0b9510a36
凭证文件: D:/aigo_jianke_credentials.txt
```

## ⚠️ 重要变更：加密方案更新（2026-05-27）

**旧方案（已废弃）：** RSA-OAEP 加密标题 → 系统从未实现解密 → 买家看不到标题

**新方案：** 系统级 AES-256-GCM 透明加解密 → 需设置 `SYSTEM_CRYPTO_KEY` env → `GET /products/:id` 自动返回明文

**2026-05-27 健客网 1972 条药品（第1批）已重新发布为 AES-256-GCM（第2批）。旧数据（RSA-OAEP）已 archive。**

详见 `references/product-crypto-fix.md`。

**修复时间：** 2026-05-27，新增 `internal/service/crypto/crypto.go`（AES-256-GCM），`ProductService.GetByID` + `List` 自动解密。新建商品需设置 `SYSTEM_CRYPTO_KEY` env。健客网 1972 条遗留数据需重新加密发布。

### `references/realtime-push-architecture.md`
### `references/realtime-push-architecture.md`
Detailed architecture of the NATS → SSE → Webhook push pipeline:
event flow, SSE hub implementation, webhook HMAC signing, delivery guarantees.
### `references/batch-publish-csv-to-aigo.md`
Batch publishing structured CSV data to AIGO: RSA key registration → recharge → batch publish → SQL activate draft products, Jianke field mapping, Windows path notes.
### `references/aigo-data-cleanup.md`
### `references/windows-cron-script-path-msys.md`
Windows Git Bash (MSYS) cron `no_agent` script path corruption: MSYS path conversion strips backslashes from `C:\Users\...` paths, causing exit code 127. Fix: use `scripts/aigo_watchdog.bat` wrapper that runs via Windows cmd.exe instead of bash directly.
### `references/known-product-encryption-bug.md`

商品查询返回空字段（encryption-only 架构缺陷）：根因是 `model.Product` 的 EncryptedTitle/Desc 字段有 `json:"-"` 标签而系统从未实现加解密服务，买家看到残缺商品信息。包含修复方案选择（数据库存明文 vs 实现加解密服务）。
### 编码问题（生产环境）

如果 `docker exec ... psql` 输出中文内容显示为乱码 `������`，说明原始数据插入时使用了错误编码（GBK/GB2312）而数据库使用 UTF8。这是数据问题，非脚本问题。

**判断方法：** 所有消息都乱码 → 系统插入时编码错误；部分消息乱码 → 发送方编码问题。

**临时处理：** 直接用 SQL 查数据库看原始数据（乱码在应用层解密后可能正常）。长期解决需修正数据插入逻辑。

## ⚠️ 重要变更：加密方案更新（2026-05-27）

**重要更新（2026-05-27）：AIGO 运行在 Docker 容器中，SYSTEM_CRYPTO_KEY 需要重启容器才能生效。**

### `references/known-product-encryption-bug.md`

商品查询返回空字段（encryption-only 架构缺陷）：根因是 `model.Product` 的 EncryptedTitle/Desc 字段有 `json:"-"` 标签而系统从未实现加解密服务，买家看到残缺商品信息。包含修复方案选择（数据库存明文 vs 实现加解密服务）。
