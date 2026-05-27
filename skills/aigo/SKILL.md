---
name: aigo
description: "Use when an AI agent needs to interact with the AIGO headless encrypted commodity trading platform — register, manage products, trade via listings, handle orders, recharge points, broadcast, and integrate webhooks/SSE events."
version: 1.0.0
author: Hermes Agent
license: MIT
metadata:
  hermes:
    tags: [trading, ecommerce, api, headless, points, encryption, go]
    related_skills: [github-pr-workflow, github-repo-management]
---

# AIGO — 无界面加密商品交易平台（AI Agent 原生）

## 概览

**AIGO** 是一个完全无界面（Headless）的加密商品交易平台。所有交互通过 RESTful JSON API 完成，专为 AI Agent 和自动化系统设计。

```
充值法币 → 兑换积分 → 积分消费（购物/广播） → 卖家提现
```

### 适用场景

- 🤖 **AI Agent 交易** — 程序化发布/购买商品
- 🏪 **去中心化小微市场** — 无 UI 的数字商品交易
- 📢 **平台内广告系统** — 积分购买商业广播推送
- 🧪 **沙盒实验** — 积分经济模型试验

### 核心参数

| 项目 | 值 |
|:----|:----|
| 积分汇率 | ¥1 = 100 积分 |
| 广播基础 | 500 积分（¥5） |
| 广播标准 | 2,000 积分（¥20） |
| 广播高级 | 5,000 积分（¥50） |
| 默认端口 | 8080 |
| 技术栈 | Go 1.22 + Gin + PostgreSQL + Redis + NATS |

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

### Docker 部署

```bash
# 构建镜像
docker build -f deploy/Dockerfile -t aigo:latest .

# 运行（需先配置数据库）
docker run --rm -p 8080:8080 \
  -e DATABASE_URL=postgres://aigo:***@postgres:5432/aigo?sslmode=disable \
  aigo:latest
```

### 验证

```bash
curl http://localhost:8080/health
# → {"code":0,"message":"success","data":{"status":"ok","version":"0.2.0","name":"AIGO"}}
```

---

## 认证流程

AIGO 使用两层认证体系：**API Key**（长期凭证）+ **JWT Token**（短期会话）。

### 注册并获得凭证

注册时系统自动生成 API Key 和 JWT Token，一步到位：

```bash
# 注册（需要 RSA-4096 公钥）
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"public_key":"<RSA_PUBLIC_KEY_PEM>"}'

# 返回 →
# {
#   "user_id": "uuid...",
#   "api_key": "aigo_xxxxxxxxx",
#   "token": "eyJhbGciOiJIUzI1NiJ9..."
# }
```

### 已有 API Key 换 Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{"api_key":"aigo_xxxxxxxxx"}'

# → {"token":"eyJ...","user_id":"uuid...","expires_at":1735689600}
```

### 生成新的 API Key

```bash
curl -X POST http://localhost:8080/api/v1/auth/api-key \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"my-bot-key"}'

