.PHONY: help build install test lint check

help: ## Show this help
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-10s %s\n", $$1, $$2}'

build: ## Build the kl binary in the repository root
	go build -o kl ./cmd/kl

install: ## Install kl into GOPATH/bin
	go install ./cmd/kl

test: ## Run all tests
	go test ./...

lint: ## Run gofmt, go vet and staticcheck
	test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

check: lint test ## Run lint and tests, same as CI
