# Build and run commands for prom2grafana

# Default target: show available commands
default:
    @just --list

# Run the application
run:
    go run main.go

# Build the application
build:
    go build -o prom2grafana main.go

# Build for multiple platforms
build-all:
    GOOS=linux GOARCH=amd64 go build -o dist/prom2grafana-linux-amd64 main.go
    GOOS=darwin GOARCH=amd64 go build -o dist/prom2grafana-darwin-amd64 main.go
    GOOS=darwin GOARCH=arm64 go build -o dist/prom2grafana-darwin-arm64 main.go
    GOOS=windows GOARCH=amd64 go build -o dist/prom2grafana-windows-amd64.exe main.go

# Install dependencies
deps:
    go mod download
    go mod tidy

# Run with example configuration
run-example:
    OPENAI_API_KEY="your-key-here" OPENAI_MODEL="google/gemini-2.0-flash" go run main.go

# Clean build artifacts
clean:
    rm -f prom2grafana
    rm -rf dist/

# Format code
fmt:
    go fmt ./...

# Run linter (requires golangci-lint)
lint:
    go tool golangci-lint run

# Run tests
test:
    go test ./...

# Development mode with air (requires air)
dev:
    air

# Setup development environment
setup:
    @echo "Please set OPENAI_API_KEY environment variable"
    go mod download

# Check if environment is properly configured
check:
    @test -n "${OPENAI_API_KEY}" || (echo "Error: OPENAI_API_KEY environment variable not set" && exit 1)
    @echo "Environment configured"

# Run goreleaser
release:
    goreleaser release --snapshot --clean

# Run goreleaser in release mode
release-prod:
    goreleaser release --clean