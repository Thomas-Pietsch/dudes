---
name: assess
description: Assess or reassess one external tool from primary evidence into TREND.md §5 and §6. Workspace DevOps only. Do not load in product role-play or inject into Go workers.
---

# Framework assess

This skill is for **one world only**: the framework (workspace `AGENTS.md` + Go engine). It is not a role-play rule and not always-on.

**Do not** run this during a product Feature/Fix flow. **Do not** add it to `recipes/<phase>.json`.

Writing the entry into `TREND.md` still requires `skills/framework/maintain`. This skill does not edit framework files during the run.

DUDES is a local-model ADE for the company product lifecycle. A **harness** is a worker. An **ADE desk** is another environment (a session shell or a task desk over harnesses). Do not score a desk as if it were a coder.

## When to use

The human named one tool, repo, or URL to assess, or requested reassessment of its existing verdict after a source, configuration or evidence change. A recommendation from `skills/framework/screen` is not permission to assess until the human picks. A catalog audit may identify stale entries without launching an assessment of every tool.

**One tool per run**, then stop. Do not start a second tool in the same run.

## Read first

`MISSION.md` §The North Star, `CORE.md` §1, §10, §12 and §15; `TREND.md` §4 evidence/axes, the relevant §5 entry, §6 patterns and §7 review record. Read the linked `FEATURES.md` rows before claiming an idea is already implemented.

Then the tool's primary docs at an identified release or commit. Record canonical identity, source and checked dates, revision, configuration and claim scope using §7. An undated mutable page is a dated observation, not a version pin. A recap is a discovery lead, not primary evidence.

## Establish evidence before scoring

- Separate documented capability, inspected implementation, measured behavior, fit judgment and forecast as §4 requires. A README is not proof of enforcement; inspect the relevant implementation/tests before claiming a guarantee. Do not label tests measured unless they ran.
- Check the configuration DUDES would use: default, configurable and unavoidable behavior are different. Assess a compatible bounded mode if documented, not only the full product's default session.
- State the strongest supported case for use, then the exact boundary or missing evidence against it. Database storage, optional cloud or optional subagents alone is not a failure. Unknown and N/A are allowed.
- For comparative superiority, use the §4 paired worker trial and preserve its evidence. Without that result, say unmeasured; the incumbent is not thereby a winner.
- Reapply the §4 erosion guard. Better results never waive frozen specs, test gates, phase isolation, local self-containment, Git evidence, sequential engine ownership or the human seal.

## Classify

Use the §5 category whose question the tool answers.

| If the tool is… | Category |
|-----------------|----------|
| A phase-machine library | A · Orchestration & Workflow |
| A coding harness for one bounded task | B · Coding Agents (Workers) |
| A spec / ticket / gate toolkit | C · Spec-Driven Development Toolkits |
| A way to shrink input or shape output | D · Context & Output Engineering |
| A model server | E · Model Infrastructure |
| Discipline delivered as prompts into a host | F · Method & Skill Packs |
| An ADE desk (session shell or task desk over other agents) | **G · ADE desks** |
| A method by which a system improves itself from its own runs (prompt, workflow, harness code, weights) | **H · Self-improvement (RSI)** |

Do not file a desk under B. Do not create a "misc" bucket. Verdict words are the §5 decision categories: Orchestrator, Worker Tool, Pattern / Reference, Infrastructure, Rejected.

## Spec vs how-to

The tool's docs mix two things. Split them. Do not paste either into this skill.

| You found | Where it goes |
|-----------|----------------|
| What is documented, what is verified, what remains unknown | The §5 entry with §7 evidence/review metadata. Do not promote an aspiration to a guarantee. |
| An idea DUDES could build into the phase machine | A §6 pattern, status `hold`, plus one architecture line. |
| Install steps, flags, screenshots, their tutorial | Their docs. Do not copy them; preserve primary URLs and revisions supporting the assessment. |
| How this DevOps run is performed | This skill, and only that. |

A how-to does not become a feature. A feature starts later, in `FEATURES.md`, after the human adopts a pattern. That is the only way into `ARCHITECTURE.md` and Go (`CORE.md` §15).

The §7 entry shape, the category questions, and any `shipped Fxx` stay in `TREND.md`. This skill points at them. It does not become a second copy.

## Reply

In this order. Use the §7 entry shape. Do not invent a second form.

1. **Index row** — tool, category, verdict, pattern names.
2. **Entry** — classifying intro, §7 evidence/review record, five axes, three pillars, and **DUDES Assessment**. Add one **Use** line: a bounded invoke, pattern, or nothing; state the configuration that preserves DUDES boundaries, or the evidenced reason none does.
3. **Patterns** — for each idea worth keeping:

| Pattern | New or extend | Theme | Status | Architecture |
|---------|---------------|-------|--------|----------------|
| name | extend an existing §6 row, or new | a §6 theme | `hold` unless it is already `shipped Fxx` | simplest substrate from `CORE.md` §11; worktree for changes, image for a pinned toolchain, whole-ADE isolation separately (§10). Or "do not build" |

Prefer extending a pattern's Source over a near-duplicate. An idea that puts the lifecycle in a chat, or that has agents hand work to each other by message, is an anti-pattern with status `hold`.

Default a new pattern to **hold**. Hold means tracked, not scheduled. Building it later is a `FEATURES.md` id, then `ARCHITECTURE.md`, after the human adopts it. This skill does not open that path and does not write Go.

## Reassessment loop

Use this same skill, not a new role, daemon or backlog.

1. **Select one affected decision.** Name the §7 revisit trigger and the old supporting source/configuration. A new version matters only when it changes a premise; before adoption, unsupported inherited claims also require review.
2. **Mark the scope.** The affected claim is stale (or unverified when support is missing). Do not invalidate unrelated evidence or a shipped feature. A stale recommendation cannot justify a new adoption; this review does not silently disable an existing worker.
3. **Check the delta.** Read changed primary evidence and any counterevidence. Repeat only the relevant checks; for performance or reliability claims use §4's paired trial. If the needed environment or goal is unavailable, explicitly retain stale/unverified status and name the missing check.
4. **Decide, without adopting.** Return retained/changed verdict, evidence and reason, strongest compatible use, unresolved risks, and the next concrete trigger/check. A negative result or an unchanged verdict is a valid outcome. Remove unsupported certainty rather than inventing evidence.
5. **Reconcile on authorized landing.** Update the §4 evidence, §5 index/entry, related §6 source implications and any dependent Harness watch recommendation together. Set current only for supported claims. Update the scoped review date; Git retains the old record. Stop after this tool.

Fresh evidence can reopen a tool verdict, not a Mission invariant. Changes to a shipped capability or runtime configuration still require human adoption and the full `CORE.md` §15 path. Do not turn a review trigger into scheduled engineering work.

## Where a run lands

The reply is the result. When the human asks to keep it, load `skills/framework/maintain` and:

- add or update the Index row and `###` entry under its category
- preserve the §7 evidence/review record and primary sources in `TREND.md`, never only in chat
- extend or add §6 patterns and their Index status
- add a category row (`TREND.md` §7 step 1) only when no category's question fits
- reconcile dependent current recommendations and update the scoped review date; do not append a history log

Do not rewrite unrelated essays. Do not edit `MISSION.md` or `CORE.md`.

Commits only when the human asked: `docs(trend): assess <tool> [🔹AI-driven DevOps Engineer]`
