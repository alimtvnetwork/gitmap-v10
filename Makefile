.PHONY: lint vet test build clean setup vulncheck all release release-dry changelog changelog-check

GO       := go
LINT     := golangci-lint
MODULE   := gitmap
BINARY   := gitmap
VERSION  ?= dev
LDFLAGS  := -s -w -X 'github.com/alimtvnetwork/gitmap-v7/gitmap/constants.Version=$(VERSION)'

all: lint test build

## Setup — install tools and git hooks
setup:
	@./setup.sh

## Lint — run golangci-lint
lint:
	@cd $(MODULE) && $(LINT) run ./... --timeout=5m

## Vet — run go vet
vet:
	@cd $(MODULE) && $(GO) vet ./...

## Test — run all tests
test:
	@cd $(MODULE) && $(GO) test ./... -v -count=1

## Build — compile for the current platform
build:
	@cd $(MODULE) && CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o ../$(BINARY) .
	@echo "Built $(BINARY) ($(VERSION))"

## Vulncheck — scan for known vulnerabilities
vulncheck:
	@cd $(MODULE) && $(GO) run golang.org/x/vuln/cmd/govulncheck@latest ./...

## Release — run full release workflow (usage: make release BUMP=patch)
BUMP ?= patch
release: lint test
	@cd $(MODULE) && $(GO) run . release --bump $(BUMP)

## Release dry-run — preview release without executing
release-dry:
	@cd $(MODULE) && $(GO) run . release --bump $(BUMP) --dry-run

## Clean — remove build artifacts
clean:
	@rm -f $(BINARY)
	@rm -rf $(MODULE)/.gitmap/release-assets
	@echo "Cleaned."
