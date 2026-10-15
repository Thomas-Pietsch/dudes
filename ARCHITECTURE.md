# DUDES — Architecture (Go Realization)

> **What is this document?** How DUDES is built technically. It realizes `CORE.md` in software.
>
> **Not this document:** Mission (why), feature catalogue (what capabilities), tool fashion (`TREND.md`), or line-by-line function specs.
>
> **Derives from:** `MISSION.md` → `CORE.md` → `FEATURES.md` (capability IDs) → **this file** → Go code.
>
> **Related:** this file is the lasting architecture baseline. Format contracts of single packages live beside their code (e.g. `internal/itempack/SPEC.md`).

---

## 1. Doc Map — Where to Look

| Question | File |
|----------|------|
| Why does DUDES exist? Non-negotiables? Erosion test? | `MISSION.md` |
| Must-have properties, lifecycle, human boundary, gates? | `CORE.md` (software PRD) |
| Which concrete capabilities does the product claim? | `FEATURES.md` (F-IDs) |
| How do we implement that in Go / workers / Git? | **`ARCHITECTURE.md` (this file)** |
| Which tools/patterns are current candidates? | `TREND.md` (side input) |
| Where does DUDES stand, and what counts as the next stage working? | `ROADMAP.md` (lanes, exit checks) |
| What are we working on now and next? | `TODO.md` (Now / Next / Later) |

**Agent minimum read (engine work):** `MISSION.md` (erosion test) → `CORE.md` §1–§4 → the `FEATURES.md` rows you touch → this §3 + §8. Do not load all of `rules/*.md` to change Go — those files are IDE-agent behaviour, not the runtime.

**Two worlds:** workspace `AGENTS.md` + this engine vs per-repo `AGENTS.md` + `rules/DUDES.md`. Workspace `AGENTS.md` loads only the DevOps rules; `rules/DUDES.md` names the role-play rule set (no `.cursor/` tree). DevOps how-to is `skills/framework/`. Instance map: `examples/sample/`.

```text
ONE USER PROMPT                    ← UX (“agents deliver, you seal”)
        │
        ▼
┌───────────────────┐
│  DUDES Go Engine  │            ← THIS document
│  phase · validate │
│  persist · gate   │
└─────────┬─────────┘
          │ invoke (bounded)
    ┌─────┼─────┐
    ▼     ▼     ▼
 Worker Worker Worker            ← Ollama / Aider / … (replaceable)
    │     │     │
    ▼     ▼     ▼
 artifacts (PRD, code, QR, …)
          │
          ▼
     Git / durable state
          │
          ▼
   Human: seal | reject
```

**“One Prompt, Full Team”** is **UX**, not the internal architecture. Internally there is one engine and many **bounded worker invokes**, not one LLM playing a debating team.

---

## 2. Architectural Invariants (from CORE)

These are not optional implementation preferences:

1. **Go owns the process** — phase order, transitions, validation, Git lifecycle, known commands.
2. **LLM is a worker** — no Markdown rule engine at runtime; model does not invent the workflow.
3. **State is durable outside chat** — artifacts + Git (+ framework execution records as needed).
4. **Context resets between phases** — no hidden conversational memory across phases.
5. **Human seals outcomes** — workers never self-seal a release.
6. **Tools are replaceable** — Aider, Ollama, OpenCode, … are adapters, not identity.
7. **Deterministic ⇒ software** — if Go can check it, do not spend tokens on it.

---

## 3. Logical Components

| Component | Responsibility | Non-responsibility |
|-----------|----------------|--------------------|
| **CLI / Entry** | Accept user prompt; load config; start/resume run | Model inference |
| **Orchestrator** | Choose start phase; sequence phases; on failure apply recovery policy | Free-form coding |
| **Phase Registry** | Declares phases, inputs, outputs, gates, worker kind | Prompt prose libraries as runtime law |
| **Artifact Store** | Know paths/types (`*_PRD.md`, tests, source, `*_QR.md`, …); load/save | Interpreting product domain |
| **Validator / Gates** | Mechanical checks (exists, frozen spec, tests before impl, command exit codes, …) | Subjective quality taste |
| **Worker Adapters** | Invoke Ollama HTTP, Aider/Pi CLI, packers (Repomix), … with **fresh** context | Owning global lifecycle |
| **Spec-set / Contract / Evidence** | Register instance docs as-is; project one increment; record software vs field proof | Renaming BM/PS/UX or treating PMF as a unit test |
| **Item resolver** (`internal/itempack`) | Resolve `[[DOC:ID]]` refs of one work item to owner definitions under a profile; render the item pack; consistency C1–C8; derive status from QR | Writing or fixing spec files; semantic trimming; any model call |
| **Git Controller** | Stage/commit/branch/tag policies for phase boundaries | Optional worker-local snapshots (see TREND: Orchestrator-Owned Commits) |
| **Run State** | Current phase, run ID, last results (files or small DB) — reconstructible after crash | Chat transcript as SoR |
| **Human Gate** | Surface QR / summary; wait for seal \| reject \| feedback | Auto-approve production seal |

