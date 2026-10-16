# Product starter

Three files that make your product repo a DUDES role-play repo: an IDE agent (Claude Code, Cursor or Cline) plays the roles in turn (PM, tester, developer, QA) under the DUDES rules, in your repo, with no engine and no GPU. What that is and is not: [Two ways in](../../README.md#two-ways-in).

| File | What it does |
|---|---|
| `AGENTS.md` | The entry. Requires the DUDES rules. Names the repo type (spec and dev in one repo), the roles, the toggles (all off) and where its release is set (`rules/DUDES.md` §6, *One Repository*) |
| `CLAUDE.md` | One line, `@AGENTS.md`. Claude Code reads `CLAUDE.md`. Cursor and Cline read `AGENTS.md` themselves and do not need it |
| `Makefile` | The goals the flow runs: `make help` and `make test`. Set `TEST` once |

## Use it

1. Put DUDES inside your repo, as a git submodule: `git submodule add https://github.com/Thomas-Pietsch/dudes.git .dudes-framework`
2. Copy the files into your repo root: `cp .dudes-framework/templates/product/AGENTS.md .dudes-framework/templates/product/CLAUDE.md .dudes-framework/templates/product/Makefile .` You have a `Makefile` already: add its `test` goal to yours instead.
3. Set `TEST` in the `Makefile` to the command that runs all your tests. In a new repo you can leave it: QA (Author) sets it for the stack the TA names.
4. Open the repo in your agent and ask `As PM & Software Architect: which files do you read before you write a spec?` The answer starts with `## 🔷 PM & Software Architect`, then the `Active rules: … · Toggles: …` line.

The guides give each host's checks and fixes: [Claude Code](../../docs/CLAUDE_CODE_SETUP.md) (it adds a `.claude/settings.json`), [Cursor](../../docs/CURSOR_SETUP.md), [Cline](../../docs/CLINE_SETUP.md). The first run, on a sample product: [the 10-minute walkthrough](../../docs/WALKTHROUGH.md).

## The one path

`AGENTS.md` holds one setting, `DUDES`: where the DUDES clone is, relative to that file. It ships as `.dudes-framework`, not `dudes/` or `.dudes/`: the engine uses those two names in a repo (`CORE.md` §13). The clone sits inside your repo on purpose: `rules/SECURITY.md` forbids reading outside the workspace, and the hosts ask or refuse for paths outside it. The submodule also pins the rule version in your history. A clone next to your repo (`../dudes`) works when the host may read it; the guides say how.

## Change it

| You want | Do |
|---|---|
| A toggle on | Edit the `Toggles:` line in `AGENTS.md`, or say `Report: ON` in a prompt for one session |
| User docs, a design, a security review | Add the role to the Role Tasks table and the Flow line (`rules/DUDES.md` §4): UI/UX and Security after the PM, Technical Writer after QA (Verify) |
| Spec and code in two repos | Give each repo its own `AGENTS.md` (spec level: the PM; dev level: QA (Author), Developer, QA (Verify)) and follow `rules/DUDES.md` §4, §6 (*Cross-Repo Product Cycle*) and §12 instead of the Release section |
| Commit rules enforced by git | `git config core.hooksPath .dudes-framework/hooks`. Then a commit without a `[🔹Role]` tag is refused, and so is a QR commit on `development`. On `release/*` only QA commits, and only the QR (§12) |
| A newer DUDES | `git submodule update --remote .dudes-framework`, read what changed in `rules/`, commit the new pointer |

The rules and this `AGENTS.md` are about 47 KB of text, read at the start of every session.
