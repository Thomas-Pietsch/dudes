# BENCH BACKLOG

## 0. How to work this backlog

1. One item per ladder level; an item is done when its acceptance tests and its DoD rows are green.

## 4. Work items

### R0.1 — Ladder

#### BL-001 · Slugs from titles
**Release.** R0.1 · **Status.** derived from QR
**Story.** As a writer I get a URL slug from a title.

**Functional requirements**
- FR-1 Package `slug` per [[TA:§3]]: `Make(title string) string` meets [[PRD:AC-SLUG-01]]. Runs of any other character become one hyphen.

**Acceptance criteria**
- AC-1 GIVEN the titles `Hello, World!`, `  Go 1.26  `, `a--b` and `!!!`, WHEN `slug.Make` runs on each, THEN it returns `hello-world`, `go-1-26`, `a-b` and the empty string.

**References.** [[PRD:AC-SLUG-01]] [[TA:§3]]
**DoD.** AC-SLUG-01 green.

#### BL-002 · Version order
**Release.** R0.1 · **Status.** derived from QR
**Story.** As a release manager I sort versions the way people read them.

**Functional requirements**
- FR-1 Package `semver` per [[TA:§3]] implements [[TA:ALG-01.1..4]].

**Acceptance criteria**
- AC-1 GIVEN `v1.10.0` and `1.9.3`, WHEN `semver.Compare` runs, THEN it returns 1.
- AC-2 GIVEN `1.2.0-rc.2`, `1.2.0-rc.10` and `1.2.0`, WHEN they are compared pairwise, THEN each sorts before the next.
- AC-3 GIVEN `1.x.0`, WHEN `semver.Parse` runs, THEN it returns an error starting with `ERR_VERSION:` that quotes `1.x.0`.

**References.** [[PRD:AC-VER-01]] [[TA:§3]] [[TA:ALG-01]]
**DoD.** AC-VER-01 green.

### R0.2 — Ladder L3

#### BL-003 · Latest release
**Release.** R0.2 · **Status.** derived from QR
**Story.** As a release manager I get the latest release of a repository from its tags.

**Functional requirements**
- FR-1 Package `release` per [[TA:§4]] implements [[TA:ALG-02.1..5]] on top of `semver` ([[TA:ALG-01]]); `semver` gains `IsCandidate`.

**Acceptance criteria**
- AC-1 GIVEN the tags `v1.9.0`, `v1.10.0-rc.1`, `nightly` and `1.10.0`, WHEN `release.Latest` runs, THEN it returns `1.10.0`.
- AC-2 GIVEN only `v2.0.0-rc.1` and `v2.0.0-rc.2`, WHEN `release.Latest` runs, THEN it returns `v2.0.0-rc.2`.
- AC-3 GIVEN no tag that parses, WHEN `release.Latest` runs, THEN it returns an error starting with `ERR_NO_RELEASE:`.

**References.** [[PRD:AC-REL-01]] [[TA:§4]] [[TA:ALG-02]] [[TA:ALG-01]]
**DoD.** AC-REL-01 green.

### R0.3 — Ladder L4

#### BL-004 · Release notes
**Release.** R0.3 · **Status.** derived from QR
**Story.** As a release manager I get a release-notes file from tagged notes, the same bytes every time.

**Functional requirements**
- FR-1 Package `notes` per [[TA:§5]] implements [[TA:ALG-03.1..6]] on top of `semver` ([[TA:ALG-01]], [[TA:§3]]) and `release` ([[TA:ALG-02]], [[TA:§4]]); both stay as they are. The committed `testdata/notes.txt` renders to the committed `testdata/release-notes.md`.

**Acceptance criteria**
- AC-1 GIVEN `testdata/notes.txt`, WHEN `notes.Parse` reads it and `notes.Render` renders the result, THEN the output equals `testdata/release-notes.md` byte for byte.
- AC-2 GIVEN an input named `notes.txt` with the lines `v1.0.0 First` and `1.x.0 Broken`, WHEN `notes.Parse` runs, THEN it returns an error starting with `ERR_NOTES: notes.txt:2:`.
- AC-3 GIVEN the notes `1.2.0 Patch`, `v1.10.0 Minor` and `v1.9.3 Fix` in that order, WHEN `notes.Render` runs, THEN the headings follow as `## v1.10.0`, `## v1.9.3`, `## 1.2.0` and the file holds the line `Latest release: v1.10.0`.

**References.** [[PRD:AC-NOTES-01]] [[TA:§5]] [[TA:ALG-03]] [[TA:ALG-01]] [[TA:ALG-02]] [[TA:§3]] [[TA:§4]]
**DoD.** AC-NOTES-01 green.

## 6. Global Definition of Done

1. `go test ./...` green; standard library only.
