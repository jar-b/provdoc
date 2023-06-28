.DEFAULT_GOAL:=help

.PHONY: build
build: clean ## Build binaries
	@go build -o provdoc

.PHONY: clean
clean: ## Clean up binaries
	@rm -f provdoc

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}'

.PHONY: regen-schema
regen-schema: ## Re-generate provider schema data
	@cd example-config && \
		terraform init && \
		terraform providers schema -json > schema.json

.PHONY: test
test: ## Run unit tests
	@go test -v -coverprofile=coverage.txt ./...