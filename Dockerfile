FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o serverimages .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata wget shadow su-exec

RUN addgroup -g 1000 appgroup && adduser -u 1000 -G appgroup -h /app -D appuser

WORKDIR /app

COPY --from=builder /app/serverimages .
COPY .env /app/.env


EXPOSE 4200

COPY <<EOF /app/entrypoint.sh
#!/bin/sh
set -e

# Ensure uploads directory exists with proper permissions
mkdir -p /app/uploads
chown -R appuser:appgroup /app
chown -R appuser:appgroup /app/uploads
chmod -R 777 /app/uploads

# Print directory permissions for debugging
echo "Directory permissions:"
ls -la /app/uploads

# Run the application as appuser
exec su-exec appuser:appgroup /app/serverimages
EOF

RUN chmod +x /app/entrypoint.sh

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:4200/images || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]