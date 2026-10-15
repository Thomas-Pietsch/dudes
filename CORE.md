# DUDES Framework — Core Design

> **What is this document?** The software-level PRD and design contract of DUDES. It defines the framework's stable behavior, lifecycle, boundaries, principles and responsibilities. It describes **what the framework must provide and how it is conceptually structured** — not the implementation details of a particular model, runtime or tool.
>
> **Derives from:** `MISSION.md` — the framework's North Star. If this document contradicts `MISSION.md`, `MISSION.md` wins.
>
> **Priority:** `CORE.md` is the authoritative design baseline. `FEATURES.md` describes the concrete feature surface. Technical realization belongs in `ARCHITECTURE.md`. Time-sensitive technology and tool choices belong in `TREND.md`.
>
> **Important:** The former `rules/*.md` rule engine is retired. DUDES no longer asks local LLMs to interpret large Markdown rule sets. The Go framework owns deterministic orchestration and enforcement wherever possible.

---

## 1. The Core Refactoring

DUDES is no longer a collection of Markdown rule sets that an LLM interprets at runtime.

**The Go framework is the process engine.**

LLMs are bounded workers inside that process — not the process controller.

```text
OLD
Human → LLM → interpret rules → decide workflow → LLM → ...

DUDES
Human
  ↓
Go Framework
  ├── state / phase
  ├── artifact loading
  ├── validation
  ├── Git / persistence
  ├── worker invocation
  └── transition
       ↓
   LLM Worker
       ↓
 concrete artifact/result
       ↓
Go Framework → validate → persist → next phase
```

The goal is to move deterministic behavior **out of model context and into software**. This reduces token usage, context pressure and behavioral drift while making the process observable, testable and reproducible.

**Tool selection consequence:** prefer workers and patterns that improve **context engineering** (drift resistance, compression, token economy, bounded invoke/exit, Git audit). Do not treat protocol checklists (e.g. MCP) as adoption criteria. Detailed scoring lives in `TREND.md` §4.

---

## 2. Context Architecture

### 2.1 Isolated Phases

DUDES executes work as bounded, sequential phases.

A phase receives:

- a defined input;
- a defined responsibility;
- the minimum required repository context;
- explicit output expectations.

It produces durable artifacts and a machine-readable result.

The conversational context is disposable.

### 2.2 State Lives Outside the LLM

Authoritative state lives in:

- repository artifacts;
- Git history;
- branches and tags;
- framework-managed execution state where required.

The chat history is **not** the system of record.

Every phase must be able to reconstruct its required context from durable state.

### 2.3 Context Reset

The framework should reset or terminate worker contexts between phases wherever technically possible.

The architectural invariant is:

> **No phase may depend on hidden conversational memory from a previous phase.**

The exact reset mechanism is an implementation concern.

**Priority:** phase reset + durable artifacts first. In-phase token trimming/summarization is optional and secondary (see `TREND.md` patterns). Do not replace artifact handoffs with long-lived multi-agent channel buses.

### 2.4 Micro-agent context: the task, nothing else

A worker gets a small system prompt and exactly three things:

- **the task** — the work-item pack with its frozen requirement and scenario IDs;
- **its goals** — the `make help` lines of the goals its phase may run (named in the phase recipe, §13);
- **the output contract** — delimited output, or a `BLOCKED:` line when the spec does not cover the task.

No framework rules, no role-play, no skills. The mechanisms (§10 hooks and worktrees, §11, §13 goals, engine gates) enforce what rules used to say, so the worker is free to use the model's strengths. DUDES then improves as models improve, without changing.

