# The goals the DUDES flow expects (DUDES, CORE.md §13): the flow runs a goal, never the raw command.
# `make help` lists them. Give every goal you add a `## what it does` comment on its line.

# Set once: the command that runs ALL your tests and exits non-zero when one fails.
# For example: go test ./...   npm test   python -m pytest
TEST = echo "make test: set TEST in the Makefile to your test command" >&2; exit 1

.PHONY: help test

help: ## List the goals
	@awk 'BEGIN{FS=":.*## "} /^[a-zA-Z0-9_-]+:.*## /{printf "  %-10s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run all tests; fails when one fails
	@$(TEST)
