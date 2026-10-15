# Terminal & Shell

> For IDE agents (Claude, Cursor, …) working in a repo. How to use the terminal is mostly answered by the repo's goals; this file keeps only what a goal cannot carry.

## Goals first

Every repeatable action is a make goal — `make help` lists them (`CORE.md` §13). Run the goal, never the raw command behind it (no hand-built `go test`, `npm run`, `docker run`, script paths). A job without a goal is a goal to add, not a command to improvise.

## Git: non-interactive

- No pager: `git --no-pager log`, `git --no-pager diff`, `git --no-pager show`.
- No interactive flags (`-i`, `rebase -i`, `add -p`) and no editor: pass the message (`-m`, `-F`).
- Stage files by name, never `git add -A` / `commit -a` in shared repos.
- Hooks are not optional: no `--no-verify`.

## Unclear output

If command output is cut off or garbled, run it once more with the output captured to a file and read that. Never take a gate decision (QA, release, seal) on output you could not read in full.

## References
- Requires: `SECURITY.md` · Required by: `PERFORMANCE.md`