Exact package layout lives in the Go module as it grows; names above are **logical**, not a mandate to create 9 packages on day one.

---

## 4. Phase Machine (Lifecycle)

Align with `CORE.md` §4. MVP may collapse some phases; names stay stable:

```text
IDEA → STRATEGY → SPEC → TEST AUTHORING → IMPLEMENTATION → QA → DOCS → HUMAN DECISION → SEAL
```

Each phase:

1. Load **only** required artifacts + config.
2. Invoke **one** worker (or deterministic Go step — e.g. create empty files, run `pytest`).
3. Validate outputs mechanically.
4. Persist + commit (policy per phase).
5. **Reset** worker context.
6. Advance, retry, wait for a durable human decision, or stop.

A phase may contain an **ordered task list** (CORE §5, F65). Each task names inputs, output, skills/standards, worker, tools, and approval. Go sequences; the model does not invent the list. If `.dudes/contract.json` has no tasks for the phase, the phase is one bounded invoke.

---

## 5. Worker Contract

A worker adapter MUST:

- accept a **bounded** task (paths, instructions file, recipe id, …);
- run with **no dependence** on prior phase chat memory;
- return machine-observable success/failure + artifacts on disk;
- exit (or end session) so the orchestrator can reset.

A worker MUST NOT:

- decide the global phase graph;
- seal a release;
- become the system of record for workflow state;
- invent a missing input — it stops with `BLOCKED:` / `CONTRACT:` (F74).

Prompts larger than a command line travel in a file under `.dudes/` (Aider `--message-file`).

**Default MVP adapters (replaceable):**

| Kind | Example | Notes |
|------|---------|--------|
| Spec / text | Ollama HTTP `generate` | Delimited or small schema outputs |
| Code edit | Aider CLI | Atomic commits optional; orchestrator may still commit |
| Context pack | Repomix / code2prompt | Phase-scoped; optional |
| Heavy multi-tool | OpenCode / Goose / OpenHands / Pi | Pi adapter shipped (M10); others still optional |

Details and eval scores: `TREND.md`.

---

## 6. Gates (Mechanical First)

Implemented in Go wherever possible (CORE §3 / §8 / §16):

| Gate | Example check |
|------|----------------|
| Artifact presence | Required file exists and non-empty |
| Spec freeze | Impl phase refuses without frozen spec-set hashes (F59) |
| Test-before-impl | Test paths exist before codegen phase (when required) |
| Command success | Framework runs `pytest` / `npm test` / `go test` — worker only fixes from stderr |
| Spec shape | PRD has `### Requirement:` + SHALL + `#### Scenario:` GIVEN/WHEN/THEN (F37) |
| Requirement coverage | Each requirement id appears in authored tests (F37) |
| Spec-set consistency | With a profile: C1–C8 zero findings before freeze (F69) |
| Definition of Ready | Item pack builds — every reference resolves (F68) |
| Two-axis converge | Spec findings vs standards findings reported separately (F38) |
| Git clean / policy | Domain paths committed before advance |
| Context budget | Item pack fits `context_budget` after trimming, else stop (F73) |
| Convergence | Optional: re-analysis passes until one changes nothing, bounded (F72) |
| Human seal | No release tag without explicit human action. The seal tag is signed when the sealer's git can sign, `seal.require_signature` refuses an unsigned seal, `dudes verify` checks one from git alone: signature, seal commit with its `Sealed-by` trailer, run state (F89) |

---

## 7. Mapping to FEATURES.md

- **FEATURES.md** lists **capabilities** (what the software provides), with IDs.
- **This file** explains **how** those capabilities are realized (components, adapters, gates).
- Legacy rows still pointing at `rules/*.md` are **migration debt**: capability intent may remain; runtime owner moves from “LLM reads rule file” → “Go enforces / worker executes.”

When adding a capability:

1. CORE (if design changes) → 2. FEATURES row → 3. Architecture note / component → 4. Go → 5. TODO done.

---

## 8. Build Order (milestone record)

Work **top-down**. A worktree isolates a change; the pinned worker image supplies its toolchain, not a complete host-security boundary. Whole-ADE isolation remains separate. Follow `CORE.md` §10: engine on the host, at most one container entry per goal, never a container per edit or Docker in Docker.

