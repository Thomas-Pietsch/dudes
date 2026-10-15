# DEMO PRD

## 2. Features

### F1 Notes

| ID | Criterion |
|---|---|
| AC-NOTE-01 | GIVEN a note, WHEN saved, THEN it shows its author. |
| AC-NOTE-02 | GIVEN a note, WHEN edited, THEN the previous text stays readable. |

**AC-PRIV-01 — No export per person**
- GIVEN any export
- THEN it contains no count per person

## 6. Tests

### TC-SEC-01 Session required
- GIVEN no session
- WHEN any write is sent
- THEN `401`

### TC-P0-01 Note saved with author
- GIVEN a mechanic
- WHEN a note is saved
- THEN the author is shown (AC-NOTE-01)
