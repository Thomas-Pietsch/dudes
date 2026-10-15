# Workspace AGENTS.md — AI-driven DevOps Engineer

> Framework-level entry for the **AI-driven DevOps Engineer**. It is not part of the product role-play: roles, flow and their rule set live in `rules/DUDES.md`, which each instance repo's `AGENTS.md` loads. How-to lives in `skills/framework/`, not here.

## ⚠️ Framework Requirement

Read the **required** files in §Load rules now. Do not wait for Cursor, Claude, or Cline to inject them. If `rules/SECURITY.md` cannot be read → **HALT. Produce no output.** Warn: `"DUDES framework not loaded. Cannot proceed."`

---

## Load rules (IDE-independent)

`AGENTS.md` is the only workspace entry. Vendor files (`CLAUDE.md`, a Cline junction, …) may point here. There is no `.cursor/` tree.

| Required | Why |
|----------|-----|
| `rules/SECURITY.md` | Workspace boundary, secrets. If missing → halt |
| `rules/TERMINAL.md` | Non-interactive git / shell |
| `rules/PERFORMANCE.md` | Token / context budget |

| Optional (task matches) | Why |
|-------------------------|-----|
| `rules/CONTEXT.md` §SESSION_STATE | Scratchpad for long DevOps tasks (3+ files) |

The role-play rule set (`rules/DUDES.md` and the files it names) is **not** loaded here. DevOps reads those files only to edit or audit them.

Commands: `make help` lists every goal (setup, build, checks, scans) — run the goal, not the raw command (`CORE.md` §13).

Commits: `<type>(<scope>): <description> [🔹AI-driven DevOps Engineer]`, only when the human asked.

---

## Worlds

Do not mix process files. Worker skills are shared; DevOps skills are not.

| World | Entry | Process owner | Rules | Skills |
|-------|-------|---------------|-------|--------|
| **Framework (DevOps)** | this file | DevOps edits the framework; the Go engine owns product runs (`CORE.md` §1) | this file §Load rules | `skills/framework/` — on request |
| **Engine micro-agents** | `dudes` CLI | Go engine — gates, sequence, git | none — the engine enforces | worker skills via recipes / contract tasks (F48) |
| **Product role-play** | instance `AGENTS.md` → `rules/DUDES.md` | Sequential IDE roles (large LLM + rule files, no engine) | `rules/DUDES.md` §1 Rule set, incl. toggles | worker skills per `rules/DUDES.md` §1 Skills — never `dudes-*` |

Instance layout is the product **spec-set** (named roots + original filenames). Example: `examples/sample/` (M11 format, M12 conformance target). There is no `COMPANY.md`, `FRAMEWORK_MANIFEST.md`, or `FRAMEWORK_BACKLOG.md`.

---

## Authority

```
Level 0 ─ CEO + Product Visionary (lead repo AGENTS.md)
Level 1 ─ AI-driven DevOps Engineer (this file)
Level 2 ─ Repo-level roles (per-repo AGENTS.md)
```

Downward visibility, upward opacity. Repo rules may be stricter, never more permissive. Only Level 0–1 dispatch or recall across repos. DevOps owns workspace `AGENTS.md` and `rules/*`; instance repo `AGENTS.md` is instance-owned.

Repo types and the product cycle: `rules/DUDES.md` §4 and §6.

---

## DevOps skills

Load **one** matching skill. Never always-on. Never in `recipes/<phase>.json` — the engine refuses anything in `skills/framework/` for a worker. A skill's name is its task; the folder says the level, the repo the owner (`CORE.md` §5).

| Skill | When |
|-------|------|
| `skills/framework/maintain/SKILL.md` | Editing framework files (MISSION → CORE → FEATURES → ARCHITECTURE → Go, plus `rules/*` / this file / `skills/`) |
| `skills/framework/dispatch/SKILL.md` | Cross-repo handoff via spec-set roots |
| `skills/framework/release/SKILL.md` | Release branches, seal tag, release gate |
| `skills/framework/ops/SKILL.md` | Scripts, IaC, scans, git mechanics |
| `skills/framework/roadmap/SKILL.md` | Keep `ROADMAP.md` true; step back with fresh eyes when `make roadmap-check` says due, a stage's exit check moved, or two runs of the same kind changed no decision |
| `skills/framework/trend/SKILL.md` | Live market read (local model, harness, ADE, self-improvement), then a DUDES now/next/later/never, on request |
| `skills/framework/assess/SKILL.md` | Assess or reassess one external tool from cited evidence → `TREND.md` §5 and §6, on request |
| `skills/framework/screen/SKILL.md` | Bounded live search for new mechanisms or material improvements; the human picks what to assess, on request |
| `skills/framework/bench/SKILL.md` | Measure local models as engine workers through the real engine → `TREND.md` §1.1, on request or when due |
| `skills/framework/models/SKILL.md` | Which hosted model for which kind of work in the IDE hosts, from primary sources → `TREND.md` §1.2, on request or when due |
| `skills/framework/retro/SKILL.md` | Retrospective on a DevOps session or engine run: which check, engine fix, pointer or tool would have prevented the lost time, on request |
| `skills/framework/delegate/SKILL.md` | Any task that splits: which subagents is the model's judgement; the principles — lead keeps the thread, self-contained briefs, code in isolated worktrees, review the diff before merging |

Direct artifacts (no product-repo handoff): Makefiles, Dockerfiles, gitignore, `rules/*`, this `AGENTS.md`, `CORE.md`, `FEATURES.md`, `skills/framework/`. Everything else → `dispatch`.