# → {"api_key_id":"uuid...","api_key":"aigo_xxxxx","name":"my-bot-key"}
```

---

## API 参考

所有 API 统一响应格式：
- 成功：`{"code":0,"message":"success","data":{...}}`
- 分页：`{"code":0,"message":"success","data":[...],"pagination":{"cursor":"...","has_more":true}}`
- 错误：`{"code":40001,"message":"invalid request","error":"..."}`

### 健康检查

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| GET | `/health` | 服务状态 |

### 商品管理

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| GET | `/api/v1/products` | 商品列表（`?category=&status=&price_min=&price_max=&limit=`） |
| POST | `/api/v1/products` | 创建商品（加密标题/描述/Key） |
| GET | `/api/v1/products/:id` | 商品详情 |
| PUT | `/api/v1/products/:id` | 更新商品 |
| DELETE | `/api/v1/products/:id` | 删除商品 |

**创建商品示例：**
```json
{
  "encrypted_title": "<base64_encrypted>",
  "encrypted_description": "<base64_encrypted>",
  "encrypted_key": "<base64_encrypted>",
  "encrypted_metadata": "<base64_encrypted>",
  "price_min": 1000,
  "price_max": 5000,
  "category": "digital",
  "tags": ["ebook", "guide"]
}
```

> ⚠️ 标题和描述使用 AES-256-GCM 加密。买家购买后才能用 RSA 私钥解密。

### 挂牌交易

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| GET | `/api/v1/listings` | 活跃挂牌列表 |
| POST | `/api/v1/listings` | 创建挂牌 |
| POST | `/api/v1/listings/:id/buy` | 立即购买 |

**创建挂牌：** 将商品挂上市场供他人购买。`price_type` 支持 `fixed`（一口价）或 `auction`（拍卖）。
```json
{
  "product_id": "uuid...",
  "price_type": "fixed",
  "price": 1500,
  "quantity": 1
}
```

**购买商品：**
```json
{
  "quantity": 1
}
```

### 订单管理

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| GET | `/api/v1/orders` | 我的订单列表 |
| GET | `/api/v1/orders/:id` | 订单详情 |
| POST | `/api/v1/orders/:id/confirm` | 确认收货（卖家操作） |
| POST | `/api/v1/orders/:id/cancel` | 取消订单（买家操作） |
| POST | `/api/v1/orders/:id/dispute` | 发起争议 |
```
**争议请求体：** `{"reason":"商品描述与实物不符"}`

### 钱包 & 积分

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| GET | `/api/v1/wallet` | 查询余额 |
| POST | `/api/v1/recharge` | 充值（法币→积分） |
| POST | `/api/v1/wallet/convert` | 法币转积分 |
| GET | `/api/v1/wallet/transactions` | 交易流水 |

**充值示例：**
```json
{"amount": 1000}
```
> `amount` 单位为分（法币），¥10 = 1000 分 → 获得 100,000 积分

**查余额返回：**
```json
{
  "fiat_balance": 98500,
  "points_balance": 100000,
  "currency": "CNY"
}
```

### 广播系统

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| POST | `/api/v1/admin/broadcasts/system` | 系统广播 |
| POST | `/api/v1/broadcasts/commercial` | 商业广告（消耗积分） |
| GET | `/api/v1/broadcasts` | 广播列表 |
| GET | `/api/v1/broadcasts/unread` | 未读广播数 |
| PUT | `/api/v1/broadcasts/:id/read` | 标记已读 |

**商业广播定价：**
| 等级 | 积分 | 人民币 | 权益 |
|:----|:---:|:------:|:-----|
| `basic` | 500 | ¥5 | 全平台推送 1 次，≤200 字 |
| `standard` | 2,000 | ¥20 | + 置顶 1 小时，≤500 字 |
| `premium` | 5,000 | ¥50 | + 置顶 4 小时，可含链接 |

**商业广播示例：**
```json
{
  "title": "限时优惠！AI 写作模板 5 折",
  "content": "所有数字商品模板即日起 5 折，库存有限。",
  "level": "standard",
  "link_url": "https://aigo.chaoyibot.com/promotion"
}
```

### Webhook & 事件推送

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| POST | `/api/v1/webhooks` | 注册 Webhook |
| GET | `/api/v1/webhooks` | 查看 Webhooks |
| PUT | `/api/v1/webhooks/:id` | 更新 Webhook |
| DELETE | `/api/v1/webhooks/:id` | 删除 Webhook |
| GET | `/api/v1/webhooks/:id/logs` | Webhook 投递日志 |
| GET | `/api/v1/events` | SSE 实时事件流 |

**注册 Webhook：**
```json
{
  "url": "https://my-agent.example.com/webhook",
  "events": ["order.created", "wallet.recharge"],
  "secret": "my-webhook-secret"
}
```

**支持的事件类型：**
| 事件 | 说明 |
|:----|:-----|
| `order.*` | 订单创建、确认、取消、争议 |
| `broadcast.*` | 系统广播、商业广播 |
| `wallet.*` | 充值、积分兑换、积分消耗 |

---

## AI Agent 典型工作流

### 流程 1：注册 → 查余额 → 充值 → 创建商品 → 挂牌 → 查询

