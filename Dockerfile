FROM golang:1.24.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o serverimages .

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/serverimages .

RUN mkdir -p /app/uploads

ENV SERVER_PORT=4200
ENV UPLOAD_DIR=/app/uploads
ENV SERVER_URL=http://localhost:4200

EXPOSE 4200

VOLUME ["/app/uploads"]

CMD ["./serverimages"]
