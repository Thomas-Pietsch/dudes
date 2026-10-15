# DUDES Framework — TODO

> The one work list of the framework. Product/business tasks belong in the lead repo's `TODO.md`.
>
> **How work moves:** `TREND.md` holds knowledge (verdicts, pattern status), never work. A trend or assess run names at most 1–3 candidates; the human pulls them here. **Now** = in progress, at most 3 (`make now`, `make todo-check`). **Next** = ordered, pull from the top; a blocked item stays in place, marked `blocked: <who/what>`. **Later** = unordered pool.
>
> **Done** → delete the line: history = `git log`, shipped = `FEATURES.md`, pattern status = `TREND.md` §6. **Never** → not here: a rejected idea is a `TREND.md` verdict or anti-pattern (§5 REJECTED, §6 anti-patterns, Harness watch "Never"); a dropped item is deleted with the reason in the commit message.
>
> **Lanes:** every line starts with the `ROADMAP.md` lane it serves — `R` the role-play delivers · `E1` `E2` `E3` the engine's stages (dev, spec, lead) · `P` public (`make roadmap-check` fails otherwise). A line that serves no lane is a side quest.
>
> Tags: `[trend]` market and tools · `[lean]` rules → mechanisms · `[engine]` Go engine · `[instance]` first instance fit · `[trial]` product trials · `[skills]` skills · `[ops]` operations and security.

## Now