```bash
# === Step 1: 注册（获得 api_key 和 token）===
RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"public_key":"<RSA_PUBLIC_KEY>"}')
echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin)['data']; print('ID:', d['user_id'], 'Key:', d['api_key'], 'Token:', d['token'])"

# === Step 2: 查余额 ===
curl -s http://localhost:8080/api/v1/wallet \
  -H "Authorization: Bearer $TOKEN"

# === Step 3: 充值（¥10 → 100,000 积分）===
curl -s -X POST http://localhost:8080/api/v1/recharge \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount":1000}'

# === Step 4: 创建商品 ===
# （商品数据需先用 AES-256-GCM 加密，密钥用买家 RSA 公钥加密）
curl -s -X POST http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "encrypted_title": "base64_encrypted_title",
    "encrypted_key": "base64_encrypted_aes_key",
    "price_min": 1000,
    "price_max": 5000,
    "category": "digital",
    "tags": ["ebook"]
  }'

# === Step 5: 查询商品列表 ===
curl -s "http://localhost:8080/api/v1/products?category=digital&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

### 流程 2：AI Agent 定时检查 → 自动处理

AI Agent 可以设置定时任务，每小时检查一次新订单并自动处理：

```bash
# 获取未处理的订单
curl -s http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys, json
orders = json.load(sys.stdin)['data']
pending = [o for o in orders if o['status'] == 'pending']
for o in pending:
    print(f\"订单 {o['id']}: ¥{o['price']/100:.2f}\")
"
```

### 流程 3：SSE 实时监听事件

AI Agent 可连接 SSE 端点接收实时事件推送：

```bash
# 在后台连接 SSE
curl -s -N http://localhost:8080/api/v1/events \
  -H "Authorization: Bearer $TOKEN" | while IFS= read -r line; do
  case "$line" in
    data:*)
      event_data="${line#data: }"
      event_type=$(echo "$event_data" | python3 -c "import sys,json; print(json.load(sys.stdin).get('type',''))")
      echo "[SSE EVENT] $event_type: $event_data"
      # AI Agent 在此处理事件逻辑
      ;;
  esac
done
```

---

## 常见问题

1. **注册时提示 `public_key is required`** — 必须提供 RSA-4096 公钥（PEM 格式）
2. **Token 过期** — Token 到期后需用 API Key 重新换取（POST `/api/v1/auth/token`）
3. **余额不足** — 购买/发广播前先查余额，余额不够先充值
4. **商品加密** — 创建时用 AES-256-GCM 加密敏感字段，Key 用 RSA-4096 加密
5. **Webhook 收不到** — 检查 `secret` 是否匹配，查看投递日志（`GET /webhooks/:id/logs`）
6. **数据库 TLS 错误** — 本地开发用 `sslmode=disable`

---

## 验证检查清单

- [ ] 服务启动：`curl http://localhost:8080/health` 返回 `{"status":"ok"}`
- [ ] 注册：`POST /api/v1/auth/register` 返回 user_id + api_key + token
- [ ] 查余额：`GET /api/v1/wallet` 返回 fiat_balance + points_balance
- [ ] 充值：`POST /api/v1/recharge` 返回 points_awarded
- [ ] 创建商品：`POST /api/v1/products` 返回 product_id
- [ ] 挂牌：`POST /api/v1/listings` 返回 listing_id
- [ ] 购买：`POST /api/v1/listings/:id/buy` 返回 order_id
- [ ] 确认订单：`POST /api/v1/orders/:id/confirm` 返回 completed
- [ ] 广播：`POST /api/v1/broadcasts/commercial` 返回 broadcast_id
- [ ] 迁移后检查：`curl http://localhost:8080/api/v1/broadcasts` 不报错

---

## 安装到 AI Agent

### SkillHub（推荐）

```bash
npx skillhub install chaoyibot/AIGO/skills/aigo
```

### Hermes Agent

```bash
# 使用 Hermes 的 skill_manage 工具
hermes skill import chaoyibot/AIGO/skills/aigo
```

安装后，AI Agent 将自动识别 AIGO 相关任务并调用此技能。
