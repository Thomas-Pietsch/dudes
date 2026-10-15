---
name: tdd
description: Turn the item's failing acceptance tests green, test first. Use in test_authoring and implementation.
---

# Test-driven discipline

DUDES already sequences test_authoring before implementation. This skill is how to work inside those phases.

1. **The spec is the source of truth.** Read the criterion in the item pack (or the frozen PRD section). Do not invent a competing spec.
2. **Keep the test's name and file.** Each acceptance criterion has its own failing test (`TestBK001AC1` in `tests/bk001ac1_test.go`; `TestReq_<id>` without an item). Replace its body in place. The name carries the ID the engine traces.
3. **Test behaviour through the public interface**, the way a caller uses it: run the binary, call the exported function, send the request. No mocks of internal parts, no private functions, no checking a database behind the interface's back.
4. **Expected values come from the spec**, written as literals (the criterion's example, a worked case), never recomputed the way the code computes them. A test that cannot disagree with the code proves nothing.
5. **A missing precondition fails; it does not skip.** `t.Skip` on a missing file or fixture hides the gap the test exists to show.
6. **Assert behaviour (GIVEN/WHEN/THEN)**, not business hypotheses (willingness to pay, PMF).
7. **Red, then green.** Write only the code that turns the test green. No speculative features, no drive-by refactors. Refactoring belongs to review.
8. **The engine runs the gate** after you exit: the task's test command. Do not invent a test runner.

Skills teach discipline. They do not choose the next phase.
