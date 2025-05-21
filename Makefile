run: ## Run the application
	go run godevmcp/main.go serve

build: ## Build the application
	go build -o output/godevmcp godevmcp/main.go

install: ## Install the application
	go install ./godevmcp
	@echo "Installed to $(shell go env GOPATH)/bin/godevmcp"

test: ## Run unit tests
	go test -v ./...

fmt: ## Run format
	gofumpt -extra -w .

lint: ## Run lint
	golangci-lint run

fix: ## Run fix
	golangci-lint run --fix

inspect: ## Run in MCP inspector
	npx @modelcontextprotocol/inspector go run ./godevmcp/main.go serve

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "%-20s %s\n", $$1, $$2}'
