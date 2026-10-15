---
name: ops
description: Run framework scripts, IaC, scans, and git mechanics. Workspace DevOps only. Do not load in product role-play or inject into Go workers.
---

# Framework ops

Use for Makefiles, Docker/Terraform, scans, hooks, and non-trivial git. Not for product-repo app code (that is `dispatch`).

**Do not** load in a product Feature/Fix flow. **Do not** add to `recipes/<phase>.json`.

## Tools

`make help` lists every framework goal — setup, build, checks, work list, scans. Run the goal, not the script behind it; a new repeatable job becomes a goal with a `## ` line, not a row in this skill (`CORE.md` §13).

Config review (Docker, Nginx, systemd — misconfig and exposure only) stays judgement, not a goal.

There is **no** `framework-sync.sh` and **no** `FRAMEWORK_MANIFEST.md`. Do not recreate them. Instance `AGENTS.md` is never overwritten from upstream.

## Git

- No pager, no `-i`, no force-push unless the human asked.
- No commit unless the human asked.
- Complex merge / rebase / spec-ref pointer updates stay here; product content stays in the named spec-set root.
