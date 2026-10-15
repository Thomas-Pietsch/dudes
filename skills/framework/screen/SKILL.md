---
name: screen
description: Bounded live search for new mechanisms or material improvements to known tools, recommending assessment or reassessment. Workspace DevOps only. Do not load in product role-play or inject into Go workers.
---

# Framework screen

This skill is for **one world only**: the framework (workspace `AGENTS.md` + Go engine). It is not a role-play rule and not always-on.

**Do not** run this during a product Feature/Fix flow. **Do not** add it to `recipes/<phase>.json`.

This skill does not edit framework files. A tool enters `TREND.md` only when the human names it and `skills/framework/assess` files it. Market direction is `skills/framework/trend`, not this loop.

DUDES is a local-model ADE for the company product lifecycle. The lifecycle is Go, the spec, and Git. A harness is a worker. A desk is another environment.

## When to use

The human asked to search for tools worth assessing, to repeat this screen, or to keep looking until something new shows up.

Run the loop **once per such request**, then stop. The reply is the result.

## Read first

`MISSION.md` §The North Star, `CORE.md` §1 and §15, `TREND.md` §4 evidence/erosion guard, §5 Index and §6 Index, and the `FEATURES.md` rows those indexes name. Use them as the fit and redundancy baseline, not as proof every existing implementation is optimal. Do not copy the indexes or §7 entry shape into the reply.

## Goal

Default budget: **up to ten tools**, then a recommendation. Honor a different breadth when the human requests it. A cycle is one named tool/configuration with a primary URL, checked date, revision when available, supported claim and comparison to the current mechanism. An empty search is not a tool.

The result is that report, not a catalog entry or a `TODO.md` line. The goal is decision value, not novelty or a filled quota.

## Loop

Work within the budget. Stop at the budget, or earlier when two distinct search angles produce no new decision-relevant evidence; report coverage and the reason. Never pad a requested count with invented tools or repeated evidence.

1. **Angle.** Name the problem or improvement sought. Change angle when a find only repeats known evidence. A concrete lead may justify another tool in the same angle; state what question the deeper search resolves.
2. **Search.** Start where others already catalogue: [agent-orchestrators.com](https://agent-orchestrators.com/) — a curated, weekly updated directory of open-source agent orchestrators by category (terminal, desktop/web, swarms, loop runners, task runners, infrastructure, assistants), with stars, growth and status, an `/api/` and `/rss.xml`, source list `andyrewlee/awesome-agent-orchestrators`. Filter it by the angle and against the §5 Index; growth and status say where to look, never what a tool does. Then read the tool's own repository at a revision — that is the evidence. Record it and distinguish documented from measured claims (§4). Unknown local performance stays unknown.
3. **Compare.** Name the §5 row, §6 pattern or shipped F-id it overlaps. Compare mechanism, enforceable boundary and possible benefit. The same job is not the same mechanism; a different implementation may materially improve task success, cost, latency, memory, maintenance or evidence quality.
4. **Record and continue.** Fold a find only when no supported decision-relevant difference remains. A vendor improvement claim is a reason to test, not a proven win. Record a candidate's strongest supported compatible use and the unresolved question. Do not run a full assessment inside the screen.

A catalog tool may qualify for **reassessment** when its version, configuration or evidence changes a premise. Skip unchanged rediscoveries. Do not count different URLs for the same evidence twice. Existing negative verdicts do not replace checking a newly documented compatible mode.

## Reply

At the stop, review the comparisons. Recommend assessment for a new mechanism or a plausible material improvement; recommend reassessment for changed evidence about a known tool. Prioritize relevance and expected value over novelty. One tool can be enough; zero is allowed.

In this order:

1. **Table** — one row per cycle:

| Cycle | Angle | Tool / revision | Source / checked date | Existing mechanism / evidence delta | Assess or reassess? |
|-------|-------|-----------------|-----------------------|-------------------------------------|--------------------|

The comparison names an existing index row, pattern or F-id, or **new mechanism**. State the evidence level and why the difference matters; do not infer a measured benefit from marketing.

2. **Recommend** — selected rows in order, with the benefit hypothesis, Mission boundary and question/check the assessment must answer. For rejected leads the table's reason is enough. State search coverage and any early-stop evidence gap.

The human decides. Naming a tool in this reply is not permission to assess it.

## Where a run lands

| Piece | Where |
|-------|--------|
| How to search, compare, change angle and stop within budget | This skill |
| The evidence-backed screen and recommendation | The reply; not sufficient evidence for adoption |
| One catalog entry, after the human names that tool | `skills/framework/assess`, then `TREND.md` §5 and §6 |
| Work on an adopted pattern | `TODO.md`, only when the human pulls it (`CORE.md` §15) |

When the human names one tool from the report, stop this skill and load `skills/framework/assess` for that tool only.

## Boundaries

| Do in the reply | Do not do |
|-----------------|-----------|
| Bounded comparisons, primary evidence, a short recommendation | Edit `TREND.md`, `FEATURES.md`, or `TODO.md` |
| Change angle on repetition; deepen a concrete lead; recommend reassessment | Assess inside the loop, or assess every candidate unasked |

Commits only when the human asked. An assessment commit stays with `skills/framework/assess`.
