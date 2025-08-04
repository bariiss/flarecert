.PHONY: build clean install test run-cert run-list run-renew run-export run-export-domain help dev-cert dev-export dev-k8s-test version release

# Build configuration
BINARY_NAME=flarecert
BUILD_DIR=./bin
MAIN_FILE=main.go

# Version information
VERSION ?= 1.1.0
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go build flags
LDFLAGS=-ldflags "-X github.com/bariiss/flarecert/cmd.version=$(VERSION) -X github.com/bariiss/flarecert/cmd.commit=$(COMMIT) -X github.com/bariiss/flarecert/cmd.date=$(DATE)"

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download
	@echo "✅ Dependencies installed"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf certs/
	@rm -rf exports/
	@echo "✅ Clean complete"

# Install binary to $GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to $$GOPATH/bin..."
	@go install $(LDFLAGS) .
	@echo "✅ Installation complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "✅ Tests complete"

# Run certificate generation (example)
run-cert:
	@echo "Running certificate generation example..."
	@go run $(MAIN_FILE) cert --domain example.com --staging --verbose

# Run certificate listing
run-list:
	@echo "Listing certificates..."
	@go run $(MAIN_FILE) list --verbose

# Run certificate renewal
run-renew:
	@echo "Running certificate renewal..."
	@go run $(MAIN_FILE) renew --verbose

# Run certificate export
run-export:
	@echo "Exporting certificates to Kubernetes YAML..."
	@go run $(MAIN_FILE) export --all --output ./exports/

# Run certificate export for specific domain
run-export-domain:
	@echo "Exporting specific domain certificate..."
	@go run $(MAIN_FILE) export --domain example.com --output ./exports/example-com-secret.yaml

# Check for required environment variables
check-env:
	@echo "Checking environment variables..."
	@if [ -z "$$CLOUDFLARE_API_TOKEN" ]; then echo "❌ CLOUDFLARE_API_TOKEN not set"; exit 1; fi
	@if [ -z "$$CLOUDFLARE_EMAIL" ]; then echo "❌ CLOUDFLARE_EMAIL not set"; exit 1; fi
	@if [ -z "$$ACME_EMAIL" ]; then echo "❌ ACME_EMAIL not set"; exit 1; fi
	@echo "✅ Environment variables are set"

# Setup development environment
setup: deps
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then cp .env.example .env; echo "📝 Created .env file from template"; fi
	@echo "✅ Setup complete"
	@echo ""
	@echo "📋 Next steps:"
	@echo "1. Edit .env file with your Cloudflare credentials"
	@echo "2. Run 'make build' to build the binary"
	@echo "3. Run './bin/flarecert cert --domain yourdomain.com --staging' to test"
	@echo "4. Run 'make dev-export' to test Kubernetes YAML export"

# Development helpers
dev-cert:
	@echo "🧪 Generating test certificate (staging)..."
	@go run $(MAIN_FILE) cert --domain test.example.com --staging --verbose

dev-export:
	@echo "🧪 Exporting test certificates..."
	@mkdir -p exports
	@go run $(MAIN_FILE) export --all --output ./exports/
	@echo "✅ Exported certificates to ./exports/"

dev-k8s-test:
	@echo "🧪 Testing Kubernetes YAML generation..."
	@mkdir -p exports
	@go run $(MAIN_FILE) export --domain test.example.com --output ./exports/test-secret.yaml
	@echo "📝 Generated test-secret.yaml"
	@cat ./exports/test-secret.yaml

# Show version information
version:
	@echo "FlareCert version: $(VERSION)"
	@echo "Build commit: $(COMMIT)"
	@echo "Build date: $(DATE)"

# Create a release build
release: clean
	@echo "Creating release build v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ Release build complete: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "📦 Version: $(VERSION)"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run
	@echo "✅ Linting complete"

# Show help
help:
	@echo "FlareCert - Let's Encrypt SSL Certificates with Cloudflare DNS-01"
	@echo ""
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  release    - Create a release build"
	@echo "  version    - Show version information"
	@echo "  deps       - Install dependencies"
	@echo "  clean      - Clean build artifacts and exports"
	@echo "  install    - Install binary to $$GOPATH/bin"
	@echo "  test       - Run tests"
	@echo "  setup      - Setup development environment"
	@echo "  check-env  - Check required environment variables"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code"
	@echo "  help       - Show this help"
	@echo ""
	@echo "Development helpers:"
	@echo "  run-cert    - Run certificate generation example"
	@echo "  run-list    - List certificates"
	@echo "  run-renew   - Run certificate renewal"
	@echo "  run-export  - Export all certificates to Kubernetes YAML"
	@echo "  run-export-domain - Export specific domain certificate"
	@echo "  dev-cert    - Generate test certificate (staging)"
	@echo "  dev-export  - Export test certificates"
	@echo "  dev-k8s-test - Test Kubernetes YAML generation"
