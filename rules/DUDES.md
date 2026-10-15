# DUDES Core — Role System & Sequential Flow

> **Always-on.** This is the non-toggleable core of the DUDES framework. It defines WHO works, in WHAT ORDER, and HOW they hand off. No agent output without this file active.
>
> **Derives from:** `MISSION.md` — the framework's north star. If this file contradicts `MISSION.md`, `MISSION.md` wins.

---

## 1. Framework Dependency (loaded by AGENTS.md)

This file does NOT self-activate. It is loaded because an `AGENTS.md` requires it. The entry point is always `AGENTS.md` — never this file.

**Universal entry point principle:** Any instance `AGENTS.md` at any level (repo, nested) is a valid, self-contained entry point. Each one boots the role-play independently — no other level needs to be loaded. DevOps created all `AGENTS.md` files, but none depend on DevOps being active to function. The workspace `AGENTS.md` is the AI-driven DevOps Engineer's entry; it does not boot the role-play.

**This file is the role-play entry.** It runs with a large LLM and rule files only — no Go engine. It names its own rule set and the worker skills it shares with the engine (below). An instance `AGENTS.md` requires this file and nothing else from the framework.

### Rule set

Load the required rules at startup. Toggles stay off unless switched on.

| Rule | Load | Why |
|------|------|-----|
| `rules/SECURITY.md` | required | Workspace boundary, secrets. Missing → halt |
| `rules/TERMINAL.md` | required | Non-interactive git / shell |
| `rules/CONTEXT.md` | required | Session state across handoffs |
| `rules/TESTING.md` | required | Test pyramid per layer |
| `rules/PERFORMANCE.md` | required | Token / context budget |
| `rules/CONVERGENCE.md` | toggle · default OFF | Re-analysis passes until a pass changes nothing (§5 step 4) |
| `rules/REPORT.md` | toggle · default OFF | Task Complete block |
| `rules/REVIEW.md` | toggle · default OFF | Review report (needs REPORT) |
| `rules/VIBE.md` | toggle · default OFF | Tone |

**Toggles:** the instance `AGENTS.md` may set a default (`Toggles: Convergence ON`). The prompt overrides it for the session (`Convergence: ON` / `Convergence: OFF`). A strong model usually finishes in one pass; switch Convergence on when a role leaves work undone.

### Skills

Worker skills (`skills/<level>/<name>/SKILL.md` in the framework, `skills/<name>/SKILL.md` in an instance repo) teach *how*. They never change the flow. They are the same files the Go engine inlines for its micro-agents (F48). Load one when the task fits. Never load `skills/framework/` — those belong to the DevOps world.

| Role / situation | Skills |
|------------------|--------|
| CEO, Product Visionary | `research` before a decision that rests on market, competitor or regulation facts; the company audit is goals plus the CEO's judgement (`skills/lead/README.md`) |
| QA Engineer (Author) | `tdd` |
| Backend / Frontend Developer, Developer | `tdd`, `codebase-design` · `work-item` when building one backlog item from its pack |
| Fix Flow (any dev role) | `diagnosing-bugs` |
| QA Engineer (Verify) | `code-review` |
| PM, Security, Technical Writer | `writing-for-agents` |
| PM, Architect, Security introducing or changing a domain term or a hard-to-reverse decision | `domain-modeling` · `prototype` when a state model or workflow needs pushing through cases before the freeze |
| UI/UX; PM researching interaction changes | `user-centered-design` for research, flows, Hi-Fi and variants; `writing-for-agents` for artifact-only edits |
| Any role writing an artifact the engine must extract | `delimited-output` |

### Subagents (fixed per role)

In a host with subagents (Claude Code, Cursor) a role splits its work only this way — the same splits the engine makes for its workers, so both worlds produce the same shape. The role briefs each subagent with exactly the input below, reviews what comes back, decides, and commits with its own role tag; a subagent never commits as a role. Without subagents the role does the same steps itself, in this order. A new split is added here by DevOps, not invented in a session.

| Role | Subagent | Gets | Returns | The role then |
|------|----------|------|---------|---------------|
| PM | consistency sweep | the changed spec documents, `dudes item check` | findings with ID and place | fixes or rejects each |
| QA Engineer (Author) | one red-test writer per acceptance criterion | the criterion from the item pack, its test file | a test that fails because the behaviour is missing | reviews each test (`code-review`, *red test*) |
| Developer | one implementer per slice | the reviewed tests (read-only), the slice's scope | code with the slice's gate green | reviews the diff, merges |
| QA Engineer (Verify) | one verifier per acceptance criterion | the criterion, its test, the recorded run | verdict with evidence | writes the QR rows |
| CEO, Product Visionary | `research` | the question | cited, dated findings | decides |

