---
name: writing-for-agents
description: Write durable artifacts a later worker can parse. Use for specs, QR, and docs.
---

# Writing for agents

1. Durable files are the handoff. Chat is disposable; an agent's memory may point to a rule but never hold it.
2. For behavioral requirements: `### Requirement: <id>`, one SHALL, `#### Scenario:` with GIVEN/WHEN/THEN.
3. For BM, PS, UX, security: keep the instance format. Do not force GIVEN/WHEN/THEN onto research or layout docs.
4. Point at paths, revisions, and section IDs. Do not summarize another document as if it were the source.
5. Wrap worker artifacts in `===BEGIN ARTIFACT===` / `===END ARTIFACT===` when the engine must extract them.
6. Rethink, do not transcribe. The human's input is rough; write it as clean current truth, consistent with the other files, with its limits named, and do not quote it. Quote verbatim only where the instance asks for it (a UI string with its anchor, a saying).
7. One language per document. The instance sets it per document type and names the exceptions where they apply. Input in another language is written in the document's language, by its meaning.

Skills teach discipline. They do not choose the next phase.
