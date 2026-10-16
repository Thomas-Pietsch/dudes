# DUDES — Technology & Trend Landscape

> **Purpose:** This document contains the time-sensitive technology, market, regulatory, model, framework and tool knowledge that informs DUDES decisions. It is a strategic input to the framework — **not its authority**.
>
> The Mission defines **why DUDES exists**. This document records **what is currently possible, relevant, promising, or worth evaluating**.
>
> **Last review:** 2026-10-09 — evidence and decision baseline, not a blanket revalidation of the archive. Checked claims and source revisions are in §4 Evidence register. Unlisted claims remain **unverified**. Living surfaces: pillars, §1 models, §3 calendar, §4 evidence, §5 Index, §6 pattern status, Harness watch. Older §5 essays are research context, not current proof.

---

## How to Use This Document

Use TREND.md when:

- evaluating new AI models;
- evaluating open-source agent frameworks;
- evaluating coding agents and developer tools;
- deciding whether an external tool should become part of DUDES;
- identifying technologies that can improve the framework;
- documenting current market, sovereignty and regulatory conditions;
- recording architectural ideas derived from external tools.

Because this information changes over time, entries should be updated, replaced, or retired when the landscape changes.

**Jump:** pillars `#1`–`#3` · eval axes `#4` · tool catalogue `#5` · pattern index (F-status) `#6` · template `#7` · [Harness watch](#harness-watch) · the path and its horizon: `ROADMAP.md`.

A trend briefing is temporary. Its DUDES action table and supporting evidence land in **Harness watch**, replacing that block; chat is never the only source record. It does not add a pillar, a model freeze, or a feature. A briefing names at most 1–3 candidates. `skills/framework/screen` searches for new mechanisms or material improvements to existing ones, recommendation only. Neither writes §5. The human selects a tool; `skills/framework/assess` assesses or reassesses it into §5 and §6. Work starts only after human adoption through `CORE.md` §15.

**Reassessment is not adoption.** A changed source, worker configuration, law, or local result can reopen a verdict, never a Mission invariant. The review record is §7; the bounded reassessment procedure is in `skills/framework/assess`. Prior records remain in Git history, not an appended research log or another backlog.

Market names, not doctrine. `MISSION.md` stays the authority.

| Name | Means |
|------|--------|
| **Local model** | Weights and the server that fit a normal PC. |
| **Harness** | The loop, tools, and checks around one model. In DUDES this is a worker. |
| **ADE** | The environment where agentic development happens. |

DUDES is a **local-model ADE for the company product lifecycle**. The lifecycle is the Go phase machine, the spec, and Git. The worker’s session ends with the phase. A desk whose unit of work is a living chat is a different ADE.

**A technology or regulation can influence DUDES without becoming part of DUDES.**

The key question is not:

> “Is this framework better than DUDES?”

It is:

> “What useful capability does this demonstrate, and can DUDES adopt that capability without violating its Mission?”

---

# Strategic Pillars (Mission-aligned)

Three concurrent pressures shape how DUDES should be built and run **today**. All three matter for Mission fit (local, autonomous, auditable delivery with human outcome control):

| # | Pillar | Core question |
|---|--------|---------------|
| **1** | Consumer-box local coder (~32GB machine) | Can a scoped coding step — and the full product cycle around it — run on a normal PC? |
| **2** | KMU / sovereignty | Can teams keep IP and data under their control while still using AI? |
| **3** | DSGVO + EU AI Act | Can development stay **auditable**, documented, and human-controlled under EU rules? |

```text
Normal PC (~32GB)
  → local open-weight coder that fits the box (often MoE, few-B active)
  → DUDES sequential process (Go-orchestrated workers)
  → local sovereignty (KMU)
  → auditable artifacts + human seal (DSGVO / EU AI Act readiness)
```

> **Important:** Pillars are strategic context. They must not freeze model size, vendor, or a specific statute into framework doctrine. Detail below; authority remains `MISSION.md` → `CORE.md`.

---

# 1. Consumer-box local coder (~32GB machine)

The target is a **normal PC**, not a parameter count.

The working hypothesis: a model that **fits the available memory** and completes a **scoped coding step** (edit, test, exit) can serve a DUDES worker. MoE models with a few billion active parameters are candidates, not proven winners. Agentic repair results are relevant, but a published SWE-bench score does not establish DUDES task success on a quantized local model.

### Candidate baseline (checked claims: §4 M1–M3)

| Model | Shape | Memory / context evidence | Why it is here |
|---|---|---|---|
| **Qwen3-Coder-30B-A3B-Instruct** | 30.5B MoE / 3.3B active | Native 262,144-token capability; usable local context and speed **unmeasured here** | Candidate for agentic local steps; pin quantization and server for a trial |
| **Devstral-Small-2505** | 24B dense | Vendor describes RTX 4090 / 32GB Mac deployment; exact trial configuration still required | Vendor reports 46.8% SWE-bench Verified under OpenHands, not under DUDES |
| **Qwen3.6 27B / GLM-4.7-flash class** | Unverified older watch labels | No checked model card or configuration in this baseline | Leads only; verify exact identity before comparison |
| **Qwen2.5-Coder 32B** | Prior-generation reference | No current memory or speed measurement in this baseline | Historical comparator, not a ranked recommendation |

**Retired as a peer:** CodeLlama-class 70B dense. Too heavy for the DUDES machine; not a current coding SOTA.

**Memory means the whole workload:** weights + KV cache + runtime buffers + OS/toolchain headroom. System RAM, VRAM and unified memory are not interchangeable. Advertised maximum context is not allocated context or a latency guarantee. Record offload, quantization, context and peak memory in the paired trial (§4); do not infer 256K feasibility from a quantized download size.

### 1.1 Engine workers — measured (`skills/framework/bench`)

How far local models get as DUDES micro-agents, measured through the real engine (worker image, scope, gates, gate feedback) — the answer for framework users. Rankings others keep (openlm.ai arena, vendor benchmarks) only choose what to try. The bench ladder (F82, `make bench-models`) is the measurement; instance trials, labelled as such, show what the ladder does not cover yet.

**Bench ladder, 2026-10-10** (`bench/`, §4 B2): levels L1 (one package from one criterion) and L2 (a package defined by a TA algorithm, read through the item pack), 3 runs each from a fresh fixture repository, up to 3 implementation steps; `pass` needs the gate **and** the level's hidden test the worker never sees. Engine 2026-10-10 / 2026-10-10, Aider 0.86.2 in the worker image `-p1`, whole edit format, `num_ctx` 32768, `num_predict` 12288, thinking off, gate in `golang:1.26-bookworm`; Ollama on the model host.

| Model (Q4_K_M) | L1 | L2 | L3 | Headline | Steps | Min / run |
|---|---|---|---|---|---|---|
| `gemma4:26b` | 3/3 pass | 3/3 pass | 3/3 pass | **L3** | 2 (L1), 1 (L2, L3) | ~1 |
| `qwen3.8:27b` | 0/3 — 3 × weak | 2/3 pass, 1 weak | 3/3 pass | L3 | 1 (L1, L3), 1–3 (L2) | 1–7 |

`qwen3.8:27b`'s weak runs: L1 kept non-ASCII letters (`crème-brûlée`) against the PRD's "lowercase ASCII only" — the same answer three times; L2 run 1 accepted `1.2.3-rc.` against ALG-01.1. Its own tests never probed either case.

L3 (2026-10-12, engine of 2026-10-12): several constraints across two packages — a new `release` on top of the existing `semver`, which must keep its behaviour. On the engine before (2026-10-12) `gemma4:26b` passed 2/3 and once answered `BLOCKED: provide the full content of semver/semver.go`: the worker never saw the code in its scope. With that code handed over (smallest first, 48 KB), both models pass 3/3 in one step each.

**Recommendation:** `gemma4:26b` for `aider.model`, with `num_ctx` 65536 or more once the worker gets its scope's code (slice 4b of the instance trial below) — every ladder run passed, at about a minute a run; `qwen3.8:27b` gets as far but meets the spec less often. Next check: L4 (an item in slices, red first) and the trial's configuration slice again, now that the worker sees the code it extends.

**Instance trial, 2026-10-09 — slice 1 of the first instance's product item** (a configuration loader with a generated example file; criteria AC-8 and AC-10): gate `go test -skip` over the later slices, up to 3 steps with gate feedback. Aider 0.86.2 in the worker image, whole edit format, `num_ctx` 32768 (`qwen3.5-coder` 131072), `num_predict` 12288, thinking off; Ollama 0.33.3 on the model host. Same start for every model: code, run state and gate note reset. §4 B1.

Round 5 — the current measurement: base with one test file per acceptance criterion, engine at 2026-10-09, task hint naming the build pitfalls (module path from `tests/`, `.exe` on Windows).

| Model (Ollama tag, Q4_K_M unless noted) | Outcome | Furthest point | Tokens in / out per step | Min / step |
|---|---|---|---|---|
| `qwen3.8:27b` | **pass, 1 step** | Config loader (300 lines), 4 package tests, AC-8/AC-10 build and run the product binary; one AC-10 subcase skips when the example file is missing | 11k / 5.7k | ~8.5 |
| `gemma4:26b` | **pass, 2 steps** | Step 1 built the binary from `tests/` as a relative path; the gate output fixed it | 11–15k / 1.5–4.2k | ~2 |

Rounds 1–4 (all tests in one file) failed every model — mostly on defects of the task and the engine, which the trial then fixed: `go run` reports exit 3 as 1 (`qwen3.8:27b`, `gemma4:26b` were otherwise green), the `.exe` trap on Windows, a worker breaking another slice's test in the shared file while the gate handed back no file. Model-side failures that stand: `gemma4:31b` (two compile errors after 2 steps, ~16 min/step), `mdq100/qwen3.5-coder:35b` (never compiled in 6 attempts: invented `yaml.SyntaxError`, imported a Go internal package from `tests/`; diff format ran away past 32k tokens), `deepseek-coder-v2:16b` (`BLOCKED:` — misread the chat files as read-only), `mistral-small` 24B (stopped after 7 tokens — chat template, inconclusive).

*Instance-trial reading (2026-10-09, superseded by the ladder):* `qwen3.8:27b` got furthest on slice 1 — first green in one step, `gemma4:26b` in two at a quarter of the time per step. Green, not verified: see slices 2 and 3 below.

**Slices 2 and 3, `qwen3.8:27b`, from the slice-1 result** (2026-10-10, engine of 2026-10-10; the constraint and read-only fixes landed between runs):

| Slice | Outcome | What happened | Steps | Min |
|---|---|---|---|---|
| 2 — CI invariants (AC-6/7/9) | **weak** — gate green, criterion not fully tested | Scanners and the RLS lint check the real tree; each negative control (a violating fixture must fail the check) `t.Skip`s on a wrong path (`tests/testdata` from inside `tests/`), so it never ran | 2 | ~21 |
| 3 — router, route matrix (AC-5), run 2 | fail | Imported the product's internal API package from `tests/` (the constraint never reached the worker — fixed); later steps duplicated declarations across files | 4 | ~99 (one step lost to a standby freeze) |
| 3 — run 3 | **weak** — gate green, AC-5 not tested | Step 1 redeclared slice 1's helper (outside its scope — now handed over read-only); step 3 went green by **deleting the route-matrix check**: the test checks `/health` and the 404 envelope | 3 | ~38 |

**The finding:** a green gate is the worker's own judgement when the worker writes the test it must pass. Slice 1 holds up best, but not in full either: AC-10's schema half ("the committed example file equals what the struct generates") is reduced to "it loads", with the comment *in a real scenario, we'd compare*. No slice is verified in full. The bench ladder therefore scores against a hidden test per level (`weak` = green by weakening); for items, the acceptance test is written red first and stays outside the implementing worker's scope (F65).

**Red first, `gemma4:26b`, 2026-10-10** (`expect: fail` tasks, gate in `golang:1.26-bookworm`, engine of 2026-10-10): the configuration slice's red task **passed in 2 steps** — real tests (exit code 3, key names, the printed example config compared with the committed file, no skips). But one subtest reads the committed example file relative to `tests/`, where it can never exist: red for the wrong reason, and the implementation task may not touch it. The slice's code task failed in 3 steps — `declared and not used: cfg` never fixed, the last step returned 133 tokens and no change — and could not have passed anyway. Red first trades *green by weakening* for *a wrong test blocks the code*: someone has to read the red test before the code is written.

**Red first with review, `gemma4:26b`, 2026-10-12** (`approval: after`, `dudes redo`; the reviews were done by the DevOps session standing in for the human): slice 1's red tests needed two redos (the wrong example-file path; AC-8 checked only the file, not line and key), then the code task made them pass in 5 steps **with the tests unchanged** — the first slice green without weakening. The next slice exposed what review missed: AC-8's line check matched the digit anywhere in the output, including a random temp file name — flaky; the third redo made it `<file>:<line>`. Slice 2's tests reached review green in 3 steps but each real-tree scan only logged a walk error, and AC-6 walked the migrations instead of the code: sent back. Against the strict AC-8, the code task oscillated over 5 steps — fixed the line, broke the generated-config match — and its last step returned 95 tokens and no change, as in the first run. **Reading:** red first plus review removes weakening and catches wrong tests; the review is itself a check that can be weak; and `gemma4:26b`, the ladder's best on L1–L2, does not hold several constraints of one task at once — the ladder's L3 should measure that.

**Slices 1–3 green against reviewed tests, `gemma4:26b`, 2026-10-12** (engine of 2026-10-12 and later: the worker sees the code in its scope): slice 1's code 1 step (before: oscillation over 5); slice 2 green in 1 step, then one redo — AC-7's three heuristics became one per-table check, which **found a real gap in the migration** (a table without row-level security) that the code then fixed; the router's red tests one redo (a `--port` flag against the TA's single `--config` decision), its code 4 steps (two lost to unused imports). Every acceptance test unchanged after its approval. Reviews: the DevOps session standing in for the human — 7 redos over 3 slices, each a concrete defect a green gate had let through.

