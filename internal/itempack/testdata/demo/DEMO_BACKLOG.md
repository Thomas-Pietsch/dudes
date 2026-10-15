# DEMO BACKLOG

## 0. How to work this backlog

1. → [[STRATEGY:§H.6]] rule M1.
6. UI strings are German.

## 4. Work items

### R0.1 — Notes

#### DM-001 · Save a note with author
**Release.** R0.1 · **Status.** derived from QR
**Story.** As a mechanic I save a note and see my name on it.

**Functional requirements**
- FR-1 Implements [[TA:API-01]] and [[TA:TBL-01]].
- FR-2 Implements [[TA:ALG-01.1..2]].
- FR-3 Implements [[UX:UXB-01.1]] with copy [[UX:W-MC-01..03]].
- FR-4 Implements the author guard of [[SC:T-DM-01]]; no claim beyond [[PS:§14.7]].

**Acceptance criteria**
- AC-1 GIVEN a note `Ölstand prüfen`, WHEN [[TA:API-01]] is called, THEN the card shows „Notiz von {Name}“ ([[UX:W-MC-03]]).

**References.** [[PS:§14.7]] [[PRD:AC-NOTE-01..02]] [[PRD:AC-PRIV-01]] [[PRD:§6]] [[TA:API-01]] [[TA:TBL-01]] [[TA:ALG-01.1..2]] [[TA:SLO-01]] [[UX:UXB-01.1]] [[UX:W-MC-01..03]] [[SC:T-DM-01]] [[SC:TC-SEC-ATTR-01]]
**DoD.** TC-P0-01, AC-NOTE-01/02, TC-SEC-01, TC-SEC-ATTR-01, TC-SEC-13 green.

### R0.2 — Search

| Item | Title | Outline |
|---|---|---|
| DM-002 | Search notes | refine after R0.1 |

## 6. Global Definition of Done

1. → [[STRATEGY:§H.6]] D1.
