# Context Persistence & Recovery Rules

## SESSION_STATE.md — Scratchpad Protocol

For any task that spans **3+ tool calls or 2+ role handoffs**, the agent MUST maintain a `SESSION_STATE.md` file in the workspace root.

### When to Create

| Condition | Action |
|---|---|
| Task involves 2+ roles | Create at first role's start |
| Multi-role flow is active | **Mandatory** — create when the flow starts |
| Task involves 3+ files being edited | Create before second edit |
| Task will take >5 tool-call rounds | Create at start |
| Simple single-file fix | Skip — not needed |

### Format

```markdown
# Session State — <task summary>

## Current Role: 🔹 <Role>
## Phase: <spec / impl / qa / release>
## Flow: <Feature | Fix | Single-Role> | Step: <current role>

## Done
- [x] <completed step + key decision made>

## In Progress
- [ ] <current step>

## Next
- [ ] <planned step>

## Key Decisions (mid-session)
- <decision>: <rationale> (decided in step N)

## Files Modified
- `path/to/file.ext` — <what changed>
```

### Rules

| Rule | Detail |
|---|---|
| **Update after every role handoff** | Outgoing role updates Done + In Progress + Flow step before handing off |
| **Update after every major decision** | Any non-obvious choice goes in Key Decisions |
| **Delete at task end** | Remove `SESSION_STATE.md` after ✅ COMPLETE + commit |
| **Never commit** | This file is ephemeral — add to `.gitignore` if not already |
| **Re-read on every resumption** | On ANY session resumption or context summary — `read_files SESSION_STATE.md` is the **first action**, before selecting a role. Not optional. |

---

## Context Recovery Protocol

If the agent detects context loss (repeated reads of already-seen files, contradictory decisions, explicit compaction notice, or conversation starts with a context summary):

1. **Read `SESSION_STATE.md`** — restores what was done, what's in progress, key decisions, and flow state
2. **Inherit flow state** — the `## Flow` field is the source of truth. Never re-derive flow status from compressed context alone.
3. **Re-read the frozen PRD** — read the PRD sections relevant to the current phase. This is the primary defence against context drift: every role re-grounds in the source spec, not in compressed context or another role's output (PRD-Direct Rule, `DUDES.md` §5).
4. **Run `git diff --stat`** — shows what files were already modified this session
5. **Continue from "In Progress"** — do not restart or duplicate completed work

---

## Micro-Checkpoint (inline)

For tasks under the SESSION_STATE.md threshold, the agent may append `<!-- CHECKPOINT: ... -->` to the active file (deleted before commit; `hooks/pre-commit` refuses a staged `<!-- CHECKPOINT`).

---

## Pre-Task Context Snapshot

Before starting a multi-role task, the **first role** captures the starting state:

1. Read relevant spec docs (PRD, TA sections for the feature)
2. Create SESSION_STATE.md with the plan
3. Begin work

---

## Interaction with Other Rules

| Rule | Relationship |
|---|---|
| `CONVERGENCE.md` | Convergence loop re-reads source docs — SESSION_STATE.md is NOT a source doc, it's a scratchpad |
| `PERFORMANCE.md` | SESSION_STATE.md updates are exempt from "no redundant writes" — they are state persistence |
| `REPORT.md` | Task Complete report replaces SESSION_STATE.md at end — delete the scratchpad |
| `DUDES.md` | Each role updates SESSION_STATE.md on handoff |

---

## API Crash Recovery

If a session crashes with a `400 invalid_request_error` referencing an orphaned `tool_result`, the conversation history is **unrecoverable**.

**Recovery steps:**
1. **Start a new chat.**
2. **Read `SESSION_STATE.md`** — contains plan, completed steps, key decisions, and flow state.
3. **Continue from "In Progress"** — do not restart completed work.
4. If no `SESSION_STATE.md` exists, run `git --no-pager diff --stat` to reconstruct context.

## References
- Requires: `DUDES.md`, `AGENTS.md` · Required by: `CONVERGENCE.md`, `REPORT.md`
