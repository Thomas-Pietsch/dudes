---
name: trend
description: Live market read across local models, harnesses, ADEs, and recursive self-improvement, then a 3-month and 1-year outlook and a now/next/later/never DUDES action table by tool. Workspace DevOps only. Do not load in product role-play or inject into Go workers.
---

# Framework trend

This skill is for **one world only**: the framework (workspace `AGENTS.md` + Go engine). It is not a role-play rule and not always-on.

**Do not** run this during a product Feature/Fix flow. **Do not** add it to `recipes/<phase>.json`.

Landing any recommendation into `TREND.md` or other framework files still requires `skills/framework/maintain`. This skill does not edit those files.

Names, not a new pillar: a **local model** is weights plus server; a **harness** is the worker around one model; an **ADE** is the environment where agentic development happens. DUDES is a local-model ADE for the company product lifecycle. The lifecycle is Go, the spec, and Git. The worker session ends with the phase.

## When to use

The human asked where the market is going for local models, harnesses, ADEs, or recursive self-improvement, or what that means for the DUDES path (`ROADMAP.md`). Which hosted model to use today is `skills/framework/models`, not a second pass.

Run **once per such request**, then stop. Present the briefing in the reply. Do not open a backlog file. Do not evaluate a single named tool here — that is `skills/framework/assess`.

## Phase 1 — Live research

Read `TREND.md` §4 evidence discipline and the current Harness watch first. Search primary public sources. Every item needs canonical identity, source date, checked date, URL, revision when available, evidence level and a one-line supported claim. Distinguish vendor claims from independent measurements. Undated items may be context, explicitly unverified as recent developments.

**Window:** the last 4 weeks from today's date, unless the human names another window. State the window in the reply.

Four lists, in this order, before any outlook. Default to **at most 12 material developments total**, unless the human requests a broader survey; no per-lens minimum. Deduplicate the underlying development, not just its URL. If a lens is thin, report the gap. A relevant mature baseline can be cited separately as outside the window. Coverage and decision value matter more than filling a quota.

| List | What belongs |
|------|----------------|
| **Local model** | Open weights, model shape (dense or mixture-of-experts, active parameters — the shape column of `TREND.md` §1), context, tool calling, quantization, tokens/s, servers (Ollama, llama.cpp, vLLM, MLX, LM Studio) on a normal PC. |
| **Harness** | The loop around one frozen model: tools, context policy, checks, eval protocol. |
| **ADE** | Environments that host agents: session desks, task desks, spec-and-ticket workspaces, hosted agent platforms. Include small and open ones, not only the largest vendors. A desk's hosted model may be named here as a lead. |
| **Self-improvement (RSI)** | The meta level over the three above: methods and frameworks in which a system improves itself from its own evaluations. For each, record the level it changes (prompt or workflow, harness code, weights, the optimizer itself), who judges (the same model or an independent check), what signal it learns from (verifiable tests or a model judge), its cost, and whether it runs on local open weights. |

Hosted (cloud) models are not a list of their own; their movement is one paragraph in each outlook cut, drawn from the ADE leads. Which model to use is judged, when the human asks, by `skills/framework/models` into `TREND.md` §1.2; local engine workers are measured by `skills/framework/bench` into §1.1. Subscriptions, keys and the machines models run on belong to the instance, not to this skill.

## Phase 2 — Market outlook

Market and technology only. No DUDES actions in this phase.

Write the **3-month** cut first, then the **1-year** cut. Inside each cut, in this order: local model, cloud model, harness, ADE, self-improvement. Each names what would falsify it. Do not repeat the 1-year text inside the 3-month cut.

If the evidence does not support a forecast, say insufficient evidence. Do not convert advertised context, active parameter count or prompt length into a local performance ranking.

## Phase 3 — DUDES actions

Read `MISSION.md` §The North Star, `CORE.md` §1, and `TREND.md` pillars, §5 verdicts, and §6.

DUDES is already the local lifecycle ADE. Map the market read onto it. Do not propose becoming a session desk.

Four buckets, in this order. Recommendations only.

| Bucket | Means |
|--------|--------|
| **Now** | This quarter. No doctrine change. |
| **Next** | The following quarter. Starts only if a Now check passed. |
| **Later** | The 1-year horizon. Watch or design. Do not build yet. |
| **Never** | A mechanism that violates the Mission/Core boundaries, not a permanent tool-name blacklist. Duplication is no reason to adopt, but a measurable improvement can justify reassessment. |

Every item states the pillar, what it must not violate (sequential execution, human outcome seal, spec and test discipline, self-containment, resistance to context drift), and **by tool**: the existing tool or `TREND.md` pattern. If no tool fits, say **engine** and why. Prefer by tool.

Map each self-improvement item onto the world it could improve (`CORE.md` §2.4 *Who splits the work*): the framework (the DevOps loop of retro, bench and trend, sealed by the human), the role-play (rules and skills — prompt level, hosted models), or the micro-agent world (the engine's worker messages and contracts, local open weights, gates and hidden bench tests as the signal). Prefer the loop DUDES already has: the bench ladder is the fitness function, a change lands only through the human seal, and no loop may change its own tests, gates, frozen spec or seal (`TREND.md` §4 erosion guard).

A single product's full catalog entry is `skills/framework/assess`, not this phase.

For every action, cite its supporting evidence and the trigger/next check that could change it. Identify any dependent stale verdict, propose one-tool reassessment, and do not silently retain incompatible current recommendations. Keep an incumbent without claiming it won a comparison; comparative changes need §4's real paired trial, not `dudes bench`. Recommend at most 1–3 candidates.

## Where a run lands

| Piece | Where |
|-------|--------|
| Method, four lists, market outlook, buckets | this skill |
| DUDES action table, supporting sources/evidence levels, review date and revisit conditions | `TREND.md` heading **Harness watch** (references to §4 records are enough; no duplicate source table) |
| Why local, KMU, and auditability matter | `MISSION.md` §Strategic Operating Context. Do not add a pillar. |

Do not create `LEAN.md` or a second roadmap: the path itself — lanes, stages, exit checks — is `ROADMAP.md` (`skills/framework/roadmap`), and these actions are what the market means for it. A market move that changes a stage or its order goes to the next roadmap review, not into Harness watch. Harness watch may name an F-id or a verdict. That link stays in `TREND.md`. This skill does not become a second copy, and it does not sit on the path from `CORE.md` to Go.

When the human asks to keep a briefing, load `skills/framework/maintain` and **replace** Harness watch with the action table and its evidence/review record. Full search lists may remain in chat, but every landed claim must have durable support. Preserve the premise and falsifier of any outlook used to justify an action. Reconcile affected living recommendations; a tool reassessment is the separate assess procedure, never an automatic runtime change. Do not edit `MISSION.md`, `CORE.md`, the §1 model table or shipped §6 status as a side effect of a briefing.

## Boundaries

| Do in the reply | Do not do unless the human asks to land it |
|-----------------|-----------------------------------------------|
| Dated lists, market outlook, now / next / later / never, reassessment recommendations | Harness watch plus its supporting evidence and review conditions |

Commits only when the human asked: `docs(trend): <summary> [🔹AI-driven DevOps Engineer]`
