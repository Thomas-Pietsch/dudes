# Spec-level skills

For the spec repos: product and project specification — PM, UX, security. No model is trained on how DUDES specifies (`CORE.md` §5, *Where skills earn their place*), so the skills here carry that method. They share this background instead of repeating it:

- **The spec-set** is the instance's original documents under named roots; DUDES reads them and never renames them.
- **Every requirement and scenario has an ID** that tests, the QR and commits carry; a statement without an ID is not a requirement.
- **A frozen spec is the contract** a dev run implements; a change after the freeze is a new freeze, never an edit under a running item.
- **The gate is the consistency check:** IDs resolve to their owner document, and the backlog, QR and specs agree. The PM checks it by reading and searching the IDs; where the engine is installed, `dudes item check` does it. The PM runs the sweep before the freeze commit and freezes only with zero open findings (`rules/DUDES.md` §6); where `dudes` is on PATH and the repo has the engine's profile, `hooks/commit-msg` refuses a PM commit of spec documents while findings are open.
- **The company's own conventions** — document names, ID families, review rules — live in the instance's spec repo (`AGENTS.md`, its own `skills/`), not here.

Which role loads which skill in the role-play: `rules/DUDES.md` §1. This background grows with the method; when a spec skill repeats a rule, the rule belongs here.
