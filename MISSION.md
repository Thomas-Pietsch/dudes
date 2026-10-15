# DUDES Framework — Mission & North Star

> **Authority:** This document is the north star of the DUDES framework. Every rule, every role, every process must serve this purpose. Changes that contradict this document are framework erosion — reject them.

---

## Vision

AI agents deliver production-ready products autonomously.

The human validates **outcomes, not process**.

Today, AI output may require iteration: the human reviews the result or QA report and sends it back until quality is sufficient. Tomorrow, when AI can deliver the required quality directly, the human only seals the result.

**DUDES is built for both states.**

The long-term destination is a digital organization in which specialized AI agents execute the complete product-development flow while the human remains the authority at the outcome boundary.

---

## The Core Promise

**You have an idea. You tell the dudes. They deliver. You seal.**

DUDES turns an idea or a piece of human feedback into an auditable, spec-driven product increment through a sequential agent flow:

**idea → strategy/spec → frozen spec → implementation → QA → report → human decision → seal**

The human does not manage the process. The agents execute it.

The human can:

- provide a new idea;
- give feedback on an existing delivery;
- read the resulting QA report;
- say **“seal it”**;
- say **“not good enough”**;
- or deliberately enter the flow at an intermediate point.

The agent always continues forward from its current point.

---

## What DUDES Is

DUDES is a **pure-agent company framework** — a digital organization where every role (PM, Developer, QA, Security, Writer, DevOps, etc.) can be an AI agent operating within the framework's controlled flow.

It connects to any real company structure:

**1 human + agents → small team → enterprise organization**

The organizational structure can change. The process principles remain.

### One sentence

> **DUDES delivers auditable, spec-driven product increments through a sequential agent flow — from idea to sealed release — with human validation at the outcome boundary and git-native traceability.**

---

## The Human-Agent Boundary

| Human | Agent |
|---|---|
| Has an idea — OR — gives feedback on the last delivery | Executes the flow: spec → implement → test → report |
| Reads the QA Report | Writes code, tests, specs and documentation autonomously |
| Says **“seal it”** or **“not good enough”** | Never seals — always delivers and waits |
| Stages the product for deployment | Optionally runs a live demo if requested |
| Can enter at a deliberate point in the flow | Continues forward from that point — one direction, sequentially |

### Two Entry Points

1. **New idea** → strategy/spec definition → freeze → development → QA → report
2. **Feedback on existing delivery** → frozen spec → development → QA → updated report on the same release branch

Each cycle adds commits to the release branch. Git history shows quality progression. The human seals when satisfied.

---

## What Must Be Human-Readable

These are the human's decision-making and accountability surfaces:

| Artifact | Why |
|---|---|
| `*_QR.md` (QA Report) | The human's decision-making input; must be understandable without code knowledge |
| `README.md` per repo | Shared contract for how human and agent interact with the repository |
| Git log | Audit trail — role-tagged and timestamped |

Everything else may be optimized for agent consumption. Human readability is a courtesy, not a requirement, for agent-internal artifacts.

---

## Core Properties — Non-Negotiable

1. **Sequential** — Roles execute in fixed order, one direction, without backtracking inside a flow.
2. **Auditable** — Every action is represented by a git commit with role attribution. Durable artifacts and the human seal/reject gate must make the delivery path reconstructable (supports accountability expectations under DSGVO / EU AI Act — see `TREND.md` §3; not legal advice).
3. **Iterative** — Cycles repeat on the same release branch until the human seals.
4. **Spec-driven** — No implementation without a frozen specification.
5. **Test-driven** — Tests are written from the PRD before implementation begins (TDD + SDD).
6. **Self-contained** — The framework works locally and does not require external services.
7. **Flexible** — It adapts to any company structure, deployment method, and team size.
8. **Context-drift-resistant** — Context pressure is mitigated through repository-level isolation, state persistence, and the PRD-Direct Rule: every role re-reads the source specification directly.
9. **Hardware-agnostic efficiency** — Agents remain stateless and single-purpose so the framework can scale across different local hardware configurations and efficient local models.

---

## Independence

DUDES is designed to avoid dependence on a particular cloud provider, platform, model vendor, or deployment method.

Local execution is a first-class capability:

- company IP can remain inside the company environment;
- local open-weight models can be used;
- the framework does not require a cloud subscription;
- deployment remains outside the framework's concern.

Cloud models may be used where appropriate. **The mission is model-agnostic and deployment-agnostic.**

---

## Strategic Operating Context

DUDES is built to succeed where three pressures arrive together (detail and current assessments live in `TREND.md` — temporal, not doctrine):

| # | Pressure | Mission link |
|---|----------|--------------|
| **1** | Capable models on a **normal PC** (current sweet spot: a coder that **fits ~32GB** and can finish a scoped step — see `TREND.md`) | Self-contained, hardware-efficient autonomy |
| **2** | **KMU / sovereignty** — AI productivity without surrendering source and data to the cloud | Local-first, flexible company fit |
| **3** | **DSGVO + EU AI Act** — development that can be **audited**, documented, and **human-overseen** | Auditable flow; human validates outcomes, never silent self-seal |

```text
Normal PC → local LLM → sequential agent/worker flow → durable Git evidence → human seal
```

These pressures explain *why* local-first, sequential, git-native, human-at-outcome-boundary design matters. They do **not** lock DUDES to a model size, vendor, or a claim of automatic legal compliance.

---

## What DUDES Is Not

- Not CI/CD — DUDES defines the development process, not the deployment pipeline.
- Not dependent on any specific cloud, platform, or external service.
- Not a human workflow-management tool — humans interact at defined boundaries.
- Not rigid about deployment — USB stick, pipeline, Docker, bare metal, or another mechanism are deployment concerns, not framework identity.
- Not a collection of autonomous agents debating with each other — the framework favors explicit roles, controlled handoffs, and a deterministic direction of execution.

---

## The North Star

When evaluating a new rule, role, tool, architecture decision, or feature, ask:

> **Does this move DUDES closer to autonomous, production-ready product delivery while keeping the human in control of the outcome?**

And then:

> **Does it preserve sequential execution, auditability, specification discipline, test discipline, self-containment, flexibility, and resistance to context drift?**

If the answer is no, the addition is framework erosion — regardless of how technically impressive it is.

**The framework exists to make the delivery of software increasingly autonomous, reliable, auditable, and human-controlled.**

---

## The Soul

- **Stay cool.** State what you did, move on.
- **Be direct.** No filler. Say the thing, then stop.
- **Flow with it.** Something breaks? Adapt, fix, continue.
- **Humor is allowed.** Never at the expense of clarity.

> **Stay cool. Ship sealed.** 🦭
