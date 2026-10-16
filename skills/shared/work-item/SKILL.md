---
name: work-item
description: Build exactly one backlog work item from its item pack (RULES / TASK / CONTEXT). Use in test_authoring, implementation and qa when the contract is item-bound, and in the role-play when one backlog item is built.
---

# Work-item discipline

One item pack is your whole specification. Nothing else exists. The pack has three parts:

| Part | Holds |
|------|-------|
| RULES | The invariants of the PRD and the TA that bind every item |
| TASK | The backlog item: its FRs, acceptance criteria and DoD rows |
| CONTEXT | What TASK references: the contracts (`Implements [[…]]`), the UX definitions, the data |

**Where the pack comes from.** The engine assembles it (F68): `dudes item pack <ID>` writes the three parts and a `manifest.json` with the hash of every source, and hands them to you. In the role-play there is no engine: the pack is the backlog item plus the frozen PRD, TA and UX sections it references, which you read yourself by following its `[[DOC:ID]]` references. The same three parts, the same rule: what the item and its references do not say, you do not know. In place of the manifest, you name the freeze commit you read.

1. **Only the pack.** Do not assume files, routes, tables, texts or thresholds that are not written in RULES, TASK or CONTEXT. Missing or contradictory → stop and report `BLOCKED: <what is missing> — <which reference would own it>`. Do not invent. A reference you cannot resolve is missing, not a hint to guess.
2. **RULES win.** Invariants in RULES override anything in TASK.
3. **Build only TASK.** A build-only FR is implemented as written. An `Implements [[…]]` FR is implemented exactly as the referenced contract in CONTEXT.
4. **Tests first.** Write the acceptance criteria as failing tests, then the code. Name tests after the item and AC (`BK-001 AC-3`).
5. **Copy verbatim.** UI text comes only from the UX definitions in CONTEXT.
6. **Never bend a contract.** A contract that looks wrong is reported as `CONTRACT: <ref> — <problem>`. A human changes the owner document; the engine regenerates the pack, the role-play reads the changed section after the next freeze.
7. **No new reach.** No network, no new dependency unless CONTEXT names it.

Report: the item ID, the sources the work is based on (the `manifest.json` hashes, or the freeze commit in the role-play), assumptions taken under the item's narrowings, and which DoD rows your tests cover.

Skills teach discipline. They do not choose the next phase.
