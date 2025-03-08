go-install-air: ## Installs the air build reload system using 'go install'
	@go install github.com/air-verse/air@latest

go-install-templ: ## Installs the templ Templating system for Go
	@go install github.com/a-h/templ/cmd/templ@latest

build:
	@go build -o ./bin/app ./cmd/app/main.go

test: 
	@go test -v ./...

run: build
	@./bin/app

migrate-new:
	@go run ./cmd/db/main.go

watch:
	@go build -o ./tmp/app ./cmd/app/main.go && air
