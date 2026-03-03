.PHONY: build run test watch clean setup help dev

## Setup & Installation
go-install-air: ## Installs the air build reload system using 'go install'
	@go install github.com/air-verse/air@latest

go-install-templ: ## Installs the templ Templating system for Go
	@go get -tool github.com/a-h/templ/cmd/templ@latest

setup: ## Install all dev dependencies
	@go mod download
	@go install github.com/air-verse/air@latest
	@go get -tool github.com/a-h/templ/cmd/templ@latest

## Build & Run
templ-generate: ## Generate templ templates
	@go tool templ generate

build: ## Build the application
	make tailwind-build
	make templ-generate
	@go build -o ./bin/main ./cmd/main.go

run: build ## Build and run the application
	@./bin/main

watch: ## Run with hot reload (air)
	@go build -o ./tmp/main ./cmd/main.go && $(shell go env GOPATH)/bin/air

## Testing
test: ## Run all tests
	@go test -v ./...

test-unit: ## Run unit tests only (skip integration)
	@go test -short -v ./...

test-integration: ## Run integration tests only
	@go test -v ./tests/integration/...

test-coverage: ## Run tests with coverage report
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## Database
migrate-new: ## Create a new migration
	@go run ./cmd/db/main.go

migrate-up: ## Run all pending migrations
	@go run ./cmd/db/main.go up

migrate-down: ## Rollback the last migration
	@go run ./cmd/db/main.go down

## Docker
docker-build: ## Build Docker image
	@docker build -t dviji-se .

docker-run: ## Run Docker container
	@docker run -p 8080:8080 --env-file local.env dviji-se

## Code Quality
fmt: ## Format Go code
	@go fmt ./...

lint: ## Run linter (requires golangci-lint)
	@golangci-lint run

## Tailwind CSS
get-install-tailwindcss: ## Installs the tailwindcss cli
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
	chmod +x tailwindcss-linux-x64
	mv tailwindcss-linux-x64 tailwindcss

tailwind-watch: ## Compile tailwindcss and watch for changes
	@npx @tailwindcss/cli -i ./web/static/css/main.css -o ./web/static/css/styles.css --watch

tailwind-build: ## One-time compile tailwindcss styles
	@npx @tailwindcss/cli -i ./web/static/css/main.css -o ./web/static/css/styles.css --minify

## Development
dev: setup ## Install deps and run server + tailwind with hot reload
	@make tailwind-watch & make watch

## Cleanup
clean: ## Remove build artifacts
	@rm -rf ./bin ./tmp coverage.out coverage.html

## Help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
