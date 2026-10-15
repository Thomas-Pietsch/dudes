---
name: work-item
description: Build exactly one backlog work item from its item pack (RULES / TASK / CONTEXT). Use in test_authoring, implementation and qa when the contract is item-bound.
---

# Work-item discipline

The engine hands you one item pack (F68). It is your whole specification. Nothing else exists.

1. **Only the pack.** Do not assume files, routes, tables, texts or thresholds that are not written in RULES, TASK or CONTEXT. Missing or contradictory → stop and report `BLOCKED: <what is missing> — <which reference would own it>`. Do not invent.
2. **RULES win.** Invariants in RULES override anything in TASK.
3. **Build only TASK.** A build-only FR is implemented as written. An `Implements [[…]]` FR is implemented exactly as the referenced contract in CONTEXT.
4. **Tests first.** Write the acceptance criteria as failing tests, then the code. Name tests after the item and AC (`BK-001 AC-3`).
5. **Copy verbatim.** UI text comes only from the UX definitions in CONTEXT.
6. **Never bend a contract.** A contract that looks wrong is reported as `CONTRACT: <ref> — <problem>`. A human changes the owner document; the engine regenerates the pack.
7. **No new reach.** No network, no new dependency unless CONTEXT names it.

Report: the item ID, the `manifest.json` source hashes the work is based on, assumptions taken under the item's narrowings, and which DoD rows your tests cover.

Skills teach discipline. They do not choose the next phase.
