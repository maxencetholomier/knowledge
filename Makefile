.PHONY: build install test lint check

build:
	go build -o kl ./cmd/kl

install:
	go install ./cmd/kl

test:
	go test ./...

lint:
	test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

check: lint test
