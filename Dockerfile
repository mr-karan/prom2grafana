# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o prom2grafana .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 -S prom2grafana && \
    adduser -u 1000 -S prom2grafana -G prom2grafana

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/prom2grafana .

# Change ownership
RUN chown -R prom2grafana:prom2grafana /app

# Switch to non-root user
USER prom2grafana

# Expose port
EXPOSE 8080

# Set environment variables
ENV PORT=8080

# Run the application
CMD ["./prom2grafana"]