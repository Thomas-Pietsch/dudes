# Model bench ladder (F82)

Frozen micro-agent tasks of rising size. Each one runs through the real engine: worker image, scope, gate, gate feedback. It runs per model on the model host, from the same start every time. The ladder answers one question for framework users: **how far does a local model get as a DUDES worker?** `skills/framework/bench` reads the results into `TREND.md` §1.1.

To run it on your own hardware and post the result: [Your 27B in 10 minutes](../docs/LOCAL_LLM_GUIDE.md). Terms used below: an *instance* is one company's use of DUDES, with its own repositories (`CORE.md` §5); an *instance trial* is a run of the engine on a real product's work items; a *slice* is one part of a work item, with its own tasks and gate; a *TA* is a spec set's technical-architecture document; an *item pack* is the context the engine builds for one work item (`dudes item pack`).

```sh
make bench-smoke MODELS=qwen3.8:27b OLLAMA=http://<model host>:11434 LEVELS=L1   # first: one worker step, how the model reacted (react=)
make bench-smoke MODELS=… OLLAMA=… LEVELS=L1 THINK=on NUM_PREDICT=32768 NUM_CTX=65536 TEMPERATURE=default   # the same with other settings of the model
make bench-models MODELS="qwen3.8:27b gemma4:26b" OLLAMA=http://<model host>:11434
make bench-models MODELS=qwen3.8:27b OLLAMA=… LEVELS=L1 REPS=1 STEPS=3 KEEP=1   # one quick run, workspace kept
make bench-models MODELS=qwen3.8:27b OLLAMA=… LEVELS=L1 REPS=1 NOTES=0           # one run, no git note
make bench-models WORKER=pi MODELS=… OLLAMA=…                                    # Pi as the code worker (sh scripts/worker.sh --build-pi <gate image> first)
```

