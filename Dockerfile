# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build without CGO (pgx is pure Go)
RUN CGO_ENABLED=0 GOOS=linux go build -o nyusu .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/nyusu .

# Copy static assets and templates
COPY --from=builder /build/html ./html
COPY --from=builder /build/static ./static
COPY --from=builder /build/sql ./sql

# Expose application port
EXPOSE 8888

# Set default environment variables
ENV PORT=8888
ENV ENVIRONMENT=production
ENV SCRAPPER_TICK=300

# Run the application
CMD ["./nyusu"]
