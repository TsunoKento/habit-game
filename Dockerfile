# syntax=docker/dockerfile:1

# ── Stage 1: Build ──────────────────────────────────────
FROM golang:1.22 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux \
    go build -ldflags="-s -w" -o /out/app ./cmd/app/

# ── Stage 2: Runtime ────────────────────────────────────
FROM debian:bookworm-slim AS runtime

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        libsqlite3-0 \
    && rm -rf /var/lib/apt/lists/*

RUN useradd -u 1000 -m appuser \
    && mkdir -p /data \
    && chown appuser:appuser /data

COPY --from=builder /out/app /usr/local/bin/app

USER appuser

VOLUME ["/data"]

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/app"]