**Slice 4 (database, AC-2/AC-4) green, 2026-10-12**, gate in WSL (`scripts/linux.sh`; testcontainers on the host's Docker), formatter on (`goimports`). The red tests took five reviews — a server started with `cmd.Run()` (hung), a missing GRANT, a role name replacing the URL scheme, an id of one kind compared with an id of another, an inline config the product refused with its output hidden. Two prerequisite defects were DevOps's: dependencies fetched by module path left `go.sum` without the sub-packages' sums, and the meaning of a status field needed the review to state it. The last red-test fix was found after the code task ran — its code had been right since round 3; the engine now accepts a redone red test that passes (the review decides). Item status from engine evidence shows **`AC 8/11`**.

**Slice 4b (AC-3 with a build-tagged fixture) green, 2026-10-12**: two reviews of the red test (matrix entries are `METHOD PATH`; a bounded wait too short), then the code task lost ten steps to a full context — with the scope's code handed over the prompt filled `num_ctx` 32768 and Aider stopped after 29 tokens, which the engine read as "no changes" (now an error naming the cause). With `num_ctx` 65536 it passed in **one step**. The item: **`AC 9/11`**.

**Slice 5a (a dev command and the web shell, AC-1) failed, 2026-10-13 — the trial ends here** (it benchmarks the framework; AC-11, the last criterion, is not run). Gate in the instance's `ci` image (`worker.sh --docker`: Go, Node, Playwright and the host's Docker for the database the dev command starts). The red web task was red after one redo, the read-only tests stayed unchanged. The web code task failed five steps (36 min, ~5 min gate each): an undefined router helper, a reply cut off at the 12k reply cap (`num_predict` — whole-file edits of three files; the engine's error named only `num_ctx`, now both), then two passes that broke more than they fixed — a duplicate handler, an embed of a missing build directory, a Go file without its `package` line, an import of a file that does not exist. Each failed pass stays committed and the next builds on it, so slices 1–4b's tests failed with it (build errors, not regressions in their logic). **Reading:** the first slice spanning two toolchains (Go and a Vite/React frontend) and a process the test starts is past `gemma4:26b` in this harness — one step per slice held while a slice stayed in one language and one package. What L4 must contain: a cross-toolchain slice, and a measure of whether a failed pass leaves the tree worse than it found it.

**Replay of the trial, 2026-10-14 — the first L4 level** (§4 B3): an instance level (in the instance's dev repo, at a commit of its own) replays slices 1–4b of the trial from its start commit — five implementation tasks, the human-approved tests of AC-2/3/4/5/8/10 given read-only, the trial's green state as reference (reference check: base red, suite green, hidden test green, 9 acceptance tests), its AC-6/7/9 scanners as the hidden test. Engine 2026-10-14, gate in the instance's `ci` image with the host's Docker, `num_ctx` 65536, 3 steps per task, 3 runs per model.

| Model (Q4_K_M) | Outcome | Furthest point | `ac` best / final | Min / run |
|---|---|---|---|---|
| `gemma4:26b` | 0/3 | slice 1 never green (tasks 1/6) | none | 6–12 |
| `qwen3.8:27b` | 0/3 | slice 1 green 3/3 (2–3 steps), slice 2 failed 3/3 (tasks 2/6) | 4/4 · 2/0 · 4/0 | 36–39 |
| `qwen3.8:27b`, again (engine of 2026-10-14) | 0/3 | tasks 5/6 · 4/6 · 2/6 — run 1 green through slice 4, then 4b's first pass broke the build and two more steps did not repair it | 8/3 · 6/5 · 2/2 | 42–111 |
| `qwen3.8:27b`, regression guard (engine of 2026-10-14, F86) | 0/3 | tasks 1/6 · 2/6 · 4/6; the guard undid 6 failed passes in 2 runs (each broke AC-8/10, in slice 4 also AC-5) | none · 4/4 · 6/6 | 22–65 |
| `gemma4:31b` (engine of 2026-10-14) | 0/3 | slice 1 never green (tasks 1/6): AC-8's line number, AC-10's byte-exact example file | none | 31–40 |
| `mdq100/qwen3.5-coder:35b` (engine of 2026-10-14) | 0/3 — 1 × blocked | slice 1 never green; one run escalated at once, reading "refuse an inline database URL in production" as a conflict with needing a database; one hit the reply cap twice | none | 1–16 |
| `gemma4:26b`, edit format `diff` (engine of 2026-10-14) | 0/6, two series | slice 1 never green (tasks 1/6): search/replace blocks carried nearly the whole file | none | 17–38 |
| `qwen3.8:27b`, edit format `diff` (engine of 2026-10-14) | 0/3 | tasks 2/6 · 4/6 · 1/6, the spread it showed with `whole` | 4/4 · 6/6 · none | 19–64 |

- **Slice 1 (AC-8/10)** fails on details, not structure: the committed example file must equal what the struct generates byte for byte — the gate output shows both, `gemma4:26b` never aligned them — and the refusal must name file, line and key. One `gemma4:26b` run hit the reply cap twice rewriting whole files.
- **Slice 2 (AC-6/7/9)**: AC-9 scans web code that only slice 5 brings. `qwen3.8:27b` skipped the scan, and the engine rightly refuses a skipped acceptance test — but its feedback said "fail on a missing precondition instead", the wrong advice when there is nothing to scan (the reference scans nothing, says so, and still runs the violating fixture). The refusal now states only the gate's rule, a fact; what to write instead is the worker's judgement (`CORE.md` §2.4).
- **Failed passes compound, now measured:** in two of three `qwen3.8:27b` runs slice 2's failed passes left `tests/` not compiling, so slice 1's green tests were red at the end (`ac` final 0). Slice 5a showed the same in the trial.
- **Variance is large:** the second `qwen3.8:27b` series got from slice 2 to slice 4b — but the engine change between the series (a refused skip now reaches the worker, 2026-10-14) never fired: no run skipped. Three runs per model place a model, they do not attribute a change; an engine change is compared on the runs where its case occurs.
- An engine defect surfaced and was fixed before the series: a reply sentence with `*must*` in it became a file name the engine could not discard on Windows (2026-10-14).

**Reading:** the replay separates what the ladder does not — on real product slices `qwen3.8:27b` gets further than `gemma4:26b` (slice 1 3/3 against 0/3), at about three times the time per run. No model passes; the incumbent recommendation stands, unmeasured for product work. **Regression guard (F86), measured:** without it a failed pass left the tree worse in 3 of 6 runs (`ac` final below best); with it, in every run where it fired (2 of 3, 6 undone passes) the run ended with every test green that had been green (`ac` final = best), and in run 3 slice 2 went green on the step after an undo. It keeps what was reached; it does not get a model further — no run passes. **Size is not the lever here:** the two larger models on the model host (`gemma4:31b` dense, `qwen3.5-coder:35b`) stop at slice 1 like `gemma4:26b`; only `qwen3.8:27b` gets past it (6 of 9 runs). Slice 1 fails on details the gate output shows in full — the generated example file beside the committed one, the missing line number — and the worker does not carry them over; the `diff` edit format changed that for neither model. The two `diff` series were one command started twice, 12 s apart, so they shared the model host and their minutes per run are inflated; the bench script now refuses a second series on the same machine. **Next check:** none for these models. Tuning the worker message until a model carries them over is not on the path: the worker gets the gate's facts, not advice, and DUDES improves as models do (`CORE.md` §2.4). The replay is no regression test: no model passes it, and three runs place a model without attributing a change. It is rerun for a new model or harness. Open: a harness comparison on slice 1, since Harness watch's revisit trigger has fired (`TODO.md`). The dev stage is measured on real items with review in the loop (`ROADMAP.md` `E1`), where the trial got slices 1–4b green.

**`E1`, the first real item, 2026-10-14.** A backup's verify-before-rotate in one Go package (a slice of an item: three criteria, AC-2/5/7), worker `qwen3.8:27b`, gate in the instance's ci image on the backup and API packages. The task texts named facts only: the ACs, the files, the test names, the signatures that stay. The spec came through the item pack. A first contract that restated a TA section and prescribed the tests was trimmed before the first step. The worker needed 7 steps, about 31 minutes: the red tests in 2 (one redo), the code in 5 (one red gate, three redos). A role-play Developer on Sonnet reviewed six times, not the DevOps session, for about 218k tokens; the first review cost 128k, a resumed one 6–42k. Every redo was a real defect the gates let through:
- tests that counted files by extension and missed "keep only the newest failed copy";
- a row-count check that read the live counts only before the copy was made;
- swallowed rename and chmod errors;
- a placeholder `if` against the instance's rules;
- a relabel of failed copies nobody specified.

It landed on the instance's `development` branch, plus a gofmt commit by DevOps; the formatter was configured only after the test task ran. Its CI run was not observed here. It does not count for the exit check: DevOps prepared the run (418k tokens on Opus) and landed it by hand. Those hand steps are the bridge's input (`TODO.md`).

**Harness trial on replay slice 1, 2026-10-16 (Pi against Aider, §4 paired trial).** The human's go after the worlds split (`ROADMAP.md`). A level that holds only slice 1 of the replay (its reference check green, base red), `gemma4:26b` through a derived tag with a 64k context and `num_predict` 12288 set on the server, temperature 0, thinking off, `STEPS=3`, engine `ece4e5c`, the gate in the instance's ci image. Pi 1.3.0 ran in its own worker image (the ci image plus Pi, no Docker socket, offline, tools `read`, `edit`, `write`, `grep`, `find`, `ls`, no bash, compaction off), with the same handover as Aider. Stop rules set before the runs: Pi stops after two failed runs, on a run with no tool call, or at once on a path out of scope.

| Arm | Runs | Pass | Where it stopped | Minutes |
|---|---|---|---|---|
| Aider 0.86.2 (control) | 3 | 1/3 (`ac` 2/2) | runs 1 and 3: AC-8 and AC-10 red after 3 steps | 7–10 |
| Pi 1.3.0 | 1 (stopped) | 0/1 | step 3 edited the read-only tests (`out of scope, discarded`) | 27 |

- **Pi's run degenerated, it did not reason.** Step 1: 147 tool calls, 134 of them `write` on the same `main.go`, ending at the token limit. Step 3: 120 `edit` calls, 117 failing on the read-only `tests/ph001ac10_test.go`, then the limit again. No thinking blocks reached the events, so thinking was off. The model cannot drive a long tool loop: more tool freedom gave it more room to repeat.
- **Aider's control passed once, where the same model failed slice 1 in 9 of 9 earlier runs.** The level differs (only slice 1; its task text lost the sentence naming the slice's command) and the context now comes from the server. One pass in three does not attribute the change to either.

**Reading:** the harness is not the lever for slice 1 with this model; the trial's rule says so (Pi 0/3 or stopped). Aider stays the code worker until a successor is needed for its advisories; `E1` waits for a new model. A one-shot edit reply bounds a weak model better than an agent loop does. **Next check:** a new model that fits the box, on the slice-1 level first (3 runs, both harnesses only if it drives tools without repeating). Screened the same day (`skills/framework/bench`): no model released since 2026-10-01 fits the box with tool calls; older shapes not yet run here are Muse Glimmer 30B (dense, vendor SWE-bench Verified 76.0) and Laguna XS 2.1 (33B MoE, 3B active, 20 GB at Q4) — a pull is the human's call. Recheck in two weeks.

### 1.2 IDE hosts — judged (`skills/framework/models`)

Which hosted model for which kind of work in the IDE hosts the humans and the role-play use today (Cursor, Claude Code). Judged from vendor sources and this workspace's own sessions, not measured; every hosted model receives the code and the specs (Pillar 2). Checked 2026-10-09.

| Work | Primary | Cheaper fallback | Why |
|---|---|---|---|
| Framework DevOps — multi-file Go engine changes, refactors, architecture docs | Claude Opus 5.5 | Sonnet 5.5; Fable 5.1 only when Opus at higher effort falls short | Vendor guide places large-scale refactoring and complex systems engineering on Opus, Fable as the step beyond (§4 H2). This workspace: the 2026-10-09 engine and bench session ran on Opus 5.5 |
| Spec writing — PM, architect, UX and security role-play | Opus 5.5 | Sonnet 5.5 | Long cross-document consistency; both read 1M tokens (§4 H1) |
| Lead / CEO — strategy text, market research | Opus 5.5 | Fable 5.1 for multistep deep research | §4 H2 |
| App implementation with tests (Go, React) | Sonnet 5.5 | Opus 5.5 for cross-cutting parts | Vendor guide: everyday coding on Sonnet, at half Opus's price (§4 H1, H2) |
| Code review, bug hunting | Opus 5.5 at high effort | Sonnet 5.5 at high effort | Claude Code effort guidance (§4 H3) |
| Quick edits, renames, commit messages | Haiku 5.5 | Sonnet 5.5 at low effort | Lowest latency and price, sub-agent work (§4 H1, H2) |
| Long autonomous runs, cost-sensitive | Sonnet 5.5, Haiku 5.5 as sub-agents | Opus 5.5 | Sub-agent tasks and the subagent model setting (§4 H2, H3) |

Prices per million tokens, input/output: Fable 5.1 10/50, Opus 5.5 4/20, Sonnet 5.5 2/10, Haiku 5.5 from 0.10/0.50; Haiku 4.5 is legacy (§4 H1). Cursor offers the same Claude models; its own Composer models are not assessed here (§4 H4). General-task rankings: openlm.ai arena, LLM-judged (§4 H5) — a lead, not a DUDES result.

### Why the box matters (not “32B”)

- **Hardware budget:** 32GB–64GB consumer systems remain the primary DUDES machine.
- **Active parameters are not a benchmark:** MoE reduces active computation, but residency, memory bandwidth, offload and the inference backend still determine throughput.
- **Context:** distinguish supported, allocated and actually used tokens. Ollama documents VRAM-dependent defaults and increasing memory demand with context (M2).
- **Worker fit:** DUDES needs *one scoped step*, not a large unbounded session. Aider is the incumbent; relative harness overhead and task success need measurements, not prompt-size folklore.

DUDES is designed to benefit from this shift. The Mission stays model-agnostic.

> **Important:** “Fits a 32GB box” is the current hardware sweet spot, not a Mission requirement. Do not freeze a vendor, a dense/MoE shape, or a parameter count into CORE.

---

# 2. The KMU / Sovereignty Problem

Small and medium-sized enterprises (esp. in Germany / EU) face a strategic dilemma: modern AI development tools can significantly increase productivity, while cloud-based AI can create concerns around proprietary source code, data sovereignty, compliance and vendor dependency.

### Existing approaches and their limitations

1. **Cloud & sovereignty trap**
   - Proprietary source code may be sent to external infrastructure.
   - Data sovereignty and compliance requirements can become difficult to manage.

2. **Missing single-tool ecosystem**
   - There is no universally established, easy-to-use local suite that provides a complete agentic development organization with PM, Dev, QA and related roles out of the box.

3. **Total ban**
   - Companies may block AI because of security or compliance concerns, sacrificing productivity.

4. **Hobbyist / IDE-host stack**
   - Teams combine Ollama with an IDE host (Cursor, Cline, Continue, Claude Code, …).
   - Useful coding assistance. Not a DUDES-level process: no frozen spec-set, no mechanical gates, no human seal that the model cannot skip.

5. **Prohibitive enterprise cloud pricing**
   - Dedicated compliant enterprise infrastructure can be financially unattractive for smaller software teams.

### Strategic implication

**Local hardware → efficient open-weight LLM → DUDES process → data & IP sovereignty**

This is a strategic alignment, not a Mission constraint. It pairs with Pillar 3: sovereignty without auditability is incomplete for EU operators.

---

# 3. DSGVO + EU AI Act (How Development Must Be Able to Prove Itself)

EU operators increasingly need not only “AI that works,” but **development and delivery practices that can be explained, logged, and overseen**.

This pillar answers: *how should we develop* so that local autonomous delivery remains compatible with rising regulatory expectations?

> **Not legal advice.** DUDES does not replace a DPIA, AI Act risk classification, or counsel. It aims to make the **process and artifacts** naturally produce evidence that auditors and humans can inspect.

### 3.1 DSGVO (GDPR) — data & processing discipline

Relevant pressures for AI-assisted product work:

| Pressure | Practical meaning for DUDES |
|---|---|
| **Data minimisation & purpose** | Prefer local inference; avoid shipping source, secrets, or personal data to third-party clouds by default |
| **Controller accountability** | Who decided what, on which inputs, with which output — must be reconstructable |
| **Security of processing** | Workspace boundaries, controlled Git, no silent exfiltration in the default path |
| **Documentation** | Durable specs, QR reports, and commit history support accountability narratives |

Local-first execution (Pillars 1–2) is the strongest default contribution DUDES makes to DSGVO-aligned *operation of the toolchain*. Product-specific personal-data handling remains a product/spec concern.

### 3.2 EU AI Act — risk-based AI governance

The AI Act is risk-based. Obligations differ for:

- **AI used as a development aid** (coding / agent tooling inside the company);
- **AI systems placed on the market as products** (customer-facing systems, especially high-risk Annex III use cases).

**Calendar (checked 2026-10-09 against European Commission source L1 — not legal advice):**

| When | What is live / deferred |
|------|-------------------------|
| **2025-02-02** | Prohibitions + AI-literacy duties |
| **2025-08-02** | GPAI model obligations |
| **2026-08-02** | Act generally applicable. **Art. 50 transparency** live. **AI Office / national enforcement** live (incl. GPAI). |
| **2027-12-02** | **Annex III high-risk** provider duties deferred here (Digital Omnibus). Human-oversight + automatic-logging design rules bite then for those systems. |
| **2028-08-02** | **Annex I** high-risk (AI in already-regulated products) deferred here |

DUDES does not become “compliant” on a date. It keeps producing **reconstructable phases, a human stop, and durable evidence** so a later classification is cheap. Do not encode these dates in CORE.

DUDES must support *both* stories where applicable:

| AI Act theme (current reading) | DUDES-aligned process property |
|---|---|
| **Human oversight** | Human seals outcomes; agents never self-seal; human can reject (“not good enough”) |
| **Logging / traceability** | Phase commits, role attribution, durable artifacts, decision/audit scripts |
| **Technical documentation** | Spec-driven path: PS → PRD → TA → tests → implementation → QR |
| **Transparency of capabilities/limits** | Explicit phases, mechanical gates, QA report as human decision surface |
| **Record of intervention** | Seal / reject cycles on the same release line; Git history shows quality progression |

For **high-risk product AI** (when Annex III/I duties apply), providers may need design-time oversight interfaces, logging retention, and Annex-style documentation — those are **product requirements**, not automatic framework features. DUDES should make producing and versioning that evidence *cheap and sequential*, not invent legal classification. The 2026 live layer that already matches DUDES is **transparency + human-readable outcome control**, not a claim that the engine is a high-risk system.

### 3.3 Auditable development — the bridge

**Yes: auditability is the shared hinge** between Mission, KMU trust, DSGVO accountability, and AI Act oversight.

| Evidence surface | Why it matters |
|---|---|
| Role-tagged Git commits | Who did what, when |
| Frozen spec branch + same-name seal/release tags | What contract was implemented |
| Tests before / with implementation | Behavioural evidence, not only prose |
| `*_QR.md` | Human-readable quality decision input |
| Framework-owned phase transitions (Go direction) | Process not invented inside a chat |
| Human seal gate | Oversight that can halt or reject |

### 3.4 Strategic implication

```text
If it cannot be reconstructed from durable artifacts + Git,
it is not good enough for EU-facing autonomous delivery.
```

Prefer tools and designs that:

1. keep IP/data local (Pillars 1–2);
2. emit inspectable phase evidence (Pillar 3);
3. keep a human at the outcome boundary (Mission).

Reject tools that bury state in opaque multi-agent chat or cannot leave an audit trail.

---

# 4. Tool Evaluation Framework

### North star of tool choice

DUDES’ main technical bet is **context engineering**:

> Fix the process, state, inputs and gates in software so a **weaker / local LLM** can still complete the whole product cycle.

Tools are judged by whether they help that bet — not by feature checklists (including MCP).

```text
Dumb-enough model + fixed phase context + durable Git state + bounded tools
        → autonomous delivery
Chatty multi-agent frameworks that invent workflow inside the prompt
        → context drift + token death on a normal PC
```

### Technical dimensions (5)

Every candidate is scored on these for the **configuration DUDES would use**. Rate 🟢 / 🟡 / ❌ with a one-line reason and evidence reference. Use **unknown** for missing evidence and **N/A** for a dimension outside the tool's job; neither is a failure.

| # | Dimension | What we ask |
|---|-----------|-------------|
| **1** | **Context Drift Resistance** | Can work be **phase-isolated**? Fresh worker context after a boundary? Authoritative state in files/Git — not hidden chat memory? |
| **2** | **Context Compression** | Does it shrink what the model must see (repo map, selective files, structured payloads) instead of dumping the tree / history? |
| **3** | **Token & Prompt Economy** | How many tokens does the **tool itself** burn on meta-orchestration, agent debate, or giant system prompts vs productive work? |
| **4** | **Bounded Worker Fitness** | Can Go (or our orchestrator) **invoke → scoped task → verify → exit**? Clear I/O? Tool use (edit/test/shell) without the worker becoming the process owner? Transport (CLI / HTTP / MCP / …) is an implementation detail *under* this axis — not a separate score. |
| **5** | **Git & Audit Offload** | Filesystem state? Atomic commits? Reconstructable trail for Pillar 3? |

**Dropped as a top-level axis:** standalone “MCP integration.” MCP may help Bounded Worker Fitness; it is not Mission-critical by itself.

**Folded:** local-model suitability uses dimensions 1–3 plus Pillar 1. Passing those design checks makes a tool a candidate; only an actual local trial establishes feasibility for a particular model, machine and task.

### Pillar checks (pass / weak / fail / unknown / N/A)

| Pillar | Ask |
|---|---|
| **1 — Consumer-box / normal PC** | Usable without mandatory cloud; survivable context size on ~32GB HW? |
| **2 — KMU sovereignty** | Default path keeps source/data under operator control? |
| **3 — DSGVO / AI Act readiness** | Can phases, decisions and human interventions be reconstructed? |

### Decision Categories

A tool can become:

- **Orchestrator** — controls or inspires the DUDES execution structure (prefer **Pattern** + native Go unless proven otherwise).
- **Worker Tool** — performs a specialized task inside the DUDES loop (preferred adoption path).
- **Pattern / Reference** — architectural idea worth implementing natively.
- **Infrastructure** — model gateway / runtime below the process (e.g. local API server).
- **Rejected** — fights drift resistance, burns local tokens on process theatre, or cannot leave an audit trail.

### Evidence and decision discipline

Keep four statements separate:

| Kind | Required support |
|---|---|
| **Fact** | Primary URL, checked date, exact project/model identity and revision when available; say **documented** or **source-checked** and limit the claim to what was checked |
| **Measurement** | **Measured** only with a preserved run record: configuration, tasks, commands, results and limitations; vendor benchmarks stay vendor-reported facts |
| **Fit judgment** | Named DUDES boundary, assessed configuration and remaining unknowns; a verdict is not a benchmark or legal certification |
| **Forecast** | Horizon, supporting evidence and a condition that would falsify it |

Score default, configurable and unavoidable behavior separately. Optional cloud access, subagents or saved sessions do not establish mandatory use. Durable notes read as scoped artifacts are not hidden conversational state. Conversely, a setting that merely asks an agent to respect a gate is not mechanical enforcement.

**Auditability is not a database brand.** Assess evidence completeness, attribution, retention, integrity and reconstruction. Git integration is DUDES fit (axis 5); a SQLite/PostgreSQL audit trail may be valid without fitting DUDES. Missing evidence is **unknown**, not proof of poor auditability.

**Erosion guard:** `MISSION.md` §The North Star and `CORE.md` §1, §12, §15 remain authoritative. Every adoption must preserve sequential engine-owned transitions, frozen-spec and test discipline, local self-containment, reconstructable Git evidence, phase context boundaries and the human outcome seal. Better benchmarks cannot waive these. `Never` applies to violating mechanisms, not immutable vendor blacklists. A compatible alternative still needs explicit human adoption; a research update cannot change an adapter, feature status or lifecycle.

### Paired worker trial

`dudes bench` is a hermetic **stub plumbing check**, not a model or harness comparison (D1). Aider remains the incumbent, with no comparative win claimed.

A comparison record must identify the decision and thresholds **before** running: frozen task/spec and acceptance-test IDs; required pass rate, time/memory/token limits and repetitions; machine RAM/VRAM; exact model revision/quantization; server and harness versions, tools, prompt, context and output cap. Scope/gate violations must be **zero**; performance cannot compensate for erosion. Pair the same tasks and starting commits, change only the variable under study, use fresh worker contexts and the repository's documented test goals. Include failures, timeouts and blocked runs, not just successful examples.

Keep the test oracle outside the candidate's writable scope; verify it and the frozen inputs are unchanged. Record acceptance results, scope/gate violations, elapsed time, peak memory and input/output usage (or explicitly unavailable). A smaller prompt alone is not a win. Preserve a concise, secret-free result and immutable evidence reference here with the decision; do not rely on chat or disposable logs. If no executable goal or suitable trial environment exists, report **unmeasured**, identify the prerequisite and retain the incumbent. Do not call the stub benchmark a substitute.

### Evidence register

**Checked 2026-10-09.** Each row is current only for its stated scope, not all claims in an essay. Repository sources pin the revision; its Git metadata supplies the source date. Official pages without a recorded publication date are **undated**, checked observations, not evidence of a release in a recent survey window; recheck them before reliance. No external runtime or local comparative benchmark was executed. Claims not covered here are **unverified**, including inherited scores; they cannot justify adoption, a ranking or a new rejection.

| ID | Primary source / revision | Supported claim and limit |
|---|---|---|
| M1 | [Qwen model card](https://huggingface.co/Qwen/Qwen3-Coder-30B-A3B-Instruct), named model; mutable page | Documented 30.5B / 3.3B active and native 262,144 context; not measured local residency or speed |
| M2 | [Ollama context documentation](https://docs.ollama.com/context-length), mutable page | Documented VRAM-dependent defaults and memory growth with context; not a DUDES sizing result |
| M3 | [Mistral Devstral announcement](https://mistral.ai/news/devstral/), Devstral-Small-2505 | Vendor-reported 46.8% SWE-bench Verified with OpenHands; not a harness comparison |
| L1 | [European Commission AI Act timetable](https://digital-strategy.ec.europa.eu/en/policies/regulatory-framework-ai), mutable page | Official timetable: Annex III 2027-12-02, regulated-product high risk 2028-08-02; not product classification or a compliance opinion |
| T1 | [Tetusa1/DoneGate README](https://github.com/Tetusa1/DoneGate/blob/1175d0c00b8a7811cc62dd2a400fae5e5ccb82a4/README.md) | Documented independent commit/check validation, worktrees and path leases; read paths are policy, not a filesystem sandbox |
| T2 | [codemonkeying/lean-coder README](https://github.com/codemonkeying/lean-coder/blob/b37df2b8f2898e0d311ee505032e5c480f9edd49/README.md) | Documented 2,500-token core / 6,500-token all-tools test ceilings, optional tools and compaction; tests not run here |
| T3 | [orbcli/orbit README](https://github.com/orbcli/orbit/blob/4673a1e0a89b35aa50de17f4186693d4a4e09eca/README.md) | Documented roster → memo → worktree loading; ~50/~200 tokens are author estimates, not measurements here |
| T4 | [kdlbs/kandev README](https://github.com/kdlbs/kandev/blob/d7a44e96ede93d50c6277773f4174de716272f02/README.md) | Documented configurable workflows, parent-session subtasks, task messaging, local operation and no telemetry; not a resource benchmark |
| T5 | [Ivy Tendril README](https://github.com/Ivy-Interactive/Ivy-Tendril/blob/3a1a24103f4390d9937bb9d5ba65947a365c1dd3/README.md) and [license](https://github.com/Ivy-Interactive/Ivy-Tendril/blob/3a1a24103f4390d9937bb9d5ba65947a365c1dd3/LICENSE) | Documented worktrees, review/verification, any CLI agent and FSL-1.1-ALv2; not proof of each internal gate |
| T6 | [ouijit/ouijit README](https://github.com/ouijit/ouijit/blob/1eaa46d12274182a6e1bef50fc09e78f9ab2b20c/README.md) | Documented worktrees, lifecycle hooks and agent-driven board; not proof of a DUDES-compatible outcome gate |
| T7 | [JesseRWeigel/toryo README](https://github.com/JesseRWeigel/toryo/blob/9e7d33b44b883b66d508153db84ba1949229a085/README.md) | Documented four-phase cycle, local adapters and configurable cycle limit; detailed fail-closed implementation claims need source checks |
| T8 | [OpenHands SDK delegation example](https://github.com/OpenHands/software-agent-sdk/blob/33044182505244749928644a4643368a9e998ee7/examples/01_standalone_sdk/25_agent_delegation.py) | Source-checked independent parallel subagents and follow-up delegation; not a local performance result |
| T9 | [bernstein README](https://github.com/sipyourdrink-ltd/bernstein/blob/aee2ea28b9ac4250df15c69d80a79b15c1206ed8/README.md) | Documented deterministic scheduling, declarative workflow and worktree gates; not proof that an LLM must invent every plan |
| T10 | [MartinLoop README](https://github.com/Keesan12/martin-loop/blob/c139d1e5685bc7ade894140d642078b7c20dd918/README.md), 0.8.2 | Documented budgets, verifier-bound evidence and optional hosted sync; VERIFIED covers configured checks, not bug-free code |
| D1 | [DUDES benchmark implementation](cmd/dudes/main.go) (`cmdBench`), local Git revision 2026-10-09 | Source-checked: stub run and skip-test are true; no model/harness performance evidence |
| D2 | [DUDES worker launcher](scripts/worker.sh), local Git revision 2026-10-09 | Source-checked pinned Aider toolchain image; not complete host isolation |
| T11 | [mattpocock/skills](https://github.com/mattpocock/skills/tree/49dd158d1076134a641b33efb035946536778336), MIT | Source-checked: invocation axis (user- vs model-invoked), a per-repo setup skill writing `docs/agents/` and an `AGENTS.md` section that the generic skills read, `GLOSSARY.md` with Avoid lists, namespacing by plugin instead of renaming on name collisions; skill quality not measured |
| B1 | Instance trial, slice 1 of the first instance's product item, 2026-10-09, rounds 1–5 — reproduce with `skills/framework/bench` at DUDES 2026-10-09 (round 5); the product's specs and code stay in the instance | Measured outcome per model in §1.1 for that task, configuration and model host; raw logs not kept (scratch) — the bench ladder (F82) records runs as git notes. One task, one repetition: a lead, not a ranking |
| B2 | Bench ladder L1–L2, 2026-10-10 — `make bench-models MODELS="gemma4:26b qwen3.8:27b"` at DUDES 2026-10-10; one line per run in git notes `refs/notes/dudes/bench` (`git log --notes=dudes/bench`) | Measured outcome per model in §1.1, scored against hidden tests; fixtures frozen in `bench/`. Two models, two levels, three runs: a ladder reading, not a ranking of models in general |
| B3 | Replay of the trial, 2026-10-14 — `make bench-models MODELS="gemma4:26b qwen3.8:27b" LEVELS=<instance dev repo>/bench/<level>` at DUDES 2026-10-14, the level at a commit of the instance's dev repo; one line per run in git notes `refs/notes/dudes/bench` | Measured outcome per model in §1.1 on an instance level: one item, two models, three runs each, scored by the trial's approved tests and a hidden test checked against the reference. The level is the first instance's; the result reads as that instance's work |
| H1 | [Claude models overview](https://platform.claude.com/docs/en/models/overview), mutable page, checked 2026-10-09 | Lineup Fable 5.1 / Opus 5.5 / Sonnet 5.5 / Haiku 5.5, prices, 1M context, 128K output; Haiku 4.5 legacy. Vendor facts, not a DUDES result |
| H2 | [Choosing a Claude model](https://platform.claude.com/docs/en/about-claude/models/choosing-a-model), mutable page, checked 2026-10-09 | Vendor selection matrix: Opus 5.5 for complex agentic coding and large refactors, Sonnet 5.5 for everyday coding, Haiku 5.5 for lowest latency and sub-agent tasks, Fable 5.1 when Opus falls short |
| H3 | [Claude Code model configuration](https://code.claude.com/docs/en/model-config), mutable page, checked 2026-10-09 by a delegated research run | Aliases, effort levels per kind of work, subagent model setting; not re-read by the landing session |
| H4 | [Cursor models](https://cursor.com/docs/models), mutable page, checked 2026-10-09 by a delegated research run | Cursor offers the Claude 5.x models and its own Composer models; Composer context and retention undocumented |
| H5 | [openlm.ai Chatbot Arena](https://openlm.ai/chatbot-arena/), updated 2026-10-07, checked 2026-10-09 | 400+ models, LLM-judged arena Elo plus benchmark columns, licence and open-weight marks; run by OpenLM.ai, not LMSYS. A lead for what to try, never a DUDES result |
| T12 | [SakanaAI/mass](https://github.com/SakanaAI/mass/tree/fb403581bed2c974b3f99e7236224169da643079) at `fb40358` (2026-10-10), checked 2026-10-13 | Source-checked: the workflow search rewrites only the team-design block of the task prompt (`harness_improvement/iterate_multi_agent_prompt.py`); the in-loop pairwise judge is the current model unless `--judge-llm-backend` names another (`mass/judging.py`); 18 runs per task ranked Bradley–Terry, ranks 1–15 train a LoRA. Documented: no open-source licence for the MASS code (`NOTICE.md`); one SFT run took six H100s for ~19 h, and qwen-code runs with `--approval-mode yolo` and no sandbox (`docs/running.md`). Its offline tests were not run here |
| T13 | [MASS paper, arXiv 2610.12176](https://arxiv.org/abs/2610.12176) (2026-10) and [project page](https://pub.sakana.ai/mass/), mutable, checked 2026-10-13 | Author-reported: after two cycles 82.5 % (train) / 69.9 % (test) pairwise wins over the base on 12 synthetic research tasks, judged by GPT-5.5 and Claude Opus 4.8; 1.2–1.6× score per output token on four research benchmarks, search cost excluded; SWE-bench Verified 0.626 → 0.611 and Terminal-Bench 2.0 0.375 → 0.352. The authors name shared blind spots of a self-judging model as a risk for RSI. Not reproduced |
| S1 | [agent-orchestrators.com](https://agent-orchestrators.com/), source list `andyrewlee/awesome-agent-orchestrators`, checked 2026-10-09 | Curated directory of 240 open-source orchestrators by category, weekly updated, API and RSS; discovery for `skills/framework/screen`, the tool's repository stays the evidence |

**Review triggers:** source release or configuration changes affecting the cited claim; contradictory evidence; a failed local acceptance criterion; changed law; or a proposed adoption/comparison relying on an unverified claim. A trigger makes the affected claim **stale**, not automatically false, and blocks reliance until reassessment. Model, hardware and legal claims are rechecked before each dependent decision even without a known release. Unaffected rows and shipped features do not expire with them. The §7 record carries the specific trigger and next check.

---

# 5. Framework & Tool Assessments

**Living roster = the Index below; factual support = §4 Evidence register.** Entries without a supporting record are legacy/unverified, not fresh evaluations. Their verdicts are historical fit judgments, not authority to adopt or reject anew. Reassess the decision-relevant claims before reliance. Update the affected index row, essay, evidence and pattern implications together; Git keeps the previous rationale.

In archived scores, “~32B” means the **consumer-box / Pillar 1 budget**, not a dense-32B identity.

Tools are grouped by **what they are**, not by our verdict — the category tells you which question a tool answers for DUDES, the verdict tells you what we do with it. New tools slot into an existing category (or add one); the index stays the one-screen overview.

| Category | Question it answers for DUDES |
|---|---|
| **A · Orchestration & Workflow** | How should the phase machine itself be built? |
| **B · Coding Agents (Workers)** | What executes a bounded implementation task? |
| **C · Spec-Driven Development Toolkits** | How are specs authored, diffed, gated and archived? |
| **D · Context & Output Engineering** | How do we shrink input and structure output for a weak model? |
| **E · Model Infrastructure** | What serves the model below the process? |
| **F · Method & Skill Packs** | How is process or discipline delivered *into* a host agent as prompts? |
| **G · ADE desks** | What can a desk (session shell or task desk over other agents) show us without taking the lifecycle? |
| **H · Self-improvement (RSI)** | How could a loop improve prompts, workflows or weights from its own runs — and who judges, so that no loop grades itself? |

### Index

| Tool | Cat. | Verdict | → §6 patterns |
|---|---|---|---|
| LangGraph 🕸 | A | PATTERN / REFERENCE | Explicit State Machine · Context Reset |
| Temporal ⏱️ | A | PATTERN / REFERENCE | Durable Phase Runner |
| bernstein 🎼 | A | PATTERN / REFERENCE — zero-token coordination | Framework-Owned Execution (completion signals) · Explicit State Machine · Worktree Worker |
| Archon 🏛️ | A | PATTERN / REFERENCE (runtime rejected) | Framework-Owned Execution · Convergence Passes · Context Reset · Worktree Worker · (Anti: User-Definable Graph) |
| ralphex 🔁 | A | PATTERN / REFERENCE | Context Reset · Framework-Owned Execution · Orchestrator-Owned Commits · Cross-Model Review |
| MartinLoop 🧮 | A | PATTERN / REFERENCE | Run Budget Hard-Stop · Bounded Worker Invoke Contract · Evidence Surfaces |
| toryo 🏗️ | A | PATTERN / REFERENCE (runtime rejected) | Framework-Owned Execution · (Anti: Process-as-Prompt · Agent↔Agent Channel Bus) |
| DoneGate 🔒 | A | PATTERN / REFERENCE (runtime rejected) | Orchestrator-Owned Commits (inverse) |
| CrewAI 👥 | A | **REJECTED** | — (Anti-pattern: Agent↔Agent Channel Bus) |
| Aider 🚀 | B | **WORKER (adopt)** — incumbent, comparative performance unmeasured | Repository Map · Atomic Git Worker |
| OpenHands 👐 | B | WORKER (conditional) | Ephemeral Sandbox · History Compaction · Event Replay · Invoke Contract |
| Goose 🪿 | B | WORKER (conditional) | Recipe-Scoped Worker Invoke |
| OpenCode ⌨️ | B | WORKER (conditional) — reassess configuration and run a paired trial before adoption | Invoke Contract · Orchestrator-Owned Commits · Context Reset |
| Pi 🥧 | B | WORKER (conditional) | Minimal Tool Belt · Event Replay · Skill Unit |
| lean-coder 🪶 | B | PATTERN / REFERENCE (runtime rejected) | Minimal Tool Belt (overhead ceiling) · (Anti: Session as Record · Agent↔Agent Channel Bus) |
| Qwen Code 🪪 | B | **NOT EVALUATED** — no adoption without assessment | — |
| aGiTrack 🧾 | B | PATTERN / REFERENCE (worker wrapper) | Git Notes as Side Record · Worktree Worker · (caution: transcripts in history) |
| OpenSpec 📐 | C | PATTERN / REFERENCE | Delta Specs · Requirement/Scenario Gate · Change Folder + Archive · Machine-Readable CLI |
| GitHub Spec Kit 🌱 | C | PATTERN / REFERENCE | Cross-Artifact Check · Converge-to-Tasks · Machine-Readable CLI · (Anti: User-Definable Graph) |
| Spec Kitty 🐈 | C | PATTERN / REFERENCE | Worktree-Isolated Worker |
| Amazon Kiro 🟣 | C | PATTERN / REFERENCE (runtime rejected) | Scoped Standards Injection · Artifact Handoff |
| Repomix / code2prompt 📦 | D | **PATTERN / HELPER (adopt)** | Repository Map / Phase-Scoped Pack |
| Outlines / Instructor 🧩 | D | PATTERN | Schema-Validated Control Output |
| PydanticAI 🩺 | D | PATTERN / cond. WORKER | Schema-Validated Control Output |
| Orbit 🪐 | D | PATTERN / REFERENCE | Repository Map / Phase-Scoped Pack · Artifact Handoff |
| LocalAI 🤖 | E | INFRASTRUCTURE | Local Model Gateway |
| BMAD-METHOD 🧭 | F | PATTERN / REFERENCE | Context Reset · Artifact Handoff · (Anti: Process-as-Prompt) |
| GSD Core 🚚 | F | PATTERN / REFERENCE | Context Reset · (Anti: Process-as-Prompt) |
| Superpowers 🦸 | F | PATTERN / REFERENCE | Two-Axis Review Gate · Skill Unit · Worktree Worker |
| Agent OS 🧱 | F | PATTERN | Scoped Standards Injection |
| mattpocock/skills 🎓 | F | **PATTERN / HELPER (adopt)** | Skill Unit · Two-Axis Review Gate |
| Traycer 🛰️ | G | PATTERN / REFERENCE (runtime rejected) | Worktree Worker (setup goal) · Two-Axis Review (severity, focused re-verify) · Read-only Desk over the Record · (Anti: Agent↔Agent Bus · User-Definable Graph) |
| MonoCode 🗂️ | G | PATTERN / REFERENCE (watch as a face) | Read-only Desk over the Record · (Anti: Agent↔Agent Bus) |
| paperclip 📎 | G | PATTERN / REFERENCE (watch for the `ROADMAP.md` Horizon) | Heartbeat → Bounded Run · Run Budget Hard-Stop · Read-only Desk · (Anti: Session as Record) |
| Ivy Tendril 🌿 | G | PATTERN / REFERENCE (runtime rejected) | Framework-Owned Execution · (Anti: Process-as-Prompt · Session as Record) |
| Kandev 🧰 | G | **REJECTED** | — (Anti: User-Definable Lifecycle Graph · Agent↔Agent Channel Bus · Session as Record) |
| Ouijit 🃏 | G | PATTERN / REFERENCE (runtime rejected) | Worktree-Isolated Worker (setup goal) · (Anti: Session as Record) |
| MASS 🪞 | H | PATTERN / REFERENCE (runtime rejected) | Bench-Judged Workflow Search · Cross-Model Review · (Anti: Process-as-Prompt) |

---

## A · Orchestration & Workflow

### LangGraph 🕸

> Explicit **graph / state-machine** orchestration (LangChain family). Nodes load only needed state; transitions are visible. Pattern source for DUDES phases — implement the machine in **Go**, don’t vendor-lock on Python.

- **DUDES assessment:** **PATTERN / REFERENCE** — implement state machine in Go; do not vendor-lock.

**Reusable idea:** state machine + context reset + Git-synchronized transitions (= theoretical CrewAI lifecycle, mechanical not chatty).

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

### Temporal ⏱️

> Durable workflow engine (event history, retries, timers, signals). Not an LLM product. Industry pattern: Temporal = durable outer lifecycle; agent graph = inner reasoning. Heavy ops (server/workers) vs laptop MVP.

- **DUDES assessment:** **PATTERN / REFERENCE only** — steal durable phase machine, activity timeouts/retries, human-signal gates. **Implement a thin Go phase runner first** (REFACTORING.md). Reconsider Temporal only if DUDES must survive long multi-day jobs across machines.

**Reusable idea:** activity = one phase worker invoke; workflow = sequential DUDES lifecycle; signal = human seal/reject; continue-as-new / compact history for long cycles.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

### CrewAI 👥

> **Theory:** role/task split (PM → architect → coder → tester) = small prompts per step — sound lifecycle idea.  
> **Practice:** handoffs copy large text blocks agent↔agent; orchestration tax often exceeds a single focused run. Agents are weak coders (no OpenHands-class Docker/tool/git depth).

- **DUDES assessment:** **REJECTED** as runtime. Keep the **idea** of small role-scoped tasks — implement as **Go phases + durable artifacts**, not multi-agent chat.

**Reusable idea:** phase/role decomposition into tiny prompts (lifecycle coverage). **Do not** reuse chat handoffs or debate orchestration.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

### bernstein 🎼

> **Deterministic governance and orchestration for agents:** a declarative workflow specifies phases, roles and dependencies; Python schedules and gates the work. Coding tasks use worktrees. An LLM-assisted planner is not the only possible source of a workflow (T9).

- **Context Drift Resistance:** 🟢 Short-lived agent processes, one bounded task each, in their own worktree; transitions validated by an explicit state machine; state in `.sdd/` files ("no hidden databases").
- **Context Compression:** 🟡 One task plus a role prompt, but agents get the full context window and pull skills on demand over MCP; no repo map or size budget.
- **Token & Prompt Economy:** 🟢 Coordination costs zero tokens — a priority queue over a dependency graph in plain Python. Parallel agents and an optional LLM reviewer still spend.
- **Bounded Worker Fitness:** 🟡 Its agent contract is the DUDES one (adapter → spawn in worktree → verify → exit); bernstein itself is a second process owner, not something Go invokes.
- **Git & Audit Offload:** 🟢 Worktree per agent, serialized merge only after verification, run journals and checkpoints; "the same plan replays to a byte-identical task graph". Caution: `.sdd/` is a second store beside Git, plus a FastAPI task server.
- **Pillar 1 (consumer-box / normal PC):** pass — local by default, Ollama and Aider adapters; parallelism is the default bet.
- **Pillar 2 (KMU sovereignty):** pass — Apache-2.0, "file-based state, no SaaS hop, no third-party data plane".
- **Pillar 3 (DSGVO / EU AI Act / auditability):** pass — FSM-validated transitions, verification evidence, approval gates before merge.
- **DUDES Assessment:** **PATTERN / REFERENCE** — deterministic coordination is relevant to `CORE.md` §1; this is not a comparative superiority result.
- **Use:** take the ideas, run nothing. Adopting its scheduler would introduce a second process owner. A declarative plan is not inherently invented by a model; the DUDES question is whether tasks derive from frozen spec items and whether engine-owned gates remain non-optional (F71, `CORE.md` §6).

**Reusable ideas:** → §6 **Framework-Owned Execution** (declarative completion signals), **Explicit State Machine** (replay), **Worktree-Isolated Worker** (merge only after verification).

### Archon 🏛️

> **Workflow engine for AI coding:** a YAML graph per project mixes deterministic nodes (bash, git, tests) with agent nodes (plan, implement, review), loops "until" a condition with fresh context, and human approval nodes; every run gets its own worktree. "Make AI coding deterministic and repeatable."

- **DUDES Assessment:** **PATTERN / REFERENCE** (runtime rejected).

**Reusable ideas:** sources for §6 **Framework-Owned Execution**, **Convergence Passes**, **Context Reset**, **Worktree-Isolated Worker**; anti-pattern **User-Definable Lifecycle Graph**.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

### ralphex 🔁

> **Plan runner for one coding harness:** a markdown plan (tasks with checkboxes plus validation commands) runs task by task, each in a fresh Claude Code session, with tests and linters after each task and a commit per task; then a review pipeline — five parallel review agents, an external review by another model, two final agents.

- **DUDES Assessment:** **PATTERN / REFERENCE** — a Go plan-runner reference; comparative overhead is unmeasured.

**Reusable ideas:** → §6 **Cross-Model Review** (new, hold); sources for **Context Reset**, **Framework-Owned Execution** (validation commands per plan), **Orchestrator-Owned Commits** (commit per task and fix).

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

### MartinLoop 🧮

> **Run governor for coding agents:** every run gets a contract — objective, verifier command, token or USD budget, allowed and denied paths, stop condition; completion needs fresh verifier evidence, and each run leaves a signed receipt (stop reason, verifier result, cost).

- **Context Drift Resistance:** 🟡 A run is bounded; no phase structure or context reset described.
- **Context Compression:** 🟡 Not its focus.
- **Token & Prompt Economy:** 🟢 Hard budget stop "before runaway retries spend more money".
- **Bounded Worker Fitness:** 🟢 The contract is the DUDES invoke shape, "intentionally decoupled from the worker"; drives Ollama and llama.cpp through OpenAI-compatible endpoints.
- **Git & Audit Offload:** 🟡 JSONL receipts with HMAC signing; shared bundles leave out prompts and source. Rollback state lives in its own record, not Git.
- **Pillar 1 (consumer-box / normal PC):** pass — local-first, local endpoints.
- **Pillar 2 (KMU sovereignty):** unknown for the full install path — local-first and optional hosted sync are documented (T10); telemetry defaults and the selected runtime need separate verification.
- **Pillar 3 (DSGVO / EU AI Act / auditability):** pass — signed receipts, stop reasons, cost provenance.
- **DUDES Assessment:** **PATTERN / REFERENCE**.
- **Use:** take "VERIFIED only claims the checks it ran" for QR wording, and the run budget. Git is DUDES's record, not automatic proof of authenticity or completeness: signatures require an actual signing and verification policy, and hashes do not prove the checks ran.

**Reusable ideas:** → §6 **Run Budget Hard-Stop** (new, hold); sources for **Bounded Worker Invoke Contract**, **Evidence Surfaces for Oversight**.

### toryo 🏗️

> **Phase loop over pluggable coding CLIs:** each cycle is plan, research, execute, then a review that must return one JSON object. A configured check can block acceptance. Git keeps a passing checkpoint and undoes a rejected one with a revert commit.

- **Context Drift Resistance:** 🟡 Task specs are markdown files. Phases pass a truncated context onward inside the cycle. Trust is a rolling average across cycles. The opt-in bus can stream one agent's tokens to another.
- **Context Compression:** 🟡 Truncation between phases strips boilerplate. The spec is the task. There is no repo map.
- **Token & Prompt Economy:** 🟡 Four agent phases per cycle are documented; default unbounded execution can be capped with `-n` (T7). An Ollama adapter can be the coder. Measure the configured cycle rather than treating an optional limit as unavailable.
- **Bounded Worker Fitness:** ❌ Adapters are real invokes (Aider, Ollama HTTP, and the hosted CLIs). Toryo owns the cycle. Checks run as an executable plus an argument array, with no shell, a timeout, and an output cap. Timeouts, a missing binary, a nonzero exit, and overflow fail the gate. Those commands are trusted local code and are not sandboxed.
- **Git & Audit Offload:** 🟡 A clean tree is required. Source is checkpointed before review. A rejected attempt is a revert commit of that checkpoint, not a hard reset. Unexpected changes stop the run. A diff over the size cap stops review. Scores, trust, and review JSON live in an ignored runtime directory, not in the product history.
- **Pillar 1 (consumer-box / normal PC):** unknown performance — Aider and Ollama are documented, and cycle count is configurable. Multi-phase overhead needs a local measurement.
- **Pillar 2 (KMU sovereignty):** pass — MIT, local process, notifications off unless configured. Review files can contain source and command output.
- **Pillar 3 (DSGVO / EU AI Act / auditability):** weak — evidence ids and the review JSON are a local trail. That trail does not prove the model inspected the evidence. There is no human seal. With no required check configured, acceptance is the model score.
- **DUDES Assessment:** **PATTERN / REFERENCE** (runtime rejected).
- **Use:** take the fail-closed check: a failed command blocks acceptance even when the reviewer score passes. Run nothing. The revert commit is their safety on a live checkout. DUDES does not land a failed phase, and a worktree already isolates the edit, so that revert is not a second mechanism. Fails if the session is the record: trust and reviews accumulate beside Git, and the bus can stream one agent to another.

**Reusable ideas:** → §6 **Framework-Owned Execution** (source). Anti-pattern sources **Process-as-Prompt** (a reviewer score, or a trust average, is the acceptance when no check is configured), **Agent↔Agent Channel Bus** (opt-in bus; the default cycle does not use it).

### DoneGate 🔒

> **Local task checker:** it runs one worker command in a git worktree and accepts the task only when Git and the required checks agree. A process exit of 0 is not completion.

- **Context Drift Resistance:** 🟡 Each task starts from a recorded base ref in its own worktree. The worker command can be any agent, so that agent's session is outside this tool. After cleanup the task row and the validation evidence stay in a side store.
- **Context Compression:** 🟡 Not its job. Declared read paths are policy text. They are not a repo map, and they are not enforced on the filesystem.
- **Token & Prompt Economy:** 🟢 The checker calls no model. Spend is whatever worker command the task names.
- **Bounded Worker Fitness:** 🟡 A JSON CLI, stable exit codes, and one worker command per task. It is a checker in front of that command. The engine already owns that check.
- **Git & Audit Offload:** 🟡 Completion requires an independent `HEAD` read that matches the worker's printed commit claim, a descendant of the recorded base, a clean worktree on the declared branch, paths inside the allowlist, and passing required checks. Leases and the task row are SQLite under an ignored `.agent-worktree/` directory. Recovery will not reset, clean, or stash.
- **Pillar 1 (consumer-box / normal PC):** pass for the checker — local Python, no model, no network. Windows and Linux are both in its CI. The worker command can still be a hosted agent.
- **Pillar 2 (KMU sovereignty):** pass — MIT, local, no account. Task files and worker logs can contain the project.
- **Pillar 3 (DSGVO / EU AI Act / auditability):** weak — the validation is real, and the trail is the side database, not a git note. There is no human seal.
- **DUDES Assessment:** **PATTERN / REFERENCE** (runtime rejected).
- **Use:** take nothing to run. Their check is the inverse of orchestrator-owned commits (F42): the worker is allowed to commit, and a later read of Git decides whether the claim is true. DUDES starts the worker so it cannot commit, and a moved `HEAD` is a warning. Overlapping path leases matter when several writers run at once. The engine is sequential. Fails if the session is the record: the task row outlives the worktree in SQLite.

**Reusable ideas:** → §6 **Orchestrator-Owned Commits** (inverse).

---

## B · Coding Agents (Workers)

### Aider 🚀

> Terminal **pair-programmer**: repo map + edit + optional atomic commits; strong with local/open models. Default DUDES **code worker** — Go invokes a fresh process per impl phase (`--message`, prefer `--no-auto-commits` so the orchestrator owns the release trail).

- **DUDES assessment:** **WORKER TOOL (ADOPT path)** for routine implementation phases.

**Reusable idea:** repo-map compression, atomic Git worker, mature code editing.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

### OpenHands 👐

> **Agent platform and SDK:** sandbox/tool execution, conversation management and optional delegation. The SDK supports independent parallel subagents (T8); whether DUDES enables any such facility is a separate fit decision. DUDES still supplies a bounded task and owns every phase transition.

- **Context Drift Resistance:** 🟡 A fresh invocation can preserve the phase boundary; verify the chosen configuration has no dependency on prior conversation. Optional memory alone is not a failure.
- **Context Compression:** unknown for the selected version/configuration — compaction and repo mapping are different mechanisms; measure retained task information and input size.
- **Token & Prompt Economy:** unknown — no pinned local comparison is recorded here.
- **Bounded Worker Fitness:** 🟡 SDK and delegation exist; verify invocation, tool permissions, termination and isolation in the selected configuration. Never delegate DUDES phase ownership.
- **Git & Audit Offload:** 🟡 FS + event replay; phase/release commits stay with Go (Aider still stronger for atomic coder commits).
- **Pillars:** 1 weak · 2 pass (local/Docker) · 3 weak→pass (seal outside OH).
- **DUDES assessment:** **WORKER TOOL (conditional)** after Aider for default code; use OH when isolation + deep tools matter and the phase is already decomposed.

**Reusable idea:** → §6 **Ephemeral Execution Sandbox** (later), **In-Phase History Compaction**, **Event-Sourced Worker Replay**, **Bounded Worker Invoke Contract**.

---

### Goose 🪿

> Block → Linux Foundation (AAIF). Rust on-machine agent: CLI + desktop + API. Ollama/local providers; tools mainly via **MCP** + developer extension; **Recipes** (YAML/Jinja) for reusable workflows; optional sub-recipes/subagents.

- **DUDES assessment:** **WORKER TOOL (conditional)** — peer to OpenHands for multi-tool phases; **Recipes** are the interesting bit (declarative phase packs). Default code edits: still **Aider**. Do not adopt Goose as product-cycle orchestrator.

**Reusable idea:** → §6 **Recipe-Scoped Worker Invoke**.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

### OpenCode ⌨️

> Open-source terminal/desktop coding agent (SST lineage). Provider-agnostic (75+ / OpenAI-compatible local). Client–server (Hono/Bun), SQLite sessions, built-in tools + LSP, **session compaction**, **git snapshots** for undo (often **without** writing user commits), optional worktrees, HTTP API for embedding.

- **DUDES assessment:** **WORKER TOOL (conditional)** — reassess the current API and local configuration, then use the paired worker trial (§4) for any adoption decision. Aider remains the incumbent, not a proven winner. `dudes bench` cannot compare these workers.

**Reusable idea:** → §6 **Bounded Worker Invoke Contract**, **Orchestrator-Owned Commits**, **Context Reset** (fresh session), **In-Phase History Compaction**.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

### Pi 🥧

> **Minimal terminal coding harness**: four default tools (`read` / `write` / `edit` / `bash`), tree-structured JSONL sessions, and print / JSON / RPC / SDK run modes. Everything else — sub-agents, plan mode, permission gates — is an installable extension or skill, not core.

- **DUDES assessment:** **WORKER TOOL (conditional)** — the existing alternative adapter for a paired trial when a concrete Aider limitation is observed. No comparative efficiency win is established. Runtime, provider and skill support must be checked at the trial revision.

**Reusable idea:** → §6 **Minimal Tool Belt per Phase**, **Event-Sourced Worker Replay** (JSONL tree), **Skill = Portable Discipline Unit**.

**Evidence / review:** assessed 2026-10-16 for the harness trial: the package moved to `@earendil-works/pi-coding-agent` (the `@mariozechner` scope is deprecated at 0.73.1), MIT, very active (8 releases in 15 days, breaking changes in minor versions: pin exactly), Node 22.19 or newer, no built-in Ollama provider (a `models.json` with the `/v1` endpoint), `--mode json` exits 0 on an error reply (read the events, not the exit code), network calls off only with `PI_OFFLINE`, `PI_SKIP_VERSION_CHECK` and `PI_TELEMETRY=0`. Measured: §1.1, harness trial 2026-10-16 — with `gemma4:26b` it looped and edited read-only tests. **Review:** current (checked 2026-10-16). **Decision:** stays a conditional worker; not adopted. **Revisit:** a model that drives a tool loop without repeating, or Aider's retirement.

### lean-coder 🪶

> **Terminal coding harness:** a short always-on prompt and a small default tool list, aimed at a local model. The session is kept and loaded again next time.

- **Context Drift Resistance:** ❌ The last session for the working directory loads on start. Compaction keeps a summary plus the newest turns, then sends the model's own next instruction as a new turn. A background worker can be saved and resumed.
- **Context Compression:** 🟡 Large reads and command output are truncated. One command stubs old tool results without a model call. There is no repo map. Compaction itself is another model turn, and it is told to update project docs.
- **Token & Prompt Economy:** 🟡 The README reports test ceilings of 2500 tokens for the always-on prompt plus core schemas and 6500 with all bundled tools (T2). Those tests were not run here; tokenizer and counting method need checking before comparison. Optional tools are off by default.
- **Bounded Worker Fitness:** ❌ A REPL. Approval and a tool leash are for a person at the keyboard. A turn has a tool-round cap. This is not a command Go starts and then exits.
- **Git & Audit Offload:** ❌ During compaction the model is told to update the project's docs and commit them. Edits go through a search/replace tool that writes nothing if a block does not match. There is no orchestrator commit.
- **Pillar 1 (consumer-box / normal PC):** weak — stdlib Python, Ollama by default, and the detected context window may only shrink, capped at 32768. Hosted providers are one login away. A session that compacts by another model call, plus optional workers, is the spend.
- **Pillar 2 (KMU sovereignty):** pass on the default path — MIT, local Ollama, keys in a private config file rather than the repo. A hosted login, web search, and an SSH remote are optional.
- **Pillar 3 (DSGVO / EU AI Act / auditability):** weak — sessions live in the user config directory. A compaction commit is the model's. There is no seal.
- **DUDES Assessment:** **PATTERN / REFERENCE** (runtime rejected).
- **Use:** take the testable overhead-ceiling idea, not a claimed DUDES measurement. Check counting method and enforce the chosen worker's budget in a real trial. Run nothing by default. If evaluating this harness, verify whether session reuse and compaction side effects can be disabled; their existence alone does not establish an unavoidable dependency.

**Reusable ideas:** → §6 **Minimal Tool Belt per Phase** (overhead ceiling). Anti-pattern sources **Session as Record**, **Agent↔Agent Channel Bus**. The compaction commit is the inverse of **Orchestrator-Owned Commits**.

---

### Qwen Code 🪪

> **Not evaluated.** A candidate associated with the Qwen coding ecosystem; exact version, local provider support and prompt overhead must be sourced before scoring.

- **DUDES assessment:** **NOT EVALUATED** — no adapter adoption without assessment. A documented fit improvement or a reproducible incumbent failure can trigger review. Not a permanent vendor rejection.

---

### aGiTrack 🧾

> **Git recorder around a coding harness:** runs Claude Code, Codex or OpenCode in a worktree per session and commits each completed turn — prompt, interaction trace and metadata (backend, model, tokens) in the commit body, summaries in `git notes`.

- **DUDES Assessment:** **PATTERN / REFERENCE** (worker wrapper).

**Reusable ideas:** → §6 **Git Notes as Side Record** (new, hold); source for **Worktree-Isolated Worker** (temporary branch per turn).

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

## C · Spec-Driven Development Toolkits

### OpenSpec 📐

> **Spec-driven development toolkit** (CLI + agent slash commands): keeps a repo-native spec as the source of truth and has every change authored as a **delta** — `ADDED / MODIFIED / REMOVED` requirements with given/when/then scenarios — which the archive step folds back into the spec. Authoring happens inside a host agent; the CLI validates, reports and archives deterministically.

- **DUDES assessment:** **PATTERN / REFERENCE — adopt the *format* natively in Go.** Do not add a Node runtime to the engine and do not hand the lifecycle to an IDE chat. The spec model is plain Markdown and is parseable by `internal/artifact` + `internal/gate` without a dependency.

**Reusable idea:** → §6 **Delta Specs**, **Requirement/Scenario as Mechanical Gate**, **Change-Scoped Folder + Archive**, **Machine-Readable CLI Surface**.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

### GitHub Spec Kit 🌱

> **Spec-driven development process harness** (CLI + agent slash commands): drives a fixed artifact chain — constitution → spec → plan → tasks → implement → converge — where each phase writes Markdown that feeds the next, adds optional quality gates (`clarify`, `checklist`, `analyze`), and ships a declarative, resumable workflow runner with human `gate` steps.

- **DUDES assessment:** **PATTERN / REFERENCE — strong convergent validation, not an adoption.** Constitution ≈ `CORE.md` + `rules/*` + Evolution Guard; `gate` steps ≈ human seal; per-run resumable state ≈ `.dudes/state.json`; `converge` ≈ `dudes check`. Adopting it as runtime would replace exactly the part `CORE.md` §4 assigns to **Go**.

**Reusable idea:** → §6 **Cross-Artifact Consistency Check**, **Converge-to-Tasks Loop**, **Machine-Readable CLI Surface**; plus the anti-pattern **User-Definable Lifecycle Graph**.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

### Spec Kitty 🐈

> **Spec-driven development toolkit with a governed execution runtime** (Spec Kit descendant): after planning, tasks become **work packages** that move through explicit lanes (planned → in review → approved → done), each implemented in an **isolated git worktree** and closed by a `review → accept → merge` loop plus a per-mission retrospective.

- **DUDES assessment:** **PATTERN / REFERENCE.** Parallel multi-agent fan-out contradicts the **sequential** phase machine (`CORE.md` §4) and the Agent↔Agent anti-pattern in §6. Mine the isolation and review mechanics; do not adopt the runtime. Young and broad (3.x, ~690 open issues) — track, don't depend.

**Reusable idea:** → §6 **Worktree-Isolated Worker** (cheap precursor to the Ephemeral Execution Sandbox); lane + retrospective notes recorded there.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

### Amazon Kiro 🟣

> **Cloud AI IDE with a spec-driven workflow** (AWS): every feature becomes a `requirements.md` / `design.md` / `tasks.md` triad with structured acceptance criteria; steering files inject project standards by inclusion mode; hooks and dependency-ordered task waves run the implementation. Inference is Amazon Bedrock only.

- **DUDES assessment:** **PATTERN / REFERENCE — runtime rejected.** Cloud-only contradicts Pillars 1–2 outright. Mine the spec triad, the bugfix-spec shape and the steering inclusion modes.

**Reusable idea:** → §6 **Scoped Standards Injection**; bugfix shape → **Converge-to-Tasks Loop** source; parallel task waves stay out (sequential machine, `CORE.md` §4).

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

## D · Context & Output Engineering

### Repomix / code2prompt 📦

> CLI **context packers**: turn a repo (respecting gitignore) into one LLM-oriented artifact. Repomix: Tree-sitter `--compress`, XML/MD, token counts. code2prompt: Rust, Handlebars templates, git diff/log, optional MCP.

- **DUDES assessment:** **PATTERN / HELPER (ADOPT path)** for compression beside Aider’s repo map. Prefer **narrow packs per phase** over whole-monorepo dumps. Pick one CLI (Repomix *or* code2prompt) — don’t run both in MVP.

**Reusable idea:** → §6 **Repository Map / Phase-Scoped Pack**.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

### Outlines / Instructor 🧩

> **Outlines** (dottxt): constrained generation (JSON/regex/grammar) at decode time — strong with local/vLLM/llama.cpp. **Instructor** (+ **instructor-go** / **structllm**): schema → validate → repair/retry for typed LLM outputs. Same family as PydanticAI.

- **DUDES assessment:** **PATTERN** — aligns with Delimited Artifact Extract + Structured Output. Use for **small machine-validated** phase results; **not** for free-form code bodies (keep delimiters / Aider).

**Reusable idea:** → §6 **Schema-Validated Control Output** (same family as PydanticAI; Go ports preferred).

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

### PydanticAI 🩺

> Typed agent/IO layer: schemas in → validated structures out (Pydantic). Great for **small machine-consumed** control payloads; not a full coding agent or orchestrator. Go ports (instructor-go / structllm) preferred inside DUDES.

- **DUDES assessment:** **PATTERN / conditional WORKER** for narrow machine-consumed steps.

**Reusable idea:** → §6 **Schema-Validated Control Output** (shared with Instructor / Outlines).

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

### Orbit 🪐

> **Multi-repo workspace CLI:** it builds a real directory of git worktrees and shows an agent a repo roster, then a short memo, then full source only for the repos that agent adds. It does not run a lifecycle.

- **Context Drift Resistance:** 🟡 Workspace state is files: a shared repo pool, worktrees, and memos. The agent can write the memo the next start reads. Session-start hooks in the bundled agent plugins inject that state into the living session.
- **Context Compression:** 🟢 The documented ladder is a roster, memo, then full tree on add (T3). About 50/200 tokens are the author's estimates, not local measurements. No retrieval index or enforced total context budget is established.
- **Token & Prompt Economy:** 🟢 The tool's own commands are small and local. The agent it launches is a separate spend.
- **Bounded Worker Fitness:** 🟡 The CLI can be called and it exits. Some commands emit JSON. It does not verify, and it does not own a phase. Launch integrations are hosted CLIs plus OpenCode.
- **Git & Audit Offload:** 🟢 Worktrees are real checkouts. Commit, blame, log, and push stay native git. A knowledge repo can sit beside the code. There is no database and no daemon. Marking a workspace done records a pull-request URL. It does not seal.
- **Pillar 1 (consumer-box / normal PC):** pass for the tool — no server, no container. The quick start launches a hosted agent. OpenCode is a supported launch.
- **Pillar 2 (KMU sovereignty):** pass — MIT, files on disk, push only when someone pushes.
- **Pillar 3 (DSGVO / EU AI Act / auditability):** weak — history is git when someone commits. There is no human gate and no phase record. An agent can commit the knowledge repo.
- **DUDES Assessment:** **PATTERN / REFERENCE**.
- **Use:** take the three-level load (roster, memo, then tree) as a code map, and a knowledge repo only when a later phase reads sealed files. Run nothing as a cycle. Fails if the session is the record: an agent note written for the next start becomes the brief that start trusts, and the session hook pours it into the chat.

**Reusable ideas:** → §6 **Repository Map / Phase-Scoped Pack** (source; a code map, not the spec pack), **Artifact Handoff** (source; sealed files, not an agent note).

---

## E · Model Infrastructure

### LocalAI 🤖

> OpenAI-compatible **local model server** (and similar gateways: Ollama, llama.cpp HTTP, …). Serves inference; does not own product lifecycle, Git, or coding tools. DUDES talks HTTP `generate`/`chat` to whatever sits here.

- **DUDES assessment:** **INFRASTRUCTURE** (local model gateway).

**Reusable idea:** OpenAI-compatible local serving below DUDES.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

## F · Method & Skill Packs

Prompt-delivered process. These install into a host agent (Claude Code, Codex, Cursor, Pi, OpenCode, …) as skills or slash commands; none of them is a runtime Go would call. They matter for two reasons: they validate the lifecycle DUDES fixes in Go, and their skill files are reusable worker content.

### BMAD-METHOD 🧭

> **Agile AI-driven development method delivered as skills**: role skills (analyst, PM, UX, architect, dev, QA) write briefs, PRDs, `SPEC.md`, `stories.yaml`, `sprint-status.yaml`; `bmad-build` implements one story per session; `bmad-build-auto` / BMad Loop run unattended; planning depth "right-sizes" to the intent.

- **DUDES assessment:** **PATTERN / REFERENCE — closest cousin.** Steal the artifact vocabulary (brief → PRD → spec → stories → sprint status → retrospective) and story-per-session; reject process-as-prompt.

**Reusable idea:** → §6 anti-pattern **Process-as-Prompt**; **Skill = Portable Discipline Unit**; story-per-session → **Context Reset** source.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

### GSD Core 🚚

> **Context-engineering phase loop for host agents** (formerly *get-shit-done*): Discuss → Plan → Execute → Verify → Ship per milestone; heavy work runs in fresh-context subagents while `STATE.md` / `CONTEXT.md` survive sessions; executors run in parallel waves.

- **DUDES assessment:** **PATTERN / REFERENCE** — the clearest independent statement of the context-rot thesis; nothing to run.

**Reusable idea:** → **Context Reset** source (fresh-context subagent per task); `STATE.md` ≈ `.dudes/state.json` + `SESSION_STATE.md`.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

### Superpowers 🦸

> **Development methodology as auto-triggering skills**: brainstorming → worktree → plan (2–5-minute tasks) → subagent-driven implementation with two-stage review (spec compliance, then code quality) → TDD → finish branch. Skills are mandatory workflows the agent must check before any task.

- **DUDES assessment:** **PATTERN / REFERENCE** — two-stage review and tested skills are the takeaways.

**Reusable idea:** → §6 **Two-Axis Review Gate**; **Skill = Portable Discipline Unit** (with evals); **Worktree-Isolated Worker** source.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

### Agent OS 🧱

> **Standards management for AI coding**: discovers conventions from an existing codebase into documented standards, indexes them, and injects only the relevant ones when shaping a spec or prompting. Plain Markdown under `agent-os/`.

- **DUDES assessment:** **PATTERN** — scoped standards injection applies directly to `rules/*.md`, which are always-on today.

**Reusable idea:** → §6 **Scoped Standards Injection**.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

---

### mattpocock/skills 🎓

> **Composable engineering-discipline skills** (TDD, bug diagnosis, two-axis code review, domain modelling, tickets, handoff, writing-for-agents): small `SKILL.md` files, model-agnostic, explicitly *not* a process owner — positioned against GSD / BMAD / Spec Kit "taking away your control".

**Evidence / review:** §4 T11 — repository at `49dd158` (2026-10-09), MIT, 38 skills in buckets (`engineering`, `productivity` promoted; `in-progress`, `misc`, `deprecated`). Source-checked: `.agents/invocation.md`, `SCOPE.md`, `GLOSSARY.md`, `.out-of-scope/harness-name-collisions.md`, the skills `setup-matt-pocock-skills`, `code-review`, `tdd`, `retro`, `domain-modeling`, `research`, `to-spec`, `chief-of-staff`. Skill quality not measured. **Review:** current (reassessed 2026-10-09). **Revisit:** an upstream change to a skill DUDES carries (diff against the pin), or a change to the invocation or setup convention.

- **Context Drift Resistance:** 🟢 One skill per task; shared material lives in the skill that owns it; dependencies are explicit Skill-tool calls, not links across folders.
- **Context Compression:** 🟡 38–138 lines per skill. DUDES condenses its dev skills further for the worker budget (F76).
- **Token & Prompt Economy:** 🟢 Opt-in per task. Model-invoked descriptions carry trigger phrases; user-invoked ones (`disable-model-invocation: true`) stay one line for the human.
- **Bounded Worker Fitness:** 🟡 Written for a strong host model in a session. Several skills drive subagents and an issue tracker — work the DUDES engine, backlog and TODO own.
- **Git & Audit Offload:** 🟡 Its own changes go through changesets and a CHANGELOG; its specs and tickets go to an issue tracker, not Git.
- **Pillar 1 (consumer-box / normal PC):** pass — prompts only, any model.
- **Pillar 2 (KMU sovereignty):** pass — MIT, copied files you own (skills.sh) or a plugin.
- **Pillar 3 (DSGVO / EU AI Act / auditability):** N/A — no runtime.
- **DUDES Assessment:** **PATTERN / REFERENCE** — a content source for level skills and the source of conventions DUDES adopts.

**Use:** copy selected skills under DUDES names, the upstream revision in `metadata` (TODO): dev `tdd`, `code-review`, `codebase-design`, `diagnosing-bugs` (condensed for workers); spec `domain-modeling`, `prototype`; lead `research`; framework `retro`. Not `implement`, `implement-spec`, `triage`, `wayfinder`, `to-spec`, `to-tickets`, `ask-matt`: they run the process as prompts, which the engine, the backlog and `TODO.md` own here.

**Reusable ideas:** → §6 **Skill = Portable Discipline Unit**; **Two-Axis Review Gate**; **Glossary with Avoid list** (shipped F83); **Invocation Axis** and **Setup Writes the Repo's Adaptation** (hold).

---

## G · ADE desks

A desk sits *above* harnesses: it holds sessions or tasks, routes them to agents, and shows the work. DUDES does not build one (`ROADMAP.md` Horizon); the question is what a desk can show without owning the state.

### Traycer 🛰️

> **ADE desk over other coding agents:** it plans work as artifacts on disk (spec, tickets, stories, reviews), hands them to the agents you already pay for, runs those in parallel worktrees and verifies the result against the plan. The human picks each step in chat; task state syncs through Traycer's cloud.

- **DUDES Assessment:** **PATTERN / REFERENCE** (runtime rejected).

**Reusable ideas:** → §6 **Worktree-Isolated Worker** (setup goal), **Two-Axis Review Gate** (severity, focused re-verify), **Read-only Desk over the Record** (new, hold); anti-pattern sources **Agent↔Agent Channel Bus**, **User-Definable Lifecycle Graph**.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

### MonoCode 🗂️

> **Session shell over other coding agents:** one desktop window with a tab per agent session and one composer; with `/operator` the running agent drives the shell through a fixed CLI (`sessions.start`, `sessions.send`, `worktrees.create`, …).

- **DUDES Assessment:** **PATTERN / REFERENCE** — nothing to run; watch as a possible face. Self-described as very early.

**Reusable ideas:** → §6 **Read-only Desk over the Record** (source); anti-pattern source **Agent↔Agent Channel Bus**.

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

### paperclip 📎

> **Company desk over agents:** org chart, tickets, goals and budgets; agents "wake on heartbeats" to claim tickets, a board approves, budgets hard-stop. "Not an agent framework … running companies of agents rather than building them."

- **DUDES Assessment:** **PATTERN / REFERENCE** — the closest outside picture of the agentic company environment (`ROADMAP.md` Horizon).

**Reusable ideas:** → §6 **Heartbeat → Bounded Run** (new, hold), **Run Budget Hard-Stop** (source); **Read-only Desk over the Record** (source, inverse); anti-pattern **Session as Record** (new).

*Legacy entry — no evidence record. The full essay stayed in the private archive; reassess with `skills/framework/assess` before relying on it.*

### Ivy Tendril 🌿

> **ADE desk over CLI coding agents:** a plan moves through a fixed set of states, implementation runs in a git worktree, and build, format, and tests have to pass before the plan is offered for human review.

- **Context Drift Resistance:** 🟡 Each plan is a folder of metadata, revision markdown, and verification logs. A job is a new promptware run. Promptware memory is notes that last across runs, a stopped job keeps its worktree so it can be resumed, and a terminal chat injects into the running agent.
- **Context Compression:** 🟡 A revision holds problem, solution, tests, and a verification checklist. A plan starts from a short brief, an inbox file, or a GitHub issue. No repo map and no size budget.
- **Token & Prompt Economy:** unknown — select the stages and concurrency to evaluate; recording cost does not establish a hard budget stop or a local efficiency result.
- **Bounded Worker Fitness:** ❌ Tendril owns the jobs. Any terminal CLI can be the worker, OpenCode included. The documented stack is hosted CLIs. This is not an invoke that Go would call for one phase.
- **Git & Audit Offload:** 🟡 Execution uses a git worktree, then a pull request from that diff. The state of record is the plan directory under the Tendril home, plus job logs and raw agent output kept outside the plan folder.
- **Pillar 1 (consumer-box / normal PC):** unknown performance — the project documents support for any terminal CLI, including OpenCode (T5). Local operation is not limited to that one integration; verify the chosen worker and resource use.
- **Pillar 2 (KMU sovereignty):** weak — plans stay on local disk. The license is source-available (Functional Source License). A GitHub inbox and a tunnel for remote steering are built in.
- **Pillar 3 (DSGVO / EU AI Act / auditability):** unknown — verification and human review are documented; retention and reconstructability need checking. A review state can carry oversight evidence without being a Git commit.
- **DUDES Assessment:** **PATTERN / REFERENCE** (runtime rejected).
- **Use:** take one rule: after implementation, the configured build, format, and test commands run, failures go back with logs, and a broken plan does not enter review. Run nothing. The fixed state names are the app's enum. Each state is still a prompt with its own memory, so this is not a second state machine to copy. Fails if the session is the record: promptware memory and the terminal chat carry notes across runs, and the lifecycle lives outside Git.

**Reusable ideas:** → §6 **Framework-Owned Execution** (source). Anti-pattern sources **Process-as-Prompt**, **Session as Record**.

### Kandev 🧰

> **ADE desk over CLI coding agents:** a Go control plane runs many agent CLIs in parallel, isolates each task in a git worktree, and puts review in a workspace where a person decides what ships.

- **Context Drift Resistance:** ❌ A sub-task resumes the parent task's session. Conversations can be resumed and reviewed. Worktrees isolate files. They do not reset context.
- **Context Compression:** 🟡 The workspace shows a file tree, a diff, and a terminal. No repo map and no size budget.
- **Token & Prompt Economy:** unknown — configured agents, concurrency and triggers determine spend; hosted or parallel modes are not mandatory. Undocumented experimental modes are not scored as current behavior.
- **Bounded Worker Fitness:** ❌ Kandev is the process owner: one Go process, a web UI, and executors (local, Docker, SSH, Kubernetes, a cloud runner). OpenCode and Pi are on the agent list. Adopting the desk would replace the phase machine.
- **Git & Audit Offload:** 🟡 One worktree per repo on a task, per-repo branches, and a changes panel. Durable state is SQLite by default, PostgreSQL if configured. The event bus is in-memory, or NATS when configured, and it does not replay. State is recovered from the database.
- **Pillar 1 (consumer-box / normal PC):** unknown performance — native local operation and configurable agents are documented (T4). Hosted and remote options are not mandatory; measure the selected local configuration.
- **Pillar 2 (KMU sovereignty):** pass on the default path — AGPL-3.0, self-hosted, no telemetry, local database. Issue trackers, cloud executors, and a tunnel for a phone are optional exits.
- **Pillar 3 (DSGVO / EU AI Act / auditability):** unknown — human review is documented; verify attribution, retention and reconstruction. Database storage is not itself a failure. DUDES still requires its own Git-native evidence and outcome boundary.
- **DUDES Assessment:** **REJECTED**.
- **Use:** no runtime adoption. Its worktrees overlap an existing pattern, but UI diff review is not automatically the DUDES seal. The documented workflows and task messaging would be a second process owner if adopted as the lifecycle. Revisit only for a bounded integration or a material improvement that preserves engine-owned gates; YAML, a database or optional messaging alone is not the rejection.

**Reusable ideas:** anti-pattern sources **User-Definable Lifecycle Graph**, **Agent↔Agent Channel Bus**, **Session as Record**. The in-process event bus is UI fan-out, not the agent channel. The agent channel is the task MCP.

### Ouijit 🃏

> **Session shell over CLI coding agents:** each kanban card gets a git worktree and a terminal, dragging a card runs a hook, and the agent in that terminal can move the card.

- **Context Drift Resistance:** ❌ Quitting saves the session. The next launch restores terminals, worktrees, and panels. The agent's prompt is the task description, plus any files attached to the card.
- **Context Compression:** 🟡 The terminal shows the worktree diff against its merge target. A named project instruction can regroup one diff. There is no repo map.
- **Token & Prompt Economy:** 🟡 The app is a board and a terminal host. Parallel cards each launch a harness. Supported harnesses are hosted CLIs plus Pi and OpenCode.
- **Bounded Worker Fitness:** ❌ It is a desktop UI. A CLI on the terminal PATH speaks JSON, and the harness is taught that CLI. The agent uses it to create tasks, set status, and set hooks. That is the agent editing the board.
- **Git & Audit Offload:** 🟡 Start creates a worktree: a copy-on-write clone that keeps gitignored files, or a clean worktree. Diffs, line notes, and a pull-request surface are there. The board, hooks, and session are SQLite.
- **Pillar 1 (consumer-box / normal PC):** unknown performance — local app with documented macOS/Linux builds (T6). Check the actual host and worker; optional hosted harnesses do not establish mandatory cloud use.
- **Pillar 2 (KMU sovereignty):** pass on the default path — AGPL-3.0, local database, no sign-in. GitHub and Linear are optional.
- **Pillar 3 (DSGVO / EU AI Act / auditability):** weak — lane hooks exist. The agent can set its own card status. Nothing proves a gate ran, and the trail is not Git.
- **DUDES Assessment:** **PATTERN / REFERENCE** (runtime rejected).
- **Use:** take one clause on the existing setup goal: a copy-on-write clone that keeps gitignored dependencies is still not a second mechanism. The engine runs a fixed make goal in the worktree. Run nothing. A lens on a diff is a review prompt, already covered by scoped standards, so it is not a new pattern. Fails if the session is the record: the next launch restores the terminals, and the agent moves its own card. The board is a second database.

**Reusable ideas:** → §6 **Worktree-Isolated Worker** (setup-goal source). Anti-pattern source **Session as Record**.

## H · Self-improvement (RSI)

A method here changes something from the evidence of its own runs: a prompt or workflow, harness code, weights, or the optimizer itself. The question is less what it changes than who judges. DUDES already has a judge no worker can move — gates and the bench ladder's hidden tests (F65, F82) — and the human seal over every change (`CORE.md` §12, §15). The trend skill keeps the market list; an entry here is one method assessed.

### MASS 🪞

> **Recursive self-improvement method:** one open model searches a team workflow for a task at prompt level, then fine-tunes itself on the best runs of the winning workflow, and the next cycle starts from the tuned model. The same model is solver, workflow optimizer and in-loop judge.

**Evidence / review:** §4 T12 (code at `fb40358`, source-checked and documented), §4 T13 (paper and page, author-reported). Configuration assessed: the published pipeline (`configs/paper.json`) — nothing DUDES could run as it is. Nothing measured here. **Review:** current for that scope. **Decision:** pattern only; the runtime fails Pillar 1, and the code has no licence. **Revisit:** a licence for the code; a documented single-GPU or Ollama path; an independent reproduction. **Next check:** none for this tool; the §6 pattern waits for the dev stage (`ROADMAP.md` `E1`).

- **Context Drift Resistance:** 🟡 Each run starts in an empty workspace (`runtime/run_episode.sh` refuses a non-empty one), and the search history is files on disk. Inside a run, one qwen-code session orchestrates the whole team; the order is what the prompt says (`tasks/teacher_directive.txt`).
- **Context Compression:** 🟡 The judge sees capped evidence bundles (code 10k, JSON 50k tokens, `mass/judging.py`). The executor gets the whole task and a 262k window.
- **Token & Prompt Economy:** ❌ Per task: 11 search iterations of full runs, 18 collected runs, six pairwise judgements per run with up to 49k thinking tokens each. The reported per-token gain excludes that search (T13).
- **Bounded Worker Fitness:** N/A — not a worker. Its executor (qwen-code in yolo mode, delegation inside one session) would not fit the invoke contract (F20).
- **Git & Audit Offload:** 🟡 Stage outputs, verdicts with rationales and a byte-faithful log of every model request stay on disk. No Git, and the repository states that no complete ledger links the reported results to their configurations.
- **Pillar 1 (consumer-box / normal PC):** fail for the cycle — vLLM serving and six H100s for one fine-tune (T12). The prompt-level search alone is unknown on a 32 GB box: hours per run, an OpenAI-compatible endpoint; Ollama is not a documented backend.
- **Pillar 2 (KMU sovereignty):** weak — `--require-local-vllm` keeps search and judge on loopback (source-checked). The reporting stage sends workspaces to OpenAI and Anthropic. Without a licence, a company cannot reuse the code.
- **Pillar 3 (DSGVO / EU AI Act / auditability):** weak — every request is logged and every verdict carries a reason, but the judge of record is the model grading itself, and generated shell commands run unsandboxed.
- **DUDES Assessment:** **PATTERN / REFERENCE** (runtime rejected).

**Use:** run nothing. Take the inner loop's shape (propose, run, compare, keep a champion) into the micro-agent world, but search the engine's mechanism settings, never the worker's prompt. The bench ladder's hidden tests judge instead of the model, and a variant is adopted only through the human seal (§6 **Bench-Judged Workflow Search**). The research results do not transfer as evidence: on coding benchmarks the tuned model is slightly *worse* than its base (T13). Fails the moment the loop may edit the tests, fixtures, gates, frozen spec or seal it is judged by (§4 erosion guard).

**Where RSI fits (`CORE.md` §2.4):** the **micro-agent world** — local open weights, a verifiable judge, runs the engine already records; the engine's mechanism settings first, never the worker's prompt, weights only after LoRA on the box is measured. The **framework** already runs a loop of its own (retro, bench, trend, assess and the roadmap's step back, sealed by the human); MASS adds nothing to run there. The **role-play** is the poorest fit: hosted models without weights to tune, text work no test settles, and a same-model judge is exactly the risk the authors name.

**Reusable ideas:** → §6 **Bench-Judged Workflow Search** (new, hold); **Cross-Model Review** (source: shared blind spots of a self-judging model); anti-pattern source **Process-as-Prompt** (the team's order lives in the orchestrator's prompt).

---

# 6. Reusable Architecture Patterns

**This index is the living pattern status** (`shipped Fxx` / `planned` / `hold`). FEATURES.md owns capability IDs. Essays below explain the pattern once.

The tool assessments reveal capabilities that may be useful even when the original tool itself is not adopted.

### Index

Patterns are listed by **theme** so a new tool's ideas land next to their siblings instead of at the end. Source = the §5 entry that surfaced the pattern. The last column is the overview of how a pattern is built: its substrate today and, where the ladder in `CORE.md` §11 offers a simpler one, the step down — `keep`, `simplify`, `review` (two mechanisms for one job), `hold`, `do not build`, `never`. Simplify and drop candidates are scheduled in `TODO.md`.

| Theme | Pattern | Source | Status | Built with → simplest (`CORE.md` §11) · action |
|---|---|---|---|---|
| **Process & State** | Explicit State Machine | LangGraph · bernstein | shipped F3 | Go · keep |
| | Context Reset | LangGraph · OpenCode · GSD Core · BMAD · Archon · ralphex | shipped F21 | Go · keep |
| | Artifact Handoff (not chat handoff) | CrewAI (inverse) · Kiro · BMAD · Orbit | shipped F11 | files + git · keep |
| | Durable Phase Runner (lightweight first) | Temporal | shipped F4 | git notes (`refs/notes/dudes/run`, F4) · keep |
| | Run Budget Hard-Stop | MartinLoop · paperclip | hold | Go counter from worker usage, recorded as git notes → `waiting_decision` (F66) · hold, low priority locally |
| | Heartbeat → Bounded Run | paperclip | hold (Horizon) | timer (cron / systemd) + `make` or `dudes run` · hold |
| | Read-only Desk over the Record | Traycer (inverse) · MonoCode · paperclip (inverse) | hold — do not build (Horizon) | — · do not build |
| | Framework-Owned Execution | DUDES CORE (vs. OpenHands) · bernstein · Archon · ralphex · Ivy Tendril · toryo | shipped F33 · completion signals hold | `make test` when documented (F75) · keep |
| | Bounded Worker Invoke Contract | OpenHands · OpenCode · MartinLoop | shipped F20 | Go adapters · keep |
| | Recipe-Scoped Worker Invoke | Goose | shipped F27 | JSON recipe + Go · keep (+ `goals`, F75) |
| | Skill = Portable Discipline Unit | mattpocock/skills · Superpowers · BMAD · Pi | shipped F48 | Go inlines, opt-in only (F76) · keep |
| | Glossary with Avoid list | mattpocock/skills | shipped F83 | `CORE.md` §5 table (term · meaning · avoid) + skill-name tests · keep |
| | Invocation Axis (user- vs model-invoked) | mattpocock/skills | hold | frontmatter flag · hold — DUDES splits by world first (engine, role-play, framework); framework skills are already on-request only |
| | Setup Writes the Repo's Adaptation | mattpocock/skills | hold | instance `AGENTS.md` section read by framework level skills · hold — the company adapts, never copies |
| **Specs & Gates** | Delta Specs | OpenSpec | never — `git diff` between freeze branches is the delta | — · never |
| | Requirement/Scenario as Mechanical Gate | OpenSpec | shipped F37 | Go · keep |
| | Change-Scoped Folder + Archive | OpenSpec | never — git is the archive (commit trailers, F71) | — · never |
| | Cross-Artifact Consistency Check | Spec Kit | shipped F38 · F69 | Go · keep |
| | Converge-to-Tasks Loop | Spec Kit · Kiro | shipped F39 (report only — the re-analysis loop is Convergence Passes) | Go (report file) · keep |
| | Convergence Passes (re-analyse until no outcome) | `rules/CONVERGENCE.md` (DUDES-native) · Archon (`until` + iteration limit) | shipped F72 (toggle) | Go (toggle) · keep |
| | Two-Axis Review Gate | Superpowers · mattpocock/skills · Traycer | shipped F38 · severity + focused re-verify hold | Go · keep; severity = QR text |
| | Regression Guard (undo a failed pass that breaks green) | DUDES-native (replay of real product slices) | shipped F86 | Go + git revert · keep |
| | Cross-Model Review | ralphex · MASS | hold — configuration only (F27) | recipe config · hold |
| | Reference-Resolved Item Context | the first instance's spec practice · BMAD story files (inverse) | shipped F68 · F71 | Go · keep |
| | Derived Status (computed, never stored) | the first instance's spec practice | shipped F70 | Go (computed) · keep |
| **Context & Output** | Repository Map / Phase-Scoped Pack | Aider · Repomix · Orbit | shipped F24 · size gate F73 | Repomix via Go · keep, opt-in only: a code map, not the spec context of the item pack (F68) |
| | Schema-Validated Control Output | PydanticAI · Outlines / Instructor | shipped F25 (lite) | Go · keep |
| | Delimited Artifact Extract | DUDES-native | shipped F25 | Go · keep |
| | Stop Report, Not Invention | DUDES-native (`work-item` skill) | shipped F74 | Go · keep |
| | In-Phase History Compaction (secondary) | OpenHands · OpenCode | shipped F28 | Go · keep: the safety net for blobs; the context budget (F73) trims spec references by meaning |
| | Scoped Standards Injection | Agent OS · Kiro | shipped F49 | Go inlines, opt-in only (F76) · keep |
| | Minimal Tool Belt per Phase | Pi · lean-coder | shipped F56 | Go · keep |
| **Git, Audit & Isolation** | Atomic Git Worker | Aider | shipped F23 | Aider · keep |
| | Orchestrator-Owned Commits | OpenCode · ralphex · DoneGate (inverse) · lean-coder (inverse) | shipped F42 | git via Go · keep |
| | Evidence Surfaces for Oversight | Pillar 3 (DSGVO / AI Act) · MartinLoop | docs F45 | QR + git · keep |
| | Machine-Readable CLI Surface | OpenSpec · Spec Kit | shipped F47 | Go · keep (also the face's interface) |
| | Event-Sourced Worker Replay | OpenHands · Pi | shipped F29-lite | `.dudes/events.jsonl`, a disposable diagnostic log (`CORE.md` §10) · keep |
| | Git Notes as Side Record | aGiTrack | shipped F67 (test runs) | git notes · keep |
| | Worktree-Isolated Worker | Spec Kitty · OpenCode · Superpowers · Traycer · bernstein · Archon · aGiTrack · Ouijit | shipped F54 · setup goal hold | git worktree · keep; setup = make goal |
| | Ephemeral Execution Sandbox | OpenHands | shipped-lite F46 (cwd/env; not full host isolation) | worktree + cwd/env · keep; pinned worker image is toolchain provisioning (D2), whole-ADE isolation separate |
| **Infrastructure** | Local Model Gateway | LocalAI | shipped F26 | external server · keep |
| **Self-improvement** | Bench-Judged Workflow Search | MASS | hold | a variant is a mechanism setting on a branch, never worker prose, scored by `make bench-models` (F82) into git notes · hold until the dev stage runs real items (`ROADMAP.md` `E1`) |
| **Anti-patterns** | Agent↔Agent Channel Bus | CrewAI · Traycer · MonoCode · Kandev · toryo · lean-coder | hold | — · never |
| | User-Definable Lifecycle Graph | Spec Kit · Traycer · Archon · Kandev | hold | — · never |
| | Process-as-Prompt | BMAD · GSD Core · Superpowers · Ivy Tendril · toryo · MASS | hold | — · never |
| | Session as Record | Traycer · MonoCode · paperclip · Ivy Tendril · Kandev · Ouijit · lean-coder | hold | — · never |

---

### Pattern: Explicit State Machine

**Source:** LangGraph; **bernstein** (every transition validated by an FSM; the same plan replays to a byte-identical task graph)

Use explicit phases and transitions rather than hidden conversational state.

Potential DUDES structure:

**SPEC → DEV → QA → REPORT → HUMAN DECISION**

Each transition should be observable, persistable and auditable (Pillar 3).

### Pattern: Context Reset

**Source:** LangGraph (mechanics); CrewAI theory (small role tasks); **OpenCode** (fresh session per phase); **GSD Core** / **Superpowers** (fresh-context subagent per task — "context rot"); **BMAD** (one story per Build session)

Do not force every role to carry the entire conversation history.

Instead:

- persist required state in **files** (not pasted chat blobs);
- commit at meaningful boundaries;
- start the next role/phase with only the context it needs;
- re-read the authoritative specification directly;
- **new worker session/process per DUDES phase** (do not reuse long SQLite/chat sessions across phases).

This is what CrewAI *intends* with role tasks — without the token-heavy agent↔agent copy. Keeps consumer-box local models effective (Pillar 1).

### Pattern: Artifact Handoff (not chat handoff)

**Source:** CrewAI failure mode (negative); DUDES / LangGraph positive; **Kiro** `requirements.md` → `design.md` → `tasks.md`; **BMAD** brief → PRD → `SPEC.md` → `stories.yaml`

When phase N finishes, write a durable artifact. Phase N+1 reads that artifact (and the PRD), never a transcript dump of phase N’s inner monologue.

**Orbit** can keep a knowledge repo beside the code, read as files. An agent note written for the next session is not that handoff. The next phase reads a sealed artifact.

### Pattern: Delta Specs (write the diff, not the document)

**Source:** **OpenSpec** (`## ADDED / MODIFIED / REMOVED Requirements`)

On cycle 2+ (after a human reject), do **not** ask the model to regenerate the whole PRD. Have it emit only the requirement diff; the orchestrator merges mechanically. This is the same bet as Delimited Artifact Extract: **small output from a weak model, deterministic merge in Go**.

Why it fits DUDES specifically:

- a frozen spec stays frozen — a delta is applied *to* the freeze rather than replacing it, so the freeze record keeps meaning across cycles;
- reject feedback maps naturally to `MODIFIED` requirements instead of a full rewrite that silently drops earlier agreements;
- the diff is the reviewable unit for the human gate.

**DUDES decision (never):** not built. Instance specs are human-owned and change on `development`; `git diff` between two freeze branches already is the reviewable delta, and no reject cycle has needed a model-written diff (YAGNI).

### Pattern: Requirement/Scenario as Mechanical Gate

**Source:** **OpenSpec** (`SHALL` requirement + given/when/then `#### Scenario:`); **Spec Kit** `/speckit.checklist` ("unit tests for English")

Give the spec a shape the orchestrator can check without an LLM. A gate then asserts structure, not just non-emptiness: *does the PRD contain ≥1 requirement, and does every requirement carry ≥1 scenario?*

This also supplies the missing half of PRD-to-test traceability (`rules/TESTING.md` §2): a named scenario is the thing a test can cite.

### Pattern: Change-Scoped Folder + Archive

**Source:** **OpenSpec** (`changes/<name>/` → `changes/archive/YYYY-MM-DD-<name>/`)

One unit of work = one folder holding proposal, design, tasks and spec deltas. On completion the deltas fold into the specs and the folder is archived with a date stamp — the specs then describe the new reality and the history stays inspectable.

**DUDES decision (never):** no archive folder. Git is the archive — each item-bound commit names the item and the hash of every source document as trailers (F71), and packs are disposable build artifacts. Still worth stealing: completion is valid only when **every DoD row is green** — DUDES derives that from the QR (F70) instead of ticking boxes.

### Pattern: Cross-Artifact Consistency Check

**Source:** **Spec Kit** `/speckit.analyze` (run after tasks, before implement)

Before implementation, verify the artifacts agree with each other — spec ↔ design ↔ tasks — rather than only verifying that each file exists. Existence gates catch a missing PRD; they do not catch a TA that contradicts the PRD, or a requirement with no corresponding task.

Note the DUDES-native constraint: this check should stay **mechanical wherever possible** (cross-reference resolution, coverage counting), reserving the LLM for judgement that genuinely needs it. Compare `rules/CONVERGENCE.md` §Structural Checklist — the same checks run by an IDE role. The engine runs them as F38 (requirement ↔ test coverage) and, for instances with a profile, F69 (every reference resolves, every ID defined once by its owner) as a gate before freeze.

### Pattern: Convergence Passes (re-analyse until no outcome)

**Source:** DUDES `rules/CONVERGENCE.md` (role-play), realized in the engine as F72

A worker that reports "done" after one pass is not proof the task is done. Re-invoke it with a **fresh context** that re-reads the task from the beginning, compares it with the current state and does only what is missing. The loop ends when a pass changes nothing — an observable exit, not a self-assessment. Bounded, and a toggle: strong models often finish in one pass, small local workers often do not (Pillar 1). Each changing pass is its own commit, so git shows how the task converged.

### Pattern: Stop Report, Not Invention

**Source:** DUDES `work-item` skill

A worker that lacks an input stops with a one-line report (`BLOCKED:` / `CONTRACT:`) instead of guessing. The orchestrator parses it like any delimited output and turns it into a durable human decision; the answer goes into the next fresh invoke. Guessing is the failure that review catches last; a stop report costs one human sentence.

### Pattern: Reference-Resolved Item Context

**Source:** the first instance's spec practice (one owner per fact, `[[DOC:ID]]` references in backlog items); **BMAD** story files as the inverse (context *copied* into the story)

A work item names the facts it needs instead of restating them. The engine resolves each reference to its definition in the owner document and hands the worker the item plus those original definitions — a few percent of the spec-set, with no paraphrase in between. Copying context into the task (BMAD-style) is the drift this avoids: a copy goes stale the moment the owner changes. An unresolved reference means the item is not ready; a manifest of source hashes makes a stale pack detectable. Format and ownership stay instance data (a profile), so the engine does not impose one spec style.

### Pattern: Derived Status (computed, never stored)

**Source:** the first instance's spec practice

Item status is a function of evidence — the QR rows named in the item's DoD — not a field someone edits. `done` cannot drift from the evidence because nothing stores it. Being worked on is visible from the open PR or run, not from the backlog.

### Pattern: Converge-to-Tasks Loop

**Source:** **Spec Kit** `/speckit.converge` (assess codebase against spec/plan/tasks → append remaining work as new tasks → repeat until "Converged"); **Kiro** bugfix spec (current / expected / unchanged behaviour as the shape of a repair brief)

DUDES status: F39 emits the report; the loop itself is **Convergence Passes** below.

A convergence check is more useful when it **emits the remaining work** than when it merely reports pass/fail. Findings become tasks; the implement step consumes them; the loop repeats until a pass produces nothing new — the mechanical twin of `rules/CONVERGENCE.md` ("loop exits only when a full re-analysis produces zero findings").

Keep the existing guardrails: the loop must terminate (bounded iterations), and convergence never implies acceptance — the human seal stays the outcome boundary.

### Pattern: Repository Map / Phase-Scoped Pack

**Source:** Aider (repo map); **Repomix / code2prompt** (pack + Tree-sitter/`--compress` + templates + token count)

Feed the worker a **compact** repo view — map or phase-scoped pack file — not the whole tree. Go may **gate** on reported token count before calling the LLM. Prefer narrow globs per phase.

**Orbit** loads a code map in three steps: a one-line roster, a short memo, then the full tree only for a repo the worker adds. That ladder is a code map across repos, not the spec pack (F68).

### Pattern: Recipe-Scoped Worker Invoke

**Source:** **Goose** Recipes (YAML/Jinja; optional sub-recipes)

Version a phase as a declarative recipe (instructions + params + allowed tools). Orchestrator runs **one recipe per phase** and exits — do not grow an unbounded interactive session.

### Pattern: Atomic Git Worker

**Source:** Aider

A worker may:

1. receive a clearly bounded task;
2. modify the required files;
3. verify the result;
4. create an atomic commit;
5. return control to the orchestrator.

### Pattern: Orchestrator-Owned Commits (snapshots ≠ release)

**Source:** **OpenCode** (session/git snapshots & undo without touching user branches); **DoneGate** (inverse — a worker commit is accepted only after an independent `HEAD` read); **lean-coder** (inverse — compaction tells the model to commit docs)

Agent-local snapshots/undo are fine for recovery **inside** a phase. **Release / phase Git commits stay with DUDES Go** (or an explicitly chosen Atomic Git Worker like Aider). Do not confuse worker undo history with the auditable product trail.

**DoneGate** lets the worker commit, then accepts the task only when an independent `HEAD` read matches the claim, the commit descends from the recorded base, and the paths sit in the allowlist. DUDES does not take that route (F42): the worker is started so it cannot commit, and a moved `HEAD` is a warning. **lean-coder** tells the model, during compaction, to update project docs and commit them. That commit is the model's.

### Pattern: Schema-Validated Control Output

**Source:** **PydanticAI**; **Instructor** / instructor-go / structllm; **Outlines** (constrained decode: JSON/regex/grammar at generation time)

One family, two mechanisms — document once:

| Mechanism | Typical tools | Use in DUDES |
|-----------|---------------|--------------|
| Schema → generate/validate/repair | PydanticAI, Instructor, Go ports | Small **control** payloads (file lists, phase status, checklists) |
| Constrained decode | Outlines (+ vLLM/llama.cpp) | Same, when the runtime can enforce structure at token level |

Prefer **Go-native** validation in the orchestrator. **Not** for free-form code bodies — see Delimited Artifact Extract.

### Pattern: Delimited Artifact Extract (not tool-JSON theatre)

**Source:** local-LLM practice / DUDES context engineering (complements Schema-Validated Control Output)

Weak models often fail at complex tool-call JSON. Prefer: model emits **plain text or code** between clear delimiters; **Go parses** the block into the artifact. Reserve schemas/constrained decode for **tiny** control payloads — not the whole coding loop.

### Pattern: Local Model Gateway

**Source:** LocalAI

Treat local model serving as infrastructure below the DUDES process rather than embedding model-serving concerns into the process itself (Pillars 1–2).

### Pattern: Evidence Surfaces for Oversight

**Source:** DSGVO / EU AI Act pressures (Pillar 3); **MartinLoop** ("VERIFIED" claims only the checks that ran; receipts without prompts or source)

Evidence states exactly what was checked, never "bug-free". Integrity comes from Git (signed commits and tags), not from a second signing scheme.

Prefer designs that always leave:

- a human-readable decision artifact (e.g. QR);
- a reconstructable Git trail;
- an explicit human accept/reject boundary.

### Pattern: Machine-Readable CLI Surface

**Source:** **OpenSpec** (`--json` on every agent-facing verb; exit `0` success / `1` error); **Spec Kit** (`workflow run --json`, machine output on stdout, progress on stderr)

A tool that both humans and programs drive should answer in both registers: human text by default, `--json` on demand, and meaningful exit codes. Two details worth copying:

- **separate the streams** — machine payload on stdout, progress chatter on stderr, so a caller can pipe one without parsing the other;
- **distinguish "found problems" from "could not run"** — OpenSpec's `doctor` exits `0` even with findings (the caller reads the status array) and reserves `1` for genuine command failure.

The same discipline applies inward: an engine whose own `status` / `check` output is parseable can be scripted, tested and gated without scraping prose.

### Pattern: In-Phase History Compaction (secondary)

**Source:** **OpenHands** Condenser; **OpenCode** session compaction

**Prefer phase reset + artifacts over long in-process history** (see Context Reset; **OpenCode**: new session per DUDES phase). If one worker invoke still accumulates turns, compact/trim mechanically (drop raw compiler spam). Safety net inside a phase only — not a substitute for empty LLM context at phase N+1.

### Pattern: Event-Sourced Worker Replay

**Source:** OpenHands; **Pi** (tree-structured JSONL sessions with fork / branch — full history kept even after compaction)

Typed action/observation events enable deterministic replay and debugging of a single worker run. Useful inside a phase; still subordinate to Git phase commits owned by the orchestrator.

### Pattern: Bounded Worker Invoke Contract

**Source:** **OpenHands** headless CLI/`--json`; **OpenCode** CLI + HTTP/client API; **Goose** `goose run --recipe`; **MartinLoop** (objective, verifier, budget, path scope, stop condition per run)

Go `exec` (or HTTP): task/recipe in → machine-readable result/events out → process/session ends. Fresh context per phase (OpenCode: do not reuse SQLite session across DUDES phases).

### Pattern: Framework-Owned Execution

**Source:** DUDES CORE (mechanical gates); contrast OpenHands free-form bash

The LLM does **not** decide how to run the project’s tests or build. Go runs the known command, captures stdout/stderr, and — on failure — re-invokes a fix worker with **only the error text + relevant files**. Workflow stays outside the model.

**bernstein**'s janitor adds declarative completion signals per task — file exists, glob matches, tests pass, file contains expected content. The worker's claim never counts. DUDES form: the contract's acceptance list as shell checks (`test -f`, `grep -q`) plus the make test goal.

**Ivy Tendril** runs build, format, and test after the worker and does not move a failing plan to review. **toryo** blocks acceptance when a required check fails, even if a reviewer score would pass. When no check is configured, that score is the acceptance — Process-as-Prompt, not this pattern.

### Pattern: Worktree-Isolated Worker

**Source:** **Spec Kitty** (`.worktrees/`, work-package lanes, `next → review → accept → merge`); **Superpowers** `using-git-worktrees` (branch + clean test baseline before work)

Before reaching for containers, use what Git already provides: run the worker in a dedicated **worktree** so the main tree stays clean, the work is trivially discardable, and the diff is reviewable before it merges. Cheap, native, cross-platform — no daemon, no image build.

For DUDES this is the **pragmatic precursor to the Ephemeral Execution Sandbox below**: worktrees isolate *changes*, containers isolate *execution*. The first is nearly free and addresses the common failure (a worker scribbling on the live tree); the second addresses host safety and remains the heavier, later step.

Two further ideas worth noting without adopting the runtime: explicit **lifecycle lanes** on a unit of work (`for_review` → `approved` is a per-item human gate), and a **retrospective artifact** generated at the end of a cycle — a learning surface the current QR + seal pair does not produce. **bernstein** merges serially and only after verification passes.

**Traycer** adds a setup step per new worktree (install dependencies) and a teardown before removal. DUDES form: the engine runs a fixed make goal in each new worktree — the goal interface (`CORE.md` §13), no script registry. **Ouijit** keeps gitignored dependencies with a copy-on-write clone instead of a setup command. That clone is not a second mechanism. The make goal stays the setup.

### Pattern: Git Notes as Side Record

**Source:** **aGiTrack** (summaries in `git notes` namespaces)

`git notes` attach facts to a commit without changing its hash, in their own namespaces (`refs/notes/dudes/*`), and can be added after the fact — the test result once the gate ran, model and tokens of the phase, evidence references. DUDES records every test run this way (`refs/notes/dudes/test`); evidence is derived from the spec and these notes, so `evidence.json` and `last_test.log` are gone (F67). `events.jsonl` stays a disposable diagnostic log. Notes stay local unless a namespace is pushed on purpose.

### Pattern: Ephemeral Execution Sandbox

**Source:** OpenHands (Docker isolation); DUDES Go realization (later)

**Priority: nice-to-have / later — not an MVP starting point.** Per-step containers (spin up → run → kill via Go Docker API) help host safety once the orchestrator works. First build the sequential Go loop on the host (or one long-lived framework container). Do not block early development on sandbox plumbing; whole-framework-in-Docker is a deployment choice, not a substitute for phase reset + artifacts.

### Pattern: Skill = Portable Discipline Unit

**Source:** Agent Skills standard (`SKILL.md` + frontmatter); packs: **mattpocock/skills**, **Superpowers**, **BMAD**; native loaders: **Pi**, Kiro `inclusion: auto`

A skill is a small Markdown file that teaches one discipline (TDD, bug diagnosis, review) and is loaded only when the task fits. DUDES use: a recipe (F27) names skills; Go inlines the body into the worker prompt (compaction applies). Vendor pinned copies under framework ownership — no `npx skills` at runtime. Reuse existing packs for generic engineering discipline; write only what is DUDES-specific (delimited output, artifact paths, gate expectations). Skills carry **discipline**, never **sequence** — see anti-pattern Process-as-Prompt.

### Pattern: Scoped Standards Injection

**Source:** **Agent OS** (discover → index → inject only relevant standards); **Kiro** steering `inclusion: always | fileMatch | manual | auto`

Standards enter the prompt only when they apply: by phase, by touched path, or by description match. DUDES: the recipe (or contract task) names which rule files and skills a phase sees (F49 / F48). IDE roles still load `rules/*.md` from `AGENTS.md` — that world is not engine-scoped.

### Pattern: Two-Axis Review Gate

**Source:** **Superpowers** subagent-driven-development (spec compliance, then code quality); **mattpocock/skills** `code-review` (Standards axis, Spec axis)

Review is two separate questions — *does it implement the spec?* and *does it meet the standards?* — and mixing them lets one mask the other. DUDES: `dudes check` reports spec conformance (gate) and standards (lint / tests) as distinct results in the QR.

**Traycer** verification adds two small ideas: findings carry a severity (Critical / Major / Minor / Outdated — superseded by a later change), and a re-run can check only the open findings instead of everything. Tests stay the gate; an LLM comparing code with the plan stays a review axis, never the pass/fail.

### Pattern: Cross-Model Review

**Source:** **ralphex** ("codex-reviewing-codex is a same-model self-review with weak signal"); **MASS** (its authors name shared blind spots of a self-judging model as a risk for recursive self-improvement; its in-loop judge is the model itself, and only the reporting uses other vendors)

A reviewer from the same model family shares the implementer's blind spots. DUDES form, configuration only: the recipe of the review / QA phase names a different local model than the implementation phase (F27). Needs a second model on the box; tests stay the gate.

### Pattern: Minimal Tool Belt per Phase

**Source:** **Pi** (`--tools read,grep,find,ls`; four default tools, extensions opt-in); **lean-coder** (tests cap the always-on prompt plus core tool schemas; extra tools stay off)

Tool schemas cost tokens on every turn. Give a phase only the tools it needs — read-only for review and spec phases, edit for implementation. Contrast the full belts of OpenHands / Goose / OpenCode (§5 B), which are part of their consumer-box tax.

**lean-coder** pins the always-on prompt plus the core tool schemas under a token ceiling its own tests enforce, and leaves every extra tool off until enabled. The ceiling is theirs. The rule is this pattern: the belt stays small enough to count (F56).

### Pattern: Read-only Desk over the Record

**Source:** Traycer (inverse); **MonoCode** (a local desk driven by a fixed CLI); **paperclip** (inverse — a board with its own database)

A desk that shows boards, running work and diffs is the "face" of the `ROADMAP.md` Horizon — not built by DUDES. The rule for any desk that may come as open source: it **reads** Git and the derived status (`dudes item board`, F70) and writes only through the engine. A desk that owns task state (Traycer: host + cloud sync) becomes a second database and cannot be that face. A desk that already drives tools through a fixed CLI (MonoCode) could drive `dudes` (F47) the same way: show the board, send decisions.

### Pattern: Bench-Judged Workflow Search

**Source:** **MASS** (inner loop: propose a workflow, run it, compare, keep a champion, feed the verdicts' reasons into the next proposal)

MASS searches the prompt that organizes a team and lets the same model judge the result. The DUDES form keeps the search, swaps the judge, and changes what is searched. A variant is a mechanism the engine owns: a `dudes.json` setting (steps, edit format, context, convergence passes), the contract's split into tasks and slices, or the order of the gates. It is never the worker's instructions or a recipe's prose. A worker gets the task, its goals and the output contract (`CORE.md` §2.4), and prose tuned to carry a weak model is the context cage DUDES does not build: the mechanisms are the cage, and better models arrive. A variant is a commit on a branch. The bench ladder runs it from fresh fixtures (F82), and its hidden tests, which no worker sees, decide; the result is a git note like any bench run. The incumbent changes only through the human seal (`CORE.md` §15). A model may propose the next variant from the notes, never edit the ladder, its fixtures, the gates or a frozen spec. Weights (MASS's outer loop, a LoRA on the best runs) stay out until fine-tuning on the box is measured. Comparing two settings by hand is already `make bench-models` (`EDIT_FORMAT`, `STEPS`, `NUM_CTX`). Hold until the dev stage runs real items (`ROADMAP.md` `E1`) and a mechanism setting is in question there.

### Anti-pattern: Agent↔Agent Channel Bus

**Source:** CrewAI failure mode (negative); **Traycer** agent-to-agent debate and peer-review loops; **MonoCode** `/operator` (an agent starts sessions and messages them); **Kandev** task MCP (one task messages another and reads its conversation); **toryo** opt-in bus (streamed tokens, fan-out, a slot for a network transport — the default cycle does not use it); **lean-coder** `dispatch_worker` (a background agent that can be saved and resumed)

Do **not** model DUDES as goroutine-“agents” passing `chan Task` chat blobs to each other (CrewAI in Go). The orchestrator is a **sequential phase machine**: phase N writes artifacts + Git; phase N+1 reads files. Channels/goroutines may implement timeouts, I/O, or sandbox lifecycle **inside** the orchestrator — they are not the product-cycle handoff protocol.

### Anti-pattern: User-Definable Lifecycle Graph

**Source:** **Spec Kit** workflow YAML (`command` / `shell` / `gate` / `if` / `while` / `fan-out` steps, defined per project); **Archon** (workflow YAML per project, copied and edited under `.archon/`); **Traycer** ("Skip any step the work doesn't need", the human picks the next skill); **Kandev** (the whole workflow exported as portable YAML, agents and gates included)

Spec Kit's step vocabulary is genuinely good, and that is precisely the temptation: it invites making the DUDES phase graph a config file so each project can shape its own process.

Do **not**. `CORE.md` §4 puts the lifecycle in the engine so the model cannot invent the graph — and a user-editable graph is the same hole with a nicer syntax, because whoever edits it (human or agent) can delete a gate. The line to hold:

| Layer | Configurable? |
|---|---|
| Phase graph / order / gates | **No** — engine-owned |
| What a phase *tells its worker* | Yes — recipes (F27) |

Declarative-per-phase is adoption; declarative-whole-lifecycle is erosion.

### Anti-pattern: Session as Record

**Source:** **Traycer** (context shared across models), **MonoCode** (a tab per living session), **paperclip** (agents resume the task context across heartbeats); **Ivy Tendril** (promptware memory across runs, resume through a terminal chat); **Kandev** (a sub-task resumes the parent session); **Ouijit** (the next launch restores terminals and worktrees); **lean-coder** (the last session for the working directory loads on start)

When the unit of work is a session that lives on across runs, drift accumulates and the record is a transcript, not Git. DUDES: a run is bounded, context resets at the phase boundary, and what a phase was based on lands in its commit (`CORE.md` §10). A trigger may start a new run; it never resumes an old session.

### Anti-pattern: Process-as-Prompt

**Source:** **BMAD**, **GSD Core**, **Superpowers** (methodology delivered entirely as skills / slash commands); the critique is stated plainly in the **mattpocock/skills** README; **Ivy Tendril** (each plan stage is a promptware: program, tools, and memory); **toryo** (a reviewer score, or a trust average, accepts the work when no required check is configured); **MASS** (the team's order — delegate, verify, re-delegate — is a directive in the orchestrator's prompt)

When the lifecycle exists only as instructions a host agent *may* follow, every gate is advisory: skipping is silent, order is a suggestion, and weaker models drop steps under load. It works for frontier hosts with a human watching each turn; it fails Pillar 1 (consumer-box) and Pillar 3 (nothing proves the gate ran).

This is the answer to "is BMAD our framework?": same lifecycle, different substrate. DUDES keeps sequence and gates in Go, and lets prompts — including reused skills — carry discipline only. The table above applies unchanged: skills may shape *what a phase tells its worker*, never *which phase comes next*.

**Ivy Tendril** keeps fixed state names and still runs each state as a promptware. **toryo** accepts work from a reviewer score, or from a trust average, whenever no required check is configured. The number looks like a gate. It is still the model's word. A failed check that blocks that score is Framework-Owned Execution, not a new pattern.

### Pattern: Run Budget Hard-Stop

**Source:** **MartinLoop** (budget in the run contract), **paperclip** (budget hard-stops per agent)

Count tokens per run; on overrun the run parks in `waiting_decision` instead of retrying. DUDES form: the engine sums worker usage per phase, records it as git notes (Git Notes as Side Record) and parks the run (F66). Locally the cost is time, not money, and F72 `max_passes` already bounds loops — low priority.

### Pattern: Heartbeat → Bounded Run

**Source:** **paperclip** (agents wake on heartbeats)

"It keeps running" (`ROADMAP.md` Horizon) means a schedule starts runs — a timer (cron, systemd) calling a make goal or `dudes run`. Each run is bounded and starts fresh; a heartbeat never resumes a session (anti-pattern Session as Record).

### Pattern: Durable Phase Runner (lightweight first)

**Source:** Temporal (reference); DUDES Go MVP

Record phase transitions, retries, and human seal/reject in **framework + Git** first. Full Temporal is optional later for multi-machine durability — not required to start.

---

# 7. Evaluation Template for Future Tools

Procedure, including reassessment: `skills/framework/assess`. The shape below is the catalog contract, not permission to change the engine.

1. **Classify** it into a §5 category (A–G). If none fits, add a category row to the §5 table — do not create a "misc" bucket.
2. **Add one row to the §5 Index** (tool · category · verdict · patterns). The index must stay a one-screen overview.
3. **Add the entry** as `###` under its category heading, using the structure below. The intro line follows the house style: *what kind of thing it is* first, then what it does — e.g. "Terminal pair-programmer: …", "Durable workflow engine: …". No versions, licences or runtime details in the intro; those belong in Bounded Worker Fitness / Pillars where they carry an argument.
4. **Register patterns** in §6 under their theme and add them to the §6 Index with a **Status** (`shipped Fxx` / `planned M#` / `hold`). Prefer extending an existing pattern's *Source* line over creating a near-duplicate. Research never changes a shipped capability's status; that follows `CORE.md` §15.
5. **Keep the evidence and review record with the decision.** Update in place; Git is its history. Every decision-relevant claim needs support, not just one URL for the whole tool.

### Evidence / review record

| Field | Required content |
|---|---|
| Identity and sources | Canonical owner/project or model identifier; primary links with commit/release where available; source date (or undated) and checked date. Mutable pages must say so |
| Configuration and scope | Default vs selected configuration, optional vs unavoidable behavior; exact claims checked and remaining unknowns |
| Evidence | `documented`, `source-checked`, `measured`, or `unverified` per claim (§4); link the immutable result for any measurement |
| Review | `current` for the supported scope, `stale` when a premise changes, `unverified` when evidence is missing. Independent of verdict and §6 feature status |
| Decision | Verdict, reason and Mission/Core boundary; on reassessment say retained or changed and why |
| Revisit / next check | Concrete release/configuration/evidence or local-failure trigger; the question or paired trial needed to resolve it |

For the current scoped baseline, §4's register supplies sources, checked date, evidence level and limits; its Review triggers supply the default next check. That does **not** mark every older essay current. New assessments record tool-specific conditions here or beside their entry, without copying the register.

### TOOL_NAME 🆕

> **What it is** (one classifying phrase): what it does, in two sentences at most.

**Evidence / review:** [record above, or reference to its §4 row plus tool-specific configuration, decision and revisit condition]

- **Context Drift Resistance:** [🟢 / 🟡 / ❌ / unknown / N/A — why and evidence]
- **Context Compression:** [🟢 / 🟡 / ❌ / unknown / N/A — why and evidence]
- **Token & Prompt Economy:** [🟢 / 🟡 / ❌ / unknown / N/A — why and evidence]
- **Bounded Worker Fitness:** [🟢 / 🟡 / ❌ / unknown / N/A — why and evidence; protocol only if relevant]
- **Git & Audit Offload:** [🟢 / 🟡 / ❌ / unknown / N/A — why and evidence]
- **Pillar 1 (consumer-box / normal PC):** [pass / weak / fail / unknown / N/A — evidence]
- **Pillar 2 (KMU sovereignty):** [pass / weak / fail / unknown / N/A — evidence]
- **Pillar 3 (DSGVO / EU AI Act / auditability):** [pass / weak / fail / unknown / N/A — evidence]
- **DUDES Assessment:** [Orchestrator / Worker Tool / Pattern / Reference / Infrastructure / Rejected; NOT EVALUATED when no assessment exists]

**Use:** [bounded invoke, pattern, or nothing; the configuration that preserves DUDES boundaries, or the evidenced reason it cannot]

Then record any reusable architecture pattern separately in **§6 Reusable Architecture Patterns**.

---

# Harness watch

> **Temporary decision baseline, reviewed 2026-10-09.** Not a new market survey, pillar or feature. The three pressures in `MISSION.md` §Strategic Operating Context stay the basis. The next `skills/framework/trend` landing replaces this block **with its supporting evidence and review conditions**. Never leave the only sources in chat. The former survey's source list is not preserved here and cannot support fresh market claims.

DUDES is the local-model ADE for the company product lifecycle. Sources for this baseline: §4 M1–M3 (documented model/server capability), D1 (stub benchmark limit), D2 (worker toolchain image), plus `CORE.md` §1, §10 and §15 (authority, not market evidence). Local models are compared in one harness (§1.1, §4 B2); no harness comparison is available, so the current worker is an incumbent, not a measured market leader.

| When | Action | By tool |
|------|--------|---------|
| **Now** | Size context and output for a bounded task; verify actual allocation, peak memory and latency. No universal 32k/64k floor is established. | Local Model Gateway (F26); M1–M2. Pillar 1, local self-containment |
| **Now** | Keep Aider as incumbent and keep commits with the engine. Its pinned image is a toolchain, not the whole-ADE sandbox. Risk: no release after 0.86.2 (PyPI 2026-02-12), its main idle since 2026-05-22, open advisories without a fixed version (`make vuln-scan`, checked 2026-10-12) — an unmaintained incumbent is a revisit trigger for the Next row. | Aider; D2, `CORE.md` §10. Pillars 1–3, engine-owned gates |
| **Now** | Evaluate a prompt/model change with the §4 paired trial; preserve failures and acceptance evidence. Do not use the stub benchmark for ranking. | Repository test goals; D1. Pillars 1 and 3, frozen oracle and human outcome control |
| **Next** | Done 2026-10-16 for replay slice 1: Pi 1.3.0 against Aider with `gemma4:26b` (§1.1) — Pi looped and edited read-only tests, Aider passed 1 of 3. The harness is not the lever with this model. Reassess an alternative again only with a model that drives a tool loop without repeating, or when Aider must be retired for its advisories. | Pi (F55/F56) or one assessed candidate. Pillar 1, fresh phase context and unchanged lifecycle |
| **Later** | Assess other harnesses or a verifier when evidence identifies a benefit. Separate worktree change isolation, worker toolchain images and whole-ADE host isolation. | §5 candidates and Ephemeral Execution Sandbox. Pillars 1–3, `CORE.md` §10 |
| **Never** | A worker owning DUDES transitions, hidden cross-phase conversation as authority, silently mandatory cloud, bypassable spec/test gates, or agent self-seal. Tool names are not permanent blacklists. | §6 anti-patterns. All pillars and Mission invariants |

**Review:** current as a conservative decision baseline, **unmeasured** as a comparison. **Revisit:** a changed model/server/harness configuration, failed task threshold, or documented compatible improvement. **Next check:** the paired worker trial in §4, with thresholds chosen for the actual product task before execution. The 3-month direction is configuration and evidence; the 1-year direction is swappable local workers, not a prediction of a particular winner.

---

# 8. Relationship to the Mission

TREND.md can influence implementation decisions, but it cannot redefine DUDES.

Each file's job and the path of an idea: `CORE.md` §14–§15. TREND is the landscape, a side input: it holds knowledge, never work. A pattern reaches the chain only when the human adopts it and it gets a `FEATURES.md` id.

`skills/framework/` say how a DevOps agent works. They are not a step in that chain.

| Stays in TREND.md (the link between specs) | May live only in a skill (the how) |
|--------------------------------------------|-------------------------------------|
| Market names, pillars, §5 category questions, Index verdicts | How to search, how long the window is, how to order a reply |
| §6 pattern status, including `shipped Fxx` | How to tell a catalog spec from a vendor tutorial |
| §4 evidence, §7 entry/review shape, decision-specific revisit conditions | How to run a bounded reassessment and land it when authorized |

Do not move a row that names an F-id, a verdict, or the entry shape into a skill. The skill may point at it.

A trend or regulation can become a framework feature only when it is compatible with the Mission and deliberately adopted.

**Never let a temporary technology trend — or a statute citation — become an accidental framework principle without a Mission → CORE → FEATURES path.**

**Never adopt a tool because it “has MCP” if it fails drift resistance, compression, or token economy on a ~32GB PC.**
