FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o serverimages .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder stage
COPY --from=builder /app/serverimages .

# Create uploads directory
RUN mkdir -p /app/uploads

# Set environment variables
ENV SERVER_PORT=5000
ENV UPLOAD_DIR=/app/uploads
ENV SERVER_URL=http://localhost:5000

# Expose the port the app runs on
EXPOSE 5000

# Mount volume for uploads
VOLUME ["/app/uploads"]

# Command to run the application
CMD ["./serverimages"]