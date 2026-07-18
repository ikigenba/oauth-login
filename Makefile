BINARY  := oauth-login
BIN_DIR := bin
PREFIX  ?= $(HOME)/.local

.PHONY: build fmt test check install clean

build: $(BIN_DIR)/$(BINARY)

$(BIN_DIR)/$(BINARY): go.mod $(wildcard go.sum) $(shell find cmd internal -name '*.go')
	install -d $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) ./cmd/oauth-login

fmt:
	go fmt ./...

test:
	go test ./...

check:
	go build ./...
	go vet ./...
	go test ./...
	@test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }

install: build
	install -d $(PREFIX)/bin
	install -m 0755 $(BIN_DIR)/$(BINARY) $(PREFIX)/bin/$(BINARY)

clean:
	rm -rf $(BIN_DIR)
	go clean
