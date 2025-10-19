.PHONY: help proto run-user run-all docker-up docker-down clean test swagger

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $1, $2}' $(MAKEFILE_LIST)

install-tools: ## Install required tools (swag, protoc plugins)
	@echo "📦 Installing tools..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "✅ Tools installed"

swagger: ## Generate Swagger documentation
	@echo "📚 Generating Swagger docs..."
	@swag init -g cmd/user-service/main.go -o docs --parseDependency --parseInternal
	@echo "✅ Swagger docs generated in ./docs"

proto: ## Generate protobuf files
	@echo "Generating protobuf files..."
	@mkdir -p pkg/pb/user pkg/pb/course pkg/pb/enrollment
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/user/user.proto
	@echo "✅ Protobuf files generated"

deps: ## Install dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed"

migrate: ## Run GORM auto-migration (creates/updates tables in NeonDB)
	@echo "🔄 Running database migration..."
	@go run cmd/user-service/main.go
	@echo "✅ Migration complete"

run-user: swagger ## Run user service locally
	@echo "Starting user service..."
	@go run cmd/user-service/main.go

run-course: ## Run course service locally
	@echo "Starting course service..."
	@go run cmd/course-service/main.go

run-enrollment: ## Run enrollment service locally
	@echo "Starting enrollment service..."
	@go run cmd/enrollment-service/main.go

run-gateway: ## Run API gateway locally
	@echo "Starting API gateway..."
	@go run cmd/api-gateway/main.go

docker-up: ## Start all services with Docker Compose
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d
	@echo "✅ Services started"
	@echo "User Service: http://localhost:8081"
	@echo "Swagger Docs: http://localhost:8081/swagger/index.html"

docker-down: ## Stop all Docker services
	@echo "Stopping services..."
	@docker-compose down
	@echo "✅ Services stopped"

docker-logs: ## Show logs from all services
	@docker-compose logs -f

docker-rebuild: ## Rebuild and restart all services
	@echo "Rebuilding services..."
	@docker-compose down
	@docker-compose build --no-cache
	@docker-compose up -d
	@echo "✅ Services rebuilt and restarted"

clean: ## Clean up generated files and caches
	@echo "Cleaning up..."
	@go clean
	@rm -rf pkg/pb/*
	@rm -rf docs/
	@echo "✅ Cleanup complete"

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Tests complete. Coverage report: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "✅ Linting complete"

format: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

dev: swagger run-user ## Development mode: generate swagger and run user service