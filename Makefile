.PHONY: lint vet test build clean setup vulncheck all release release-dry changelog changelog-check goldens-regen goldens-verify

GO       := go
LINT     := golangci-lint
MODULE   := gitmap
BINARY   := gitmap
VERSION  ?= dev
LDFLAGS  := -s -w -X 'github.com/alimtvnetwork/gitmap-v10/gitmap/constants.Version=$(VERSION)'

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

## Changelog — regenerate CHANGELOG.md and src/data/changelog.ts from
## Conventional Commits since the latest annotated git tag.
## Usage:
##   make changelog VERSION=v3.92.0
##   make changelog VERSION=v3.92.0 SINCE=v3.90.0          # partial backfill
##   make changelog RELEASE_TAG=v3.91.0 SINCE=v3.90.0      # rebuild a past release
SINCE       ?=
RELEASE_TAG ?=
changelog:
	@cd scripts/changelog && $(GO) run . -mode=write -version=$(VERSION) -repo=../.. -since=$(SINCE) -release-tag=$(RELEASE_TAG)

## Changelog-check — fail (exit 3) when the on-disk changelogs drift
## from the regenerated output. Wire into CI. Forwards SINCE / RELEASE_TAG
## so partial-update PRs can verify only their slice.
changelog-check:
	@cd scripts/changelog && $(GO) run . -mode=check -version=$(VERSION) -repo=../.. -since=$(SINCE) -release-tag=$(RELEASE_TAG)

## Goldens-regen — regenerate golden fixtures for a specific test pattern.
## REQUIRES RUN=<pattern>. Delegates to `gitmap regoldens`, which is the
## ONLY sanctioned entry point that may unlock the golden-update gate
## (see spec/05-coding-guidelines/21-golden-fixture-regeneration.md §6).
## PKG defaults to ./... but should be narrowed for speed.
## Usage:
##   make goldens-regen RUN=TestStartupListJSONContract
##   make goldens-regen RUN=FindNextJSONContract PKG=./cmd/...
PKG ?= ./...
goldens-regen:
	@if [ -z "$(RUN)" ]; then \
		echo "ERROR: RUN=<test pattern> is required (e.g. make goldens-regen RUN=TestFooContract)"; \
		exit 2; \
	fi
	@echo "▸ Regenerating goldens via gitmap regoldens: pattern=$(RUN) pkg=$(PKG)"
	@cd $(MODULE) && $(GO) run . regoldens --run '$(RUN)' --pkg '$(PKG)'

## Goldens-verify — re-run the same pattern WITHOUT the gating env vars to
## confirm regenerated fixtures pass cleanly. This is the mandatory second
## pass: if it fails, the writers are non-deterministic or drift remains.
## Usage:
##   make goldens-verify RUN=TestStartupListJSONContract
goldens-verify:
	@if [ -z "$(RUN)" ]; then \
		echo "ERROR: RUN=<test pattern> is required (e.g. make goldens-verify RUN=TestFooContract)"; \
		exit 2; \
	fi
	@echo "▸ Verifying goldens (no env gates): pattern=$(RUN) pkg=$(PKG)"
	@cd $(MODULE) && unset GITMAP_UPDATE_GOLDEN GITMAP_ALLOW_GOLDEN_UPDATE && \
		$(GO) test $(PKG) -run '$(RUN)' -count=1 -v