**Coding workers first.** This shape fits the dev repos: models are trained hardest on code, and goals plus tests decide the outcome. Spec work (PM, UX, security) and lead work (CEO, strategy) is text and judgement that no test settles; their micro-agents get their own shape when they come — discipline skills that carry the DUDES method, which no model is trained on (§5 *Where skills earn their place*), next to the goals those repos already have (the spec repo's gate is the consistency check, `dudes item check`). Until then those roles run as IDE role-play.

**The worker never writes the test it must pass.** A green gate is only the worker's own judgement when it wrote the test: in an instance trial every green slice was green partly by weakening its own test. An acceptance test is therefore written red first by a task of its own (`expect: fail`), read by a human before any code (`approval: after`; `dudes redo` sends it back), and handed read-only to the task that makes it pass; a skipped acceptance test fails the gate (F65, F66). The worker also sees the code it extends — the tracked files of its scope — since writing it anew from the spec made it oscillate.

**Who splits the work.** Three worlds, three ways: in **framework DevOps** the model decides which subagents to start — dynamic, and it improves with the models (`skills/framework/delegate` holds only the principles); in the **role-play** the split is fixed per role, the same splits the engine makes (red test and code apart, one verifier per criterion — `rules/DUDES.md` §1, *Subagents*); in the **engine** it is the contract's tasks and slices, and a worker stays one process with one task. Spec and lead work have no micro-agent shape yet; the role-play carries them until one is invented.

Two contexts stay rich on purpose:

- **Framework DevOps** works *on* the framework, outside the role-play: detailed `skills/framework/` hold the complex prompts that would otherwise be repeated.
- **IDE role-play** (cloud agents in Cursor or Claude today) loads `rules/*` — the transition until workers carry the flow. `rules/SECURITY.md` stays while those agents run unsandboxed on the local machine.

---

## 3. Framework Responsibilities

The Go framework owns deterministic lifecycle behavior.

### Orchestration

The framework:

- accepts the human request;
- determines the starting phase;
- invokes the appropriate worker;
- validates results;
- persists state;
- performs Git operations;
- advances, repeats or stops the lifecycle.

### Artifact Management

The framework knows which artifacts belong to each phase.

Typical artifacts include:

```text
*_PS.md  → strategy
*_PRD.md → requirements
*_TA.md  → technical architecture
tests    → test authoring
source   → implementation
*_QR.md  → quality result
```

The authoritative feature registry is `FEATURES.md`.

### Mechanical Enforcement

If a requirement can be checked deterministically, the framework should check it.

Examples:

- required artifact exists;
- valid phase transition;
- specification is frozen before implementation;
- tests exist before implementation;
- worker exited successfully;
- QA result satisfies the release gate;
- Git state is valid;
- workspace boundaries are respected;
- **known commands** (e.g. `pytest`, `npm test`, linters) are executed by the framework — the worker receives failure text to fix, it does not invent the test-runner workflow;
- worker outputs that must be durable artifacts are extracted by the framework (delimited plain text/code preferred for local models; strict schemas only for small machine-validated control payloads).

> **Do not spend LLM tokens asking a model to enforce something Go can enforce reliably.**

---

## 4. Product Lifecycle

Every unit of work follows the same conceptual lifecycle:

```text
IDEA
  ↓
STRATEGY
  ↓
SPECIFICATION
  ↓
TEST AUTHORING
  ↓
IMPLEMENTATION
  ↓
QA / VERIFICATION
  ↓
DOCUMENTATION
  ↓
HUMAN DECISION
  ↓
SEALED RELEASE
```

The framework may repeat a cycle after human feedback.

### Human boundary

The human:

- provides an idea or feedback;
- owns consequential decisions (investment, positioning, Go/No-Go, scope, high-impact architecture, final seal);
- reviews the delivered result / QA report;
- decides **seal it** or **not good enough**.

The framework performs the process between those boundaries.

The human does not manually orchestrate role-to-role handoffs, and does not approve every internal handoff. Missing business decisions produce a durable **waiting-for-decision** state — workers must not invent assumptions.

Autonomy expands by **delegating tasks** (with an approved envelope: scope, writable paths, permitted actions, budget, escalation), never by delegating a whole job title, and never by an agent promoting itself.

Skills answer **how** to perform a task. They do not grant authority or decide when work is approved.

---

## 5. Roles Are Capabilities, Not Runtime Personalities

This table is the glossary: one name per concept. A word in *Avoid* means the same thing — use the term instead, in prose, names and commits.

| Term | Means in DUDES | Avoid |
|------|----------------|-------|
| **World** | Who acts, under which rules: **framework** (DevOps works *on* DUDES — workspace `AGENTS.md`), **role-play** (IDE agents play the instance's roles — instance `AGENTS.md`, `rules/DUDES.md`), **engine** (Go runs micro-agents — no rules, mechanisms) | DevOps world, outer part, IDE world, Claude/Cursor world |
| **Instance** | One company's use of DUDES: its repos, specs and adaptations | company instance, product repos (when the whole is meant) |
| **Level** | A repo type of the instance (`rules/DUDES.md` §4): **lead** (strategy, vision), **spec** (product and project specification — PM, UX, security), **dev** (code and tests) | C-level, PM level, project-management level, developer level |
| **Role** | A capability (PM, QA, DevOps, …) — not an OS process and not a license to run unattended | persona, agent |
| **Phase** | An engine step in the lifecycle (`CORE.md` §4) | stage, step (for the lifecycle) |
| **Worker** | A replaceable adapter (Ollama, Aider, Pi, stub) invoked by Go | agent, model |
| **Micro-agent** | One worker invocation by the engine: the task, its goals, the output contract — nothing else (§2.4) | sub-agent, bot |
| **Skill** | A discipline file (`SKILL.md` in a directory of its name) — answers *how*, never *what next*. Workers get none by default (§2.4): a make goal carries a procedure, a gate carries a discipline | prompt pack, rule (a rule is `rules/*`) |
| **Skill name** | The task, plainly, in the [agentskills.io](https://agentskills.io/specification) charset (lowercase, digits, single hyphens, ≤ 64), name = directory. Every other fact has one place: **the repo says who owns it** — this repo what DUDES delivers, an instance repo what the company owns; **the folder says the level** — this repo holds every level and groups them in `skills/framework/`, `skills/lead/`, `skills/spec/`, `skills/dev/` and `skills/shared/` (every level), each folder with a `README.md` for what its skills share; an instance repo *is* one level, so its skills sit directly in `skills/` (`<lead repo>/skills/lab-check`). Skills load by name — the engine for workers (F48), the role-play rules, DevOps on request — so a name is unique in the framework, and an instance skill that adapts one takes its own name and names the skill it builds on instead of copying it | prefixes (`dudes-`, a level) in the name |
| **Framework skill** | A skill in `skills/framework/`: DevOps on DUDES. Never given to a worker (the engine refuses its folder), never always-on | DevOps skill |

**Where skills earn their place.** Models are trained on code, not on DUDES. On the **dev** level a worker needs few skills — opt-in for particular tasks such as a refactoring or a review; the goals and tests carry the rest, the model gets as much freedom as possible, and DUDES improves as models do. On the **spec** and **lead** levels no model learns the DUDES method anywhere else: there the skills *are* the know-how (how a backlog item, a frozen spec or a strategy decision is made here), next to the level's gate.

DUDES still models responsibilities such as:

- CEO / Product Vision;
- Product / Strategy;
- PM / Specification;
- Developer;
- QA;
- Security;
- Technical Writer;
- DevOps.

These are **capabilities and responsibilities**, not necessarily independent AI agents.

One worker model may execute several roles across separate phases. Multiple models or worker tools may also be used.

DUDES therefore does not depend on:

- one model per role;
- one process per role;
- multi-agent conversations;
- autonomous agent debate.

The framework controls the role boundary. The worker performs the task.

Roles are capabilities bound to phases, not OS processes and not licenses to run unattended. A role contains tasks with different risk and different verification. A task may be human-performed, agent-drafted pending approval, or agent-executed inside an approved envelope. "CEO is always human, developer is always autonomous" is too coarse.

---

## 6. Spec-Driven + Test-Driven Development

### Specification is the contract

No implementation begins without an applicable specification.

```text
PS → PRD → TA → tests → implementation
```

Companion artifacts (business model, UX, security, roadmap, backlog, and others the instance owns) sit beside that chain. They are **not** collapsed into the PRD. The instance keeps original filenames; the spec-set registers them. Missing companions do not block a freeze — they join the freeze when the files exist.

For existing products:

```text
frozen specification set → tests → implementation → QA
```

Bug fixes and refactors may follow a reduced path where no new requirement specification is necessary.

### No unspecified work

A worker cannot be stopped from assuming, but unspecified work can be made impossible to merge. Traceability runs both ways:

| Guard | Mechanism |
|---|---|
| Every scenario has a test | coverage gate (spec → test) |
| Every acceptance test (BDD / E2E scenario) names the scenario or requirement ID it proves; unit tests serve the implementation and carry none | test-ID check at the test gate (profile `acceptance_tests`: paths, ID pattern, report or gate): an unnamed acceptance test is unspecified work |
| Every commit names its work item | `Item:` trailer, checked by `commit-msg` while an item is bound |
| A change stays inside the work item's paths | `pre-commit` against the scope file the engine writes from the contract (`.dudes/writable`) |
| The spec does not cover the task | the worker answers `BLOCKED:`, the run waits for a decision |

Scenarios stay markdown in the spec (GIVEN/WHEN/THEN), owned by PM and readable by humans. Executable Gherkin (godog, playwright-bdd, …) is an instance choice where it pays off (§13).

### Three layers (product knowledge ≠ execution ≠ evidence)

DUDES distinguishes:

| Layer | Purpose | Authority |
|---|---|---|
| **Product knowledge** | Vision, research, business model, PRD, UX, architecture, security | Instance-owned documents. Humans and agents may draft; humans accept consequential claims. |
| **Execution contract** | Approved scope for one delivery cycle: source references (path + revision + section/ID), constraints, dependencies, acceptance, evidence required | Versioned **projection** of those documents — never a second PRD to maintain by hand. |
| **Delivery evidence** | Tests, implementation, reviews, measurements, QA | Reproducible results tied to the contract and a code revision. |

Workers receive the relevant **original** sections, not another agent's summary of them (direct specification access below).

Use GIVEN/WHEN/THEN where it describes **behavior**. Other claims need other checks: architecture constraints, performance budgets, threat mitigations, research evidence, or an explicit human decision. Software tests must not be treated as proof of willingness-to-pay, field PMF, or similar business hypotheses.

"Finalized specification" and "verified product" are different states. A QR written before implementation is a baseline, not a seal.

### Complete context per work item

Large spec-sets do not fit a bounded worker, and a summary of them is drift. Where the instance declares **who owns which fact** and a **machine-resolvable reference syntax** (a spec-set *profile* — instance data, not framework doctrine):

- An increment is one backlog **work item**. Its worker input is the item plus every owner definition its references resolve to — assembled by the framework from the original text, never paraphrased.
- An item whose references do not resolve is **not ready**. It cannot start.
- Spec-set **consistency** (every reference resolves, every fact defined once by its owner) is a mechanical gate before freeze.
- Item status is **derived** from the QR evidence rows named in the item. Nobody maintains it by hand.
- The execution contract for the increment is projected from the item, not retyped.

Without a profile, the whole-document path above still applies.

### Test-first

Tests are authored from the specification before implementation wherever new behavior is introduced.

### Direct specification access

Workers receive or read the authoritative specification directly. They should not depend on another worker's conversational summary.

This is a framework-level context-drift defense.

---

## 7. Sequential Execution

DUDES deliberately favors explicit, deterministic transitions:

```text
phase N
  ↓
validate
  ↓
persist / commit
  ↓
reset worker context
  ↓
phase N+1
```

There is no requirement for agents to debate or negotiate.

If a phase fails:

```text
failure
  ↓
framework records result
  ↓
defined recovery — retry (bounded), stop, or wait for an explicit human decision
  ↓
worker receives required context only (no invented defaults)
```

**The framework decides the transition. The model does not invent the workflow.**

---

## 8. Quality and Convergence

Completion is not based solely on a worker saying it is finished.

DUDES distinguishes:

- worker result;
- mechanical verification;
- QA result;
- human acceptance.

A release cannot be sealed until its required quality gates pass.

Where a gate can be deterministic, it belongs in Go rather than in a prompt.

**Convergence (toggle).** A worker that says "done" after one pass may have left work behind — small local models especially. When switched on, the framework re-invokes the worker with a fresh context that re-reads the task from the beginning, compares it with the current state and does only what is missing. The loop ends when a pass changes nothing, and it is bounded. Off by default: a strong model often finishes in one pass.

**Stop, do not invent.** A worker that finds its input missing or contradictory stops with a report instead of guessing. The framework turns the report into a durable waiting-for-decision (§4) and hands the human's answer to the next invoke.

---

## 9. Security and Boundaries

The framework enforces hard boundaries independently of model compliance wherever technically possible.

At minimum:

- workers operate inside an allowed workspace;
- file access is bounded;
- external integrations are explicit;
- destructive operations require explicit framework permission;
- Git operations are controlled by the framework;
- secrets are not intentionally exposed to workers.

Prompt instructions may supplement these controls, but must not be the only enforcement layer for deterministic properties.

---

## 10. Git and Release Integrity

Git is part of the DUDES state and audit model.

The framework is responsible for lifecycle-related Git behavior such as:

- traceable commits;
- branch transitions;
- frozen spec branches;
- seal and release tags;
- release integrity;
- auditable phase progression.

The exact implementation belongs in `ARCHITECTURE.md` and the Go code.

### Git is the database

The DUDES backend sits on a local git repository and uses it as its database. Git features are chosen by the job the framework needs done, not by the usual collaboration conventions (pull requests, branching models, upstream/downstream). Today it is one person, one flow, sequential and all local: no parallel runs, no review round-trips. That is what keeps it fast.

| Database job | Git feature | Engine today |
|---|---|---|
| Record: what happened, by whom, based on what | commit with role tag and trailers (work item, hash of every source document) | `gitx.AddCommit`, item-pack trailers |
| Query: history, diff, provenance | `log`, `show`, `diff`, `describe`, `merge-base` | `internal/gitx` |
| Immutable state: freeze, seal, release | branch `release/<product>-vX.Y` (frozen), annotated tags | `dudes seal X.Y.Z` → tag `<product>-vX.Y.Z`; `commit-msg` keeps the spec branch frozen |
| Transaction: isolate, then commit or roll back | worktree on a run branch; merge = commit, remove = rollback | F54 `internal/worktree` |
| Constraint: refuse a bad write | hooks (`commit-msg`, `pre-commit`, …) | `hooks/` via `core.hooksPath` (`make hooks`, F78) |
| Annotation: facts about a commit, added later without changing it | `git notes` in own namespaces (`refs/notes/dudes/*`) | run state (`refs/notes/dudes/run`), spec freezes (`dudes/freeze`), test runs (`dudes/test`) — the newest note in HEAD's history counts, so the record follows the history; evidence is derived from spec and test notes |
| Replication: connect the instance elsewhere | remote, `fetch` / `push` | none; optional, added by the instance |

- **One writer, one adapter.** The engine owns lifecycle commits, branches and tags, and reaches git only through its adapter (`internal/gitx`, `internal/worktree`) over the git CLI. No second store.
- **The commit is the record.** `.dudes/` holds only working state (events log, context packs, worker prompts, scope, worktrees) and instance data (spec-set, contract, profile); working state is disposable. What a phase was based on lands in its commit, not in archive files.
- **Local first.** A remote is replication. An instance may add one (a git server on the home network, a hosted mirror) and may check it in its own scripts, but the engine never needs a network to run a lifecycle.
- **Isolation ladder.** A worktree isolates a change. Running code is isolated by the instance's goals — a gate calls e.g. `make ci`, which may run the tests in a container with `--network none`; the engine needs no Docker of its own. A goal enters a container at most once; inside it, goals run natively — never Docker in Docker. The engine stays on the host (a static binary that needs git, make and sh) and calls goals as a human does; a worker's toolchain may be pinned the same way — one container per worker process, the worktree mounted. If the host itself needs protection (agents running unattended), the whole ADE runs in one sandbox — a VM, WSL distro or OS user, not a container, whose goals would need the host's Docker socket — never a container per edit.
- **Core tools before new mechanisms.** Branch and tag names carry the trace between repos: the frozen spec is a branch, the seal and the release are tags with the same name. Documents state the current truth; their history is `git log`. A new file, format or service must first show what git, make, shell tools or Docker cannot do.

> **If the framework can know it, the framework should enforce it.**

### Regulatory-aligned evidence (design intent)

`TREND.md` records three strategic pillars: consumer-box local models, KMU sovereignty, and DSGVO / EU AI Act pressures. CORE does not encode statutes or a parameter count. It requires process properties that make EU-facing accountability *possible*:

| Property | Framework responsibility |
|---|---|
| Reconstructable phases | Persist + commit at phase boundaries; no hidden chat-as-SoR |
| Human oversight at outcome | Seal / reject only by human; workers never self-seal |
| Decision surfaces | QA / report artifacts humans can read without code knowledge |
| Controlled data path | Default local workers; external/cloud calls explicit and bounded |

Product-specific AI Act risk class, DPIA, or Annex documentation remain product/legal concerns. DUDES makes producing versioned evidence cheap and sequential.

---

## 11. Framework Engineering Principles

These apply to DUDES and to products developed with DUDES.

| Principle | Application |
|---|---|
| **YAGNI** | Do not implement behavior before it is needed. |
| **DRY** | One source of truth for each concept. |
| **KISS** | Prefer simple deterministic mechanisms. |
| **SOLID** | Keep components focused, replaceable and testable. |
| **Modularity & cohesion** | One job per component; what changes together lives together. |

### Pillars: core tools before new mechanisms

Use what exists, elegantly, before reinventing or choosing a tool. Each step down needs a reason the step above cannot carry.

| Order | Pillar | Use it for | Leave it |
|-------|--------|------------|----------|
| 1 | Shell and Linux tools (bash, grep, sed, find; Git Bash on Windows) | A command, a check, a gate that greps | A second workflow |
| 2 | git | The database (§10): record, history, freeze, seal, isolation, constraints | Collaboration ceremony the job does not need, or a second database |
| 3 | make, Docker | The goal system (§13): every repeatable action is a self-documented goal. Docker pins a toolchain; when the host needs protection, the whole ADE runs in one sandbox | A container per edit. Isolation of a change is a worktree |
| 4 | Go, or an existing open-source tool | The phase machine; what 1–3 cannot keep (gate, artifact, audit) | A new tool because it is familiar |

Why the old tools: make, grep and git have done their job for decades, and every model knows them from training — a small local model knows `make test` better than any new tool schema. A mechanism costs no tokens and cannot be forgotten, and whatever an agent runs, a human can run again by hand. That is how DUDES uses AI efficiently: the model does the thinking, proven tools do the rest.

### Rules become mechanisms

A rule an agent must remember costs context on every call and can be forgotten. Whatever can refuse mechanically moves into a mechanism — git hook, make goal, sandbox, engine gate — and its prose is deleted. Text keeps the *what* (the spec) and the *where* (the flow step). Security rules stay text as long as agents run unsandboxed on the host.

Additional constraints:

- deterministic behavior belongs in a mechanism from the pillars above; the phase machine is Go;
- worker behavior is bounded by explicit interfaces;
- state must be externally recoverable;
- tools are replaceable workers, not framework identity;
- technology choices remain replaceable.

---

## 12. Self-Evolution

DUDES can improve itself, but every change remains subordinate to `MISSION.md`.

```text
observe
  ↓
identify improvement
  ↓
design change
  ↓
implement in Go / framework
  ↓
test
  ↓
review
  ↓
commit
```

Markdown remains valuable for documentation, specifications, feature descriptions and architectural decisions.

Markdown is **not a substitute programming language for the orchestration engine**.

---

## 13. Technology Independence

CORE defines capabilities, not vendor choices.

DUDES may use Go, local LLMs, Git, Ollama, Aider or other technologies, but CORE does not require a specific:

- model;
- model size;
- inference runtime;
- coding worker;
- provider;
- hardware configuration;
- MCP implementation;
- cloud region or legal opinion.

### Fixed process, open product technology

DUDES fixes how a product is made and leaves what it is made with to the instance (company, product repos):

| Fixed by DUDES — same for every instance | Open per instance |
|---|---|
| The cycle and its gates: spec → freeze → tests first → implementation → QA report → human seal | Product tech stack, test tools |
| The git contract: `release/<product>-vX.Y` is the frozen spec, only the QR changes there; seal tag `<product>-vX.Y.Z`, same name in the dev repos | Versioning of internal builds, deployment, update system, signing, rings |
| One entry point: humans, CI, agents and the framework drive a repo through its make goals (`build`, `ci`, `release`, `version`, …) | What each make goal does |
| Human seal, role tags, requirement IDs in tests | Which roles and toggles an instance enables |
| Item preparation for the engine: `dudes do <item>` with the role's slicing — a worktree per item, the spec pinned at its freeze, a red-test task and an implementation task per slice with facts only, the freeze — from the repo's tracked `dudes/` folder. The item's branch merges as it is: its commits name the task, and each approved task is one commit | What its `dudes/` names: the spec format and roots, the worker, the gate image, the formatter — a folder per product when the repo holds several, the item's ID finding its product. The model host and other machine values come from the user's environment, never from a repo |

Preparation is process, so the engine does it. A prompt that re-explains it costs a model call per item and is advisory (`TREND.md` §6 Process-as-Prompt), and a makefile that re-implements it drifts per instance. The instance's makefile may name it in `make help`, without logic of its own: the framework ships those names as goals, `make/bridge.mk`, which a dev repo includes by its path, so the role-play reaches the engine through goals, not rule prose.

Few toggles, no free-form flows: a fixed contract keeps the gates strong and keeps the context a worker needs small (work item, frozen requirements, the make goal to run), which is what makes small local models viable. The same contract serves cloud role-play today, local workers later.

### Goal interface

Every instance repo is driven through its makefile. The makefile documents itself; no skill explains a procedure a goal can run.

- **Self-documented:** each goal carries a `## <what it does>` comment; `make help` lists them. That list is the only how-to an agent gets.
- **Fixed names** where the repo has the job: `help` (always), `build`, `test`, `ci`, `version`, `release`. Further goals (role goals such as `backend` or `web`, pipeline stages) are the instance's.
- **Selected per phase:** a phase recipe names the goals its worker may run; the engine passes only their `make help` lines (§2.4). The descriptions stay in the makefile, the selection in the recipe.
- **A missing goal stops the gate** with a clear message. The worker never improvises the command.
- **The test gate is a goal:** without a configured command the engine runs `make test` when the makefile documents it.
- **The seal tag is the hand-off:** DUDES ends at the human seal — the tag `<product>-vX.Y.Z` in the spec repo. It never drives an instance release. Each repo's `release` goal pulls: it refuses without the seal tag and decides what a release is there (one app, an OS image carrying apps, a multi-app repo).

In MCP terms, without a server process or JSON-RPC (a goal call is a process call):

| MCP | DUDES |
|---|---|
| Host — runs the session, holds the context, decides what is allowed | the engine: runs the phase, builds the context, the recipe allows goals. Today in the IDE: Cursor or Claude Code |
| Client — lists and calls tools | the engine's goal runner: `make help` filtered per recipe, `make <goal> VAR=…` in the worktree, exit code and log back. Today in the IDE: the agent's general shell, which can bypass the goals — hence `rules/SECURITY.md` until the engine is the only client |
| Server — exposes tools | the makefile of each instance repo; Dockerfile stages and scripts are its implementation |
| Resources / prompts | git (frozen spec, history, tags) / recipes |

One difference: in MCP the model chooses every call. Here the host also calls goals itself — gates such as `test` or `ci` run whatever the worker thinks. Humans, CI and agents use the same interface.

A self-documented makefile already is a list of tool calls; MCP standardised that idea for tools behind a network. MCP earns its place where make cannot reach: a remote service, typed inputs, consent per call when tools come from untrusted parties. Add it for that one service, never as the foundation of local work. A thin MCP adapter over the goals is possible if a harness ever needs one.

Current technology, sovereignty and regulatory assessments belong in `TREND.md` (including the three strategic pillars). Mission-level *why* those pillars matter is in `MISSION.md` §Strategic Operating Context.

Concrete technical realization belongs in `ARCHITECTURE.md`.

---

## 14. Documentation Hierarchy

Each file has one job. This is the only place that defines them; other files point here.

| File | Job | In the chain (§15)? |
|---|---|---|
| `MISSION.md` | WHY — north star, erosion test | yes, first |
| `CORE.md` | DESIGN — the stable contract (software PRD) | yes |
| `FEATURES.md` | WHAT — capabilities with ID and status (`planned` → `shipped`) | yes |
| `ARCHITECTURE.md` | HOW IT IS BUILT — Go realization, milestone record | yes |
| Go code | the engine | yes, last |
| `TREND.md` | LANDSCAPE — models, tools, regulation, verdicts, pattern status. Knowledge, never work | no — side input |
| `ROADMAP.md` | PATH — the lanes (the role-play delivers, the engine takes over flow steps stage by stage, public), each stage's exit check, where DUDES stands, the horizon. No work items, verdicts or design | no — orders the work: every `TODO.md` line names its lane |
| `TODO.md` | WORK — the one work list: Now (at most 3) / Next / Later | no — schedules the chain |
| `skills/framework/` | how DevOps edits | no |
| `Makefile` (framework root) | the framework's own DevOps goals (`setup`, `build`, `test`, `ci`, `doctor`, `bench`, the worker image) — built for working *on* DUDES. Not an instance makefile: the goal interface (§13) binds the instance repos, whose makefiles the instance designs | no |
| `AGENTS.md`, `rules/*` | IDE entry, not the runtime. Two worlds: the workspace `AGENTS.md` is the framework DevOps entry; each instance repo's `AGENTS.md` plus `rules/*` carries the role-play | no |

This separation keeps `CORE.md` from becoming an implementation dump, keeps temporary technology or statute citations from becoming doctrine, and keeps work out of the knowledge files.

---

## 15. Framework Evolution

The former `rules/*.md` hierarchy is **retired**.

There is no longer:

```text
MISSION → CORE → rules/*.md → AGENTS.md
```

The new derivation is:

```text
MISSION.md
    ↓
CORE.md
    ↓
FEATURES.md
    ↓
ARCHITECTURE.md
    ↓
Go implementation
```

`AGENTS.md`, if retained, is an entry/context document — **not the runtime process engine**. The operational checklist for editing framework files is `skills/framework/maintain` (DevOps only). There is no `FRAMEWORK_MANIFEST.md` or `FRAMEWORK_BACKLOG.md`.

`TREND.md`, `ROADMAP.md`, `TODO.md` and `skills/framework/` do not sit between `CORE.md` and Go (§14).

The path of every idea or requirement:

1. **Source** — a trend or assess run (`TREND.md`, status `hold`), or a need from a product run.
2. **Adopt** — the human decides. Not adopted: it stays `hold`, or becomes a verdict or anti-pattern in `TREND.md` (the "never").
3. **Design** — check `MISSION.md`; record the stable design impact in `CORE.md`, if any.
4. **Capability** — a `FEATURES.md` row, status `planned`.
5. **Schedule** — a `TODO.md` line in Next, naming the `ROADMAP.md` lane it serves; pulled to Now when there is room. Which stage comes next, and when the work no longer moves one, is the roadmap review's question (`skills/framework/roadmap`).
6. **Build** — realization in `ARCHITECTURE.md`, then Go with tests.
7. **Close** — `FEATURES.md` `shipped`, `TREND.md` pattern status `shipped Fxx`, the `TODO.md` line deleted. History is `git log`.

---

## 16. The Architectural Test

For every proposed framework behavior:

> **Can this be implemented deterministically in software?**

If yes:

**Put it in the framework.**

If no:

**Give the worker the smallest possible responsibility and keep the surrounding process deterministic.**

And always:

> **If the framework can know it, validate it in the framework. If the framework can enforce it, do not spend LLM context enforcing it.**

This is the central architectural consequence of the DUDES refactoring.

---

## 17. North-Star Alignment

CORE turns the Mission into an executable software design.

The direction is:

```text
less hidden model state
        ↓
less prompt overhead
        ↓
fewer LLM-enforced rules
        ↓
more deterministic Go behavior
        ↓
clearer durable artifacts
        ↓
stronger auditability
        ↓
more reliable autonomous delivery
```

The ultimate objective remains:

> **AI delivers the product. The human validates the outcome.**
