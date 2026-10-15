# Convergence Passes

> **Toggle — default OFF.** Switch on with `Convergence: ON` in the prompt, or set a default in the instance `AGENTS.md` (`rules/DUDES.md` §1 Rule set). A strong model usually finishes in one pass; switch it on when a role leaves work undone. The Go engine runs the same mechanism for its micro-agents (`convergence` in `dudes.json`, F72).

A task is done when a fresh look finds nothing left to do — not when the first pass says it is done.

---

## The Passes

Pass 1 is the work itself. When it ends, the role does not declare completion. It starts a new pass:

```
PASS n (n ≥ 2):
  1. FRESH START — re-read the task and ALL source docs from the beginning.
     Never work from memory of pass n−1. Re-read the outputs as they are now.
  2. RE-ANALYSE — what does the task require? What is still missing, wrong,
     or inconsistent? Run the Structural Checklist below.
  3. FINDINGS TABLE (visible):
     | # | Finding | Action |
     or: "Findings: 0"
  4. ACT — fix every finding.
  5. EXIT when a pass has zero findings and changes nothing.
     Otherwise → pass n+1.
```

Pass 2 compares the task with what pass 1 left behind; it is not a review of pass 1's reasoning. As the role, ask: *if a senior received this output cold, what would they flag?* Re-read your role's Workflow and Definition of Done in the repo `AGENTS.md`.

---

## Bounds

| Rule | Detail |
|------|--------|
| **Max 5 passes** | Pass 5 still has findings → stop, list them, ⏸ WAITING |
| **No repeat** | The same finding two passes in a row → stop, ⏸ WAITING with that finding (`PERFORMANCE.md` loop prevention) |
| **Commit after the loop** | `DUDES.md` §5 step 6 — not between passes |

---

## Structural Checklist

| # | Check | Pass condition |
|---|---|---|
| 1 | **Scope coverage** | Every requirement / AC has been addressed |
| 2 | **Cross-references** | Every ID (TC-xx, ERR_xx, field, endpoint) resolves correctly |
| 3 | **Consistency** | Enums, field names, state values identical across all sections |
| 4 | **Missing schema** | Every referenced entity has a DDL or schema definition |
| 5 | **Missing error codes** | Every error condition has a catalogue entry |
| 6 | **Missing test cases** | Every new behaviour has at least one TC |
| 7 | **Dangling prose** | No "see below" where the target is absent |
| 8 | **Gate items** | Hard gates are inside the gate block, not floating after it |

---

## Signal

```
✅ COMPLETE — converged after N passes (pass N: zero findings, no change).
```

## References
- Requires: `DUDES.md` · Toggled by: `DUDES.md` §1 Rule set
