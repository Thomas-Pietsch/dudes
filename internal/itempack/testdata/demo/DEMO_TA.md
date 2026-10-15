# DEMO TA

## 5. Data Model

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
BEGIN SELECT RAISE(ABORT, 'author_id immutable'); END;

## not a heading — inside a code fence
```

## 6. API

| ID | Method | Path | Purpose |
|---|---|---|---|
| API-01 | `POST` | `/api/v1/demo/notes` | Create a note |
| API-02 | `GET` | `/api/v1/demo/notes/{id}` | Read a note |

#### API-01 · `POST /api/v1/demo/notes`

Request `{text}`; response `201` with `author`. Errors: `400 ERR_EMPTY`.

## 11. SLOs

| SLO | Target |
|---|---|
| SLO-01 | Save P95 ≤ 200 ms |

## 24. Behaviour Contracts

### ALG-01 · Note validation

- **ALG-01.1** *Length* — text 3–2000 chars after trim.
- **ALG-01.2** *Author* — author comes from the actor header, never from the body.
