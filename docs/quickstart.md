# Quick Start Guide

## Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Git

## 1. Clone and Setup

```bash
git clone https://github.com/aigo/aigo
cd aigo
```

## 2. Start Infrastructure

```bash
make dev
```

This starts PostgreSQL, Redis, and NATS.

## 3. Run Migrations

```bash
make migrate-up
```

## 4. Start Server

```bash
make run
```

Server starts on http://localhost:8080.

## 5. Test the API

```bash
# Health check
curl http://localhost:8080/health

# Register a user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"public_key": "test-public-key"}'

# Get API key (from JWT token first)
# See agent-integration.md for full flow
```

## 6. Run Tests

```bash
make test
```

## Common Commands

| Command | Description |
|---------|-------------|
| `make build` | Build binary |
| `make run` | Start server |
| `make test` | Run tests |
| `make dev` | Start infrastructure |
| `make down` | Stop infrastructure |
| `make migrate-up` | Run migrations |
| `make clean` | Clean all data |
