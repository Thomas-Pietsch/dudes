---
name: requirements-review
description: Read the spec-set the way its first builder will, before the freeze. A fixed question set over each item and the requirements it references; every finding rated H, M or L and resolved in its owner document; an open H blocks the freeze. Use when the PM prepares a freeze, after a large spec change, or when a builder came back with a question the spec should have answered.
---

# Requirements review

Read `skills/spec/README.md` first. The PM loads this before the freeze.

The consistency sweep (`rules/DUDES.md` §6) finds what a search can: dangling IDs, double definitions, a backlog the milestones and the QR do not match. This review finds what only reading finds: a requirement that resolves cleanly and still cannot be built without asking. Run the sweep first; the review reads what it left.

## The questions

Ask them of every item in the release and of each requirement it references:

1. **Unclear or missing?** Two readers would build two different things, or a case has no answer.
2. **Too complex?** It does more than the release needs; a narrower cut ships the same value.
3. **Ready for implementation?** It meets the Definition of Ready (`backlog-item`).
4. **Would another engineer understand it at once,** from the item and the sections its references resolve to, and nothing else?

## Severity

| Rate | Means |
|---|---|
| **H** | An agent cannot implement it without asking |
| **M** | It will be built, then reworked |
| **L** | Polish |

## Findings

- Each finding goes to the instance's review log (its repo names it), with an ID of its review family, the place (document and ID), the rate, the resolution and where it landed.
- A finding is resolved in its **owner document**, never in the log or the item alone, and the change ripples in the same commit (README).
- A choice that is the CEO's (scope, price, market) is named and stops for the human.
- A clean answer to questions 3 and 4 is logged too, with what was read, so the next review starts from it.

## The freeze

The PM freezes only with no open H finding touching an item of the release being frozen. An M or L may stay open with its owner named. Nothing counts open H findings yet; once the review log has a fixed shape, a check could, beside the consistency sweep.

## Not here

- ID resolution and double definitions: the consistency sweep (`rules/DUDES.md` §6).
- How the item under review is written: `backlog-item`.
- A state model that needs pushing through its cases: `prototype`.
