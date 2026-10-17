# The 10-minute walkthrough

One run of the whole flow on the sample product, Bookshelf ([`examples/sample/`](../examples/sample/)): you give an idea, the roles work in turn, you decide. You see what each role writes, what it leaves in git, and where the agent's word ends and your check starts.

**Time.** Setup takes about two minutes. The roles needed 4 minutes 35 seconds from the idea to your decision in the run quoted here, and 4 minutes 51 seconds in a second run. Reading what they left and deciding is the rest. Measured on 2026-10-15 with Claude Code v2.1.267, the `sonnet` alias, medium effort; your model and settings change it. Claude Code's own counter showed $1.08 for the run at API list prices.

**What you need.** One setup guide read through its *Verify* step ([Claude Code](CLAUDE_CODE_SETUP.md), [Cursor](CURSOR_SETUP.md) or [Cline](CLINE_SETUP.md)), git, make, a POSIX shell (on Windows: Git Bash) and a current Go. The idea names Go. Name your own language if you like: the flow is the same.

## 0. Set up the repo

```sh
mkdir bookshelf-demo && cd bookshelf-demo
git init -b development
git submodule add https://github.com/Thomas-Pietsch/dudes.git .dudes-framework
cp .dudes-framework/templates/product/AGENTS.md .dudes-framework/templates/product/CLAUDE.md .dudes-framework/templates/product/Makefile .
mkdir .claude
printf '{\n  "claudeMdExcludes": ["**/.dudes-framework/CLAUDE.md", "**/.dudes-framework/AGENTS.md"]\n}\n' > .claude/settings.json
printf 'SESSION_STATE.md\n' >> .gitignore
git config core.hooksPath .dudes-framework/hooks
git add -A
git commit -m "chore: DUDES starter [🔹human]"
```

