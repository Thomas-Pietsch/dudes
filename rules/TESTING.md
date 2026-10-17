# Testing Protocol

> Framework-level testing infrastructure. Defines HOW each test pyramid layer executes. Instance `AGENTS.md` declares WHAT to test (product-specific inputs). Any DUDES instance inherits this — no reinvention needed.
>
> **TDD+SDD Model:** QA splits into two phases — **Author** (write tests from PRD before implementation) and **Verify** (execute tests after implementation). Every PRD section has corresponding tests AND implementation.

---

## 1. Scope

| Layer | Protocol defined here | Instance AGENTS.md declares |
|-------|----------------------|----------------------------|
| Unit | §2 (QA Author writes, devs extend) | Test commands (`make unit`) |
| Integration | §2 (QA Author writes from PRD) | Test commands (`make integration`) |
| E2E — Functional | §3 | Which critical journeys |
| E2E — Visual Regression | §4 | Which views, which mockup files |

Future layers (performance, accessibility) slot into this file when added.

---

## 2. Test Authoring — QA Author Phase

> **When:** First role in the dev-repo flow, before any implementation begins.
> **Input:** Frozen PRD (spec repo branch `release/<product>-v<X.Y>`). Never another role's code or output.
> **Output:** Compilable test suite that fails (no implementation exists yet).

### What it is

QA Author reads each PRD section and writes the corresponding test cases. This is the TDD half of TDD+SDD: tests define the expected behaviour before code exists. Tests are the executable specification.

### Protocol

1. QA Author reads the frozen PRD section by section
2. For each PRD section with testable behaviour:
   a. Write **integration tests** covering the section's acceptance criteria
   b. Write **unit test stubs** for domain logic described in the section
   c. Write **E2E test skeletons** for user journeys (functional, per §3)
3. Tests must compile/parse but are **expected to fail** (no implementation yet)
4. Commit test suite: `test(scope): write tests from PRD §X [🔹QA Engineer]`
5. Hand off to Dev(s)

**Boundaries.** A test for a length or size limit names the limit's unit (characters, bytes, runes) and includes non-ASCII input: a limit counted in bytes passes every ASCII test. The test asserts both sides of the boundary, the last accepted value and the first refused one.

### PRD-to-Test Traceability

Every test file or test block references the PRD section it derives from:

```
// PRD §3.2 — User Authentication
// AC: User can log in with email + password
test("user can log in with valid credentials", ...)
```

### Gate

**Hard gate.** QA Author must commit a test suite before any dev role begins; the commit order in `git log` is the record, read at the review and the QR, and no hook judges which commit begins the dev work. Dev implementation without pre-existing tests = critical violation (`DUDES.md` §18).


---

## 3. E2E — Functional Testing (acceptance)

Automated browser journeys end to end — real clicks, navigation and API calls against a running app; Playwright is the framework standard. They are the acceptance tests: each names the scenario or requirement ID it proves (F77); unit tests carry none.

| Role | Does |
|---|---|
| QA Author | writes a failing journey test for each critical journey of the PRD or work item, committed before dev begins |
| QA Verify | runs the instance's acceptance goal (`make e2e`; the goal starts and stops the app) — every critical journey passes; result per journey in `*_QR.md` §E2E |

**Hard gate.** Any failing journey = 🔴 Critical; the release cannot be sealed. The engine runs it as a gate when the instance profile names the goal (`acceptance_tests.goal`, F35), and checks the IDs from the same profile. Critical journeys are defined in the instance `AGENTS.md`.

---

## 4. E2E — Visual Regression

### What it is

Two complementary checks: design-fidelity review against the approved Hi-Fi references, then screenshot regression against a reviewed running-application baseline. Neither proves that a user journey works; functional acceptance remains §3.

### Design inputs and coverage

The UI/UX discipline is [user-centered-design](../skills/spec/user-centered-design/SKILL.md) (F81). The existing UX spec owns a compact mapping from journey/scenario IDs to selected patterns, states, modes, rendered references and test checkpoints. Select critical journeys, distinct patterns and material variants; equivalent screens may reuse a reference plus their differences.

Missing or unreadable references, undecided modes and uncovered required states are explicit gaps, never a visual PASS. Prose-only mockup descriptions and generated images with illegible labels are not implementation-ready Hi-Fi evidence.

### Tooling

- **Playwright** — screenshot capture at defined viewports
- **Design-fidelity source:** Hi-Fi mockups in the spec repo (`ux/` directory, `*_UX.md` design specs), tied to the applicable frozen spec revision
- **Regression source:** Screenshots captured from the reviewed implementation, tied to its code revision; never blindly baseline the current app or pixel-diff a generated design image against browser rendering

### Protocol

1. QA checks the implementation against the selected Hi-Fi references and transition/invariant definitions; record deviations before accepting initial application baselines.
2. Playwright captures screenshots at defined viewports:
   - **Mobile:** 375px width
   - **Tablet:** 768px width
   - Additional viewports per instance `AGENTS.md` if needed
