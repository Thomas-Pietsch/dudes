# The goals the DUDES flow expects (DUDES, CORE.md §13): the flow runs a goal, never the raw command.
# `make help` lists them. Give every goal you add a `## what it does` comment on its line.

# Set once, for your stack. Each command exits non-zero when it finds a problem.
# TEST: runs ALL your tests.      For example: go test ./...   npm test   python -m pytest
# FMT:  fails on unformatted code, changes nothing.   For example: test -z "$$(gofmt -l .)"   npx prettier --check .   ruff format --check
# LINT: static checks (vet, lint, types).   For example: go vet ./...   npx eslint .   ruff check
# A tool you do not use: set its command to `true`, on purpose.
TEST = echo "make test: set TEST in the Makefile to your test command" >&2; exit 1
FMT  = echo "make check: set FMT in the Makefile to your format check" >&2; exit 1
LINT = echo "make check: set LINT in the Makefile to your static checks" >&2; exit 1

.PHONY: help test check

help: ## List the goals
	@awk 'BEGIN{FS=":.*## "} /^[a-zA-Z0-9_-]+:.*## /{printf "  %-10s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run all tests; fails when one fails
	@$(TEST)

check: ## The gate: format check, static checks, all tests; fails on the first problem
	@$(FMT)
	@$(LINT)
	@$(TEST)
