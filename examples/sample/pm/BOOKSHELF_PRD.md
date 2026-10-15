# BOOKSHELF PRD

A sample product in the reference format: a small service that keeps a reading list. It is made up. Its IDs, scenarios and references show how a spec set is written, resolved and checked.

## 1. Scope

One reader, one shelf. A book has a title, an author and a read date. R0.1 adds books to the shelf, R0.2 marks them read. Search, tags and accounts are out of scope.

## 2. Features

### F1 Shelf

| ID | Criterion |
|---|---|
| AC-SHELF-01 | GIVEN a title and an author, WHEN a book is added, THEN the shelf lists it as unread, after the books already there. |
| AC-SHELF-02 | GIVEN a book on the shelf, WHEN it is marked read, THEN it shows its read date and no longer counts as unread. |
| AC-SHELF-03 | GIVEN a title that is empty or longer than 200 characters, WHEN a book is added, THEN the request is refused and the shelf stays as it was ([[TA:ALG-01]]). |

## 3. Requirements

A requirement holds a SHALL line and at least one GIVEN/WHEN/THEN scenario. `dudes check` reads this shape (F37). The criteria of §2 are what the work items cite.

### Requirement: REQ-ADD Add a book

The system SHALL list a book it accepts at the end of the shelf, unread (AC-SHELF-01).

#### Scenario: A book is added
- GIVEN an empty shelf
- WHEN the reader adds `Sample Book` by `Sample Author`
- THEN the shelf lists one unread book

### Requirement: REQ-TITLE Refuse a bad title

The system SHALL refuse a book whose title is empty or longer than 200 characters, and leave the shelf unchanged (AC-SHELF-03).

#### Scenario: An empty title
- GIVEN a shelf with one book
- WHEN the reader adds a book with an empty title
- THEN the request is refused
- AND the shelf still holds one book

### Requirement: REQ-READ Mark a book read

The system SHALL record the read date when a book is marked read, and show the book as read from then on (AC-SHELF-02).

#### Scenario: A book is marked read
- GIVEN an unread book
- WHEN the reader marks it read
- THEN it shows today's date as its read date
