---
name: domain-modeling
description: Sharpen the product's language and record the decisions that shape it while a spec is written — challenge terms against the glossary, test them with edge-case scenarios, write the glossary entry and the ADR the moment they settle. Use when a PM, architect or security author introduces or changes a domain term, or a design choice is hard to reverse.
---

# Domain modeling

Read `skills/spec/README.md` first: a term or decision is part of the spec-set, under the instance's own document names and ID families.

This is the *active* discipline: changing the model, not just reading it. Reading the glossary for vocabulary is a habit every role has; this skill is for when a term or a boundary moves.

## Where it lands

- **Terms:** the glossary the instance keeps (its PRD glossary section or a `GLOSSARY.md` in the spec repo; the profile names the document). Created lazily, when the first term is resolved. One row per term: *Term · Means · Avoid*, as in `CORE.md` §5. Pick one word, list the others under *Avoid*. One or two sentences, what it **is**, not what it does. Only this product's terms, never general programming concepts. No implementation detail: the glossary is not a spec.
- **Decisions:** an ADR in the document that owns the ADR family (usually the TA, e.g. `ADR-07`). A short paragraph: context, decision, why. *Considered options* and *Consequences* only when they earn their place.

## During the session

1. **Challenge against the glossary.** A term used against its entry is called out at once: "The glossary defines *site* as X, you seem to mean Y. Which is it?"
2. **Sharpen fuzzy words.** For an overloaded word, propose the precise term: "*account*: the tenant or the user?"
3. **Stress-test with scenarios.** Invent the edge case that forces a boundary: two tenants, a deleted user, a retried request.
4. **Cross-reference.** Check the other spec documents and, when a dev checkout is at hand, the code. A contradiction is surfaced, not resolved silently.
5. **Write it down when it settles.** Update the glossary row then and there, not in a batch. A changed term ripples: find its uses with `dudes item check` and the ID search, and fix them in the same commit.

## When an ADR is worth it

Offer one only when all three hold:

1. **Hard to reverse.** Changing your mind later costs a lot.
2. **Surprising without context.** A later reader would ask "why on earth this way?"
3. **A real trade-off.** There were genuine alternatives.

This typically covers architecture shape, integration between contexts, technology with lock-in, boundary and scope decisions (the explicit *no* included), deliberate deviations from the obvious path, and constraints the code cannot show. If any of the three is missing, skip the ADR.

A decision that is the CEO's or the Product Visionary's (scope, price, market) is not settled here: name it and stop for the human.
