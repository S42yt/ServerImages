FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o serverimages .

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/serverimages .
COPY --from=builder /app/.env.template /app/.env

RUN mkdir -p /app/uploads

ENV PORT=4200
ENV UPLOAD_DIR=/app/uploads
ENV SERVER_URL=http://0.0.0.0:4200
ENV MAX_UPLOAD_SIZE=524288000
ENV CACHE_MAX_AGE=86400
ENV ALLOWED_MIME_TYPES=image/

EXPOSE 4200

VOLUME ["/app/uploads"]

CMD ["./serverimages"]