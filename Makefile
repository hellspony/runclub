.PHONY: build run test lint clean docker-build generate

APP_NAME := runclub
BUILD_DIR := ./bin

# Build the binary
build:
	CGO_ENABLED=1 go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

# Run the server locally
run: build
	./$(BUILD_DIR)/$(APP_NAME)

# Run tests with race detector
test:
	go test -race -coverprofile=coverage.out ./...

# Run linter
lint:
	golangci-lint run ./...

# Show test coverage
coverage: test
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

# Build Docker image
docker-build:
	docker build -t $(APP_NAME) .

# Run with docker-compose
docker-up:
	docker compose up -d

# Stop docker-compose
docker-down:
	docker compose down

# Install frontend dependencies and build
frontend-build:
	cd web/admin && npm ci && npm run build

# Run frontend dev server
frontend-dev:
	cd web/admin && npm run dev

# Generate mocks
generate:
	go generate ./internal/domain/repository/...

# Generate Go dependencies
deps:
	go mod tidy

# Run migrations manually
migrate:
	go run ./cmd/server
