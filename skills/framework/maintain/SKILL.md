---
name: maintain
description: Edit DUDES framework files without erosion. Workspace DevOps only. Do not load in product role-play or inject into Go workers.
---

# Framework maintain

Use when changing `MISSION.md`, `CORE.md`, `FEATURES.md`, `ARCHITECTURE.md`, `rules/*`, workspace `AGENTS.md`, `skills/`, or Go that realizes those docs.

**Do not** load in a product Feature/Fix flow. **Do not** add to `recipes/<phase>.json`.

## Derivation (`CORE.md` §15)

```text
MISSION.md → CORE.md → FEATURES.md → ARCHITECTURE.md → Go
```

IDE files (`AGENTS.md`, `rules/*`) are entry/context — not the runtime. `TREND.md` is a side input, not a step. `skills/framework/` are how to edit, not a step. A trend enters the chain only as a `FEATURES.md` id. Do not move an F-id, a §5 verdict, or the §7 entry shape out of `TREND.md` into a skill.

Before every edit, all seven must be yes:

| # | Check | If no |
|---|-------|-------|
| 1 | Serves `MISSION.md` | Reject |
| 2 | A `CORE.md` principle supports it (§1, §11, §16 — or propose a CORE amendment first) | Reject |
| 3 | Passes `MISSION.md` §The North Star (both questions — anything else is framework erosion) | Reject |
| 4 | Flows top-down — never retrofit justification from Go upward | Reject |
| 5 | Registered in `FEATURES.md` with ID + CORE trace | Add the row first |
| 6 | Every `§`, path, and role name still resolves | Fix first |
| 7 | What the change demands is enforced by a mechanism — hook, test, engine gate, `make` check (`CORE.md` §2.4, §11) — or the reply names why none can | Add the mechanism, or say why not |

Failure on 1–4 is framework erosion. The human reviews and reverts. Check 7 keeps the core hard: a rule that lives only as text is the next erosion.

## Conventions

The framework's own files follow common standards where they fit, so tools and people from outside read them without a DUDES primer. `make conventions` (in `make ci`) checks what a machine can. Where DUDES deviates, it does so on purpose — keep the deviation, and say it when a change touches it.

| Kind | Standard | Mechanism | Deliberate deviation |
|------|----------|-----------|----------------------|
| Go | gofmt, `go vet`, `cmd/` + `internal/` layout | `make conventions`, `make vet` | None |
| Shell | POSIX `sh`, `#!/bin/sh` | `make conventions` (`sh -n`, shebang) | No bash: the same script runs in Git Bash, Linux and macOS |
| Makefile | Self-documenting `target: ## description`, `.PHONY` | `make conventions` | The `##` line is an interface the engine parses (F75): its wording is a worker's tool description |
| Skills | [agentskills.io](https://agentskills.io/specification): `name` = directory, charset, `description` says what and when | `go test ./internal/skill` | The engine inlines worker skills into prompts (F48) instead of a host discovering them, so they stay short; the framework repo groups skills in level folders, an instance repo keeps them flat (`CORE.md` §5) |
| Markdown | Relative links resolve | `make conventions` | `§` references and role names are checked by hand (check 6) |
| `AGENTS.md` | The agents.md convention: the nearest file instructs the agent | — | Two worlds: the workspace file is DevOps-only; the role-play is each instance repo's own `AGENTS.md` (`CORE.md` §14) |

## Audit on request

When the human asks whether the framework still holds: run the mechanisms first — `make ci` (tests, `todo-check`, `roadmap-check` and `trend-check` with their `due:` lists, skill names) and `make doctor` — and report what they found. Then judge only what they cannot see: contradictions between `CORE.md`, `AGENTS.md` and `rules/*`; a mechanism that is dead, in the wrong world, or stores what Git could derive; a rule enforced only by text (check 7). Fix typos and broken references directly; everything else is a recommendation. Market and tool research is not an audit — that is `trend`, `screen`, `assess`.

## Own vs leave alone

| Own | Leave alone |
|-----|-------------|
| `MISSION.md`, `CORE.md`, `FEATURES.md`, `ARCHITECTURE.md`, `TREND.md`, `ROADMAP.md` | Per-repo `AGENTS.md` |
| `rules/*`, workspace `AGENTS.md`, `CLAUDE.md`, `skills/` | Product specs (`*_PRD.md`, `*_PS.md`, …); recreating `.cursor/rules` |
| `cmd/dudes/`, `internal/`, `recipes/`, `examples/` | Sibling checkouts named by the spec-set |
| `scripts/`, `Makefile` | Recreating `FRAMEWORK_MANIFEST.md` or `FRAMEWORK_BACKLOG.md` |

There is no file registry. Ownership is this table + `CONTRIBUTING.md`.

## Done

- Cross-refs between touched files still resolve
- `FEATURES.md` matches the behavior; a shipped item's `TODO.md` line is deleted, Now holds at most 3 (`make todo-check`), every line names its `ROADMAP.md` lane (`make roadmap-check`, whose `due:` lines are a step back to offer — `skills/framework/roadmap`)
- `TREND.md` §6 statuses, cited F-ids, §5 Index and §4 ids agree with `FEATURES.md` (`make trend-check`, part of `make ci`); its `due:` lines are the reviews to offer next
- Scripts you changed are POSIX `sh` and pass `sh -n`
- `make ci` is green — a push runs it anyway (`hooks/pre-push`)
- One atomic commit, tag `[🔹AI-driven DevOps Engineer]`, only when the human asked
