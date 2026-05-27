# AIGO 🔒

Headless encrypted commodity trading platform — AI Agent friendly, points-powered economy.

## Quick Start

```bash
# 1. Start dependencies
make dev

# 2. Run database migration
make migrate-up

# 3. Start server
make run
```

## Architecture

```
┌──────────────────────────────────────┐
│              API Gateway              │
│     JWT Auth · Rate Limit · Logging   │
├──────────────────────────────────────┤
│            Service Layer              │
│  Auth · Product · Trading · Order    │
│  Wallet · Recharge · Broadcast       │
├──────────────────────────────────────┤
│        PostgreSQL + Redis + NATS      │
└──────────────────────────────────────┘
```

## Core Modules

- **Points Economy**: ¥1 = 100 points. Recharge fiat → get points. Spend points on products.
- **Encrypted Products**: End-to-end AES-256-GCM encryption for product data.
- **Broadcast System**: System announcements (free) + Commercial ads (costs points).
- **AI Agent Ready**: RESTful API, Webhook events, SSE real-time stream.

## API Overview

| Module | Endpoints |
|--------|-----------|
| Auth | POST /register, POST /token, POST /api-key |
| Products | GET/POST/PUT/DELETE /products |
| Listings | GET/POST /listings, POST /listings/:id/buy |
| Orders | GET /orders, POST /confirm, POST /cancel, POST /dispute |
| Wallet | GET /wallet, POST /recharge, POST /convert |
| Broadcasts | GET /broadcasts, POST /commercial, POST /system |
| Webhooks | POST/GET/PUT/DELETE /webhooks |
| Events | GET /events (SSE) |

## Tech Stack

- **Language**: Go 1.22
- **Framework**: Gin
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Messaging**: NATS (JetStream)
- **Auth**: JWT (HS256) + API Key
- **Encryption**: AES-256-GCM, RSA-4096

## License

MIT
