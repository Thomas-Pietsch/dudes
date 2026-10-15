# The bridge (F88): how the role-play reaches the engine in a dev repo that tracks dudes/ — each goal
# is one engine command, two at most, with no logic of its own (CORE.md §13). A dev repo includes this
# file by its path in the DUDES clone, never a copy, so its goals cannot drift:
#     -include .dudes-framework/make/bridge.mk
# -include: a checkout without the clone (a CI) still runs its own goals. The engine is the clone's
# binary (make -C <clone> build) unless DUDES_BIN names another; DUDES_CONFIG is your own dudes.json
# (machine values: the model host, the worker's launcher), never the repo's.

BRIDGE_ROOT := $(dir $(lastword $(MAKEFILE_LIST)))..
BRIDGE_DEFAULT_GOAL := $(.DEFAULT_GOAL)
DUDES_BIN ?= $(BRIDGE_ROOT)/bin/dudes$(if $(filter Windows_NT,$(OS)),.exe)
DUDES_CONFIG ?= dudes.json
SLICE_CONFIG = .worktrees/$(SLICE)/.dudes/dudes.json

.PHONY: slice review approve redo land

slice: ## Prepare ITEM's slice from SLICING and run it to its first pause (ITEM=BK-001 SLICING=slices/bk-001a.json)
	@"$(DUDES_BIN)" -config "$(DUDES_CONFIG)" do $(ITEM) -slice "$(SLICING)" -run

review: ## What SLICE waits on: its run and the waiting task's review packet (SLICE=bk-001a)
	@"$(DUDES_BIN)" -config "$(SLICE_CONFIG)" status

approve: ## Approve SLICE's waiting TASK as ROLE, then run to the next pause (SLICE= TASK= ROLE= [WHY=])
	@"$(DUDES_BIN)" -config "$(SLICE_CONFIG)" decide $(TASK) -by "$(ROLE)" $(WHY)
	@"$(DUDES_BIN)" -config "$(SLICE_CONFIG)" run

redo: ## Send SLICE's TASK back as ROLE with WHY, then run it to its next pause (SLICE= TASK= ROLE= WHY=)
	@"$(DUDES_BIN)" -config "$(SLICE_CONFIG)" redo $(TASK) -by "$(ROLE)" "$(WHY)"
	@"$(DUDES_BIN)" -config "$(SLICE_CONFIG)" run

land: ## Merge the ended SLICE into this branch with ROLE's tag, then remove its worktree (SLICE= ROLE=)
	@"$(DUDES_BIN)" land $(SLICE) -by "$(ROLE)"

# the including Makefile keeps its own default goal, wherever it includes this file
.DEFAULT_GOAL := $(BRIDGE_DEFAULT_GOAL)
