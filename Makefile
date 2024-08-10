GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
BINARY_NAME := helm-valgrade
BINARY_UNIX := $(BINARY_NAME)_unix

.PHONY: all build test clean run deps build-linux lint install-lint install

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/valgrade

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

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/valgrade

lint:
	golangci-lint run

install-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.50.1

install: build
	mkdir -p $(HOME)/.helm/plugins/helm-valgrade
	cp $(BINARY_NAME) $(HOME)/.helm/plugins/helm-valgrade/
	cp plugin.yaml $(HOME)/.helm/plugins/helm-valgrade/
