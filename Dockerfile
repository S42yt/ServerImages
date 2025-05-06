# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o serverimages .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/serverimages .

RUN mkdir -p /app/uploads && \
    chown -R appuser:appgroup /app/uploads

COPY .env /app/.env

USER appuser

EXPOSE 4200

VOLUME ["/app/uploads"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:4200/images || exit 1

ENTRYPOINT ["./serverimages"]