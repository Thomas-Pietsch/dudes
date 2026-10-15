---
name: release
description: Release branches, seal tag and release-gate checks — which mechanism enforces each step, and what DevOps still checks by hand. Workspace DevOps only. Do not load in product role-play or inject into Go workers.
---

# Framework release

Use at spec freeze or release gate.

**Do not** load in a product Feature/Fix flow. **Do not** add to `recipes/<phase>.json`.

Git is the record — no reference file. Branch and tag names carry the trace. The cycle (freeze → implement → QR → human seal → tags) is `rules/DUDES.md` §6, *Cross-Repo Product Cycle*; do not restate it.

## What enforces each step

Run or trust the mechanism; do not re-check by hand what it already refuses.

| Step | Mechanism |
|------|-----------|
| On `release/<product>-v<X.Y>` of the spec repo only QA commits, only `*_QR.md` | `hooks/commit-msg`, installed in the instance repos by `make hooks` (F78) |
| The dev run reads the frozen spec, unchanged | Freeze gate: `refs/notes/dudes/freeze` re-hashed before implementation and by `dudes check` (F31, F59) |
| The human seal | `dudes seal [X.Y.Z]` — refused unless the run converges; with a version it tags the engine workspace `<product>-vX.Y.Z`; the tag is signed when the sealer's git can sign, and `dudes verify` checks it (F89) |
| A dev repo's release tag comes after the PM seal tag | The instance goal (for example `make release` with a `ci/release.sh`) refuses without the tag in the spec repo |

One tag path per dev repo: when the instance has its own release goal (changelog, signing, push), seal the run without a version (`dudes seal`) and let the goal set `<product>-vX.Y.Z` — both would collide on the name.

## What DevOps still checks

The hook runs only where it is installed, and `--no-verify` skips it. At the release gate:

- `git diff --name-only development...release/<product>-v<X.Y>` in the spec repo lists only `*_QR.md` — anything else **blocks** the release.
- The QR names the dev commit the release tag will point at.
