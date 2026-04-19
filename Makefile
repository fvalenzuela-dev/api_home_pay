.PHONY: build run swag release

VERSION=$(shell cat VERSION)
LDFLAGS=-X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/api ./cmd/api

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/api

swag:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go --output docs

release: swag build
	@echo "Released version $(VERSION)"
