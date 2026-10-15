---
name: code-review
description: Review a change against its frozen spec and the repo's standards, on two separate axes. Use in QA verify and before a merge.
---

# Code review

1. **Pin the diff.** Review `git diff <fixed point>...HEAD` and its commits (the item's commits carry `Item:` and `Spec-Source:` trailers), not a chat summary. A bad ref or an empty diff stops the review.
2. **Two axes, reported apart; never merge or re-rank them:**
   - **Spec:** what the frozen spec or item asked for that is missing or partial; behaviour nobody asked for (scope creep); what looks implemented but wrong. Quote the spec line (its ID) for each.
   - **Standards:** where the diff breaks the repo's documented standards (`AGENTS.md`, its conventions; cite the rule). Skip what tooling already enforces.
3. **Smells are judgement calls**, named as such ("possible duplicated code") and never as violations. A documented repo standard overrides them. Worth naming: a mysterious name, duplicated logic, speculative generality (hooks for needs the spec does not have), a primitive standing in for a domain concept, one change scattered over many files.
4. **Field and business claims** in BM/PS are not software defects. Do not demand a unit test for PMF.
5. **A missing consequential default** is a human decision. Ask; do not invent product policy.
6. **Reviewing a red test** (a test task's `approval: after` pause): it must fail *only* because the behaviour is missing.
   - Paths resolve from where the test runs (`go test`: the package directory); every flag or command it calls is in the spec, none invented.
   - Each case the criterion names is asserted with spec values; a check matches only what it means (`file:line`, not a digit anywhere in the output).
   - One detection function per rule, used for the violating fixture, the clean case and the real tree.
   - A failure fails: no skip, no log-and-continue; only a root that does not exist yet may be passed over, explicitly.
   - Answer `dudes decide <task>` or `dudes redo <task> "<the concrete defect>"`.
7. **Output:** concrete `file:line` findings under `## Spec` and `## Standards`, then one line per axis: the number of findings and the worst one. No phase-graph advice.

Skills teach discipline. They do not choose the next phase.
