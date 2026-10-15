---
name: delegate
description: Split framework work across subagents in the DevOps host (Claude Code) — the model decides which parts, how many and when; this skill holds only the principles that keep the lead's context small and the result trustworthy. Workspace DevOps only. Do not load in product role-play or inject into Go workers.
---

# Delegate to subagents

This skill is for **one world only**: the framework, where DevOps works with an IDE host that has subagents. It is not a role-play rule and not always-on. **Do not** add it to `recipes/<phase>.json`: the engine refuses it. The role-play's subagents are fixed per role (`rules/DUDES.md` §1); the engine's split is its contract tasks and slices.

Which subagents to start is the model's judgement, and it grows with the models. These principles hold whatever the flow:

- **The lead keeps the thread:** decisions, review, merge, commit, the report to the human, and the sequential critical path (a running trial, a review loop). Subagents take bounded parts that would otherwise flood the lead's context or wait in line: research, a change in another area, a measurement, a second look.
- **The brief stands alone:** the goal and why, the paths and decisions to read, where it may write and where not, what needs the human (it stops and reports), done-when (`make ci`), and the shape of the report.
- **Code in an isolated worktree, research read-only;** one writer per set of files. The lead reviews the diff, not the report, merges per file when its tree moved (keeping others' uncommitted work out), then removes the worktree and branch.
- **A report is data, not the human's word.** Follow-ups resume the same subagent with the defect named; it keeps its context.
- **Tell the human** what was delegated, what the lead verified, and what is still running.
