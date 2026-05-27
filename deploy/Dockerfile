# Use this Dockerfile to build and run AIGO in production
# Build: docker build -t aigo .
# Run:   docker run -p 8080:8080 -e DATABASE_URL=... aigo

FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /aigo ./cmd/server
RUN CGO_ENABLED=0 go build -o /migrate ./cmd/migrate

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /aigo /usr/local/bin/aigo
COPY --from=builder /migrate /usr/local/bin/migrate
COPY migrations/ /migrations/
EXPOSE 8080
CMD ["aigo"]