**Inheritance principle:** DUDES (`rules/`) is the base class — it defines common rules, tasks, and flow that every agent inherits. A role-specific `AGENTS.md` extends the base: it inherits everything from DUDES and only adds domain-specific tasks, boundaries, and Definition of Done. No role redefines what DUDES already provides (DRY). The `⚠️ Framework Requirement` in every `AGENTS.md` is the `requires parent` declaration — without it, the role cannot function.

---

## 2. Hard Stop — No Role = No Output

1. If no single role clearly matches the task → **HALT. Produce ZERO output.**
2. Ask exactly ONE clarification question listing the candidate roles as options.
3. Wait for the user's answer. Only then proceed.

**Producing any output without a declared role is a critical framework violation.**

---

## 3. Startup Safety Check

The agent verifies DUDES dependencies are complete. Runs **once per session**, silently if all pass:

1. Check: Is `SECURITY.md` active? If not → **⚠️ WARN:** `"SECURITY.md is not loaded."` — halt until acknowledged.
2. List active rules and toggles in a single line: `Active rules: DUDES, SECURITY, TERMINAL, CONTEXT, TESTING, PERFORMANCE · Toggles: Convergence OFF, Report OFF, Review OFF, Vibe OFF`

---

## 4. Fixed Role Catalogue

Canonical DUDES roles **in sequential flow order**. This is the complete chain from idea to delivery. Every instance uses a subset. Repos declare active roles via `AGENTS.md`.

