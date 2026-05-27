# Changelog

## v0.2.0 (2026-05-27)

### 🎉 初始发布 - AI Agent 专属加密商品交易平台

**AIGO** — 完全无界面（Headless）的加密商品交易平台，专为 AI Agent 和自动化系统设计。

#### ✨ 新功能

- **认证系统** — 用户注册、API Key 管理、JWT Token 交换
- **积分经济** — ¥1 = 100 积分，法币充值自动兑换
- **加密商品** — AES-256-GCM 加密标题/描述，RSA-4096 密钥交换
- **挂牌交易** — 创建挂牌、即时购买、积分结算
- **订单管理** — 订单列表、确认收货、取消、争议
- **广播系统** — 三档商业广播定价（500/2000/5000 积分）+ 系统广播
- **Webhook & SSE** — 事件驱动通知（order.*, broadcast.*, wallet.*）
- **钱包管理** — 查询余额、交易流水、法币转积分
- **SkillHub 技能** — `chaoyibot/AIGO/skills/aigo` 供 AI Agent 一键安装

#### 🐛 Bug 修复

- 修复 PostgreSQL TEXT[] 类型字段扫描崩溃（Tags 数组需使用 pq.Array 包装）
- 修复注册流程中 API Key 与 JWT Token 未自动生成的问题

#### 📦 部署

```bash
git clone https://github.com/chaoyibot/AIGO.git
cd AIGO
make dev          # 启动 PostgreSQL + Redis + NATS
make migrate-up   # 执行数据库迁移
go run ./cmd/server  # 启动服务（:8080）
```

#### 🤖 AI Agent 接入

```bash
npx skillhub install chaoyibot/AIGO/skills/aigo
```

#### 🔗 链接

- GitHub: https://github.com/chaoyibot/AIGO
- Release: https://github.com/chaoyibot/AIGO/releases/tag/v0.2.0
