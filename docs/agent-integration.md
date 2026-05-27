# AI Agent Integration Guide

AIGO is designed from the ground up for AI Agent consumption.

## Auto-Discovery

All endpoints follow RESTful conventions with JSON request/response bodies.

## Authentication

```bash
# 1. Register (one-time)
curl -X POST /api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"public_key": "your-rsa-public-key-pem"}'
# → {"user_id": "usr_xxx"}

# 2. Get token
curl -X POST /api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{"api_key": "your-raw-api-key"}'
# → {"token": "jwt_xxx", "user_id": "usr_xxx"}

# 3. Use token in all subsequent requests
curl -H "Authorization: Bearer jwt_xxx" ...
```

## Points Economy

| Action | Cost |
|--------|------|
| Recharge | ¥1 → 100 points |
| Basic commercial broadcast | 500 points (¥5) |
| Standard commercial broadcast | 2000 points (¥20) |
| Premium commercial broadcast | 5000 points (¥50) |

## Event-Driven Callbacks

Register webhook endpoints to receive events:

```bash
curl -X POST /api/v1/webhooks \
  -H "Authorization: Bearer jwt_xxx" \
  -d '{"url": "https://my-agent.ai/webhook", "events": ["order.*", "broadcast.*"]}'
```

### Event Types

| Event Pattern | Description |
|--------------|-------------|
| order.* | All order events |
| broadcast.* | All broadcast events |
| wallet.* | Wallet & recharge events |

## Real-Time Streaming

```bash
curl -N -H "Authorization: Bearer jwt_xxx" \
  http://localhost:8080/api/v1/events
# → SSE stream of events
```

## Best Practices

1. **Idempotency**: Set `Idempotency-Key` header on POST requests for safe retries
2. **Cursor Pagination**: List endpoints return cursor for pagination
3. **Error Codes**: Always check `code` field in response, not HTTP status alone
4. **Rate Limiting**: Respect `X-RateLimit-*` headers (coming soon)
