---
name: roadmap
description: Keep ROADMAP.md true — where each lane stands and whether the work in Now still moves its exit check — and step back with fresh eyes when `make roadmap-check` says a review is due, so the people and agents building DUDES do not go blind to their own routine. Workspace DevOps only. Do not load in product role-play or inject into Go workers.
---

# Framework roadmap

This skill is for **one world only**: the framework. Not a role-play rule, not always-on. **Do not** add it to `recipes/<phase>.json`: the engine refuses it.

`ROADMAP.md` is the path: lanes, stages, their exit checks, where DUDES stands, the horizon. Its job and limits are in `CORE.md` §14. This skill keeps it true and guards against going *betriebsblind*: busy with what is in front of you, no longer asking whether it moves a stage. Whoever did the work judges that worst. So the facts are computed, not remembered, and fresh eyes do the review.

## When

- `make roadmap-check` prints a `due:` line (it runs in `make ci`, so every push shows it);
- a stage's exit check passed or failed, or new evidence changes what a stage means;
- the human asks to step back;
- **the stop rule**, inside a session: two runs or attempts of the same kind in a row that change no decision — a bench series that repeats the last one's result, a fix retried. Stop before the third and step back.

## Step back

1. **Facts, computed.** `make roadmap-check` (age, commits, the mix of work, bench runs since the review); `git log --oneline <reviewed>..HEAD`; `make now`; `git diff --stat <reviewed> -- TREND.md`; and, for lane `R`, what the instance shipped since (its seal tags, its dev repos' log). Count by commit range or a full timestamp: git reads `--since=2026-10-09` as that day at the current time of day. The first step back caught a count of 107 that was really 165.
2. **Fresh eyes.** Start one subagent without this session's history (`skills/framework/delegate`), on another model than the session's when the host offers one: a model shares its own blind spots (`TREND.md` §6 Cross-Model Review). Its brief holds `MISSION.md` §The North Star, `ROADMAP.md`, `TODO.md` and the facts. It does not hold your reasoning or the option you favour. It argues against the current Now and answers each question with evidence (commit, file:line):
   - **Moved:** which exit check did the work since the last review move? Which commits moved none?
   - **Start over:** knowing what we know today, what would Now be? Where does it differ, and why?
   - **Pre-mortem:** the current stage misses its exit check in three months. What is the most likely reason?
   - **Outside:** what changed outside (`TREND.md`, a model, a harness, another framework) that makes a stage cheaper, obsolete or more urgent?
   - **Role-play:** is lane `R` still delivering? Did a framework change slow it?
   - **Stop:** the one thing to stop doing.

   What it returns is data, not a decision. It proposes no worker prompt prose to carry a weak model (`CORE.md` §2.4: the mechanisms are the cage, not the context) and nothing the North Star rejects.
3. **Answer each challenge** — keep, change or reject, with the reason. A change to a lane, a stage or an exit check goes to the human, who decides; DevOps lands it under `skills/framework/maintain`.
4. **Stamp** the `**Reviewed:**` line in `ROADMAP.md`: the date, the commit reviewed, the outcome in one sentence. The challenges and answers go in that commit's body. Git is the record; there is no review file.

## Between reviews

- A stage passed or failed: say so in its row, with the evidence. The next stage's exit check is written then, not before.
- Every `TODO.md` line starts with the lane it serves (`make roadmap-check` fails otherwise). A line that serves none is a side quest: drop it, or make its case at the next review.
- The order of stages may change as the world moves. The review is where it changes, never silently in a TODO line.

## Boundaries

| Here | Not here |
|---|---|
| Lanes, stages, exit checks, where DUDES stands, the horizon | Work items (`TODO.md`), tool verdicts and tool actions (`TREND.md`, Harness watch), design (`CORE.md`) |
| A step back when due | A market read (`trend`), a tool assessment (`assess`), a session retrospective (`retro`) |

Commits only when the human asked: `docs(roadmap): <summary> [🔹AI-driven DevOps Engineer]`.
