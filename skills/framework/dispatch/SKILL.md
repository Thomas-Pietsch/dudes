---
name: dispatch
description: Hand work to the spec-set root that owns the artifact. Workspace DevOps only. Do not load in product role-play or inject into Go workers.
---

# Framework dispatch

Use when the task belongs in a product repo (lead / spec / dev), not in the framework tree.

**Do not** load in a product Feature/Fix flow. **Do not** add to `recipes/<phase>.json`.

## Map

1. Read `.dudes/specset.json`, or the sample `examples/sample/specset.json`.
2. `examples/sample/README.md` says what the sample holds. Roots: `lead` (BM, PS, ROADMAP), `pm` (PRD, TA, SC, UX, QR, BACKLOG). Register companions even if the files are still being written.
3. Hand off to the named root that owns the file. Do not invent `COMPANY.md`.
4. A vanilla clone with no sibling roots **skips** dispatch. Do the work only if it is framework-owned.

## Stay in lane

| Here (framework) | There (named root) |
|------------------|--------------------|
| Engine, `rules/*`, workspace `AGENTS.md`, `skills/framework/` | Product `AGENTS.md`, specs, app code |
| `examples/` (the made-up sample) | Live instance git (do not commit unless the human names that repo) |

Repo types and the freeze → implement → QR → human gate cycle: `rules/DUDES.md` §4 and §6. Do not restate them.
