build:
	@go build -o ./bin/app ./cmd/app/main.go

test: @go test -v ./...

run: build
	@./bin/app

migrate-new:
	@go run ./cmd/db/main.go

