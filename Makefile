GO_IMAGE = golang:1.22-alpine
GO_RUN = docker run --rm -v "$(PWD):/app" -w /app $(GO_IMAGE) go
DB_URL = postgres://aigo:aigo@host.docker.internal:5432/aigo?sslmode=disable

.PHONY: build run test lint dev-db dev down clean migrate-up deps

build:
	$(GO_RUN) build -o /app/bin/aigo /app/cmd/server

run:
	$(GO_RUN) run /app/cmd/server

test:
	$(GO_RUN) test ./... -v -count=1

lint:
	docker run --rm -v "$(PWD):/app" -w /app golangci/golangci-lint:v1.60 golangci-lint run ./...

# Start dev dependencies
dev-db:
	docker run -d --name aigo-postgres -e POSTGRES_USER=aigo -e POSTGRES_PASSWORD=aigo -e POSTGRES_DB=aigo -p 5432:5432 postgres:16-alpine

dev:
	docker compose -f deploy/docker-compose.yml up -d

down:
	docker compose -f deploy/docker-compose.yml down

docker-build:
	docker build -f deploy/Dockerfile -t aigo:latest .

docker-run:
	docker run --rm -p 8080:8080 --network aigo_default \
		-e DATABASE_URL=postgres://aigo:aigo@postgres:5432/aigo?sslmode=disable \
		aigo:latest

clean:
	docker compose -f deploy/docker-compose.yml down -v
	rm -rf bin/

# Rebuild the AIGO binary inside the Docker image
docker-build:
	docker build -f deploy/Dockerfile -t aigo:latest .

# Run AIGO with the crypto key (fill in the hex key below before running)
docker-run:
	docker run --rm -p 8080:8080 --network aigo_default \
		-e DATABASE_URL=postgres://aigo:***@postgres:5432/aigo?sslmode=disable \
		-e SYSTEM_CRYPTO_KEY=c03c2fc76363d5296d1f404b41acc3c180c4b45dc62c112c31ea2e5eb5b72eb0 \
		aigo:latest

migrate-up:
	$(GO_RUN) run /app/cmd/migrate

deps:
	$(GO_RUN) mod tidy
