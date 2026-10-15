# BOOKSHELF TA

## 3. Architecture

One Go module, `bookshelf`. An HTTP JSON API in front of one SQLite file (`modernc.org/sqlite`, pure Go). Otherwise the standard library. The binary takes `--addr <host:port>` and `--db <file>`; the form of the UX document is a static page the same binary serves later.

## 5. Data Model

| ID | Table | Purpose | Key Columns |
|---|---|---|---|
| TBL-01 | `books` | The shelf | `id`, `title`, `author`, `position`, `read_at` |

```sql
CREATE TABLE books (
  id       TEXT PRIMARY KEY,
  title    TEXT NOT NULL,
  author   TEXT NOT NULL,
  position INTEGER NOT NULL,
  read_at  TEXT
);
```

## 6. API

| ID | Method | Path | Purpose |
|---|---|---|---|
| API-01 | `POST` | `/api/v1/shelf/books` | Add a book |
| API-02 | `POST` | `/api/v1/shelf/books/{id}/read` | Mark a book read |
| API-03 | `GET` | `/api/v1/shelf/books` | List the shelf |

#### API-01 · `POST /api/v1/shelf/books`

Request `{title, author}`; response `201` with the book. Errors: `400 ERR_TITLE` with `{code, message}`, where `message` is the text of [[UX:W-MC-03]].

#### API-02 · `POST /api/v1/shelf/books/{id}/read`

No body; response `200` with the book. Errors: `404 ERR_NO_BOOK`.

#### API-03 · `GET /api/v1/shelf/books`

Response `200` with an array of the books in position order, each with `id`, `title`, `author`, `position` and `read_at` (`null` while unread).

## 24. Behaviour Contracts

### ALG-01 · Title validation

- **ALG-01.1** *Length* — a title is 1 to 200 characters after trimming blanks; anything else is refused with `ERR_TITLE`.
- **ALG-01.2** *Position* — a new book gets the next position: one more than the highest on the shelf, 1 on an empty shelf.

### ALG-02 · Read date

- **ALG-02.1** *Date* — marking a book read sets `read_at` to today's date, `YYYY-MM-DD`, in UTC.
- **ALG-02.2** *Once* — a book that is already read keeps its first `read_at`; marking it again changes nothing.
