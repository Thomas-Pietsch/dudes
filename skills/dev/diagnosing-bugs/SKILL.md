---
name: diagnosing-bugs
description: Find the cause of a failing test or a reported defect with a tight feedback loop before changing code. Use in the fix flow.
---

# Diagnosing bugs

1. **Start from the durable failure:** the recorded run (`git notes --ref dudes/test show <commit>`) and the failing test's name.
2. **Build the loop first.** One command that goes red on *this* symptom, already run once: the narrowest `go test -run` / the task's gate, or a fixture input through the binary. It must be deterministic and take seconds. Without it, no theory: reading code to guess is the failure this skill prevents. If no loop can be built, say what you tried and ask.
3. **Minimise.** Cut inputs, steps and configuration one at a time, re-running the loop after each cut, until every remaining part is needed for the failure.
4. **Rank 3–5 falsifiable hypotheses** before testing any: "if X is the cause, changing Y makes it disappear". A hypothesis without a prediction is discarded.
5. **One variable per probe.** Prefer a debugger or one targeted log at the boundary that separates two hypotheses. Tag every debug line (`[DEBUG-a4f2]`) so one grep removes them all.
6. **Regression test before the fix**, at a seam where the real bug pattern occurs; watch it fail, fix, watch it pass, rerun the original loop. No correct seam is itself a finding: say so.
7. **Requirement IDs:** if the log does not mention a requirement ID, say so. Do not green-wash coverage.
8. **Finish clean:** debug lines removed, the confirmed cause stated in the commit message. Stop when the failing test is explained or a human decision is needed. No unrelated edits.

Skills teach discipline. They do not choose the next phase.
