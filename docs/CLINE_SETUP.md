# Cline setup

Run the DUDES role-play in your own product repo with [Cline](https://docs.cline.bot), the open-source coding agent for VS Code: Cline plays the roles in turn (PM, tester, developer, QA) under the DUDES rules and stops for your word. No engine, no GPU. Cline lists `AGENTS.md` as a rule type it reads, so the starter's `AGENTS.md` is the whole adapter. Other hosts: [Claude Code](CLAUDE_CODE_SETUP.md), [Cursor](CURSOR_SETUP.md).

**Status.** The starter was run and checked on Claude Code ([how](CLAUDE_CODE_SETUP.md#4-verify)). This guide follows Cline's documentation, linked where it matters, and has not been run in Cline yet. If a step differs for you, [open an issue](https://github.com/Thomas-Pietsch/dudes/issues/new/choose).

## What you need

| | |
|---|---|
| Cline | The extension in VS Code, or another editor it supports ([install](https://docs.cline.bot/getting-started/installing-cline)): open Extensions, search `Cline`, install |
| A provider and model | Set in Cline's settings. [`TREND.md`](../TREND.md) §1.2 says which hosted model fits which work, with sources and the check date. Every hosted model receives your code and your specs |
| Your repo | A git repo, any language. Not the DUDES clone |
| make | The flow runs `make test`. On Windows: Git Bash and GNU make ([Platforms](../README.md#platforms)) |

A local model through Cline (Ollama) runs as an agent, but the role-play asks a model to hold about 47 KB of rules and a long flow. That has not been measured with DUDES. The measured local path is the engine, [door 2](../README.md#two-ways-in).

## 1. Put DUDES inside your repo

```sh
cd your-repo
git submodule add https://github.com/Thomas-Pietsch/dudes.git .dudes-framework
```

Inside the repo, because Cline works on the folder you open, `rules/SECURITY.md` keeps the agent inside it, and Cline's auto-approve treats files outside the workspace as a separate setting (*Read all files*). The submodule also pins the rule version in your history.

## 2. Copy the starter

```sh
cp .dudes-framework/templates/product/AGENTS.md .
cp .dudes-framework/templates/product/Makefile .
printf 'SESSION_STATE.md\n' >> .gitignore
```

You have a `Makefile` already? Skip the second line and add its `test` goal to yours. Give every goal a `## what it does` comment: `make help` lists them, and that list is all the agent knows about your commands. `SESSION_STATE.md` is the agent's scratchpad (`rules/CONTEXT.md`), never committed. The starter's `CLAUDE.md` is for Claude Code; Cline does not need it.

## 3. Set the one path and the test command

| Where | Setting | Leave it when |
|---|---|---|
| `AGENTS.md`, table under *Framework Requirement* | `DUDES`: the DUDES clone, relative to `AGENTS.md` | You followed step 1 (`.dudes-framework`) |
| `Makefile` | `TEST`: the command that runs all your tests | You want QA (Author) to set it for the stack the TA names |

Then commit: `git add -A && git commit -m "chore: DUDES starter [🔹human]"`.

## 4. Verify

1. Open the root of your product repo as the VS Code folder.
2. Open the Cline panel. Click the scale icon at the bottom of it: the Rules panel lists every rule Cline detected, each with a toggle ([Cline Rules](https://docs.cline.bot/features/cline-rules)). `AGENTS.md` must be listed and on.
3. Switch to **Act** mode. In Plan mode Cline "cannot modify any files or execute commands" ([Plan and Act](https://docs.cline.bot/core-workflows/plan-and-act)), and the flow writes files and runs `make test`.
4. Start a new task:

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

## 5. Auto-approve what the flow runs

Cline asks before it acts. The flow edits files and runs `git` and `make` all the time. In Cline's auto-approve menu ([Auto Approve](https://docs.cline.bot/features/auto-approve)) turn on *Read project files*, *Edit project files* and *Execute safe commands*. Commands Cline does not mark safe still ask: approve them one by one, or turn on *Execute all commands* only for a repo you can lose. Leave YOLO mode off: its own page calls it dangerous and says it disables all safety checks.

## Day to day

| You say | What happens |
|---|---|
| `Idea: … Build it.` | Feature Flow: PM, QA (Author), Developer, QA (Verify), then the human gate |
| `Bug: …` | Fix Flow: QA (Author) writes the failing test that shows it, the Developer fixes it, QA (Verify) checks |
| `As PM & Software Architect: …` | One role only |
| `Report: ON` | A toggle, for this session. The default is the `Toggles:` line in `AGENTS.md` |
| `Seal it` · `Skip demo` · `Not good enough: …` | Your answers at the human gate |

Start a new task per piece of work: the state is in the files and in git, not in the conversation. Commit rules enforced by git instead of by the agent: `git config core.hooksPath .dudes-framework/hooks`. The first run, end to end: [the 10-minute walkthrough](WALKTHROUGH.md). To run it in Cline and report what differs: [the IDE run sheet](IDE_RUN_SHEET.md).

## Other layouts

DUDES next to your repo (`git clone https://github.com/Thomas-Pietsch/dudes.git ../dudes`, `DUDES` set to `../dudes`) needs Cline to read outside the workspace, and `rules/SECURITY.md` tells the agent not to. Cline's documentation does not describe multi-root workspaces. Use the submodule here, or [Claude Code](CLAUDE_CODE_SETUP.md#other-layouts), which adds a folder with one flag.

## If it does not work

| You see | Cause | Fix |
|---|---|---|
| An ordinary answer, no `## 🔷` line | `AGENTS.md` was not read, or is toggled off | Rules panel: is `AGENTS.md` listed and on? Is the folder opened the repo root? Start a new task |
| `DUDES framework not loaded. Cannot proceed.` | `<DUDES>/rules/DUDES.md` or `SECURITY.md` cannot be read | Empty folder: `git submodule update --init`. Wrong path: fix `DUDES` |
| The `Active rules` line, then a question which role | The prompt names no role (`rules/DUDES.md` §2) | Name one, or give the idea |
| It speaks as the DevOps engineer | The Rules panel lists the clone's own `AGENTS.md` too | The starter says to ignore it. Toggle it off in the Rules panel. Cline's documentation does not say whether it reads a nested `AGENTS.md` |
| It ignores a rule, or contradicts one | A global `~/.agents/AGENTS.md`, or global rules, apply in every workspace ([docs](https://docs.cline.bot/features/cline-rules)) | Toggle the global ones off for this repo |
| It cannot edit or run anything | Plan mode | Switch to Act mode |
| It stops after a few steps | A command waits for approval, or the tool-call budget in `rules/PERFORMANCE.md` | Approve it, or say `continue` |
