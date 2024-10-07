GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
BINARY_NAME := helm-valgrade
VERSION := $(shell awk '/version:/ {print $$2}' plugin.yaml | tr -d '"')

.PHONY: all build test clean run deps build-all lint install-lint install

all: test build

build:
	CGO_ENABLED=0 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME) -v ./cmd/valgrade

build-all:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME)-darwin-amd64 -v ./cmd/valgrade
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME)-darwin-arm64 -v ./cmd/valgrade
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME)-linux-amd64 -v ./cmd/valgrade
	GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME)-linux-arm64 -v ./cmd/valgrade
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME)-windows-amd64.exe -v ./cmd/valgrade

release: build-all
	mkdir -p release
	tar czf release/$(BINARY_NAME)-darwin-amd64-$(VERSION).tar.gz $(BINARY_NAME)-darwin-amd64 plugin.yaml
	tar czf release/$(BINARY_NAME)-darwin-arm64-$(VERSION).tar.gz $(BINARY_NAME)-darwin-arm64 plugin.yaml
	tar czf release/$(BINARY_NAME)-linux-amd64-$(VERSION).tar.gz $(BINARY_NAME)-linux-amd64 plugin.yaml
	tar czf release/$(BINARY_NAME)-linux-arm64-$(VERSION).tar.gz $(BINARY_NAME)-linux-arm64 plugin.yaml
	zip -q release/$(BINARY_NAME)-windows-amd64-$(VERSION).zip $(BINARY_NAME)-windows-amd64.exe plugin.yaml

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

run: build
	./$(BINARY_NAME)

deps:
	$(GOGET) ./...
	$(GOMOD) tidy

lint:
	golangci-lint run --timeout 5m

install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2

install: build
	mkdir -p $(HOME)/.helm/plugins/helm-valgrade
	cp $(BINARY_NAME) $(HOME)/.helm/plugins/helm-valgrade/
	cp plugin.yaml $(HOME)/.helm/plugins/helm-valgrade/