| Step | Deliverable | Done when |
|------|-------------|-----------|
| **M0** | Repo module + CLI stub `dudes` | **Done** — `go build ./cmd/dudes`; `-version`; loads `dudes.json` / defaults |
| **M1** | Phase registry + run state on disk | **Done** — `dudes init|status|advance|phases`; state in `.dudes/state.json`, since moved into git notes (F4) |
| **M2** | Artifact load/save + Git commit helper | **Done** — `dudes step` writes stub artifact + `chore(phase): … [🔹engine]` commit, then advances |
| **M3** | Ollama adapter + context reset | **Done** — single-shot `/api/generate`; prior context = files only; `-stub` offline |
| **M4** | Framework-owned test command + fix loop hook | **Done** — `dudes test`; `-fix` → `{product}_FIX.md` from log only |
| **M5** | Aider (or chosen) code worker adapter | **Done** — `internal/aider`; impl `dudes step` runs Aider (or `-stub`); orchestrator commits |
| **M6** | Human gate CLI | **Done** — `dudes seal` / `dudes reject`; step/advance blocked at human_decision. F89: signed seal tag, `Sealed-by` trailer, `seal.require_signature`, `dudes verify` |
| **M7** | Phase-scoped context packer | **Done** — `internal/packer` (Repomix); `dudes pack` / `step -pack`; stub offline |
| **M8** | Spec as mechanical contract | **Done** — F37 shape (`internal/spec`) · F38 two-axis `dudes check` · F39 `{product}_TASKS.md` · F53 `dudes bench`. F37 is for **behavioral** slices, not a rewrite of BM/PS/UX. |
| **M11** | Real-spec bridge | **Done** — F57 spec-set registry · F59 multi-input freeze with provenance · F58 execution contract · F65 in-phase tasks · F66 `dudes decide` · F67 evidence classes. Example: `examples/sample/` (modifies no repo). |
| **M9** | Worker prompt economy | **Done** — F47 `--json` · F48 recipe `skills` · F49 recipe `rules`. Worker skills: tdd, diagnosing-bugs, code-review, codebase-design, writing-for-agents (+ delimited-output). `skills/framework/` are DevOps-only. |
| **M12** | Work-item context | **Done** — F68 item pack (`dudes item pack`) · F69 consistency gate (freeze + `dudes check`) · F70 derived status (`item list/board`) · F71 item-bound contract + pack as worker spec input. Port of the context-pack/1 kit; RE2 needs explicit token boundaries (no look-around). Conformance: `examples/sample/` (zero findings, every item packable). |
| **M13** | Micro-agent hardening | **Done** — F72 convergence passes (`internal/converge`, toggle) · F73 context budget · F74 worker escalation → `waiting_decision` · F47 `item check -json` · git trailers instead of archive files · Aider prompt via message file. |
| **M10** | Isolation + second worker | **Done (lite)** — F54 worktree · F55 Pi adapter · F46 sandbox-lite (cwd + env marker, not Docker) · F56 honour recipe `tools` |
| **M14** | Item handoff | **Part 1 done** — F88 preparation: `dudes do <ITEM> -slice <file.json> [-base <ref>]` (`internal/handoff`, worktree helpers in `internal/gitx`) from the repo's tracked `dudes/` folder; the model host from `DUDES_OLLAMA_URL`; acceptance tests named by the contract (`tests`); `!` exclusions in the writable patterns (`internal/scope`, `hooks/pre-commit`). **The item branch done** — a folder per product in `dudes/` when the repo holds several, the item's product found by the profile's ID family (`handoff.ReadRepo`, `itempack.Profile.OwnsItem`); commit subjects in the item's terms with a `Task:` trailer (`handoff.Names`); `dudes decide` lands an approved task as one commit (`handoff.TaskCommits`, `gitx.Squash`), its passes kept under the local ref `refs/dudes/passes/<slice>/<task>/<n>`; the slice ends with its last task (`runstate.StatusEnded`). Open: review packet in `status` |

**Post-M7 extras** are in `FEATURES.md` (shipped rows). Do not duplicate that catalogue here.

**Active work tracking:** `TODO.md`. Do not duplicate the roadmap into FEATURES. Do not recreate `FRAMEWORK_BACKLOG.md`.

---

## 9. Out of Scope (Here)

- Product-specific PRDs / apps (instance repos)
- Choosing the forever LLM vendor
- Full Temporal cluster (pattern only — `TREND.md`)
- Re-homing a Markdown rule interpreter as the engine
- Becoming a general human project-management tool (upstream product work may stay in instance repos)

---

## 10. Evolution

Architecture changes that alter CORE invariants require a CORE edit first (Evolution Guard). Adapter swaps (Aider → OpenCode) that preserve the worker contract are architecture/adapter updates only — note in this file + TREND if evaluation changes.
