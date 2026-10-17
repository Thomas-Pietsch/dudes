# Claude Code setup

Run the DUDES role-play in your own product repo with [Claude Code](https://code.claude.com): one agent plays the roles in turn (PM, tester, developer, QA) under the DUDES rules and stops for your word. No engine, no GPU. Claude Code reads `CLAUDE.md`; the starter's one-line `CLAUDE.md` loads the shared `AGENTS.md`, so Claude Code, [Cursor](CURSOR_SETUP.md) and [Cline](CLINE_SETUP.md) use the same entry file.

## What you need

| | |
|---|---|
| Claude Code | Checked on v2.1.267. Since v2.1.277 Claude Code also reads an `AGENTS.md` alone when there is no `CLAUDE.md` ([memory docs](https://code.claude.com/docs/en/memory#agents-md)); the starter does not depend on that |
| Your repo | A git repo, any language. Not the DUDES clone |
| make | The flow runs `make check`. On Windows: Git Bash and GNU make ([Platforms](../README.md#platforms)) |
| A model | [`TREND.md`](../TREND.md) §1.2 says which hosted model fits which work, with sources and the check date. Every hosted model receives your code and your specs |

## 1. Put DUDES inside your repo

```sh
cd your-repo
git submodule add https://github.com/Thomas-Pietsch/dudes.git .dudes-framework
```

Inside the repo, because `rules/SECURITY.md` forbids reading outside the workspace, and Claude Code reads the folder you start it in. The submodule also pins the rule version in your history. Other layouts are at the end.

## 2. Copy the starter

```sh
cp .dudes-framework/templates/product/AGENTS.md .dudes-framework/templates/product/CLAUDE.md .
cp .dudes-framework/templates/product/Makefile .
mkdir -p .claude
printf '{\n  "claudeMdExcludes": ["**/.dudes-framework/CLAUDE.md", "**/.dudes-framework/AGENTS.md"]\n}\n' > .claude/settings.json
printf 'SESSION_STATE.md\n' >> .gitignore
```

- You have a `Makefile` already? Skip its line and add the starter's `test` and `check` goals and the `TEST`, `FMT`, `LINT` variables to yours. Give every goal a `## what it does` comment: `make help` lists them, and that list is all the agent knows about your commands.
- `.claude/settings.json` keeps the clone's own entry out of the session. When the agent reads a rule file, Claude Code also loads that folder's `CLAUDE.md` ("nested traversal"; the `InstructionsLoaded` hook shows it). The clone's `CLAUDE.md` points to the DevOps entry of DUDES, which is not a role in your repo. Merge the line if you have a `.claude/settings.json`.
- `SESSION_STATE.md` is the agent's scratchpad (`rules/CONTEXT.md`), never committed.

## 3. Set the one path and the test command

| Where | Setting | Leave it when |
|---|---|---|
| `AGENTS.md`, table under *Framework Requirement* | `DUDES`: the DUDES clone, relative to `AGENTS.md` | You followed step 1 (`.dudes-framework`) |
| `Makefile` | `TEST`, `FMT`, `LINT`: the commands that run all your tests, the format check and the static checks; a tool you do not use: `true`, on purpose | You want QA (Author) to set it for the stack the TA names |

Then commit: `git add -A && git commit -m "chore: DUDES starter [🔹human]"`.

## 4. Verify

Start Claude Code in the repo root and ask:

```text
As PM & Software Architect: which files do you read before you write a spec?
```

The answer starts like this. The wording after it is the model's:

```text
## 🔷 PM & Software Architect

Active rules: DUDES, SECURITY, TERMINAL, CONTEXT, TESTING, PERFORMANCE · Toggles: Convergence OFF, Report OFF, Review OFF, Vibe OFF

Flow: Single-Role
```

The role tag, the toggles line and the flow are the first lines. Also check:

- `/context` lists `CLAUDE.md` and `AGENTS.md` of your repo under *Memory files*.
- With an empty `.dudes-framework` (a submodule you did not initialise) the whole answer is `DUDES framework not loaded. Cannot proceed.` That is the rule working: an agent that cannot read the rules does not play along.

What was run (2026-10-15, Claude Code v2.1.267, the `sonnet` alias, medium effort): the starter as shipped printed the role tag and the toggles line first in 8 of 8 asks (two rounds of four). The same starter without its *First response* line printed the tag first in 3 of 4 and the toggles line in 1 of 4. That is a handful of runs, not a measurement. A model may still skip a line: ask `Which rules and toggles are active?`. A whole flow, idea to your decision: [the walkthrough](WALKTHROUGH.md).

## 5. Allow what the flow runs

Claude Code asks before it edits a file or runs a command ([permissions](https://code.claude.com/docs/en/permissions)). The flow runs `git` and `make` all the time. Answer "Yes, and don't ask again" once per kind, or write the rules down in the same `.claude/settings.json`, next to the exclusion:

```json
{
  "claudeMdExcludes": ["**/.dudes-framework/CLAUDE.md", "**/.dudes-framework/AGENTS.md"],
  "permissions": {
    "allow": ["Bash(make *)", "Bash(git add *)", "Bash(git commit *)", "Bash(git checkout *)", "Bash(git switch *)", "Bash(git branch *)", "Bash(git merge *)"],
    "deny": ["Bash(git push *)"]
  }
}
```

- The read-only forms of `git` (`status`, `log`, `diff`, `show`) run without a rule. Add your toolchain (`Bash(go *)`, `Bash(npm *)`). Allow narrow.
- A chained command (`a && b`) needs every part allowed. In one headless run a chained command was refused for that reason, and the agent ran its parts one by one.
- The `deny` line backs the starter's "never push": Claude Code refuses the common form. Its docs call such patterns fragile (`git -C . push` does not match), so it is a guard rail, not a lock.
- `git tag` stays out of the allow list on purpose. The seal is yours: Claude Code asks before the agent creates the tag.

## Day to day

| You say | What happens |
|---|---|
| `Idea: … Build it.` | Feature Flow: PM, QA (Author), Developer, QA (Verify), then the human gate |
| `Bug: …` | Fix Flow: QA (Author) writes the failing test that shows it, the Developer fixes it, QA (Verify) checks |
| `As PM & Software Architect: …` | One role only |
| `Report: ON` | A toggle, for this session. The default is the `Toggles:` line in `AGENTS.md` |
| `Seal it` · `Skip demo` · `Not good enough: …` | Your answers at the human gate |

A new chat per task: the state is in the files and in git, not in the conversation. Commit rules enforced by git instead of by the agent: `git config core.hooksPath .dudes-framework/hooks` (every commit needs a `[🔹Role]` tag). The first run, end to end: [the 10-minute walkthrough](WALKTHROUGH.md).

## Other layouts

| Layout | Do | Then |
|---|---|---|
| DUDES next to your repo, one clone for many products | `git clone https://github.com/Thomas-Pietsch/dudes.git ../dudes` | Set `DUDES` to `../dudes`. Start with `claude --add-dir ../dudes`, or run `/add-dir ../dudes` in a session |

Without the folder added, Claude Code cannot read the clone: a non-interactive run printed `DUDES framework not loaded. Cannot proceed.`, and an interactive session asks you for the permission. With it, Claude Code does not load the clone's own `CLAUDE.md` (it does so for an added folder only when `CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD` is set), so the exclusion in step 2 is not needed.

## If it does not work

| You see | Cause | Fix |
|---|---|---|
| An ordinary answer, no `## 🔷` line | The entry was not loaded | `/context`: are `CLAUDE.md` and `AGENTS.md` there? Start `claude` in the repo root. Keep the starter's `CLAUDE.md`: on v2.1.267 an `AGENTS.md` alone did not start the role-play |
| `DUDES framework not loaded. Cannot proceed.` | `<DUDES>/rules/DUDES.md` or `SECURITY.md` cannot be read | Empty folder: `git submodule update --init`. Wrong path: fix `DUDES`. Clone outside the repo: `--add-dir` |
| The `Active rules` line, then a question which role | The prompt names no role (`rules/DUDES.md` §2) | Name one, or give the idea |
| It speaks as the DevOps engineer, or reads the sample | The clone's own `CLAUDE.md` or `AGENTS.md` loaded | Step 2's `.claude/settings.json`. Check with `/context` |
| It stops after a few steps | A permission prompt, or the tool-call budget in `rules/PERFORMANCE.md` | Allow the command, or say `continue` |
