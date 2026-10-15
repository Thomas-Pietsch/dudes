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
