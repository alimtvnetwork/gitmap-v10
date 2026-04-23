.PHONY: build run test update no-pull no-deploy clean help generate

SHELL := /bin/bash

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

# `generate` runs every //go:generate directive in the module. Today the
# only generator is completion/internal/gencommands (writes
# completion/allcommands_generated.go), but new generators added later
# will be picked up automatically. Made a prerequisite of build/run/test
# so the generated file can never silently drift from the Cmd*
# constants — exactly the failure class the CI generate-check guards
# against. Using a phony target instead of a file target keeps it
# dependency-free and always runs (cheap: ~100ms locally).
generate: ## Regenerate completion + any other go:generate outputs
	@go generate ./...

build: generate ## Build and deploy (full pipeline)
	@bash ../run.sh

run: generate ## Build, deploy, and run with optional ARGS (make run ARGS="list --json")
	@bash ../run.sh -r $(ARGS)

lint: ## Check for duplicate type declarations across package files
	@bash scripts/check-duplicate-types.sh .

test: generate ## Run tests with report
	@bash ../run.sh -t

update: ## Pull, build, deploy, and sync PATH binary
	@bash ../run.sh --update

no-pull: generate ## Build without git pull
	@bash ../run.sh --no-pull

no-deploy: generate ## Build without deploying
	@bash ../run.sh --no-deploy

clean: ## Remove build artifacts
	@rm -rf bin/
	@echo "Build artifacts removed."
