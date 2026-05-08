BINARY     := ai-cmd
MODULE     := github.com/wesele/demo2
CMD_PATH   := ./cmd/ai-cmd
BIN_DIR    := ./bin
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -s -w -X main.version=$(VERSION)

# Build targets
TARGETS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64

.PHONY: all build test lint clean release tidy

all: build

## build: compile binary for the current platform into ./bin/
build:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) $(CMD_PATH)
	@echo "Built $(BIN_DIR)/$(BINARY)"

## test: run all unit tests
test:
	go test -v -race ./...

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## tidy: tidy and verify Go modules
tidy:
	go mod tidy
	go mod verify

## clean: remove build artifacts
clean:
	rm -rf $(BIN_DIR)

## release: cross-compile binaries for all target platforms
release:
	@mkdir -p $(BIN_DIR)
	$(foreach TARGET,$(TARGETS),\
		$(eval OS=$(word 1,$(subst /, ,$(TARGET))))\
		$(eval ARCH=$(word 2,$(subst /, ,$(TARGET))))\
		$(eval EXT=$(if $(filter windows,$(OS)),.exe,))\
		CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build \
			-ldflags "$(LDFLAGS)" \
			-o $(BIN_DIR)/$(BINARY)-$(OS)-$(ARCH)$(EXT) \
			$(CMD_PATH) && echo "Built $(BIN_DIR)/$(BINARY)-$(OS)-$(ARCH)$(EXT)";)
