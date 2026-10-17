---
name: quality-report
description: Write the QR in its two parts, the contract rows before the code and the verdict rows after the run. Every row names its requirement or scenario ID, a row turns green only with a test ID, the verified commit and the run reference, and the file states the current run only. Use when the PM shapes the QR before the freeze or QA (Verify) writes the verdict.
---

# Quality report

Read `skills/spec/README.md` first: every row carries an ID, and the gate is the consistency check.

The QR exists before the code, as the executable QA backlog: what the release must show, row by row. The PM loads this to **shape** the contract part; QA (Verify) loads it to **write** the verdict part. Which tests exist and how a run is made is `rules/TESTING.md`.

## Two parts, two owners, two branches

| Part | Holds | Written by | On |
|------|-------|------------|----|
| Contract | the release exit sets, the test matrix with its IDs, each row planned | the PM | `development`, before the freeze |
| Verdict | the results, the verified commit, releasable or not | QA (Verify) | `release/*` |

`hooks/commit-msg` splits them by branch: it refuses QA's staged QR on `development` and keeps `release/*` to QA and the QR (`rules/DUDES.md` §12). No hook tells the parts apart inside the file; the review does. A contract row changed after the freeze is a new freeze, never an edit by QA.

## A row

- **It names its ID:** the requirement or scenario it verifies, as its owner document defines it. A row with no ID verifies nothing the spec asked for. The shape is `examples/sample/pm/BOOKSHELF_QR.md`: ID, row, wave, status.
- **It exists before its code,** planned. A requirement with no row is a gap the consistency sweep finds.
- **It turns green only with three things:** the test ID that ran and the run reference, in the row, and the verified commit, in the verdict. Without them the row is a claim and stays red.

## The status ladder

| Mark | Means |
|------|-------|
| ⬜ | Planned: the row exists, nothing has run |
| 🔴 | The test fails in the current run |
| 🟢 | The test passes, with its evidence |
| ⚠️ | Blocked: the row names its owner and the condition that unblocks it |

The PM writes ⬜; QA writes the rest. A blocked row reads `⚠️ <owner>: <unblock condition>`; one without an owner is a finding. The marks are the instance profile's (`examples/sample/profile.json`); the derived item status reads them (`backlog-item`).

## One verdict per version

The QR states the current run only: the verdict, the verified commit and the rows. A re-run updates it in place; each run is one QR commit and `git log` of the QR is the run history. `hooks/pre-commit` refuses new dated "run N" rows (`rules/TESTING.md` §5).

## The sweep reads the rows

The consistency sweep reads the QR rows (`rules/DUDES.md` §6): every DoD ID of the backlog has one, every ID resolves, and the milestones and the rows agree. The PM runs it before the freeze commit, so the contract part is whole by then.

## Not here

- The tests, their layers and the run itself: `rules/TESTING.md`.
- The work item and its DoD IDs: `backlog-item`.
- Whether the spec can be built without asking: `requirements-review`.
- The seal: the human's, after the verdict (`rules/DUDES.md` §6).
