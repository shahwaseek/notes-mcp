# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary (pure Go SQLite, no CGO needed)
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o notes-mcp ./cmd/notes-mcp

# Final stage
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create directory for SQLite storage
RUN mkdir -p /data

COPY --from=builder /app/notes-mcp /usr/local/bin/notes-mcp

# Expose data directory as volume
VOLUME ["/data"]

# Default database location inside container
ENTRYPOINT ["/usr/local/bin/notes-mcp"]
CMD ["--db", "/data/notes.db"]
