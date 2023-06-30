.DEFAULT_GOAL:=help

BIN=provdoc
DEMODIR=demo

.PHONY: build
build: clean ## Build binaries
	@go build -o $(BIN)

.PHONY: clean
clean: ## Clean up binaries
	@rm -f $(BIN)

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}'

.PHONY: lint
lint: ## Lint source code
	@golangci-lint run ./...

.PHONY: record-demo
record-demo: ## Record a demo
	@go build -o $(BIN)
	@mv $(BIN) $(DEMODIR)/
	@cd $(DEMODIR) && \
		vhs demo.tape && \
		rm $(BIN)

.PHONY: regen-schema
regen-schema: ## Re-generate provider schema data
	@cd $(DEMODIR) && \
		terraform init && \
		terraform providers schema -json > schema.json

.PHONY: test
test: ## Run unit tests
	@go test -v -coverprofile=coverage.txt ./...

