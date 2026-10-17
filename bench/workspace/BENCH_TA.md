# BENCH TA

## 3. Packages

| Package | Import path | Holds |
|---|---|---|
| `slug` | `bench/slug` | `Make(title string) string` |
| `semver` | `bench/semver` | `Parse(s string) (Version, error)`, `Compare(a, b string) (int, error)` |

## 4. Release packages (BL-003)

| Package | Import path | Holds |
|---|---|---|
| `semver` | `bench/semver` | adds `IsCandidate(s string) bool` — true for a valid `-rc.N` version |
| `release` | `bench/release` | `Latest(tags []string) (string, error)` — uses `semver`, never a second version parser |

## 5. Notes package (BL-004)

| Package | Import path | Holds |
|---|---|---|
| `notes` | `bench/notes` | `type Note struct{ Tag, Text string }`, `Parse(name string, r io.Reader) ([]Note, error)`, `Render(notes []Note) (string, error)`; orders and classifies through `semver` (`Parse`, `Compare`, `IsCandidate`) and `release` (`Latest`), never a second version parser |

## 24. Behaviour Contracts

### ALG-01 · Version order

- **ALG-01.1** *Form* — `MAJOR.MINOR.PATCH`, each a decimal number without sign, optionally prefixed by `v` and optionally followed by `-rc.N` (N a decimal number). Anything else is refused with an error whose text starts with `ERR_VERSION:` and quotes the input.
- **ALG-01.2** *Order* — MAJOR, then MINOR, then PATCH, compared as numbers: `1.10.0` is greater than `1.9.3`.
- **ALG-01.3** *Release candidates* — with equal MAJOR.MINOR.PATCH a release candidate sorts before the release, and candidates by N as a number: `1.2.0-rc.2` < `1.2.0-rc.10` < `1.2.0`.
- **ALG-01.4** *Result* — `Compare` returns -1, 0 or 1 like `strings.Compare`; the `v` prefix does not change the order.

### ALG-02 · Latest release

- **ALG-02.1** *Usable tags* — a tag counts when `semver.Parse` accepts it; every other tag is skipped, not an error.
- **ALG-02.2** *Finals first* — the latest is the highest final release (no `-rc.N`, `semver.IsCandidate` false); a release candidate is the answer only when no final release exists.
- **ALG-02.3** *As given* — the answer is the tag exactly as it appears in the input, `v` prefix included.
- **ALG-02.4** *Nothing* — no tag parses: an error whose text starts with `ERR_NO_RELEASE:`.
- **ALG-02.5** *One parser* — `release` orders and classifies with `semver` (`Compare`, `IsCandidate`); `semver`'s behaviour under ALG-01 does not change.

### ALG-03 · Release notes

- **ALG-03.1** *Note line* — the input holds one note per line: `<tag> <text>`. The tag runs to the first space and must be a version per ALG-01.1. The text is the rest of the line with blanks at both ends removed and must not be empty. Empty lines are skipped. Lines count from 1, empty lines included.
- **ALG-03.2** *Refusal* — the first line that breaks ALG-03.1 ends the parse with an error whose text starts with `ERR_NOTES: `, followed by `<name>:<line>: ` (the input's name as given, then the line number) and the reason. `Render` refuses a note whose tag ALG-01.1 refuses with an error starting with `ERR_NOTES: `.
- **ALG-03.3** *Headings* — one heading `## <tag>` per distinct tag, newest first per ALG-01 through `semver.Compare`; the tag exactly as it was given, `v` prefix included (as ALG-02.3). Tags that compare equal keep the order of their first notes. Under its heading each note is one line `- <text>`, in input order.
- **ALG-03.4** *Candidates* — a release candidate (`semver.IsCandidate`) is listed like any tag, its heading ending in ` (candidate)`. The line `Latest release: <tag>` names what `release.Latest` returns for the tags (ALG-02.2: a final release beats every candidate), not the first heading.
- **ALG-03.5** *Layout* — the file is these lines, each ended by one newline: `# Release notes`; with at least one note an empty line, `Latest release: <tag>`, and per heading an empty line, the heading, an empty line and its note lines. Nothing follows the last note line, so the file ends with exactly one newline. Without notes the file is the title line alone.
- **ALG-03.6** *One parser* — `notes` parses, orders and classifies tags through `semver` and `release` alone, never with a parser of its own; `semver` (ALG-01) and `release` (ALG-02) do not change.
