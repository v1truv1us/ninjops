# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ninjops ./cmd/ninjops

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -g 1000 -S ninjops && \
    adduser -u 1000 -S ninjops -G ninjops

WORKDIR /app

COPY --from=builder /app/ninjops /usr/local/bin/ninjops

RUN chmod +x /usr/local/bin/ninjops

USER ninjops

ENV NINJOPS_SERVE_LISTEN=0.0.0.0
ENV NINJOPS_SERVE_PORT=8080

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["ninjops", "serve"]
