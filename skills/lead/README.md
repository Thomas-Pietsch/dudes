# Lead-level skills

For the lead repo: the CEO (strategy, vision, Go/No-Go, idea intake) and the Product Visionary (market, competition, go-to-market, product-market fit). Lead work is text and judgement no test settles, and no model is trained on how a DUDES company decides (`CORE.md` §5, *Where skills earn their place*) — so the skills here carry that method. They share this background instead of repeating it:

- **A decision is a commit.** It carries its role tag (`[🔹CEO]`), lands in the lead repo's documents and states its premise and what would falsify it. The decision log is derived from those commits (`make audit`), never kept by hand.
- **The record comes first.** Before a judgement, collect what the mechanisms already know — the decision log, each product's derived item status (`dudes item board`, F70), the consistency check (`dudes item check`), the roadmap's milestones. A judgement without that record is an opinion.
- **Sources are primary and dated.** A market claim names its source, its date and whether it is documented or measured; a forecast names what would falsify it.
- **The company's own** — its lab, infrastructure, company document, local conventions — lives in the instance's lead repo: what can be checked as a goal of its makefile, what needs judgement as its own `skills/`. A company skill that extends one here names it instead of copying it.

**Company audit — no skill.** The record is programmatic: `make audit ARGS='--since=<last audit>'` (the decisions), `dudes item board` and `dudes item check` per product (derived status, consistency), `dudes status` (runs waiting for a decision), and the instance lead repo's own goals for what only the company knows (its lab, its infrastructure). The judgement on that record is the CEO's: which premise of a past decision no longer holds, what work serves no goal and which goal has no work, what passed its review date untouched, and which decision unblocks a late milestone. A finding that needs a decision becomes one — a role-tagged commit.

Which role loads which skill in the role-play: `rules/DUDES.md` §1.
