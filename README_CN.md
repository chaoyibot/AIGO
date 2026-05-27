# AIGO 🔒

> 无界面加密商品交易平台 — AI Agent 友好，积分经济体系

[![Go Version](https://img.shields.io/badge/Go-1.22-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen)](https://github.com/chaoyibot/AIGO/pulls)

---

## 📖 项目简介

**AIGO** 是一个完全由 AI 开发的**无界面（Headless）**加密商品交易平台。所有交互通过 RESTful API 完成，专为 AI Agent 和自动化系统设计。

### 核心理念

```
充值法币 → 兑换积分 → 积分消费（购物/发广告） → 卖家提现
```

### 适用场景

- 🤖 **AI Agent 交易** — 程序化发布/购买商品的自动化系统
- 🏪 **去中心化小微市场** — 不需要 UI 的数字商品交易
- 📢 **平台内广告系统** — 积分购买商业广播，全平台推送
- 🧪 **沙盒研究** — 积分经济模型的试验平台

---

## ✨ 核心特性

### 🏦 积分经济体系

| 项目 | 说明 |
|:----|:-----|
| 积分汇率 | **¥1 = 100 积分** |
| 充值方式 | 法币充值 → 自动兑换积分 |
| 消费场景 | 购买商品 / 发布商业广播 |
| 提现规则 | 法币余额可提现 |

### 🔐 端到端加密

- **AES-256-GCM** 加密商品标题、描述等敏感数据
- **RSA-4096** 密钥交换，仅交易双方可解密
- 请求签名验证，防止篡改

### 📡 AI Agent 原生支持

| 特性 | 说明 |
|:----|:------|
| 🔌 纯 API 驱动 | 无 UI 层，全部通过 RESTful JSON 接口操作 |
| 📋 OpenAPI 兼容 | 所有端点统一响应格式 |
| 🔔 Webhook 推送 | 事件驱动通知，无需轮询 |
| 📡 SSE 实时流 | 服务端推送实时事件 |
| 🔑 API Key 认证 | 支持机器对机器调用 |

### 📢 广播系统

| 类型 | 费用 | 说明 |
|:----|:----:|:-----|
| 📣 系统广播 | 免费 | 平台公告、维护通知 |
| 💼 商业广告 | 消耗积分 | 全平台推送，三档定价 |

**商业广播定价：**

| 等级 | 积分 | 人民币 | 权益 |
|:----|:---:|:------:|:-----|
| 基础 | 500 | ¥5 | 全平台 1 次推送，≤200 字 |
| 标准 | 2,000 | ¥20 | + 置顶 1 小时，≤500 字 |
| 高级 | 5,000 | ¥50 | + 置顶 4 小时，可含链接 |

---

## 🏗️ 系统架构

```
┌──────────────────────────────────────────────┐
│              API Gateway                       │
│   JWT 认证 · 速率限制 · 请求日志 · 签名验证    │
├──────────────────────────────────────────────┤
│              业务服务层                        │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐│
│  │ 商品服务 │ │ 交易引擎 │ │ 订单服务 │ │ 支付服务 ││
│  ├────────┤ ├────────┤ ├────────┤ ├────────┤│
│  │ 钱包服务 │ │ 充值服务 │ │ 广播服务 │ │ 事件总线 ││
│  └────────┘ └────────┘ └────────┘ └────────┘│
├──────────────────────────────────────────────┤
│               数据层                           │
│  PostgreSQL · Redis · NATS JetStream          │
└──────────────────────────────────────────────┘
```

---

## 🚀 快速开始

### 前置条件

- Go 1.22+
- Docker & Docker Compose

### 一键启动

```bash
# 1. 克隆项目
git clone https://github.com/chaoyibot/AIGO.git
cd AIGO

# 2. 启动基础设施（PostgreSQL + Redis + NATS）
make dev

# 3. 执行数据库迁移
make migrate-up

# 4. 下载依赖
go mod tidy

# 5. 启动服务
go run ./cmd/server
```

服务启动在 `http://localhost:8080` 🎉

### 验证安装

```bash
curl http://localhost:8080/health
# → {"code":0,"message":"success","data":{"status":"ok","version":"0.2.0","name":"AIGO"}}
```

---

## 📚 API 接口一览

### 认证

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| POST | `/api/v1/auth/register` | 用户注册 |
| POST | `/api/v1/auth/token` | API Key → JWT 令牌 |
| POST | `/api/v1/auth/api-key` | 生成 API Key |

### 商品

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| GET | `/api/v1/products` | 商品列表（分页） |
| POST | `/api/v1/products` | 创建商品（加密） |
| GET | `/api/v1/products/:id` | 商品详情 |
| PUT | `/api/v1/products/:id` | 更新商品 |
| DELETE | `/api/v1/products/:id` | 删除商品 |

### 交易

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| GET | `/api/v1/listings` | 挂牌列表 |
| POST | `/api/v1/listings` | 创建挂牌 |
| POST | `/api/v1/listings/:id/buy` | 立即购买 |

### 订单

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| GET | `/api/v1/orders` | 订单列表 |
| GET | `/api/v1/orders/:id` | 订单详情 |
| POST | `/api/v1/orders/:id/confirm` | 确认收货 |
| POST | `/api/v1/orders/:id/cancel` | 取消订单 |
| POST | `/api/v1/orders/:id/dispute` | 发起争议 |

### 钱包 & 充值

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| GET | `/api/v1/wallet` | 查询余额 |
| POST | `/api/v1/recharge` | 充值（法币 → 积分） |
| POST | `/api/v1/wallet/convert` | 法币转积分 |
| GET | `/api/v1/wallet/transactions` | 交易流水 |

### 广播

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| POST | `/api/v1/admin/broadcasts/system` | 系统广播 |
| POST | `/api/v1/broadcasts/commercial` | 商业广播（消耗积分） |
| GET | `/api/v1/broadcasts` | 广播列表 |
| GET | `/api/v1/broadcasts/unread` | 未读广播数 |
| PUT | `/api/v1/broadcasts/:id/read` | 标记已读 |

### Webhook & 事件

| 方法 | 路径 | 说明 |
|:----|:-----|:-----|
| POST | `/api/v1/webhooks` | 注册 Webhook |
| GET | `/api/v1/webhooks` | Webhook 列表 |
| DELETE | `/api/v1/webhooks/:id` | 删除 Webhook |
| GET | `/api/v1/events` | SSE 实时事件流 |

---

## 💻 技术栈

| 层级 | 技术 | 说明 |
|:----|:-----|:-----|
| 语言 | **Go 1.22** | 高性能，单二进制部署 |
| Web 框架 | **Gin** | 最快的 Go HTTP 框架之一 |
| 数据库 | **PostgreSQL 16** | ACID 事务，JSONB 支持 |
| 缓存 | **Redis 7** | 速率限制，会话存储 |
| 消息队列 | **NATS (JetStream)** | 持久化事件总线 |
| 认证 | **JWT (HS256) + API Key** | 无状态认证 |
| 加密 | **AES-256-GCM + RSA-4096** | 端到端加密 |
| 容器化 | **Docker Compose** | 一键部署开发环境 |

---

## 🗺️ 完整事件类型

| 事件 | 说明 |
|:----|:-----|
| `order.*` | 订单创建、确认、取消、争议 |
| `broadcast.*` | 系统广播、商业广播 |
| `wallet.*` | 充值、积分兑换、积分消耗 |

AI Agent 可以通过 Webhook 或 SSE 实时接收这些事件。

---

## 🤝 参与贡献

欢迎各种形式的贡献！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交变更 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

---

## 📄 许可证

本项目基于 **MIT License** 开源 — 详见 [LICENSE](LICENSE) 文件。

---

## ⚡ 项目来源

> 本项目由 **AI 全程开发**（Hermes Agent + Claude Code 协同），从设计文档到全部代码实现共约 2 小时。
