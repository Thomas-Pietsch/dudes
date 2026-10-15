---
name: user-centered-design
description: Research user needs and connect Hi-Fi designs, variants, flows and regression evidence. Use spec-to-mockup mode to extract a copy-ready design-agent briefing from specs.
---

# User-centered design

Derives from [CORE](../../../CORE.md) sections 2.4 and 6 (instance-owned product knowledge, evidence, bounded discipline); registered as F81. This skill teaches design work, not phase order or approval authority. Keep the instance's original artifact names and IDs.

## 1. Research before styling

- Read the current PS personas/JTBD, PRD scope, UX and relevant feedback. Existing design is evidence to inspect, not proof of user fit.
- Record user context: device, light/noise, attention, accessibility, domain language, current workaround and failure/recovery costs.
- In the existing PS/UX owner sections, link dated sources or observed sessions to the need, design decision and expected benefit. Distinguish observed, externally sourced and assumed claims. Never invent interviews, task-success rates or user validation.
- Compare relevant current interaction patterns and alternatives. An innovation needs a specific user problem, a simpler alternative, an expected measurable benefit and a test/rejection criterion. Novel appearance alone is not a benefit.
- Missing field access is an explicit research gap with owner and validation task, not a claim that the design is validated. Keep participant data private; use synthetic data in designs.

## 2. Select coverage by risk and reuse

Map critical journeys before drawing screens. Include entry, action, result, back/cancel, failure/retry, permissions and what data/context must survive each transition.

Select Hi-Fi coverage for:

- the shared shell and each distinct screen/interaction pattern;
- critical or high-risk transitions;
- states, modes and responsive layouts that materially change the experience.

Reuse a reference plus documented differences for equivalent screens. Every in-scope feature maps to a pattern or an explicit coverage gap; not every feature needs a separate image. Record the selection rationale and exclusions.

## 3. Deliver actual Hi-Fi and comparable variants

- Deliver rendered images and a locally viewable/clickable prototype for the selected journeys, not only prose or image-generation prompts. Use realistic synthetic content, legible final labels, defined viewports, tokens and component states. Correct generated nonsense text before accepting a reference.
- Connect selected state images with annotated actions and recovery paths. Images specify appearance; PRD/UX references specify behavior.
- Honor the product's requested variant count. If four are requested, provide four named, directly switchable variants of the same representative journey using the same fixture and viewport. Document each variant's user/context rationale and token/component differences.
- Distinguish alternative design concepts, supported presentation/accessibility modes and role-specific workflows. Concepts need a selection decision; supported modes need ongoing test coverage. A role change is not a theme: changes to permissions, fields or workflow need explicit PRD scope and separate scenarios.
- Switching presentation modes must not reset the route, selected record, entered data or permissions. Specify preference persistence and precedence with device/accessibility settings. Do not create four independent copies of the app.
- Use shared semantic tokens and components. Extend selected mode references to every distinct affected pattern, not just the home page. Do not multiply every state by every viewport without a risk reason.

## 4. Durable handoff and changes

Keep one compact coverage matrix in the existing UX artifact:

| Journey/scenario ID | User need/source | Pattern/state/mode | Hi-Fi/prototype reference | Transition/invariant | Acceptance test + screenshot checkpoint |
|---|---|---|---|---|---|

Reference the canonical requirement, microcopy and token definitions; do not duplicate them. Before implementation, identify shared-component consumers and affected journeys, update intended behavior through the normal spec/freeze flow, and record any unresolved selection decision. Tests are authored from that specification before implementation.

For subsequent UI changes, update only the impacted references and tests, but expand regression scope when shared navigation, tokens or components change. Design-only changes must preserve behavioral acceptance criteria.

Verification follows [TESTING](../../../rules/TESTING.md) sections 3-4: first review design fidelity, then protect the reviewed application with screenshot baselines and functional journeys. Human design judgment and observed usability evidence are not replaced by passing screenshots. Report gaps explicitly; never claim that this skill supplies an automated gate.

## 5. Spec-to-mockup mode

Use when asked to extract a design prompt, prepare a mockup briefing or hand specifications to an image/design agent. This mode produces a handoff, not the images or product implementation. It does not need an engine, API key or a particular vendor.

Example invocation (natural language, not a CLI command):

