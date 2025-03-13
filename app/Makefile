go-install-air: ## Installs the air build reload system using 'go install'
	@go install github.com/air-verse/air@latest

go-install-templ: ## Installs the templ Templating system for Go
	@go install github.com/a-h/templ/cmd/templ@latest

templ-generate:
	templ generate

build:
	make tailwind-build
	make templ generate
	@go build -o ./bin/main ./cmd/main.go

test: 
	@go test -v ./...

run: build
	@./bin/main

migrate-new:
	@go run ./cmd/db/main.go

watch:
	@go build -o ./tmp/main ./cmd/main.go && air

## tailwind css
get-install-tailwindcss: ## Installs the tailwindcss cli
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
	chmod +x tailwindcss-linux-x64
	mv tailwindcss-linux-x64 tailwindcss

tailwind-watch: ## compile tailwindcss and watch for changes
	./tailwindcss -i ./web/static/css/main.css -o ./web/static/css/styles.css --watch

tailwind-build: ## one-time compile tailwindcss styles
	./tailwindcss -i ./web/static/css/main.css -o ./web/static/css/styles.css --minify