- [ ] P · `[ops]` **Going public** — the human's gate: as soon as the repo is clean (`ROADMAP.md` `P`). The instance's work goes on beside it.

  Done on 2026-10-14/15:
  - Identity "the seal" (`docs/BRAND.md`).
  - The tree is clean: `make public-check` runs in `make ci`, with the instance's names in the untracked `.dudes/private-markers`.
  - Apache-2.0 (`LICENSE`, `NOTICE`). The skills that started from mattpocock/skills are rewritten as our own; no third-party text remains (checked line by line on 2026-10-15).
  - A README with `make try`; no Git LFS.
  - The first instance moved out into a neutral sample product.

  Also done on 2026-10-15: a second public-face cycle from three outside reviews (a plain first screen, the two doors, a tour with a real refusal, a contributor door), the verifiable seal (F89) and the IDE door (F90).

  Left: publish, under the owner's account, branch `development` (the human's choice, 2026-10-15).
  1. Rebuild the fresh root from `development`: one commit of the tree under the owner's identity, then the roadmap restamp, verified in a clone of only those commits. The backup of the full history is in `C:\Work\backup`.
  2. A new repository under the same name (the human's choice, 2026-10-15): the old one is renamed `dudes-private` and stays private, the full record. In place was dropped: a force-pushed history stays in GitHub's storage, a public repository opens a commit by a 4-character SHA prefix, and GitHub Support purges only secrets. The owner renames the old repository and creates the empty new one; DevOps pushes the fresh root as `development`, and `C:\Coding` switches to it.
  3. The owner makes it public and turns on private vulnerability reporting (SECURITY.md links to it).
  The docs name engine versions by date; only third-party citations keep their commit SHAs.
- [ ] E1 · `[engine]` **Dev stage: item handoff (F88), then three items through it** — the first real item (a backup verification in one Go package) ran by hand on 2026-10-14 (`TREND.md` §1.1: 20 hand steps and 418k cloud tokens to prepare).
  1. F88 makes those steps the engine's:
     - `dudes do <item>` with the role's slicing, from the repo's tracked `dudes/`;
     - tests where the language needs them, and commit subjects that name the task;
     - the review packet in `status` and the local tokens in the events;
     - a redo labelled by who decided, and the formatter from the repo.
  2. `rules/DUDES.md` §1 *Subagents*: in a repo with `dudes/`, the red-test writer and the implementer are the engine's local worker. The role reviews (`dudes decide` / `redo`, sending back spec defects and breaks of a documented standard) and merges.
  3. Items 2–4 go through it: the factory-quality product's second item onward. The CEO allowed it on 2026-10-14 against its frozen spec, its documents untouched. The remaining slices of the first real item count as one item.

  Model comparisons stay paused.

## Next

- [ ] R · `[lean]` **What the starter's first runs showed in the rules** (F90, 2026-10-15): `rules/DUDES.md` §6 and §12 know only a spec repo and a dev repo, so the starter defines its own one-repo release; and a QR went green with a defect: a 200-character limit was counted in bytes and tested with ASCII only (`docs/WALKTHROUGH.md` shows it). Decide whether the one-repo release moves into the rules, and whether `rules/TESTING.md` asks for a limit's unit and non-ASCII input.
- [ ] P · `[engine]` **A workspace inside another repository** — `dudes` run in a folder that is not its repository's top level works on the repository around it: its commits and notes land there. `make try` fences the tour with `GIT_CEILING_DIRECTORIES`; the engine itself should refuse, or say plainly where it writes.
- [ ] E1 · `[trend]` **Harness trial on replay slice 1** — blocked: the human's go (benchmarks paused 2026-10-14). Harness watch's revisit trigger has fired (Aider: no release after 0.86.2, its main idle since 2026-05-22), and slice 1 is a reproducible incumbent limitation: 3 of 4 models never get it green (`TREND.md` §1.1). Compare Pi (adapter F55/F56) with Aider on slice 1 only, same model and settings (§4 paired trial). A guess to test: a harness that runs tools can generate the example config file with the product's own command, where a single edit reply cannot.
- [ ] R · `[lean]` **Make visual/flow evidence mechanically checkable (F81 follow-up)** — blocked: the instance's first UX coverage matrix (a lead-repo item) and the human's time for UX. DevOps; use that first UX coverage matrix as the input, not a new parallel spec format. Add a registered engine capability and make goal that detects missing local design assets, unresolved journey/state/mode references and missing required evidence before reporting visual PASS; test missing-file, stale-reference and valid-reuse cases. Why: written design rules alone did not prevent absent mockups. Research quality and aesthetic approval remain human/field judgments, not file-existence checks.
- [ ] R · `[lean]` **Reconcile instance UX instructions with F81** — blocked: the human's time for UX (the conflicting wording is in the instance's spec repo `AGENTS.md`). DevOps; remove conflicting "mockup every feature" / prose-only delivery wording from instance role instructions in favor of the shared skill and TESTING protocol. Keep product personas, mode names and journeys instance-owned. Done when role entry points load the discipline without duplicate checklists and a second product can reuse it; product design/implementation work stays in the instance lead repo's `TODO.md`, not here.
- [ ] E2 · `[skills]` **The DUDES method as spec and lead skills** — blocked: human (the next days): extract from the early chats in which specs and the framework were invented over many cycles how a backlog item, a frozen spec, a QR and a strategy decision are made here (today partly in instance documents, e.g. the backlog method in the instance lead repo's `STRATEGY.md` H.6, and the evidence ladder with its evidence-before-edit rule in §E — the lead one goes into `skills/lead/README.md`) and write them as `skills/spec/` and `skills/lead/` skills with their READMEs. The lead skills include the step back for a product roadmap, in the shape of `skills/framework/roadmap`: computed facts, fresh eyes, the human decides. Then research what is still missing (`skills/framework/screen`, agent-orchestrators.com) and take it the same way.

## Later

- [ ] E2 · `[engine]` **Mockup-generation worker integration** — only after the spec-to-mockup skill and manual product handoff prove useful. Reuse the bounded prompt/artifact contract for a suitable design/image-capable harness; evaluate capabilities and output quality before selecting a provider/model. Define asset return, revision trace, text/flow verification, privacy boundary and failure handling. Do not build a design-agent orchestration system now or assume a GPT-branded agent necessarily has image-generation tools.
- [ ] E2 · `[engine]` **Micro-agents for spec and lead repos** — after the dev-repo workers prove out (`ROADMAP.md` `E1`) and the DUDES method skills exist (Next): which goals gate them (spec repo: `dudes item check`, item packs), which discipline skills they need, what a human decides.
- [ ] E1 · `[lean]` **Rules → mechanisms, the rest** — only once workers carry the flow or agents run unattended (`CORE.md` §11): commit-before-handoff, completion tokens and handoff protocol (DUDES §9–§11) → engine commits at the phase boundary; SESSION_STATE scratchpad (`rules/CONTEXT.md`) → engine state + `git log` (F4); workspace boundary (`rules/SECURITY.md`) → one sandbox for the whole ADE. Each mechanism ships, then its prose is deleted.
- [ ] E2 · `[engine]` **Convergence for text phases** — F72 covers code workers. Ollama artifact phases (QA, docs) have no reliable "no change" signal yet; decide before adding.
- [ ] E1 · `[ops]` **Known vulnerabilities in what DUDES ships** — rechecked 2026-10-12 (`make vuln-scan`): Go engine clean (go1.26.9); worker image `aider-0.86.2-p1` — aider-chat GHSA-7w7m-v5vp-w699 and GHSA-hchg-qm84-cj9p (low) and diskcache GHSA-w8v5-vhqr-4h9v (medium) have no fixed version; litellm's rows stay at Aider's pin (proxy-server side, the worker is a client). Recheck when PyPI shows aider-chat above 0.86.2; Aider's maintenance risk is in `TREND.md` Harness watch.
