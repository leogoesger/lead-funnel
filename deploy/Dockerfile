# --------------------------------------------------------------------
# Build + install Kronk
# --------------------------------------------------------------------
FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/ardanlabs/kronk/cmd/kronk@latest

RUN kronk libs --install \
    --local \
    --os linux \
    --arch arm64 \
    --processor cpu

COPY . .

RUN go build -o api ./cmd/api

# --------------------------------------------------------------------
# Runtime
# --------------------------------------------------------------------
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y \
        ca-certificates \
        libstdc++6 \
        libgomp1 \
        libgcc-s1 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Your application
COPY --from=builder /app/api .

# Application assets
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/static ./static

# Kronk runtime
COPY --from=builder /root/.kronk /root/.kronk

# Kronk executable (optional, if you want it available)
COPY --from=builder /go/bin/kronk /usr/local/bin/kronk

EXPOSE 8080

CMD ["./api"]