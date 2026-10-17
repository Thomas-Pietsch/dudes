# BENCH PRD

A tiny Go library the model bench builds, one work item per ladder level. Standard library only.

## 2. Features

### F1 Slugs

| ID | Criterion |
|---|---|
| AC-SLUG-01 | GIVEN a title, WHEN it is turned into a slug, THEN the slug holds only lowercase ASCII letters, digits and single hyphens, with no hyphen at either end. |

### F2 Versions

| ID | Criterion |
|---|---|
| AC-VER-01 | GIVEN two version strings, WHEN they are compared, THEN the order follows [[TA:ALG-01]], not the string order. |

### F3 Releases

| ID | Criterion |
|---|---|
| AC-REL-01 | GIVEN a repository's tags, WHEN the latest release is asked for, THEN the answer follows [[TA:ALG-02]]. |

### F4 Release notes

| ID | Criterion |
|---|---|
| AC-NOTES-01 | GIVEN a notes file of tags and short texts, WHEN the release notes are rendered, THEN the file follows [[TA:ALG-03]] byte for byte. |
