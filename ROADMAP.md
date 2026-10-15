# DUDES Framework — Roadmap

> **The path:** the lanes DUDES moves in, the stage each lane is at, what counts as that stage working, and where it all leads. Work items are in `TODO.md` (every line names its lane), verdicts and tool actions in `TREND.md`, design in `CORE.md`. Each file's job: `CORE.md` §14. This file changes as fast as the world does; `git log` is its history. `make roadmap-check` ties `TODO.md` to it and says when a step back is due (`skills/framework/roadmap`).

**Reviewed:** 2026-10-15 · `2e81fa9` · carried over to the public history, unchanged since the first step back on 2026-10-14, which fresh eyes on another model challenged: no exit check moved since 2026-10-09, and lane `R` had no Now line. `R`'s release friction moved to Now. `E1`'s exit check now names distinct items, an outside reviewer and a ledger, and the replay is no longer called a regression test. The lanes and exit checks await the human's word.

## Lanes

The role-play delivers today. The engine takes over the flow's steps one stage at a time. Going public speeds both up, and it starts now. A stage moves only on its exit check.

| Lane | Stage | Status | Works when (exit check) |
|---|---|---|---|
| `R` | **The role-play delivers.** Cloud models in the IDE run the whole flow for the first instance, a small company (`rules/DUDES.md`, its toggles). Now: the first instance's workshop product, from its first sealed release to its first real-world use | now | The release passes its QA report and runs in that use. Dates and items are the instance's (its lead repo's roadmap) |
| `E1` | **Engine: dev.** A local worker writes the red tests and the code of an item's Go part from the approved spec. The cloud role-play reads the tests, reviews the code and decides (`dudes decide`, `dudes redo`). The code is written locally; the cloud only reads and decides. An item's UI part stays with the role-play until a slice across two toolchains passes (`TREND.md` §1.1, the instance trial's fifth slice) | now | Three distinct real items in a row (slices of one item count once), handed over by the role-play's Developer through the documented bridge, land on `development` with the instance's gates green. Their code and tests come from engine commits. The reviewer is the role-play or the human, never the DevOps session driving the engine, and DevOps fixed nothing while they ran. Each item has a ledger line: the review's cloud tokens against a control run of the same item in the role-play, local minutes, DevOps touches |
| `E2` | **Engine: spec.** PM, UX and security drafts as micro-agents, gated by the consistency check (`dudes item check`), carrying the DUDES method as skills (`CORE.md` §2.4) | later | Written when `E1` passes |
| `E3` | **Engine: lead.** Strategy and C-level work as micro-agents | later | Written when `E2` passes |
| `P` | **Public.** The framework repo opens on GitHub as soon as it is clean (the human's gate, 2026-10-14). Its offer is honest bench notes and a product delivered under the flow, so outside runs bring the models and harnesses `E1` lacks | now | The repo is public, with no secret, no personal data and no first-instance data in its tree, history or notes, under the Apache-2.0 license, and with a README an outsider can start from. `make public-check` in `make ci` keeps it so. The first instance keeps working from its own repos, its engine setup in their `dudes/` (F88) |

## What the lanes hold to

- **`R` must not break while `E` grows.** A framework change that slows or breaks an instance's run is erosion; the run's need enters as a source (`CORE.md` §15). The role-play keeps its rules and toggles. They exist for cloud models in the IDE, never for engine workers.
- **In `E` the mechanisms are the cage, not the context.** A worker gets the task, its goals and the output contract (`CORE.md` §2.4). When a model falls short, the stage waits for a better model or changes a mechanism (the slicing, a gate, a setting). It never grows the worker's prompt.
- **`E1` runs on what `R` specifies.** It needs spec-ready Go items. The first instance's workshop product has one item ready; the others wait for refinement. Its factory-quality product has a longer backlog of written items, many of them Go only, and the instance trial already built the first one's Go part; they are the better supply while the others wait. The CEO allowed it on 2026-10-14: the frozen product's code may move through the engine against its frozen spec, and its documents stay untouched.
- **`E1`'s case is measured, not assumed.** The ledger shows what the review costs in the cloud against the role-play doing the same item. While the cloud reads the code, `E1` saves cloud work but not sovereignty. Keeping the code on the box needs a local or human review, a later stage's question.
- **The bench follows the stages.** The engine's regression tests are its Go tests and the stub lifecycle (`make ci`, `make bench`). The replay of real product slices (a level in the instance's dev repo) is not one: no model passes it, and three runs place a model without attributing a change. It measures a new model or harness, or an engine mechanism on the runs where its case occurs; its reference check (`MODELS=reference`) shows the level still holds. Each item `E1` ships can become a replay level the same way, a bench made of real work. Models and harnesses are compared when a stage needs it (`skills/framework/bench`).
- **The order may change.** The stages run dev, then spec, then lead, because models are trained hardest on code and tests settle it (`CORE.md` §2.4). When the world moves, the step back is where the order changes.

## Horizon

Direction, not a stage. Today DUDES is a local-model ADE for one product lifecycle. The human still seals the outcome.

**Where it goes.** An agentic company environment: the same engine, still local, doing the repeated work of the company and not only one increment. It keeps running. You see what is actually in progress, and you instruct it by voice or by text.

Text stays the spec. It is the precise contract. A picture is the dense view — shape, color, and status together — for a person watching the run, and maybe for a worker when one image replaces a long dump. Whether that saves tokens is something to measure. It is not a reason to drop the spec.

**What we rely on.** Git is the database. The engine is IDE-independent and OS-independent: no editor tree and no host vendor is the product. Today the doing still sits in Claude and Cursor, and the UI is text and prompting. Workers take over pieces of that work, lane `E`. The backend we are moving toward is a local process in the DUDES way: Go, a local model, Git as the record.

**What we do not build now.** The face that speaks, understands what the company needs, and shows monitors of the running work. Open-source desks are candidates, not a guarantee that a compatible face will arrive. Do not schedule one without evidence of an unmet need and human adoption; reassess when a desk can operate over DUDES's record without owning the lifecycle. Nor does DUDES become a personal assistant: "it keeps running" means a trigger starts bounded runs with state in Git; an assistant is at most the face. An OS or appliance around DUDES is a deployment choice, not DUDES.

**How a pattern is built.** Simplest fit first: shell, then git, then make and Docker, then Go or an existing tool — the pillars in `CORE.md` §11. The phase machine stays Go. Git as database: `CORE.md` §10. Workers get the task and their goals, nothing else: `CORE.md` §2.4. An instance's product toolchain (for example a CI container) is the instance's choice (`CORE.md` §13).
