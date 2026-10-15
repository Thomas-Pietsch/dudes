# Contracts for DM-001

> Generated 2026-10-05 by context_pack 1.0.0. Build artifact — do not edit; regenerate.

## PS

### [[PS:§14.7]]

### 14.7 Claim discipline

- Never say „zertifiziert“.
- Never frame the product as reducing staff.

> *Shortened to fit the budget of 500 characters; full text: [[PS:§14.7]].*

## PRD

### [[PRD:AC-NOTE-01..02]]

| ID | Criterion |
|---|---|
| AC-NOTE-01 | GIVEN a note, WHEN saved, THEN it shows its author. |
| AC-NOTE-02 | GIVEN a note, WHEN edited, THEN the previous text stays readable. |

> *Shortened to fit the budget of 500 characters; full text: [[PRD:AC-NOTE-01..02]].*

### [[PRD:AC-PRIV-01]]

**AC-PRIV-01 — No export per person**
- GIVEN any export
- THEN it contains no count per person

> *Shortened to fit the budget of 500 characters; full text: [[PRD:AC-PRIV-01]].*

### [[PRD:§6]]

## 6. Tests

### TC-SEC-01 Session required
- GIVEN no session
- WHEN any write is sent
- THEN `401`

### TC-P0-01 Note saved with author
- GIVEN a mechanic
- WHEN a note is saved
- THEN the author is shown (AC-NOTE-01)

> *Shortened to fit the budget of 500 characters; full text: [[PRD:§6]].*

## TA

### [[TA:API-01]]

| ID | Method | Path | Purpose |
|---|---|---|---|
| API-01 | `POST` | `/api/v1/demo/notes` | Create a note |

#### API-01 · `POST /api/v1/demo/notes`

Request `{text}`; response `201` with `author`. Errors: `400 ERR_EMPTY`.

> *Shortened to fit the budget of 500 characters; full text: [[TA:API-01]].*

### [[TA:TBL-01]]

| ID | Table | Purpose | Key Columns |
|---|---|---|---|
| TBL-01 | `notes` | Notes | `id`, `author_id`, `text` |

```sql
CREATE TABLE notes (
  id        TEXT PRIMARY KEY,
  author_id TEXT NOT NULL,
  text      TEXT NOT NULL
);

CREATE TRIGGER trg_notes_author BEFORE UPDATE OF author_id ON notes

> *Shortened to fit the budget of 500 characters; full text: [[TA:TBL-01]].*

### [[TA:ALG-01.1..2]]

- **ALG-01.1** *Length* — text 3–2000 chars after trim.

- **ALG-01.2** *Author* — author comes from the actor header, never from the body.

> *Shortened to fit the budget of 500 characters; full text: [[TA:ALG-01.1..2]].*

### [[TA:SLO-01]]

| SLO | Target |
|---|---|
| SLO-01 | Save P95 ≤ 200 ms |

> *Shortened to fit the budget of 500 characters; full text: [[TA:SLO-01]].*

## UX

### [[UX:UXB-01.1]]

- **UXB-01.1** *Form* — one text field, save button W-MC-01.

> *Shortened to fit the budget of 500 characters; full text: [[UX:UXB-01.1]].*

### [[UX:W-MC-01..03]]

| ID | Kontext | Text |
|---|---|---|
| W-MC-01 | Speichern-Button | „Speichern“ |
| W-MC-02 | Toast | „Notiz gespeichert“ |
| W-MC-03 | Autorzeile | „Notiz von {Name}“ |

> *Shortened to fit the budget of 500 characters; full text: [[UX:W-MC-01..03]].*

### [[UX:W-MC-03]]

| ID | Kontext | Text |
|---|---|---|
| W-MC-03 | Autorzeile | „Notiz von {Name}“ |

> *Shortened to fit the budget of 500 characters; full text: [[UX:W-MC-03]].*

## SC

### [[SC:T-DM-01]]

| ID | STRIDE | Threat | Measure |
|---|---|---|---|
| T-DM-01 | Tampering | Author of a note rewritten later | `author_id` trigger (TA TBL-01) |

> *Shortened to fit the budget of 500 characters; full text: [[SC:T-DM-01]].*

### [[SC:TC-SEC-ATTR-01]]

| ID | Test | Priority |
|---|---|---|
| TC-SEC-ATTR-01 | Direct SQL update of `author_id` aborts | P0 |

> *Shortened to fit the budget of 500 characters; full text: [[SC:TC-SEC-ATTR-01]].*
