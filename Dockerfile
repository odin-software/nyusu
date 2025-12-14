# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies for CGO and SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled for SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o nyusu .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# Create data directory for SQLite database
RUN mkdir -p /data

# Copy binary from builder
COPY --from=builder /build/nyusu .

# Copy static assets and templates
COPY --from=builder /build/html ./html
COPY --from=builder /build/static ./static
COPY --from=builder /build/sql ./sql

# Expose application port
EXPOSE 8888

# Set default environment variables
ENV DB_URL=/data/nyusu.db
ENV PORT=8888
ENV ENVIRONMENT=production
ENV SCRAPPER_TICK=60

# Run the application
CMD ["./nyusu"]