The `.claude` lines are for Claude Code ([why](CLAUDE_CODE_SETUP.md#2-copy-the-starter)); Cursor and Cline can skip them. The `core.hooksPath` line turns on the git hooks of DUDES: from now on git refuses a commit without a `[🔹Role]` tag. Open the folder in your agent. In Claude Code, allow `git`, `make` and `go` when it asks ([guide](CLAUDE_CODE_SETUP.md#5-allow-what-the-flow-runs)).

## 1. The idea

```text
Idea: Bookshelf, release 0.1. A reader adds books to a shelf and sees them in the order they were added, so they find a book again. A book has a title and an author. A title that is empty or longer than 200 characters is refused, and the shelf stays as it was. A small HTTP JSON service in Go with SQLite (modernc.org/sqlite, pure Go), no UI. Marking a book read is release 0.2: not now.
```

One prompt. From here the roles hand off to each other without you (`rules/DUDES.md` §6, *Autonomous handoff*), so you read at the end, at the gate. To ask for one role only, name it: `As PM & Software Architect: …`. Each role prints its header, the toggles line and the flow, works, commits, and hands off:

```text
## 🔷 PM & Software Architect
Active rules: DUDES, SECURITY, TERMINAL, CONTEXT, TESTING, PERFORMANCE · Toggles: Convergence OFF, Report OFF, Review OFF, Vibe OFF
Flow: Feature
…
✅ COMPLETE
→ Handing off to **QA Engineer (Author)** — proceeding now.
```

## 2. The PM writes the spec

It read the sample's PRD (the requirements), TA (the technical architecture), backlog and QR (the QA report) as the format, wrote `spec/BOOKSHELF_PRD.md`, `_TA.md` and `_BACKLOG.md`, and committed them: `feat(spec): freeze Bookshelf R0.1 PRD, TA and backlog [🔹PM & Software Architect]`. That commit is the freeze: a spec change after it starts a new cycle.

Open the backlog. The work item is the unit of work, and its criteria cite the PRD and the TA by ID, as in [`examples/sample/pm/BOOKSHELF_BACKLOG.md`](../examples/sample/pm/BOOKSHELF_BACKLOG.md):

```markdown
#### BK-001 · Add a book to the shelf
**Release.** R0.1 · **Status.** derived from QR
**Story.** As a reader I add a book to my shelf, so that I find it again.

**Acceptance criteria**
- AC-1 GIVEN an empty shelf, WHEN `Sample Book` by `Sample Author` is added through [[TA:API-01]], THEN the response is `201` and the listing of [[TA:API-02]] shows one book.
- AC-2 GIVEN a shelf with one book, WHEN a second is added, THEN it is listed after the first ([[TA:ALG-01.2]]).
- AC-3 GIVEN an empty title, WHEN a book is added through [[TA:API-01]], THEN the response is `400 ERR_TITLE` and the shelf is unchanged.
- AC-4 …
```

It took `Sample Book` from the sample, because the sample is the format it was told to match. A product of your own gets its own wording. What to look for: every requirement has an ID, a SHALL line and a scenario; every criterion points at a PRD criterion and a TA entry; nothing you did not ask for.

## 3. QA (Author) writes tests that fail

It wrote one test per acceptance criterion, named after the item and the criterion (`TestBK001AC1`), stubbed the code so the tests compile (a function that panics with `not implemented`), set `TEST = go test ./...`, `FMT` (`gofmt`) and `LINT = go vet ./...` in the `Makefile`, ran `make check` and saw it fail in the tests only, then committed: `test(shelf): write tests from PRD §3 REQ-ADD and REQ-TITLE [🔹QA Engineer]`. Check it yourself in step 7.

## 4. The Developer writes the code

It built from the spec, not from the tests: in its step it opened no test file. It wrote the code until `make check` was green, changed no test file, and committed: `feat(shelf): implement add/list shelf API per BOOKSHELF_TA [🔹Developer]`.

## 5. QA (Verify) writes the report

It ran the suite again, cut `release/bookshelf-v0.1` from the verified commit, wrote the QR there and committed it (`rules/DUDES.md` §6, *One Repository*, and §12: the QR goes on the release branch, nothing else does). `rules/TESTING.md` §5 sets its shape:

```markdown
**Verdict.** ✅ Pass · **Verified commit.** eb975e2 · **Command.** `make check`

| PRD Section | Tests Written (QA Author) | Tests Passing (QA Verify) | Status |
|---|---|---|---|
| §3 REQ-ADD | 2 (TestBK001AC1, TestBK001AC2) | 2 | ✅ |
| §3 REQ-TITLE | 2 (TestBK001AC3, TestBK001AC4) | 2 | ✅ |
```

Then it stopped with an offer, not a decision:

```text
Want me to spin that up so you can check it live, or is this good to seal as `bookshelf-v0.1.0`?
```

Some runs close this message with `✅ COMPLETE`, some with `⏸ WAITING`. Either way nothing is sealed until you say so. `git log --oneline --graph --all --decorate` now shows the record:

```text
* 3a4f210 (release/bookshelf-v0.1) test(qr): verify Bookshelf R0.1 against commit eb975e2 [🔹QA Engineer]
* eb975e2 (HEAD -> development) feat(shelf): implement add/list shelf API per BOOKSHELF_TA [🔹Developer]
* 8368066 test(shelf): write tests from PRD §3 REQ-ADD and REQ-TITLE [🔹QA Engineer]
* b447ce3 feat(spec): freeze Bookshelf R0.1 PRD, TA and backlog [🔹PM & Software Architect]
* e0c6248 chore: DUDES starter [🔹human]
```

## 6. You decide

`rules/DUDES.md` §6 (*Live Demo Gate*) fixes the end: QA (Verify) offers a demo and waits, and you answer.

| You say | The rules | What happened |
|---|---|---|
| `Yes, show me the demo` | QA picks the most practical way to see the change | It built the service, started it on `127.0.0.1`, added two books with `curl` (`201`), listed them in order, sent an empty title (`400 ERR_TITLE`, the shelf stayed at two), stopped the service and deleted its files. 1 minute 32 seconds, and it asked you again |
| `Seal it` | QA closes. The PM tags `<product>-vX.Y.Z` on the release branch | The PM role tagged the tip of the release branch, the QR commit, `bookshelf-v0.1.0`: annotated, naming the verified commit. It did not push. 1 minute 3 seconds, in a new chat |
| `Skip demo` | QA closes without a demo | Not run |
| `Not good enough: …` | The feedback goes to the role that owns it, then QA runs again | See below. 2 minutes 27 seconds, and it asked you again |

After `Seal it`, `git log --oneline --all --decorate` shows `(tag: bookshelf-v0.1.0, release/bookshelf-v0.1)` on the QR commit, and `git tag -n` its message: `Bookshelf R0.1 — verified commit eb975e2 (QA QR: ✅ Pass, make check)`.

A rejection is a loop through the flow, not a patch. We sent this, in a new chat:

```text
Not good enough: a title of 150 letters é is refused as too long. The limit is 200 characters, not 200 bytes.
```

QA (Author) wrote and committed a failing test for it first (`TestBK001AC1Multibyte`, 150 × `é`, expects `201`) and saw it fail. The Developer changed the check to count characters; `make check` went green. QA (Verify) merged `development` into the release branch, ran the suite again and updated the QR in place: a merge commit and one QR commit, the verified commit changed, a row added. The new chat did not matter: the state is in git, not in the conversation.

A request outside the frozen spec starts at the PM instead. In an earlier run we answered `Not good enough: I also want to look up one book by its id. Add that.` The agent said it was new scope, handed it to the PM, who changed the spec on `development`; QA (Author) wrote the red tests, the Developer made them green, and QA (Verify) re-verified and updated the QR.

## 7. Check it yourself

Two minutes, and the part that matters. The agent says ✅; these commands say it too, or do not:

```sh
git log --oneline --graph --all --decorate         # who committed what, in which order
git switch --detach <the QA (Author) commit>       # the test(…) commit
make check                                         # red: FAIL, panic: not implemented
git switch development
make check                                         # green
git show release/bookshelf-v0.1:spec/BOOKSHELF_QR.md
```

In our run `make check` at the QA (Author) commit printed `--- FAIL: TestBK001AC1` and `panic: not implemented`: red because the behaviour is missing, which is the point of writing tests first. Your run words it differently; it must still fail, and for that reason. Then read the Developer's diff against the spec's words, not against the tests.

## Where the flow leans on the agent

This is the role-play: rule files an agent follows, one agent in one conversation. What holds it, what does not:

| The rule | What enforces it | What we saw, and the gap |
|---|---|---|
| The roles run in order, none skipped, each hands off | The agent | Four roles, in order, each with its header and handoff line. Nothing stops an agent from skipping one. The log shows who committed what |
| Tests before code | The agent | The commit order shows it. Step 7 proves it: check out the test commit and run `make check` |
| A `[🔹Role]` tag on every commit | `hooks/commit-msg`, once you set `core.hooksPath`; otherwise the agent | With the hooks on, an untagged commit is refused. Without them, nothing refuses it |
| The QR only on the release branch | `hooks/commit-msg`, once you set `core.hooksPath`; otherwise the agent | With the hooks on, a commit on `release/*` by any role but QA, or of any file but the QR, is refused; QA's merge of `development` is not. The hook knows the repo type from the `Repo type` row of the starter's `AGENTS.md`. QA's QR commit on `development` is refused too; the PM's contract rows are not |
| Roles read only their own inputs (`rules/DUDES.md` §5, Role Isolation Mandate) | The agent: one model, one conversation | "Forget" is an instruction, not a wall. The Developer's context still held the tests. Where the host has subagents, `rules/DUDES.md` §1 fixes narrower splits |
| ✅ only from a run | The agent | It ran the suite itself before it wrote ✅. Run it yourself |
| Never push, never tag before your word | The agent. Claude Code can refuse the common forms ([guide](CLAUDE_CODE_SETUP.md#5-allow-what-the-flow-runs)) | It tagged only after `Seal it`. The demo repo has no remote, so this run does not test the push. Anyone with a shell can run `git tag`: the seal is not a lock |
| The seal | You say it. The PM tags | A local, annotated, unsigned tag under your git identity |

The gates catch what the tests test. In both runs we made, the Developer checked the title with `len(title)`, which counts bytes in Go. The PRD says characters and the tests used ASCII only, so QA (Verify) wrote ✅ and the defect reached your gate. A report is only as good as the tests behind it. That is why you decide, and why step 7 asks you to read the diff against the spec's words.

Two more things we saw:

- The role header is meant to be the first line of the first response (`rules/DUDES.md` §8), but `rules/PERFORMANCE.md` says to call tools at once. The agent read the rule files first and printed the header after them.
- In a new chat with a short message (the rejection above) it did the whole fix and printed its header and a summary at the end, not at the top.

The engine ([Two ways in](../README.md#two-ways-in), door 2) turns most of these rows into code that a model cannot argue with: the sequence, the gates, and the seal at the end. The tour in [Five minutes, no model](../README.md#five-minutes-no-model) shows the same ending with `dudes seal`.

## What next

- Your own product: [`templates/product/`](../templates/product/) says what to change (roles, toggles, two repos instead of one).
- The rules in full: [`rules/DUDES.md`](../rules/DUDES.md).
- Something broke or differs: open an [issue](https://github.com/Thomas-Pietsch/dudes/issues/new/choose) with what you saw.
