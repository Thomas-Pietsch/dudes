---
name: prototype
description: Answer one design question with throwaway code before the spec freezes — a single HTML file in which a PM or domain expert drives a state model through its awkward cases. Use when a state machine, workflow or data shape in a spec looks right on paper but nobody has pushed it through real cases yet.
---

# Prototype

Read `skills/spec/README.md` first. A prototype is **throwaway code that answers a question** before the spec is frozen: a bug found in the idea now costs one edit, after the freeze a new freeze and a rerun.

This skill covers the **logic** question: *does this state model or workflow hold?* The **look** question (*what should this screen look like?*) belongs to `user-centered-design` with its flows, Hi-Fi and variants.

## Build it

1. **State the question first**, in one paragraph, visible at the top of the page. A prototype that answers the wrong question is waste. Name the **kill date** with the question, and write it beside the question: at that date the prototype ends, answered or not, and one that lives on is scope pull. It runs beside the items that do not depend on its answer; an item that does is frozen after the answer or after the kill date, whichever comes first. Nothing checks the date yet; written as a `due:` line, a gate goal could.
2. **Keep the logic a small, pure module** in one `<script>` block: a reducer `(state, action) => state`, a state machine when "which actions are legal now" is the question, or pure functions over plain data. No DOM inside it. The page calls it, nothing flows back.
3. **One self-contained HTML file.** No framework, bundler or server; it opens by double-click and survives being mailed. Top to bottom it shows:
   - the title and the question;
   - the **full state**, as labelled fields, re-rendered after every click, with what just changed called out;
   - **free-play buttons**, one per action, always available;
   - **guided scenarios**, one per tab: a plain-language description, then the buttons to press in order, each reset to the same start. Cover the happy path, the awkward edge case, and one attempt at something that must be refused.
4. **Domain language throughout.** Labels use the glossary's terms (`domain-modeling`), never code names, so a domain expert can drive it alone.
5. **No tests, no persistence, no polish.** State lives in memory. Restrained styling, no animation.
6. **Synthetic data only.** No real customer, user or device data touches a prototype.

## After it answered

- The **result** is a hypothesis, not evidence (the evidence ladder, `skills/lead/README.md`): it shapes the spec and never turns a QR row green.
- The **answer** goes into the spec: the state table, the transition or the scenario it settled, with its ID, before the freeze. A surprising answer may earn an ADR (`domain-modeling`).
- The **prototype** is kept as a primary source, off the main branch: commit it to a throwaway branch in the spec repo, and point to that branch from the item or the commit that took the decision.
- The pure module may become the starting point of the dev slice; the HTML shell never ships.
