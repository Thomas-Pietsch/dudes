# Cursor setup

Run the DUDES role-play in your own product repo with [Cursor](https://cursor.com): the Agent plays the roles in turn (PM, tester, developer, QA) under the DUDES rules and stops for your word. No engine, no GPU. Cursor reads `AGENTS.md` by itself, so the starter's `AGENTS.md` is the whole adapter. Other hosts: [Claude Code](CLAUDE_CODE_SETUP.md), [Cline](CLINE_SETUP.md).

**Status.** The starter was run and checked on Claude Code ([how](CLAUDE_CODE_SETUP.md#4-verify)). This guide follows Cursor's documentation, linked where it matters, and has not been run in Cursor yet. If a step differs for you, [open an issue](https://github.com/Thomas-Pietsch/dudes/issues/new/choose).

## What you need

| | |
|---|---|
| Cursor | A current version, with Agent chat |
| Your repo | A git repo, any language. Not the DUDES clone |
| make | The flow runs `make test`. On Windows: Git Bash and GNU make ([Platforms](../README.md#platforms)) |
| A model | The model picker of the Agent chat. [`TREND.md`](../TREND.md) §1.2 says which hosted model fits which work, with sources and the check date. Every hosted model receives your code and your specs |

## 1. Put DUDES inside your repo

```sh
cd your-repo
git submodule add https://github.com/Thomas-Pietsch/dudes.git .dudes-framework
```

Inside the repo, because Cursor works on the folder you open and `rules/SECURITY.md` keeps the agent inside it. The submodule also pins the rule version in your history. Other layouts are at the end.

## 2. Copy the starter

```sh
cp .dudes-framework/templates/product/AGENTS.md .
cp .dudes-framework/templates/product/Makefile .
printf 'SESSION_STATE.md\n' >> .gitignore
```

You have a `Makefile` already? Skip the second line and add its `test` goal to yours. Give every goal a `## what it does` comment: `make help` lists them, and that list is all the agent knows about your commands. `SESSION_STATE.md` is the agent's scratchpad (`rules/CONTEXT.md`), never committed. The starter's `CLAUDE.md` is for Claude Code; Cursor does not need it.

## 3. Set the one path and the test command

| Where | Setting | Leave it when |
|---|---|---|
| `AGENTS.md`, table under *Framework Requirement* | `DUDES`: the DUDES clone, relative to `AGENTS.md` | You followed step 1 (`.dudes-framework`) |
| `Makefile` | `TEST`: the command that runs all your tests | You want QA (Author) to set it for the stack the TA names |

Then commit: `git add -A && git commit -m "chore: DUDES starter [🔹human]"`.

## 4. Verify

1. **File → Open Folder**: the root of your product repo. The clone is part of it.
2. Open a **new Agent chat**, not Tab autocomplete: the flow edits files and runs `make`.
3. Ask:

```text
As PM & Software Architect: which files do you read before you write a spec?
```

The answer starts like this. The wording after it is the model's:

```text
## 🔷 PM & Software Architect

Active rules: DUDES, SECURITY, TERMINAL, CONTEXT, TESTING, PERFORMANCE · Toggles: Convergence OFF, Report OFF, Review OFF, Vibe OFF

Flow: Single-Role
```

The role tag is the first line and the toggles line says which rules are on. A tag without the toggles line means the entry loaded and the model skipped a line: ask for it. No tag means `AGENTS.md` was not read: see the table at the end. To see the other side, empty `.dudes-framework` (a submodule you did not initialise): the whole answer is `DUDES framework not loaded. Cannot proceed.` That is the rule working.

## 5. Let the flow run its commands

The flow runs `git` and `make` all the time. Cursor's **Run Mode** decides when Agent asks before a command ([Run Modes](https://cursor.com/docs/agent/security/run-modes)). In the default, Auto-review, allowlisted commands run at once, other shell commands run in a sandbox when possible, and the rest go to a classifier that allows or blocks them. Allowlist narrow: `make`, and the `git` commands you accept. Run Everything has no sandbox and no classifier. Do not use it on a repo you care about.

## Day to day

| You say | What happens |
|---|---|
| `Idea: … Build it.` | Feature Flow: PM, QA (Author), Developer, QA (Verify), then the human gate |
| `Bug: …` | Fix Flow: QA (Author) writes the failing test that shows it, the Developer fixes it, QA (Verify) checks |
| `As PM & Software Architect: …` | One role only |
| `Report: ON` | A toggle, for this session. The default is the `Toggles:` line in `AGENTS.md` |
| `Seal it` · `Skip demo` · `Not good enough: …` | Your answers at the human gate |

Start a new chat per task: the state is in the files and in git, not in the conversation. Commit rules enforced by git instead of by the agent: `git config core.hooksPath .dudes-framework/hooks`. The first run, end to end: [the 10-minute walkthrough](WALKTHROUGH.md).

## Other layouts

| Layout | Do | Then |
|---|---|---|
| DUDES next to your repo, one clone for many products | `git clone https://github.com/Thomas-Pietsch/dudes.git ../dudes` | Set `DUDES` to `../dudes`. **File → Add Folder to Workspace**, pick the clone, **File → Save Workspace As**, open the saved workspace |

The clone must be a workspace folder: `rules/SECURITY.md` forbids reading outside the workspace. Cursor's forum has reports of wrong paths and "external file" prompts in multi-root workspaces; they are not in the docs. The clone inside your repo avoids them.

## If it does not work

| You see | Cause | Fix |
|---|---|---|
| An ordinary answer, no `## 🔷` line | `AGENTS.md` was not applied | Folder opened is the repo root, with `AGENTS.md` in it? A new Agent chat? Cursor applies `AGENTS.md` in the project root and its subfolders ([Rules](https://cursor.com/docs/context/rules)) |
| `DUDES framework not loaded. Cannot proceed.` | `<DUDES>/rules/DUDES.md` or `SECURITY.md` cannot be read | Empty folder: `git submodule update --init`. Wrong path: fix `DUDES`. Clone outside the workspace: add it as a folder |
| The `Active rules` line, then a question which role | The prompt names no role (`rules/DUDES.md` §2) | Name one, or give the idea |
| It speaks as the DevOps engineer | It applied the clone's own `AGENTS.md`: Cursor applies a nested `AGENTS.md` to the files in its folder | The starter says to ignore it. Say so once, or start a new chat. `.cursorignore` hides files from the agent; we have not tried it for this |
| It ignores a rule, or contradicts one | Your User Rules (Cursor Settings) are merged with the project's; where they conflict, earlier sources win and the place of `AGENTS.md` is not documented | Remove the User Rules that contradict DUDES |
| It stops after a few steps | A command waits for approval, or the tool-call budget in `rules/PERFORMANCE.md` | Approve it, or say `continue` |