> Use user-centered-design in spec-to-mockup mode for Bookshelf. Read its spec-set, target the add-a-book journey of the current release, prepare three comparable visual concepts with exact German text, and write a copy-ready briefing. Mark unresolved branding/mode decisions as proposals. Do not upload anything.

### Read and resolve

1. Resolve product roots through the spec-set. Read the original persona/JTBD and research sections (PS), behavior (PRD), visual/copy definitions (UX), and release/work-item scope (BACKLOG/ROADMAP when registered). Follow only relevant references, not entire unrelated specs.
2. Record source paths, section/requirement IDs, repo revisions and content hashes for the inputs actually used. Mark dirty inputs explicitly. The generated brief is a derived snapshot: regenerate after relevant source changes, never maintain it as another authoritative UX spec.
3. Apply the instance's explicit scope/precedence rules. Do not illustrate a future vision as the current MVP. Distinguish observed user evidence, strategy hypotheses, specified behavior and proposed visual treatments.
4. Resolve customer-facing branding, supported modes versus exploration concepts, language, viewport and selected journey. An undecided brand/mode is a decision gap. Exploration may proceed with clearly labeled proposals when requested; implementation-ready output cannot silently choose them.
5. Inventory referenced assets with purpose and availability. Existing collages may be quality references, not permission to copy another product's identity, content or stale scope. Missing assets must be reported; include enough written constraints for the receiving agent to work without private repo access.

### Write one briefing file

Keep local provenance and the gap/review tables outside one clearly bounded, copy-ready prompt. Use the caller's output location; otherwise the instance's design-artifact location. A framework example belongs in `examples/<product>/`, not in canonical product specs.

The copied prompt must stand alone; no "see UX section..." in place of required content. Include:

| Block | Required content |
|---|---|
| Task and status | Product, current increment, images/prototype requested, exploration or approved design, no implementation authority |
| User and benefit | Persona, working conditions, JTBD, research rationale and unvalidated success hypotheses |
| Brand | Leading name/logo, endorsement/hidden-brand rule, supplied assets; explicit proposal if undecided |
| Scope | Selected patterns/states, exclusions, exact behavior and transition invariants |
| Visual system | Canonical token subset, type, spacing, touch targets; separately labeled proposed deltas |
| Variants | Requested count, names, rationale and differences; same fixture/viewport/journey for comparison |
| German copy | Exact source strings with IDs where available, synthetic field values, separately labeled proposed copy |
| Deliverables | Named individual renders, overview sheet, annotated flow; dimensions, state/mode captions outside UI, editable source if possible |
| Limits and review | No invented features/claims/data; report tool limits and text defects; image review is not usability validation |

German copy must be idiomatic de-DE: real umlauts and ß, consistent domain nouns and form of address, German date/number formatting, no lorem ipsum or English UI filler. Preserve canonical strings exactly, even when proposing a better alternative separately. Required new labels go into the proposal/gap table for the UX owner; never masquerade as quotations. Synthetic example values are not product requirements or repair advice.

For an image-capable agent, request actual images, not another description of images. For code/browser-capable agents, allow locally rendered HTML/SVG and screenshot export; require disclosure of font fallbacks. If neither is available, report that limitation rather than claiming assets were generated. A static image agent cannot prove switching or navigation: request a separate clickable prototype only where supported, otherwise record it as outstanding.

### Verify before handoff

- Every normative behavior/copy/token has a source; every creative addition is marked proposed.
- Resolve all source/asset paths; list missing ones as gaps, not attachments.
- Check copied German strings against the source and fill every fixture placeholder consistently. Do not turn a fixed test timestamp or timer into a measured result.
- The selected screens form a coherent journey; mode changes never imply different permissions, lost drafts or new features.
- Each gap records impact, owner and whether it blocks exploration or implementation.
- Include a returned-asset checklist: actual output files/dimensions, German glyphs/labels, no clipping, state/variant consistency, no unsupported features, and separate prototype/field-test evidence.
- Do not upload automatically. Keep private specs and personal data local; a manually shared prompt must contain only material approved for that destination.

Worked extraction: [Bookshelf mockup brief](../../../examples/sample/mockup-brief.md). Its three concept proposals are not approved product modes. Generating the briefing completes this mode; returned-image and user validation remain the product's design/QA work.