3. For each selected checkpoint, capture its required states (populated, empty, error, loading where applicable) and documented modes. Use fixed synthetic data, locale, clock, fonts and animation settings. Record the viewport and mode with the snapshot.
4. Compare against the reviewed application baseline using the instance's documented screenshot-diff thresholds. Unexpected differences block; an intentional change needs design review and baseline-update rationale, not automatic snapshot acceptance.
5. Run the affected §3 journeys as well. For supported presentation modes, verify switching and the critical happy path in each mode; add error/recovery and responsive combinations where behavior or layout differs. Switching must preserve current route, selected record, draft and permissions per spec. Avoid an exhaustive Cartesian product without a risk reason.
6. Shared shell/navigation, token or component changes expand checks to their consumers. Include back/cancel, failed save/retry and retained context where the affected journey requires them.

### Hard-Gate vs Human-Flag Boundary

| Check type | Gate | What it means |
|-----------|------|---------------|
| **Layout structure** | Hard | Grid, flex layout, section ordering matches mockup |
| **Component presence** | Hard | All designed components are rendered and visible |
| **Responsive behaviour** | Hard | Layout adapts correctly at each viewport |
| **State coverage** | Hard | Empty, error, loading states exist and render |
| **Variant and flow invariants** | Hard | Required modes render; switching and affected journeys preserve the specified state and behavior |
| **Colour accuracy** | Human flag | Exact colour matching — flagged, not blocked |
| **Animation / transitions** | Human flag | Motion design — flagged, not blocked |
| **Subjective aesthetics** | Human flag | "Feels right" — flagged, not blocked |

**Hard-gate failure** = 🔴 Critical. **Human-flag** = documented in `*_QR.md` §E2E-Visual for human gate review. An unexpected screenshot diff requires review; an intentional colour-only change can be accepted with rationale, while structural or behavioral failures cannot be waived as styling.

These are the QA protocol's gate requirements, not a new automatic Go validator. Instances provide executable goals; the engine runs a configured acceptance goal through F35. Manual review and field-usability results must be reported as such.

### Output

Per-view pass/fail matrix documented in `*_QR.md` §E2E-Visual:
- View name, Viewport tested, State tested, Hard-gate result (🟢/🔴), Human-flag notes (if any)
- Journey/scenario ID, mode, spec/code revision, design reference, baseline/checkpoint, linked functional result, and rationale for any accepted baseline change


---

## 5. Test Verification — QA Verify Phase

> **When:** After all dev roles have completed implementation and committed.
> **Input:** Frozen PRD + test suite (from QA Author) + implemented code (from devs).
> **Output:** `*_QR.md` with pass/fail per PRD section.

### Protocol

1. QA Verify re-reads the frozen PRD (PRD-Direct Rule — never rely on dev's interpretation)
2. Run the full test suite: unit, integration, E2E
3. For each PRD section: record pass/fail against its corresponding tests
4. If tests fail: hand back to responsible dev role with failure details → dev fixes → QA Verify re-runs
5. All tests pass → write the verdict rows of the `*_QR.md` and commit on the release branch

### PRD Section Coverage Matrix

The `*_QR.md` states the current run only: verdict, the dev commit that was verified, and the matrix below. Each run is one QR commit — `git log` of the QR is the run history; no history table in the file (`hooks/pre-commit` refuses new dated "run N" rows).

The `*_QR.md` must include:

| PRD Section | Tests Written (QA Author) | Tests Passing (QA Verify) | Status |
|-------------|--------------------------|--------------------------|--------|
| §X.Y | count | count | ⬜/🟢/🔴/⚠️ (`skills/spec/quality-report`) |

Every PRD section with testable behaviour must have a row. Missing rows = incomplete QA.

---

## 6. Instance Specialization

The instance `AGENTS.md` (QA Workflow) declares:

| What | Example |
|------|---------|
| Critical journeys list | "Onboarding, navigation, CRUD flows" |
| Views to screenshot | "Home, Shelf, Settings" |
| Coverage and modes | UX journey/pattern/state matrix, supported modes and switching invariants; reuse/exclusion rationale |
| Mockup file locations | `<spec repo>/ux/<product>_hf_*` |
| Visual baseline policy | Fixture/viewport setup, baseline location, diff thresholds, review evidence and goal |
| Additional viewports | "1440px desktop" |
| App-specific commands | `make e2e APP=bookshelf` |
| PRD-to-test mapping | Which PRD sections map to which test directories |

The framework provides the protocol. The instance provides the inputs.

---

## References
- Requires: `DUDES.md` §5 (TEST GATE + PRD-Direct Rule), `CORE.md` §6 (test-first) · `rules/TESTING.md` owns the pyramid layers
- Required by: Instance `AGENTS.md` (QA Workflow)
