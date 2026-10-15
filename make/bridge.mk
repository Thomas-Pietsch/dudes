# The bridge (F88): how the role-play reaches the engine in a dev repo that tracks dudes/ — each goal
# is one engine command, two at most, with no logic of its own (CORE.md §13). A dev repo includes this
# file by its path in the DUDES clone, never a copy, so its goals cannot drift:
#     -include .dudes-framework/make/bridge.mk
# -include: a checkout without the clone (a CI) still runs its own goals. The engine is the clone's
# binary (make -C <clone> build) unless DUDES_BIN names another; DUDES_CONFIG names your own dudes.json
# (machine values: the model host, the worker's launcher) when it is not ./dudes.json, never the repo's.

BRIDGE_ROOT := $(dir $(lastword $(MAKEFILE_LIST)))..
BRIDGE_DEFAULT_GOAL := $(.DEFAULT_GOAL)
DUDES_BIN ?= $(BRIDGE_ROOT)/bin/dudes$(if $(filter Windows_NT,$(OS)),.exe)
DUDES_CONFIG ?=

# A goal's words reach the engine as typed. make would expand a $(…) in them, and running it when it
# exports a word: so it takes each word's raw text ($(value …)), exports that copy, and keeps the word
# itself out of the environment. A recipe quotes the copy ("$$BRIDGE_WHY"), and the shell never parses
# its text again, whatever it holds: ' " # ; ` $.
BRIDGE_ITEM := $(value ITEM)
BRIDGE_SLICING := $(value SLICING)
BRIDGE_SLICE := $(value SLICE)
BRIDGE_TASK := $(value TASK)
BRIDGE_ROLE := $(value ROLE)
BRIDGE_WHY := $(value WHY)
export BRIDGE_ITEM BRIDGE_SLICING BRIDGE_SLICE BRIDGE_TASK BRIDGE_ROLE BRIDGE_WHY DUDES_BIN DUDES_CONFIG
unexport ITEM SLICING SLICE TASK ROLE WHY

.PHONY: slice review approve redo land

slice: ## Prepare ITEM's slice from SLICING and run it to its first pause (ITEM=BK-001 SLICING=slices/bk-001a.json)
	@"$$DUDES_BIN" $(if $(DUDES_CONFIG),-config "$$DUDES_CONFIG") do "$$BRIDGE_ITEM" -slice "$$BRIDGE_SLICING" -run

review: ## What SLICE waits on: its run, the waiting task's review packet or the worker's question (SLICE=bk-001a)
	@"$$DUDES_BIN" -config ".worktrees/$$BRIDGE_SLICE/.dudes/dudes.json" status

approve: ## Approve SLICE's waiting TASK as ROLE — or answer its question with WHY — then run to the next pause (SLICE= TASK= ROLE= [WHY=])
	@"$$DUDES_BIN" -config ".worktrees/$$BRIDGE_SLICE/.dudes/dudes.json" decide "$$BRIDGE_TASK" -by "$$BRIDGE_ROLE" "$$BRIDGE_WHY"
	@"$$DUDES_BIN" -config ".worktrees/$$BRIDGE_SLICE/.dudes/dudes.json" run

redo: ## Send SLICE's TASK back as ROLE with WHY, then run it to its next pause (SLICE= TASK= ROLE= WHY=)
	@"$$DUDES_BIN" -config ".worktrees/$$BRIDGE_SLICE/.dudes/dudes.json" redo "$$BRIDGE_TASK" -by "$$BRIDGE_ROLE" "$$BRIDGE_WHY"
	@"$$DUDES_BIN" -config ".worktrees/$$BRIDGE_SLICE/.dudes/dudes.json" run

land: ## Merge the ended SLICE into this branch with ROLE's tag, then remove its worktree (SLICE= ROLE=)
	@"$$DUDES_BIN" land "$$BRIDGE_SLICE" -by "$$BRIDGE_ROLE"

# the including Makefile keeps its own default goal, wherever it includes this file
.DEFAULT_GOAL := $(BRIDGE_DEFAULT_GOAL)