**Smoke first.** A series costs hours, and a model that does not answer the worker's request wastes it: it echoes one of Aider's canned turns, reasons until the reply cap, or answers in another shape. `make bench-smoke` runs one step of one run (`REPS=1 STEPS=1 NOTES=0`) and the gate once on what the model wrote; the `react=` field of its line says how the model reacted ([Results](#results)). Start a series when it reads `edit:<n>` with n covering the task's files. Otherwise change one setting (`THINK`, `NUM_PREDICT`, `NUM_CTX`, `TEMPERATURE`) and smoke again, or drop the model. Run the series with the settings the smoke passed with. The smoke's outcome is no result.

| Level | Item | What the worker must do |
|---|---|---|
| L1 | `BL-001` Slugs | One package file and the real test of one acceptance criterion |
| L2 | `BL-002` Version order | A package whose behaviour lives in a TA reference (`ALG-01`). The worker reads it through the item pack. Three criteria |
| L3 | `BL-003` Latest release | Several constraints in one task across two packages: a new package (`release`, TA `ALG-02`) on top of an existing one (`semver`, committed from `L3/testdata/base/`) that gains a function and must keep its behaviour. The hidden test replays L2's cases |
| L4 | `BL-004` Release notes | One item in a real item's shape: the acceptance tests are given, approved and read-only (`L4/testdata/base/tests/`), one implementation task, the writable scope the new package alone. Its details are the ones every model has failed on in a real item (`TREND.md` §1.1): a release-notes file rendered from the committed `testdata/notes.txt` (a new package `notes`, TA `ALG-03`) that must equal the committed `testdata/release-notes.md` byte for byte, an error that names the input and the line (`ERR_NOTES: <name>:<line>:`), a contract across packages (`notes` orders and classifies through `semver` and `release`, given read-only from `L4/testdata/base/`), and the old packages kept, outside the writable scope. The hidden test checks the trailing newline, a `v` prefix kept as given, a candidate's placement, an empty notes file, the imports of `notes`, and replays L2's and L3's cases. The first ladder level with a reference solution (`L4/testdata/reference/`) |

From L4 up the rungs are the bench product's items in a real item's shape: approved read-only tests, a reference solution, a hidden test. A rung counts once its reference check is green and at least one configuration passes it 2 of 3 while another fails it; a rung every configuration passes or none does measures nothing and is reworked before the next rung. Replay levels of real items are external levels, named after the item they replay and kept in the instance that owns it ([External levels](#external-levels) below; the first ran 2026-10-14, `TREND.md` §1.1): the hard comparison, not rungs. An instance trial (ended after its fifth slice, `TREND.md` §1.1) shows what such a level must hold: a cross-toolchain slice and a measure of whether a failed pass leaves the tree worse. A level adds its own TA section and release to the shared spec set, so the packs of the levels before it stay unchanged.

**External levels.** A level built from a real product lives in the instance that owns its specs and code, never here. Name its directory in `LEVELS` (an absolute path; `make bench-models … LEVELS=C:/Coding/<instance>/bench/<level>`). Its name in result lines and logs is the directory's basename. See [External levels](#external-levels).

## How a run works

`scripts/bench-models.sh` creates a fresh git repository from `workspace/` (the spec set, `go.mod`, and `dudes/` as `.dudes/`) plus `<level>/contract.json` (an external level: from its own repository, below). The run then goes through these steps:

1. `dudes init`: item-bound, so the run starts at test_authoring.
2. `freeze`.
3. One step for test_authoring, which writes one failing test file per criterion. The level's `testdata/base/` is then copied over again and committed when anything changed: a criterion whose approved test the level gives keeps it (L1–L3 give none; L4 gives all three of its item).
4. Steps until `dudes status -json` counts every contract task done (`tasks_done` = `tasks_total`), with up to `STEPS` steps per task. A task that waits for a decision (`approval`, F66) is approved with `dudes decide`, since a bench has no human; the log says so, and a pause before a task is not a step.

The gate is `go test ./...` (or the level's `TEST_COMMAND`), run in `golang:1.26-bookworm` (or the level's `GATE_IMAGE`) through `test_launch`. The worker runs in the worker image with the whole edit format (`EDIT_FORMAT=diff` for search/replace blocks; the result line names it), `num_ctx` 32768 (`NUM_CTX`), `num_predict` 12288 (`NUM_PREDICT`), temperature 0 (`TEMPERATURE`) and thinking off (`THINK`). The result line names thinking, reply cap and temperature. They are Aider's settings, written to the generated `aider-models.yml`; Pi keeps its own (its line says `n/a`).

**Level settings:** a level that needs more declares it in `<level>/level.env`, one `KEY=value` per line. The script reads the file without sourcing it and refuses unknown keys. L1–L4 have none.
- `GATE_IMAGE`: the gate's toolchain image (default `golang:1.26-bookworm`). The hidden test runs in it too.
- `GATE_DOCKER=1`: the gate gets the host's Docker (`scripts/worker.sh --docker`), for tests that start containers.
- `STEPS`: the level's steps per task (default 3).
- `TEST_COMMAND`: the full suite, words split on blanks, no quotes (default `go test ./...`). It is the engine's `test_command`, so the gate of a task without its own `test` — the last one — runs it. The final rerun adds `-v` after `go test`.
- `NUM_CTX`: the worker model's context window (default 32768) — a level whose task packs need more says so here, and the result is read with it.
- `THINK` (`on`, `off` or `none`, default off) and `NUM_PREDICT` (default 12288): the worker model's thinking and reply cap. `on` also sets `streaming: false` (the pinned litellm's streaming parser can drop answer text that arrives beside a thinking key). `none` writes no thinking setting, for a model without the thinking capability: Ollama answers HTTP 400 to `think: true` there.
- `TEMPERATURE` (default unset): Aider sends temperature 0 unless told otherwise, which overrides every model tag's own default (a vendor's 0.7 or 1.0, say), so the vendor's sampling is a setting of a series, not of a Modelfile. Unset writes nothing (temperature 0); `default` lets the tag's own temperature apply; a number such as `0.7` sets it.
- `BASE_REPO`, `BASE_REV`, `SPEC_<root>`: an external level's fixture (below).

**Precedence.** `STEPS`, `NUM_CTX`, `THINK`, `NUM_PREDICT` and `TEMPERATURE` are also variables, and for these five the order is the same: the variable on the command line, then the level's `level.env`, then the default. A variable set to the default's own value still beats the level's, so `make bench-smoke` runs one step on a level that declares `STEPS=3`, and `NUM_CTX=32768` overrides a level's 65536. The result line names `think=`, `predict=` and `temp=` but not the context: read that from your command or the level. Every other key comes from `level.env` alone.

**Outcomes:**
- `pass`: every task done, the last task's gate (the full suite) green **and** the level's hidden test green. `<level>/testdata/verify_test.go` (or the package `<level>/testdata/verify/`, with its own test data) is checked against the spec; the worker never sees it, and it is copied into the run's `verify/` only after the gate.
- `weak`: every task done and the gate green, the hidden test red. The item's own tests missed part of the spec, either never tested (L1: a non-ASCII title kept its letters against "lowercase ASCII only") or weakened until green (an instance trial: a deleted route-matrix check, a negative control that silently skipped).
- `blocked`: the worker escalated.
- `fail`: a task used up its steps, or the last task's gate was a slice rather than the full suite.

For L1–L3 (an author-tests task, then one implementation task) and L4 (one implementation task against the given tests) this is the same as before: the run ends at the first green gate.

**Headline per model:** the highest level it passes in at least 2 of 3 runs.

## Results

There is one line per run, appended as a git note on the framework commit that ran it (`refs/notes/dudes/bench`). Notes stay in the clone of the runner who ran the bench and are never pushed, so this repository holds no run lines. The summaries that count, failures included, are in `TREND.md` §1.1. Each line records the date, level, model, run, outcome, steps, minutes, tokens per step, engine commit, worker image and model host. It also records two fields:

- `tasks=<done>/<total>`: how many of the contract's tasks the run finished.
- `ac=<best>/<final>`: how many of the item's acceptance tests were green. The acceptance tests are the top-level tests named in the acceptance-run notes (`refs/notes/dudes/acceptance`, F65), which the engine writes only after a green `go test` gate.
  - `best` is the most that were green at any green gate.
  - `final` is the count at the end. It comes from the last note when the run's last commit carries one. Otherwise the run ended on a red pass, and the script reruns the suite once verbosely in the gate image, logged under `=== final acceptance rerun`. Code that no longer builds counts as 0 green.
  - `best` above `final` means a later pass made the tree worse; an instance trial's last slice would read `9/0`. `ac=none` means no gate was ever green.

Lines written before these two fields existed lack them. `NOTES=0` skips the note, for smoke runs.

A line has this shape (`<…>` stands for a value, `[+…]` for further steps):

```text
<date> <level> <model> run=<n>/<reps> <outcome> steps=<n> tasks=<done>/<total> ac=<best>/<final> min=<n> tokens=<sent>/<received>[+…] engine=<commit> image=<worker image> worker=<aider|pi> host=<model host> edit=<whole|diff|n/a> think=<on|off|none|n/a> predict=<n|n/a> temp=<0|default|n|n/a> react=<class>[+…] [stop=<rule>]
```

`worker=` names the code worker (`WORKER=aider|pi`, default `aider`); lines written before it existed are Aider's. `edit=` is Aider's edit format, `n/a` for Pi. A Pi run that a stop rule ended carries `stop=no-tool-call` (a run in which Pi ran and made no tool call) or `stop=scope-or-gate-violation` (the engine discarded a path out of scope). `tokens=` are the engine's worker events of the run, for both workers: `dudes status -json` sums them (`worker_calls`, `prompt_tokens`, `completion_tokens`) and the bench takes each step's change. Aider's calls include its reflection calls, in Aider's rounded count (0.1k under 10k, 1k above); Pi's are the sums of the assistant messages' usage in its JSON events. The first step's count includes the test authoring that runs before it.

`think=`, `predict=` and `temp=` are the Aider worker's thinking, reply cap and temperature in that run (`THINK`, `NUM_PREDICT`, `TEMPERATURE`; `temp=0` when none was set), `n/a` for Pi. `react=` is how the model reacted to the worker's request: one class per step, in the order the steps ran, joined with `+` like `tokens=` (`n/a` for Pi, whose tool calls and stop rules are `stop=`). Lines written before these fields lack them. The classes:

- `edit:<n>`: Aider applied edits to n different files.
- `none`: the code worker produced no file changes — the reply held no file listing (an echoed canned turn, reasoning prose, a question).
- `cut`: the reply hit a token limit (`num_predict` or `num_ctx`); Aider applies nothing from a cut reply.
- `blocked`: the worker escalated (`BLOCKED:` or `CONTRACT:`, F74).
- `err`: any other worker error (a timeout, Aider's exit status), or an output none of the classes explains.

A reaction is no outcome: a run can react `edit:6` and fail its gate. It is what `make bench-smoke` reads.

- Read your own runs with `git log --notes=dudes/bench`. The notes live in your clone only; this repository publishes no run lines.
- The same lines are in `.dudes/bench-models/<stamp>-results.txt`. The logs and the Aider chat of every step sit beside them. That directory is gitignored: raw logs stay out of the repository.
- **To contribute a result, open an issue with your result lines**: the [bench result form](https://github.com/Thomas-Pietsch/dudes/issues/new?template=bench-result.yml) asks for them, your hardware and your versions. Do not push notes. Maintainers decide what goes into the table in `TREND.md` §1.1.
- `host=` holds the URL you gave as `OLLAMA`. An issue is public: replace the URL with a short label of your machine (`host=RTX-4090-24GB`) before you post it.

## External levels

An external level has the layout of `bench/<level>/` plus `dudes/` with `profile.json` and `specset.json`, which become the fixture's `.dudes/`. Its `level.env` says where the fixture comes from:

- `BASE_REPO` and `BASE_REV`: the fixture is that repository's tree at that commit (`git archive`), committed once into a fresh repository, then `testdata/base/` overlaid. It is one commit, not a clone with history: later commits, and with them the reference solution, are out of the worker's reach, and the acceptance count walks one short history. `BASE_REV` must predate the level, so that the level itself is not in the fixture.
- `SPEC_<root>=<repo>@<rev>`: the spec root `<root>` pinned. The runner makes a shared clone of `<repo>` at `<rev>` beside the fixture, at `../spec-<root>`; the level's `specset.json` names the root that way. The clone writes nothing into `<repo>` and is removed with the run.

Relative paths start at the level directory as it sits in its repository's main checkout. A level on a linked worktree (`.worktrees/<branch>/bench/<level>`) therefore resolves as it will after the merge. `TEST_COMMAND` names the suite; the contract's last task carries no `test` of its own, so its gate is that suite.

## Reference check

```sh
make bench-models MODELS=reference LEVELS=C:/Coding/<instance>/bench/<level>
```

`MODELS=reference` runs no worker and needs no model host. L4 is the first ladder level with a reference solution: `make bench-models MODELS=reference LEVELS=L4`. Per level it builds the fixture through test_authoring, as a run does. It then runs the full suite in the gate image twice: on the level's base alone, which must be red, so the level is not trivially green; and with `testdata/reference/` overlaid and committed, which must be green with no acceptance test skipped. Last, it runs the hidden test, which must be green. One line per level, `<level> reference pass|fail base=red suite=green hidden=green ac=<green acceptance tests> min=…`, with the log under `.dudes/bench-models/<day>-<level>-reference.log`. There is no git note and no headline. A level without `testdata/reference/` reports `reference none`.

## Changing the ladder

A level is frozen. Once a level has results, change its fixture only by adding a new level, or bump the level and say so in `TREND.md` §1.1, because old and new results are not comparable. The fixture follows the instance format (profile, spec set, item-bound contract). It is a small spec repo, so `dudes item check` must report 0 findings on it.
