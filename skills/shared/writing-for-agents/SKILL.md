---
name: writing-for-agents
description: Write durable artifacts a later worker can parse. Use for specs, QR, and docs.
---

# Writing for agents

1. Durable files are the handoff. Chat is disposable.
2. For behavioral requirements: `### Requirement: <id>`, one SHALL, `#### Scenario:` with GIVEN/WHEN/THEN.
3. For BM, PS, UX, security: keep the instance format. Do not force GIVEN/WHEN/THEN onto research or layout docs.
4. Point at paths, revisions, and section IDs. Do not summarize another document as if it were the source.
5. Wrap worker artifacts in `===BEGIN ARTIFACT===` / `===END ARTIFACT===` when the engine must extract them.

Skills teach discipline. They do not choose the next phase.
