---
name: models
description: Recommend which hosted model to use for which kind of work in the IDE hosts (Cursor, Claude Code, …) from primary vendor sources and this workspace's own experience, and land the table in TREND.md §1.2. Workspace DevOps only. Do not load in product role-play or inject into Go workers.
---

# DUDES models

This skill is for **one world only**: the framework (workspace `AGENTS.md` + Go engine). It is not a role-play rule and not always-on. **Do not** add it to `recipes/<phase>.json` — the engine refuses it.

It answers: **which hosted model for which kind of work, today, in the IDE hosts the humans and the role-play use?** Local models as engine workers are measured by `skills/framework/bench`, not judged here. Landing still requires `skills/framework/maintain`.

## When to use

The human asks which model to use, or `make trend-check` lists §1.2 as `due:`, or a premise changed: a vendor releases or retires a model, prices or limits change, an IDE host changes what it offers, or work in this workspace showed a model failing a kind of work. Run once per request, then stop.

## Read first

`TREND.md` §1.2 (current table), §4 evidence discipline, Pillar 2 (sovereignty: a hosted model receives the code and the specs). `rules/DUDES.md` §4 for the kinds of work the roles do.

## Method

- **Rows are kinds of work, not roles:** framework DevOps (multi-file engine changes), spec writing, lead and market research, app implementation with tests, code review, quick edits, long autonomous runs. Add a row only for work that recurs.
- **Primary sources only** for facts: vendor model pages, pricing, model cards and release posts, the IDE host's own model docs. Record URL, checked date, and whether the page is dated or mutable. Vendor benchmarks are vendor-reported, never ranked as measured.
- **Rankings others keep are leads, not proof.** [openlm.ai Chatbot Arena](https://openlm.ai/chatbot-arena/) (400+ models, arena Elo judged by an LLM, plus coding, MMLU-Pro and other benchmark columns, licence and open-weight marks; run by OpenLM.ai, not LMSYS) shows which models to look at and how vendors compare on general tasks. Cite its date; never present an arena rank as a DUDES result.
- **Own experience counts as evidence** when it is specific: which session, which task, what failed or held (a commit, a TODO line). "Felt better" is not evidence.
- Separate what is documented (context, price, availability) from the judgement (why this model for this work).

## Reply

1. **Table** — work kind · primary · cheaper fallback · why (cite source ids) · cost tier.
2. **Sources** — id, URL, checked date, dated/mutable, the claim it supports.
3. **Changes** — rows that differ from the current §1.2 and why; unknowns left open.

## Where a run lands

When the human asks to keep it, load `skills/framework/maintain` and replace the §1.2 table and its source rows, each with `checked YYYY-MM-DD`. `docs/CURSOR_SETUP.md` only points at §1.2 — never a second copy. Git keeps the old table.

## Boundaries

| Do | Do not |
|----|--------|
| Recommend per kind of work, with sources | Change IDE settings, rules or role files |
| Name the data a hosted model receives | Treat a hosted model as a DUDES engine worker (that is local, `bench`) |
| Leave market movement to `skills/framework/trend` | Keep a second model catalogue — an instance points here |

Commits only when the human asked: `docs(trend): models <summary> [🔹AI-driven DevOps Engineer]`
