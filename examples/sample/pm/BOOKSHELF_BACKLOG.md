# BOOKSHELF BACKLOG

## 0. How to work this backlog

1. A work item is one vertical slice. The item text and the definitions its references resolve to are the complete context; nothing else is read.
2. Each acceptance criterion becomes a failing test before any code, named after the item and the criterion (`TestBK001AC1`).
3. UI text comes from the UX document by reference, never typed again. A quoted UI string without a `[[UX:…]]` anchor is a finding.
4. An item is done when every DoD row of the QR is green.

## 4. Work items

### R0.1 — Shelf

#### BK-001 · Add a book to the shelf
**Release.** R0.1 · **Status.** derived from QR
**Story.** As a reader I add a book to my shelf, so that I find it again.

**Functional requirements**
- FR-1 Implements [[TA:API-01]] and [[TA:API-03]] on [[TA:TBL-01]], started with the flags of [[TA:§3]].
- FR-2 Validates the title and sets the position by [[TA:ALG-01.1..2]].
- FR-3 Later slice: the form of [[UX:UXB-01.2]] with the copy [[UX:W-MC-01..03]] shows what the service answers.

**Acceptance criteria**
- AC-1 GIVEN an empty shelf, WHEN `Sample Book` by `Sample Author` is added through [[TA:API-01]], THEN the response is `201` and the listing of [[TA:API-03]] shows one unread book.
- AC-2 GIVEN a shelf with two books, WHEN a third is added, THEN it is listed last ([[TA:ALG-01.2]]).
- AC-3 GIVEN an empty title, WHEN a book is added through [[TA:API-01]], THEN the shelf is unchanged and the error message is „Der Titel fehlt oder ist zu lang.“ ([[UX:W-MC-03]]).

**References.** [[PRD:AC-SHELF-01]] [[PRD:AC-SHELF-03]] [[TA:§3]] [[TA:API-01]] [[TA:API-03]] [[TA:TBL-01]] [[TA:ALG-01]] [[UX:UXB-01.2]] [[UX:W-MC-01..03]]
**DoD.** AC-SHELF-01, AC-SHELF-03 green.

### R0.2 — Reading

#### BK-002 · Mark a book as read
**Release.** R0.2 · **Status.** derived from QR
**Story.** As a reader I mark a book read, so that the shelf shows what is left.

**Functional requirements**
- FR-1 Implements [[TA:API-02]] on [[TA:TBL-01]] by [[TA:ALG-02.1..2]].
- FR-2 Shows the button of [[UX:UXB-01.1]] with the copy [[UX:W-MC-04]].

**Acceptance criteria**
- AC-1 GIVEN an unread book, WHEN it is marked read through [[TA:API-02]], THEN the response is `200` and the book carries today's date as its read date.
- AC-2 GIVEN a read book, WHEN it is marked read again, THEN its first read date stays ([[TA:ALG-02.2]]).
- AC-3 GIVEN an unknown book id, WHEN it is marked read, THEN the response is `404` with `ERR_NO_BOOK`.

**References.** [[PRD:AC-SHELF-02]] [[TA:API-02]] [[TA:TBL-01]] [[TA:ALG-02]] [[UX:UXB-01.1]] [[UX:W-MC-04]]
**DoD.** AC-SHELF-02 green.

## 6. Global Definition of Done

1. `go test ./...` green; every acceptance test of the item passes and names its criterion.