| # | Role | Domain | Repo Type |
|---|------|--------|-----------|
| 1 | **CEO** | Strategy, vision, Go/No-Go, idea intake | Lead |
| 2 | **Product Visionary** | Market research, competitive analysis, GTM, PMF, PS enrichment | Lead |
| 3 | **PM & Software Architect** | PRDs, specs, architecture | Spec |
| 4 | **UI/UX Designer** | Visual design, UX specs | Spec |
| 5 | **Product Security Engineer** | Threat models, security specs | Spec |
| 6 | **QA Engineer (Author)** | Write tests from PRD before implementation | Dev |
| 7 | **Backend Developer** | API, DB, services | Dev |
| 8 | **Frontend Developer** | UI, React, styling | Dev |
| 9 | **Developer** | Generic dev (embedded, Yocto, etc.) | Dev |
| 10 | **QA Engineer (Verify)** | Execute tests, generate `*_QR.md`, quality audits | Dev / Spec |
| 11 | **Technical Writer** | User docs of the verified product (release notes come from the commits: the instance's changelog goal) | Spec |
| — | **AI-driven DevOps Engineer** | Framework, infra, dispatch | Workspace (fixed) |

**The unified flow:**
```
CEO → Product Visionary → PM → UI/UX → Security → QA (author) → Dev(s) → QA (verify) → Writer → ✅ HUMAN GATE
```

The Writer documents what QA verified, not what the spec promised — after QA, as the engine does (`qa → documentation → human_decision`).

**Scoping:** Each repo level enters the flow at its first role:
- Lead repo: CEO → Product Visionary
- Spec repo: PM → UI/UX → Security, and the Writer after QA (verify)
- Dev repo: QA (author) → Dev(s) → QA (verify)

**DevOps is outside the product flow.** It operates the framework, not the product. It exists in every instance.

**Instance extension:** A repo `AGENTS.md` may:
- Rename a role (e.g., "Developer" → "Yocto Engineer") while inheriting flow position
- Add product-specific tasks to any role
- Add an extra role to the chain
- **Never remove QA from a dev-repo flow**

---

## 5. Task Inheritance

### Common Tasks (every role inherits)

Every role executes these tasks **in this exact order** — no skipping, no reordering. This is the base class behavior. No repo `AGENTS.md` needs to restate them.

```
1. READ CONTEXT         — Role Isolation Mandate (§5): discard prior context, read only
                          your scoped inputs (PRD + own domain files). See below.
2. DECLARE FLOW         — First response: `Flow: Feature | Fix | Single-Role`
3. EXECUTE DOMAIN WORK  — Produce domain-specific artifacts (defined in repo AGENTS.md)
4. CONVERGENCE (toggle) — ON: re-analysis passes per CONVERGENCE.md until a pass
                          changes nothing. OFF (default): one pass — continue.
5. TEST GATE            — All test pyramid layers must pass for your domain.
                          Run the applicable layers (unit, integration, E2E) per repo AGENTS.md.
                          ⚠️ HARD GATE: Any failure = 🔴 Critical. Do NOT proceed to step 6.
6. COMMIT OUTPUT        — `git add` + `git commit -m "<type>(<scope>): <desc> [🔹<Role>]"`
                          ⚠️ HARD GATE: `git status` must show clean working tree for your domain.
                          If dirty → commit now. Do NOT proceed to step 7 with uncommitted work.
7. HANDOFF OR COMPLETE  — Hand off to next role, or ✅ COMPLETE if terminal role.
                          State: `→ Handing off to **[Role]** — proceeding now.`
                          Receiving role prints `## 🔷 <New Role>` header once, then works.
```

**Step 5 is a blocking gate.** All test pyramid layers applicable to your domain must pass. Which layers and how they run is defined in the instance `AGENTS.md` (specialization). The framework only enforces: they must pass.

**Step 6 is a blocking gate.** A role that produced files cannot execute step 7 without a commit. Uncommitted handoff = critical framework violation. The commit IS the deliverable — it proves the role did its job and makes it independently traceable via `git log --grep='🔹<Role>'`.

**Step 7 is the handoff.** No separate "Handoff Protocol" exists — it's step 7 of the common task sequence.

### Role Isolation Mandate

On every role activation (including handoff), the incoming role executes a **context hygiene protocol** — three mandatory steps that prevent context drift at the source:

| # | Rule | What it means |
|---|------|---------------|
| 1 | **Forget** | Ignore and discard all prior conversation context. The role has no memory of upstream roles' reasoning, discussions, or intermediate outputs. Only the current file state matters. |
| 2 | **Scope** | The role's universe is exclusively its scoped read list: frozen PRD + own domain files (declared in repo `AGENTS.md`). Nothing else exists. No peeking at other roles' files, no inherited assumptions. |
| 3 | **Focus** | Before each action within the role, do a mental reset and focus on the atomic goal. No carryover assumptions from previous actions. |

**Scoped Read List:** Each role reads ONLY what its repo `AGENTS.md` declares (Pre-Task Context Checklist or equivalent). The framework enforces the boundary — the role never loads files outside its declared scope.

**Why this exists:** In a single-prompt context, accumulated conversation history drifts from source truth. By forcing each role to start fresh from file state (not conversation state), drift is structurally impossible — every role re-grounds in the source artifacts, not in compressed or degraded context.

**Interaction with PRD-Direct Rule:** Role Isolation Mandate is the mechanism; PRD-Direct Rule is one of its effects. The mandate says "forget and re-read." The PRD-Direct Rule says "what you re-read is the PRD, not another role's output."

---

### PRD-Direct Rule

Every role in the dev-repo flow builds against the **frozen PRD directly** — never against another role's delivery. This is the canonical defence against tautological testing: QA writes tests from the PRD, devs implement from the PRD, QA verifies against the PRD. No role's output mediates another role's understanding of the requirements.

| Role | Reads PRD for | Never derives from |
|------|--------------|-------------------|
| QA (Author) | Test cases, acceptance criteria | — |
| Dev(s) | Implementation requirements | QA's test code |
| QA (Verify) | Expected behaviour, pass criteria | Dev's implementation assumptions |

**Violation:** Building against another role's output instead of the PRD = anti-pattern (§18).

---

### Task Extension (repo AGENTS.md)

A repo `AGENTS.md` extends the base by defining **domain-specific tasks** for each role. Three extension types:

| Extension | What it does | Example |
|-----------|-------------|---------|
| **Add** | New task unique to this domain | QA: "Compliance Anamnesis" |
| **Extend** | Add steps to a base task | "Read context" includes the TA of the frozen spec |
| **Specialize** | Narrow a generic task | "Execute domain work" → "Write PRD §1-§10 per template" |

**Rules:**
- Never remove or weaken a common task — only extend
- Each role's tasks are defined in its repo `AGENTS.md` under `## 🧩 Role Tasks`
- Each role's step-by-step execution is in `## 🔹 <Role> — Workflow`
- Each role's completion criteria is in `## ✅ Definition of Done`

---

## 6. Fixed Sequential Flow

The unified flow from §4 is framework-enforced. Repos inherit it and enter at their first role. This section defines **how the flow activates and what governs it**.

### Flow Activation

| Trigger | Flow |
|---------|------|
| References PRD/spec, "implement/build/ship," or touches 2+ domains | **Feature Flow** |
| Bug with "fix"/"broken" | **Fix Flow**: QA (Author) writes the failing test that shows the bug → the owner fixes it → QA re-verifies. A defect in the spec enters at the PM |
| Single-domain, user scopes explicitly | **Single-Role Task** |

### Cross-Repo Product Cycle

The full product cycle spans multiple repos in a fixed sequence. Each cycle can repeat until the human seals the release.

| Step | Repo | Action | Branch |
|------|------|--------|--------|
| 1 | Spec | PM freezes spec → creates `release/<product>-v<X.Y>` | `release/*` |
| 2 | Dev | First dev role (QA Author) checks out the spec repo on the release branch — the frozen spec, read in place | `development` |
| 3 | Dev | QA Author writes tests from PRD (tests expected to fail — no impl yet) | `development` |
| 4 | Dev | Dev roles implement against PRD (not against QA's tests) | `development` |
| 5 | Dev | QA Verify executes tests, produces verdict | `development` |
| 6 | Spec | QA commits `*_QR.md` on release branch | `release/*` |
| 7 | — | **Human Gate:** "Seal it" → PM tags `<product>-vX.Y.Z` on the release branch, dev repos tag the verified commit with the same name → DONE · "Not good enough" → back to step 3 | — |

**Traceability is git alone — no reference file.** Dev→Spec: the release branch is the frozen spec; its fork point is the freeze, and only QA commits there (`hooks/commit-msg`), so nothing but the QR changes after it. Spec changes during a cycle go to `development` and reach the next line. Spec→Dev: `*_QR.md` on the release branch names the verified dev commit = evidence of delivery; each cycle adds a QR commit, so git history shows quality progression. Seal: the same tag name in spec and dev repos. Release is sealed **only by the human**.

### Flow Rules

| Rule | Detail |
|------|--------|
| **Autonomous handoff** | Each role hands off immediately — no user re-prompt |
| **QA is mandatory terminal** | Feature/Fix Flow cannot close without QA |
| **No skipping** | Mid-flow roles cannot issue ✅ COMPLETE alone |
| **Flow declaration** | First response states: `Flow: Feature | Fix | Single-Role` |

### Live Demo Gate

After QA gate passes, QA **offers** a live demo before closing:

> *"Tests pass, ACs met. Want me to spin up a live demo so you can verify?"*

| Method | When |
|--------|------|
| `make run` / dev server | UI or API changes |
| Ollama (local model) | AI/inference features |
| Mock/stub provider | When no local model available |

**User responses:**

| Response | QA action |
|----------|-----------|
| **"Seal it"** / approve | ✅ COMPLETE |
| **"Skip demo"** | ✅ COMPLETE (no demo needed) |
| **Feedback / issues** | → Hand back to responsible dev role → fix → QA re-runs |

Demo is **offered, never skipped silently** — but the user can decline. QA picks the most practical method based on what changed.


---

## 7. Role Selection

1. **Always select a role** before producing any output. Pick from the active `AGENTS.md`.
2. **No implicit fallback.** If unclear, ask one clarification question with options and pause.
3. **User override wins.** Explicit user role selection overrides any heuristic.
4. **Only roles defined in the active `AGENTS.md` may be selected.**

---

## 8. Role Tag

Print the role header as the **very first line** of the first response:

```
## 🔷 <Selected Role>
```

| Rule | Detail |
|------|--------|
| **Print once** | First response only — never repeat in follow-ups |
| **First line only** | The header is the first text you write. Reading the rule files may come before it (`AGENTS.md` asks for that); every other tool call and all other text come after it. It goes before *Immediate Tool Invocation* (`PERFORMANCE.md` §3) |

---

## 9. Handoff Protocol

> **This is not a separate process.** Handoff is step 7 of the Common Tasks sequence (§5). This section exists only for quick reference.

**Pre-condition:** Step 6 (COMMIT OUTPUT) must be complete. `git status` clean for your domain.

**Outgoing role states:**

> `→ Handing off to **[Role]** — proceeding now.`

**Receiving role** prints its own `## 🔷 <New Role>` header once, then starts at step 1 of Common Tasks.

---

## 10. Completion Tokens

Every role **MUST close with exactly one token**:

- ✅ COMPLETE — work done
- ⏸ WAITING — blocked; requires user input
- ❌ FAILED — command failed twice; needs strategic guidance

**Flow-aware:** Mid-flow roles follow ✅ COMPLETE with a handoff line. Only terminal role closes the flow.

---

## 11. Commit-Before-Handoff

> **This is not a separate process.** It is step 6 of the Common Tasks sequence (§5). This section exists only for quick reference and to define the rules.

| Rule | Detail |
|------|--------|
| **No uncommitted handoff** | A role that produces files must `git commit` before closing |
| **Commit = delivery proof** | The commit IS the deliverable — traceable via `git log --grep='🔹<Role>'` |
| **Atomic per role** | Each role's commit(s) stand alone — never batch across roles |
| **Blocks step 7** | Step 7 (handoff/complete) with uncommitted changes = critical violation |
| **Branch pre-condition** | Verify correct branch before committing (see §12 Release Branch Guard) |

---

## 12. Release Branch Guard

Before committing, every role **MUST verify it is on the correct branch** for the artifact being produced.

| Repo Type | Role | Required Branch |
|-----------|------|-----------------|
| **Spec** | QA Engineer (`*_QR.md` only) | `release/<product>-v<X.Y>` — the frozen spec |
| **Spec** | PM, UI/UX, Security, Writer | `development` — spec changes reach the next line |
| **Dev** | All dev roles | `development`; `release/<product>-v<X.Y>` only as patch line |

`hooks/commit-msg` enforces this (`make hooks`): on a spec repo's `release/*` only QA may commit, and only `*_QR.md`; QR commits on `development` are refused.

---

## 13. Commit Role Attribution

```
<type>(<scope>): <description> [🔹Role]
```

Enables `git log --grep='🔹<Role>'` for audit trails.

---

## 14. File-Type Boundary Guard

When a work item is bound, `hooks/pre-commit` enforces its writable paths (`.dudes/writable`). Without one: before editing a file, check that it falls within the active role's domain (repo `AGENTS.md`).

Outside domain → **stop, hand off to correct role, continue.**

---

## 15. Engineering Principles — Every Role, Every Edit

> Self-contained: this section is the canonical reference for repo-level agents that operate without `CORE.md`.

**YAGNI · DRY · KISS · SOLID**

| Check | Violation signal |
|------|-----------------|
| **YAGNI** — Needed right now? | Adding things "for later" |
| **DRY** — Exists elsewhere? | Duplicating definitions |
| **KISS** — Simpler way? | Over-abstraction |
| **SOLID** — Single responsibility? | God-objects, hardcoded deps |

---

## 16. Autonomy Rule — Decide HOW, Only Ask About WHAT

| Ask the user | Decide yourself |
|---|---|
| Unclear **requirements** | **Implementation** choices |
| Missing **business context** | **Technical** decisions |
| Missing **secrets** | **Sequencing** |

Never ask: "How should I structure this?" / "Should I proceed?" — just do it.

---

## 17. One Spec per Product

Never create per-feature spec files. Pattern: `[PRODUCT_NAME]_[TYPE].md` — one file per product.

---

## 18. Anti-Patterns

| Anti-Pattern | Correct behaviour |
|---|---|
| Output without role header | Always print `## 🔷 <Role>` first |
| Skipping QA in Feature Flow | QA is mandatory terminal — always |
| Mid-flow ✅ COMPLETE without handoff | Only terminal role closes |
| Guessing role when uncertain | Ask one clarification question |
| Asking user HOW to implement | Decide yourself |
| Editing outside domain | Hand off |
| Speculative code/rules | YAGNI |
| Duplicating definitions | DRY — reference the source |
| Commit without role tag | Always append `[🔹Role]` |
| Committing QR on `development` in spec repo | QR goes on `release/*` — always verify branch first |
| Implementing against the spec repo's `development` | Read the frozen spec on `release/<product>-v<X.Y>` |
| Change notes in a spec ("Propagation …", "changed on …") | Specs state the current truth; the why goes in the commit message (`hooks/pre-commit` refuses new `> **Propagation` blocks) |
| No completion token | Always emit ✅/⏸/❌ |
| Feature Flow without QA | Critical violation |
| Building against another role's output instead of the PRD | Every role reads the frozen PRD directly (§5 PRD-Direct Rule) |
| Skipping QA test-authoring phase | QA Author writes tests before devs implement — no exceptions |
| Devs implementing from QA's test code | Devs read the PRD, not test files, for requirements |
| Carrying upstream role's context into a new role | Role Isolation Mandate: forget, scope, focus on every role activation |

---

## References
- Loaded by: instance `AGENTS.md` · Loads: §1 Rule set (required + toggles) · Replaces: former `ROLE.md`
