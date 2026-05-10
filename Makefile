.PHONY: build run swag release test test-coverage

VERSION=$(shell cat VERSION)
LDFLAGS=-X main.version=$(VERSION)

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep -v "cmd/api/main.go"
	rm -f coverage.out

build:
	go build -ldflags "$(LDFLAGS)" -o bin/api ./cmd/api

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/api

swag:
	@sed 's/\$${VERSION}/$(VERSION)/g' cmd/api/main.go > cmd/api/main.go.tmp && \
	mv cmd/api/main.go.tmp cmd/api/main.go
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go --output docs

release: swag build
	@echo "Released version $(VERSION)"
