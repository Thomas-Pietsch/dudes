---
name: backlog-item
description: Write one backlog work item its builder can take without asking. Each part of the item holds its facts by reference to their owner, the item meets the Definition of Ready, and its status is left to be derived. Use when the PM cuts a new work item, changes one, or checks before the freeze that one is ready.
---

# Backlog item

Read `skills/spec/README.md` first: every fact has an ID and one owner, scope enters top-down, the gate is the consistency check.

The PM loads this to **write** an item. How a developer **builds** one is `work-item`: what the item and its references do not say, the builder does not know. Write for that reader.

## What an item holds

Only what is its own. The parts, in the shape of the reference backlog (`examples/sample/pm/BOOKSHELF_BACKLOG.md`); the instance's template may name them differently:

1. **Release:** the milestone it ships in, which exists before the item.
2. **Story:** who, what and why, for this slice.
3. **Functional requirements:** a build-only behaviour no owner document has and nobody outside the code needs (file placement, check order, retry, fixtures, test hooks), or `FR-n Implements [[DOC:ID]]` for a fact an owner holds. Never the owner's fact restated.
4. **Acceptance criteria:** code-level GIVEN/WHEN/THEN that verify the referenced contracts in this slice. Fixture data may be concrete; thresholds, codes and copy are referenced.
5. **Narrowings:** only by reference to a recorded narrowing. An item never narrows silently.
6. **Edge cases, risks, References** (the machine-readable `[[…]]` list) and **DoD** (the QR row IDs that must be green).

**The owner test:** a sentence that would still describe the product after the item is closed belongs in an owner document, not in the item.

## Definition of Ready

An item is ready when:

- every reference resolves, its release is a milestone, and every DoD ID has a QR row (the consistency check finds each gap);
- every acceptance criterion is testable without a human judgement call;
- no open H finding of the requirements review touches it (`requirements-review`).

The first line is mechanical; the other two are the PM's reading.

## Status is derived

An item holds no status kept by hand. It is derived from the DoD rows in the QR (where the engine is installed, `dudes item board` derives it; the rules and their order are CP-FR-06 in `internal/itempack/SPEC.md`: a DoD ID with no QR row makes the item untracked before anything else). Without the engine, read the same rows. Work in progress shows in git (the branch, the open change), not in the backlog.

## Not here

- Building the item: `work-item`.
- The review that decides readiness: `requirements-review`.
- A term or decision the item needs first: `domain-modeling`.
