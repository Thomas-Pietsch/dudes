---
name: retro
description: Retrospective on a DevOps session or an engine run — where the agent or the worker lost time or made a mistake, and which mechanism, pointer or tool would have prevented it. Workspace DevOps only, on request. Do not load in product role-play or inject into Go workers.
---

# DUDES retro

This skill is for **one world only**: the framework. It is not a role-play rule and not always-on. **Do not** add it to `recipes/<phase>.json`: the engine refuses it.

It answers one question: **what in the environment should change so the next session or run goes better?** It improves the environment, not the people, and not the model. Model choice is `skills/framework/bench`.

## Read the primary sources

The session the human names, else the current one. Stay inside the workspace (`rules/SECURITY.md`):

- the conversation;
- `git log` of the session's commits, including their messages and reverts;
- for an engine run: `.dudes/events.jsonl`, the run's git notes (`dudes/run`, `dudes/test`, `dudes/acceptance`, `dudes/bench`), the step logs, the bench logs in `.dudes/bench-models/`.

Transcripts outside the workspace are out of reach unless the human hands them over.

## Look for candidates

| Category | Use when | DUDES answer, in this order |
|---|---|---|
| **Automated checks** | A mistake a check could have caught, or a check exists but is not wired | A test, a `make ci` step, a hook (`CORE.md` §11: mechanisms over rules). Read the existing goals first: an unwired check is the finding, not a new one |
| **Engine defects** | A worker failed on something the engine controls: scope, gate feedback, a hint, a flag | An engine fix with a test (`maintain`). The trial measures the engine as much as the model |
| **Navigation** | Long searches for a file or a fact | One pointer where the agent looks first: `AGENTS.md`, a README, a skill's *Read first*. Never a copy |
| **Rules** | A rule was missed or misread | Mechanical → a check. Judgement → sharpen the rule in one place. Never-read → delete it (a no-op instruction) |
| **Tool economy** | Expensive or repeated tool calls, polling, oversized output | A make goal, a narrower command, a quieter output |
| **Information access** | A needed fact was not available: a log, a state, a remote's status | A goal or a doctor check that surfaces it |

A mechanical violation always gets a deterministic check. A rule is only for a genuine judgement call.

## Reply

Candidates ordered by severity: for each, what happened (with the evidence: commit, log line, step), the change, and its kind (check, engine fix, pointer, rule edit, deletion). At most the top 3–5.

The human pulls what to do into `TODO.md`; nothing changes during the retro. Landing follows `skills/framework/maintain`.
