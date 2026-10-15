# DUDES engine — the framework's goals. `make help` lists them: every repeatable action is a goal
# (CORE §13) — humans, CI and agents run the goal, never the raw command. Arguments are make
# variables, e.g. `make audit ARGS=--since=2026-01-01`, `make test-ids ID='BK-?[0-9]{3}' DIR=../app`.

EXE :=
ifeq ($(OS),Windows_NT)
EXE := .exe
endif

BIN := bin/dudes$(EXE)
NOW  = sed -n '/^\#\# Now/,/^\#\# Next/{/^- /p;}' TODO.md
ARGS ?=

.PHONY: help setup tools config hooks build test vet ci version doctor bench clean \
        conventions now todo-check roadmap-check public-check trend-check test-ids audit license-scan vuln-scan worker-image \
        shell bench-models linux

help: ## List the goals
	@awk 'BEGIN{FS=":.*## "} /^##@ /{printf "\n%s\n", substr($$0, 5)} /^[a-zA-Z0-9_-]+:.*## /{printf "  %-13s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

##@ Setup
setup: tools config hooks build ## First start on a machine: check tools, dudes.json, hooks, build (idempotent)
	@echo "setup done - next: make try (the whole lifecycle with stub workers), then make doctor (probes ollama, aider, packer)"

tools: ## Check the required tools: go, git, make, sh (git-lfs is optional: a warning, since instance repos may need it)
	@for t in go git make sh; do command -v $$t >/dev/null 2>&1 || { echo "missing: $$t" >&2; exit 1; }; done; command -v git-lfs >/dev/null 2>&1 || echo "warning: git-lfs not found - optional, only instance repos that track large files need it" >&2; echo "tools: ok"

config: ## Create dudes.json from dudes.example.json if missing (never overwrites)
	@[ -f dudes.json ] && echo "config: dudes.json kept" || { cp dudes.example.json dudes.json && echo "config: dudes.json created - edit product and models"; }

hooks: ## Point this repo and the sibling instance repos at hooks/ (core.hooksPath, F78)
	@sh scripts/install-hooks.sh

##@ Try it
.PHONY: try
try: build ## Five-minute tour: the stub lifecycle in a fresh scratch repo under .dudes/try, no model, one real gate refusal on the way — stops at the human gate and prints the seal command
	@BIN=$(BIN) sh scripts/try.sh

##@ Build and check
build: ## Build the dudes binary into bin/
	go build -o $(BIN) ./cmd/dudes

test: ## Run the Go tests
	go test ./...

vet: ## Static check (go vet)
	go vet ./...

conventions: ## Common standards of the framework's own files: gofmt, POSIX sh, documented goals
	@sh scripts/conventions.sh

ci: vet conventions test todo-check roadmap-check public-check trend-check ## The framework's own gate: vet, conventions, tests, TODO WIP limit and lanes, no private data, TREND consistency

version: build ## Print the dudes version
	@./$(BIN) version

doctor: build ## Probe git, ollama, aider, packer and config
	./$(BIN) doctor

bench: build ## Hermetic stub lifecycle → .dudes/bench/last.json
	./$(BIN) bench

bench-models: build ## Model bench ladder (F82): MODELS="tag …" OLLAMA=http://host:11434 [LEVELS= REPS= STEPS= NOTES=0] → git notes dudes/bench; MODELS=reference checks a level
	@BIN=$(BIN) sh scripts/bench-models.sh

clean: ## Remove the built binary
	rm -f bin/dudes bin/dudes.exe

##@ Worker image
worker-image: ## Build the pinned code-worker image (Aider + git; Docker, on Windows the one in WSL)
	@sh scripts/worker.sh --build

shell: ## Shell in the worker image, DIR mounted as the worker sees it (DIR=<worktree>, default here)
	@cd $(or $(DIR),.) && sh $(CURDIR)/scripts/worker.sh bash

linux: ## Run ARGS in the host's Linux (WSL on Windows) in DIR, e.g. ARGS="--go 1.26.9 go test ./..." (--node <version> for npm) — the test gate's test_launch
	@cd $(or $(DIR),.) && sh $(CURDIR)/scripts/linux.sh $(ARGS)

##@ Work list
now: ## Show the work in progress (TODO.md Now)
	@$(NOW)

todo-check: ## Fail when TODO.md Now holds more than 3 items (WIP limit)
	@n=$$($(NOW) | wc -l); [ $$n -le 3 ] || { echo "TODO.md Now holds $$n items, WIP limit is 3" >&2; exit 1; }; echo "Now: $$n of 3"

public-check: ## Fail when a tracked file holds private data: private IPs, home paths, e-mail addresses, the local .dudes/private-markers (ROADMAP lane P)
	@sh scripts/public-check.sh

roadmap-check: ## Fail when a TODO.md line names no ROADMAP.md lane; report when a step back is due (DAYS=14 COMMITS=30 RUNS=12)
	@sh scripts/roadmap-check.sh $(or $(DAYS),14) $(or $(COMMITS),30) $(or $(RUNS),12)

##@ Scans and reports
trend-check: ## TREND.md against FEATURES.md (fails on a broken link) and what is due for review (AGE=days, default 90)
	@sh scripts/trend-check.sh $(AGE)

test-ids: ## List acceptance tests (BDD/E2E) that name no requirement ID (ID='<regex>' DIR=<e2e path>, F77)
	@sh internal/testids/check-test-ids.sh "$(ID)" $(DIR)

audit: ## Decision log of the workspace, derived from role-tagged commits, on stdout (ARGS='--since=… [repo …]')
	@sh scripts/audit.sh $(ARGS)

license-scan: ## Licences of what DUDES ships: the Go engine and the worker image's packages (instances: their own goals)
	@sh scripts/license-scan.sh

vuln-scan: ## Known vulnerabilities in what DUDES ships: govulncheck + the worker image (pip-audit)
	@sh scripts/vuln-scan.sh
