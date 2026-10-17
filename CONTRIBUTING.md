# Contributing to DUDES

Contributions are welcome: code, docs and measurements. The most useful thing an outsider can send is a measurement, and it needs no code.

## Good first contributions

| What | Why it helps | How to start |
|------|--------------|--------------|
| **Bench results on other models or hardware** | Every number in [`TREND.md`](TREND.md) §1.1 comes from one model host. Outside runs are what the bench lacks most | [Your 27B in 10 minutes](docs/LOCAL_LLM_GUIDE.md), then open a [bench result issue](https://github.com/Thomas-Pietsch/dudes/issues/new?template=bench-result.yml) with your result lines. Failures count. `WORKER=pi` runs the second harness ([`bench/README.md`](bench/README.md)); the one paired trial so far is `TREND.md` §1.1, 2026-10-16, Pi stopped |
| **A non-Go test-evidence adapter**, JUnit XML first | The evidence gate reads per-test results from `go test -v` output (`internal/testrun`). An adapter that turns JUnit XML into the same pass, fail and skip results lets the gate read pytest, Jest, JUnit and others | Open an issue with the format and a sample first. The engine uses the Go standard library only (`go.mod` has no dependencies) |

## How to contribute

1. Fork the repo.
2. Create a branch from `development` (`feat/your-thing` or `fix/your-thing`).
3. Run `make setup` once. Make your changes. Run `make ci`, and `make doctor` if you touch the CLI or the config.
4. Open a pull request against `development`.

Bug reports and ideas: open an issue, with the [bug form](https://github.com/Thomas-Pietsch/dudes/issues/new?template=bug.yml) for a bug. Be specific: what you expected, what happened, which files or rules are involved. A vulnerability goes to [`SECURITY.md`](SECURITY.md), privately, never into an issue.

## What `make ci` checks

`make ci` is the gate every change passes, on your machine and on GitHub ([`.github/workflows/ci.yml`](.github/workflows/ci.yml)). It takes a few minutes; `go test` is the long part. `make help` lists every goal.

| Goal | Fails when |
|------|------------|
| `vet` | `go vet ./...` reports a problem |
| `conventions` | A Go file is not gofmt-clean, a script or hook is not POSIX `sh`, a Makefile goal has no `##` description, or a relative Markdown link does not resolve |
| `test` | A Go test fails (`go test ./...`, which includes the skill naming rules) |
| `public-check` | A tracked file holds private data: a private IPv4 address, a home path with a user name, an e-mail address outside `example.com` and the noreply domains, or a pattern from your own `.dudes/private-markers`. Use `192.0.2.x` and `<user>` in examples |
| `todo-check` | The Now list in `TODO.md` holds more than 3 items |
| `roadmap-check` | An open `TODO.md` line does not start with a lane ID (`R`, `E1`, `P`, … the first column of the table in `ROADMAP.md`), or the review stamp in `ROADMAP.md` names a commit your clone does not have (use a full clone, not `--depth 1`) |
| `trend-check` | `TREND.md` and `FEATURES.md` disagree: an F-ID that does not exist, a `shipped` status that `FEATURES.md` does not confirm, a missing index entry or evidence row |

The last three read the planning documents. They pass unless you edit `TODO.md`, `ROADMAP.md`, `TREND.md` or `FEATURES.md`; if you do, keep the rules above. `roadmap-check` and `trend-check` also report what is due for review (an old review date, a step back that is due). That is a report, not a failure.

## What's in scope (core framework)

| Area | Files |
|------|-------|
| Go process engine | `cmd/dudes/`, `internal/`, `go.mod`, `Makefile` |
| Capability registry | `FEATURES.md` |
| Design / architecture | `CORE.md`, `ARCHITECTURE.md` |
| Mission / trends | `MISSION.md`, `TREND.md` |
| IDE-agent behaviour | `rules/*.md`, workspace `AGENTS.md` (entry), `templates/product/` (the starter entry for a product repo). `CLAUDE.md` is a thin Claude Code pointer, not a second process. No `.cursor/` tree. |
| Framework-agent skills | `skills/framework/` (DevOps only; not recipes) |
| Automation scripts | `scripts/*.sh`, `hooks/` |
| Documentation | `README.md`, `CONTRIBUTING.md`, `SECURITY.md`, `docs/`, `bench/README.md` |
| Repository automation | `.github/` (CI workflow, issue forms) |

## What's out of scope

- Instance layout: named roots in `examples/<product>/specset.json`. Example: `examples/sample/`. There is no `COMPANY.md`.
- Repo-level `AGENTS.md` files: instance-specific, grown per product from the starter in `templates/product/`.
- Product specs (`*_PRD.md`, `*_PS.md`, etc.): your business, not ours.
- `FRAMEWORK_MANIFEST.md` and `FRAMEWORK_BACKLOG.md`: retired; do not recreate.
- `.cursor/rules`: retired; rules load from `AGENTS.md`.

## Style

- **Markdown:** follow the existing formatting. Short sentences, facts, tables over prose, no hype.
- **Shell scripts:** POSIX `sh` (`#!/bin/sh`, `set -eu`), no bashisms: the same script runs in Git Bash, Linux and macOS. `make conventions` parses every script with `sh -n`.
- **Go:** gofmt-clean and `go vet`-clean, standard library only unless you can say what the standard library cannot do.
- **Commits:** `<type>(<scope>): <description> [🔹<Role>]`, with a type such as `feat`, `fix`, `chore`, `docs`, `ci` or `test`, and a short subject. The role tag (🔹 is the small blue diamond, U+1F539) is the audit trail: `git log --grep='🔹'` finds who did what. `make setup` installs this repo's git hooks. The `commit-msg` hook refuses a message without a `[🔹…]` tag, and the `pre-push` hook runs `make ci` first. As a contributor, use `[🔹Contributor]`.

## Engineering principles

All contributions must respect: **YAGNI · DRY · KISS · SOLID**

- Don't add features "for later"
- Don't duplicate what exists: reference it
- Keep it simple
- Respect clean architecture

## License

By contributing, you agree that your contributions are licensed under the [Apache License 2.0](LICENSE), as its section 5 says.
