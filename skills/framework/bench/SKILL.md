---
name: bench
description: Measure local models as engine workers — the same frozen tasks through the real engine (worker image, scope, gates) — and land the result in TREND.md §1.1 with a model recommendation for the engine's workers. Workspace DevOps only. Do not load in product role-play or inject into Go workers.
---

# DUDES bench

This skill is for **one world only**: the framework (workspace `AGENTS.md` + Go engine). It is not a role-play rule and not always-on. **Do not** add it to `recipes/<phase>.json` — the engine refuses it.

It answers one question with measurements: **how far do local models get as DUDES micro-agents, and which one should the engine's workers use?** Market reading is `skills/framework/trend`; hosted IDE models are `skills/framework/models`. Landing the result still requires `skills/framework/maintain`.

## When to use

The human asks for a model benchmark, or `make trend-check` lists a §1.1 row as `due:`, or a premise of a §1.1 row changed: a new model or quantization on the model host, a new Aider / engine / worker image version, new hardware, a changed context or output setting. Run once per request, then stop.

An instance level built from shipped work (a replay of real product slices) measures a new model or harness, or an engine mechanism on the runs where its case occurs. It is no regression test: three runs place a model without attributing a change. The engine's regression tests are its Go tests and `make bench`. Model and harness comparisons serve a stage in `ROADMAP.md` lane `E`; a comparison no stage needs is not run.

One series at a time: `make bench-models` refuses to start while another series holds its lock, and a stop ends the series and releases the lock. Check that a stop ended the series (`make roadmap-check` counts the bench runs since the review).

**Stop rule:** a series whose result repeats the last one changes no decision. Stop before the next and step back (`skills/framework/roadmap`). Do not chain series because the host is free. On 2026-10-14, six series on four models took almost 14 hours on the model host (832 run-minutes), which also runs the instance's CI. The two engine fixes and the regression guard all trace back to the first series. The others only confirmed where the models stop.

## The model runs first

Before a model enters any series, it answers the worker's request by itself on the model host. `make bench-smoke MODELS=<tag> OLLAMA=<host> LEVELS=<level> [THINK=on|none NUM_PREDICT=<n> NUM_CTX=<n> TEMPERATURE=<n>|default]` runs one real worker step through the harness on the level (Aider's whole-format request: its two canned assistant turns, the level's read-only files and task, temperature 0 unless `TEMPERATURE` says otherwise, `num_ctx` from `NUM_CTX` or the level) and prints the model's reaction in `react=`. It passes when the step is `edit:<n>` with n covering the files the task names, and not `cut` (the reply hit a token limit), `none` (no file listing: a canned turn echoed, placeholder lines, reasoning prose), `blocked` or `err`. A model that fails gets one setting changed (`THINK`, with `none` for a model without the thinking capability; `NUM_PREDICT`; `NUM_CTX`; `TEMPERATURE`, the vendor's sampling) and the smoke again, or is dropped; it never enters a series to find out, and a series runs only with the settings the smoke passed with. The smoke attributes a failure to the model and its settings, never to the engine. The level's gate runs once on what the model wrote and adds only a first outcome, which is no result: the smoke decides who runs, not who wins. Record it in `TREND.md` §1.1 beside the series it allows (the first, 2026-10-17, was a hand-built request).

## Read first

`TREND.md` §1.1 (current rows), §4 **Paired worker trial** (the protocol this skill executes) and the Evidence register; `CORE.md` §2.4 (the micro-agent shape) and §10 (worker image). The model host and its connection are configuration (`ollama.base_url`), never the laptop the engine runs on.

## Fix the comparison before running

Write these down before the first run; they do not change during it:

- **Tasks.** The bench ladder (F82): `make bench-models MODELS=… OLLAMA=…` runs levels L1–L3 (`bench/README.md`; L4 and up are external levels), three runs each, from a fresh fixture repository. Beyond the ladder, an instance's own level (`bench/README.md` *External levels*): pass its directory in `LEVELS`, run `MODELS=reference` first — a level that does not pass its reference check measures nothing — and label its rows as that instance's work.
- **Same start for every model.** Code at the base commit, a clean run state, no recorded gate on that commit, no files left by the previous model. A run that inherits another model's state is void.
- **Same configuration.** Harness and version, worker image tag, edit format, context (`num_ctx`), output cap (`num_predict`), thinking on/off, steps allowed. Change one variable at a time.
- **Candidates.** Models the human names, or what the model host serves and fits its memory. To find the newest open-weight models worth pulling, start from rankings others keep — [openlm.ai Chatbot Arena](https://openlm.ai/chatbot-arena/) marks open weights and licence and has a coding column — and filter by what fits the model host (most top open-weight entries are server-class MoE). A rank is why a model is tried, never a result. Exact tag and quantization (`ollama show`).

## Run and record

One fresh worker process per step, as the engine does. The ladder records each run as a git note (`git log --notes=dudes/bench`). Per model and task record: **outcome** (`pass` = the task's gate green; `blocked` = worker escalated; `fail` = steps exhausted), steps used, tokens in/out per step, minutes per step, edits the engine discarded as out of scope, edit-format errors, output-cap hits. A crash of the engine or harness is an engine finding — fix it under `skills/framework/maintain`, then rerun that model; do not score it.

Failures, blocks and timeouts are results. Keep them. A smaller prompt or a faster reply is not a win unless the outcome holds.

## Reply

1. **Setup** — tasks, base commit, the fixed configuration, host and hardware.
2. **Table** — one row per model: outcome, steps, tokens, minutes/step, discards, format errors, cap hits.
3. **Recommendation** — the model for the code worker (`aider.model`) and for the text phases (`ollama.model`), each with the evidence row it rests on, or **keep the incumbent: unmeasured** when no model passed. Name the next check that could change it.
4. **Findings** — engine or harness defects the run exposed, with what was fixed.

## Where a run lands

When the human asks to keep it, load `skills/framework/maintain` and:

- replace the affected §1.1 rows (one per model + configuration) with outcome, the immutable evidence reference (commit or note) and `checked YYYY-MM-DD` — Git keeps the old rows;
- add or update the Evidence register row for the measurement (level `measured`);
- change `dudes.example.json` model defaults only when the human adopts the recommendation (`CORE.md` §15).

Raw logs stay out of the repository; the evidence reference must let someone rerun the same comparison.

## Boundaries

| Do | Do not |
|----|--------|
| Run inference on the model host named in the configuration | Start or use a model on the engine's machine |
| Measure through the real engine — image, scope, gates | Rank models by vendor benchmarks, parameter count or prompt length |
| Recommend; record unmeasured honestly | Change worker defaults, contracts or the lifecycle without human adoption |

Commits only when the human asked: `docs(trend): bench <models> [🔹AI-driven DevOps Engineer]`
