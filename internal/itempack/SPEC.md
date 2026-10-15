# Context Packs for Micro Agents — Feature Specification (context-pack/1)

> **Purpose.** Given a spec set written in the reference format (STRATEGY.md §H) and one work item, produce a self-contained, deterministic *context pack* that a small agent can implement from without reading any other document. This file is the contract for `internal/itempack` (FEATURES F68–F71). The Python reference implementation it was ported from is retired: its output is kept as `testdata/golden/` (byte-parity test) and `itempack_test.go` is the conformance suite. `STRATEGY.md` is the strategy document of the first instance, the origin of the format.
>
> **Status:** ready for implementation · **Source format:** STRATEGY.md §H.1 (ownership), §H.2 (work item contract), §H.3 (reference syntax + resolution contract), §H.4 (consistency rules), §H.6 (backlog method) — summarised in §2–§3 below; the fixture `testdata/demo/STRATEGY.md` carries §H.6.

---

## 1. Terms

| Term | Meaning |
|---|---|
| **Spec set** | All Markdown documents of one vertical plus the shared documents, located by a *profile* |
| **Profile** | `.dudes/profile.json` — ID families and the few format conventions that differ per vertical. Document keys are spec-set entry IDs (`.dudes/specset.json`), so files may live in any named root |
| **Owner** | The one document that defines an ID (STRATEGY §H.1) |
| **Reference** | `[[DOC:ID]]`, `[[DOC:ID..ID]]`, `[[DOC:§n]]` inside a work item |
| **Definition** | The text an ID resolves to in its owner (§2.2) |
| **Work item** | A `#### XX-nnn · Title` section in the BACKLOG items section |
| **Complete context** | The work item plus all definitions its references resolve to (STRATEGY §H.3) |
| **Pack** | The complete context rendered as files + manifest for one work item |

## 2. Functional requirements

### CP-FR-01 Find the work item
- Work items are level-4 headings matching `backlog.item_heading` inside the level-2 section starting with `backlog.items_section`.
- Release = value of `backlog.release_field` in the item; fallback for outline rows: the enclosing `backlog.release_heading`.
- Outline rows (`backlog.outline_row`) are listed and shown on the board, but **cannot be packed** (they are not implementation-ready).
- Item text = the heading plus everything up to the next heading of level ≤ 4.

### CP-FR-02 Resolve references
For `[[DOC:TARGET]]`, look only in document `DOC` (never search other documents).

| # | Target form | Definition returned |
|---|---|---|
| a | ID in the first cell of a table row | that row **plus the table header and separator row** |
| b | ID at the start of a heading (after an optional section number such as `6.2`), or in an ID list before ` · ` in a heading | the heading and its section up to the next heading of equal or higher level |
| c | ID as bold lead of a list item (`- **ID** …`) or bold paragraph lead (`**ID — Title**`) | the list item line / the paragraph up to the next bold lead, heading or `---` |
| d | Range `X-01..05` / `X-01..X-05` / sub-rule range `ALG-01.1..3` | each ID in order, zero padding preserved; consecutive rows of the same table are emitted under **one** header |
| e | `§n` / `§n.m` / `§4a` | the first heading whose text starts with that number followed by `.` or a space |
| f | Table ID (`table_ddl.id_prefix`, e.g. `TBL-nn`) | its row plus, from the same document, the `CREATE TABLE <name>` statement, every `CREATE TRIGGER … ON <name>` and every `CREATE INDEX … ON <name>`, as one `sql` block |
| g | IDs in `multi_definition_allowed` (e.g. `API-`) | **all** definitions: index row and schema heading |
| h | Lines inside fenced code blocks | never count as headings or definitions |

An ID token is bounded: not preceded by `[\w-]`, not followed by `\w` or `.digit` (so `ALG-03` never matches `ALG-03.2`).

### CP-FR-03 Fail loudly
- An unknown document key, a missing section or an undefined ID raises an error naming document and ID.
- A document listed in the profile whose file does not exist is simply not loaded (not every vertical has every document, e.g. no BM yet); only a reference into it is an error.
- A pack is **never** written with unresolved references; the item is reported as not ready (STRATEGY §H.6 Definition of Ready).

### CP-FR-04 Render the pack
Output folder `<out>/<ITEM-ID>/`:

| File | Content | Consumer |
|---|---|---|
| `RULES.md` | The sections named in `rules_sections`, in profile order (STRATEGY §H.6 method, BACKLOG §0, BACKLOG §6) | system prompt / constitution |
| `TASK.md` | The work item **verbatim** | task |
| `CONTEXT.md` | One `## DOC` group per document in `context_order`, then any other document alphabetically; inside a group one `### [[DOC:TARGET]]` block per reference in first-appearance order; duplicates once | reference material |
| `PACK.md` | `RULES` + `TASK` + `CONTEXT` | single-prompt runners |
| `manifest.json` | §CP-FR-05 | orchestrator |

Every file carries one generated/do-not-edit line. Output is deterministic except for that date line and `manifest.generated`.

### CP-FR-05 Manifest and budget
- `manifest.json` validates against `testdata/manifest.schema.json`: item, title, release, derived status, DoD IDs, missing DoD IDs, budget, `over_budget`, file sizes, every reference with source document, size and `trimmed` flag, SHA-256 of **every loaded source file**.
- **Budget** (characters, optional): `TASK` and `RULES` are never shortened. While the pack is over budget, the largest not-yet-shortened `CONTEXT` block is cut to its first 12 lines plus a pointer line naming the reference. If still over budget after all blocks are shortened, the pack is written with `over_budget: true` (the orchestrator decides: bigger model or split item).
- **Staleness:** a pack is stale when any `sources` hash differs from the current files; regenerate, never patch.

### CP-FR-06 Derived status
- DoD IDs = IDs in the item's `backlog.dod_field`, ranges and slash lists (`AC-X-08/09`) expanded, BACKLOG-owned IDs (other work items) ignored.
- QR index: for every table row of `status.doc`, the status marks in cells that *start with* a mark (`⬜`, `🟢`, `🔴`, `⚠️`), attached to every ID in the row (ranges expanded).
- `done` = every DoD ID has only done-marks · `blocked` = any blocked-mark · `untracked` = a DoD ID without QR row · `open` = otherwise · `outline` for outline rows. Status is computed, never stored in a spec.

### CP-FR-07 Consistency check (STRATEGY §H.4)

| Code | Rule |
|---|---|
| C1 | Every `[[…]]` in every document resolves (references inside inline code are examples and are skipped) |
| C2 | Every owner-family ID referenced in a product document is defined in its owner, exactly once (except `multi_definition_allowed`); shared documents and lines containing `skip_line_marker` are not scanned for references |
| C3 | Work items contain no route shape outside `[[…]]`, no DDL marker, and no quoted German UI string without a UX anchor on the same line |
| C4 | The PRD contains no route table row and no DDL |
| C5 | Every route matching `route_pattern` in any document occurs in `route_owner` |
| C6 | Every work item has a release, and the release is a ROADMAP milestone |
| C7 | Every ROADMAP milestone has at least one work item or outline row |
| C8 | Every DoD ID of a work item has a QR row |

Exit code 1 on any finding; one finding per line: `<code> <location> <message>`.

### CP-FR-08 Board and list
- `list`: id, release, kind (`item`/`outline`), title — tab separated.
- `board`: per release, a table item · title · number of DoD IDs · derived status.

## 3. Profile (per vertical)

Everything product-specific is data, not code: shared documents, ordered owner families (first full match wins — put narrow exceptions first), the backlog regexes (plus optional `backlog.ac_line`, default `- AC-n` / `- **AC-n**`), rules sections, status marks, roadmap milestone regex, table-DDL settings, context order and the consistency patterns. Document locations come from the spec-set, not the profile. Profiles are RE2 regexes (Go): no look-around, no back-references. Example: `examples/sample/profile.json`.

## 4. Integration in DUDES

```text
dudes item check                 C1–C8; with a profile, dudes freeze refuses on findings
dudes item board                 pick: status == open, release == current, DoR met
dudes item pack <ID> [-budget N] .dudes/items/<ID>/ — over_budget → larger tier or split the item
dudes item contract <ID>         contract.json projected from the item (item, sources, acceptance)
dudes step / run                 item-bound contract → workers get PACK.md (skill: work-item)
```

The PR records `manifest.sources`; QR rows turn green → the next derive says `done`.

| Go call | Returns |
|---|---|
| `itempack.Open(workspace, product)` / `New(profile, docs)` | loaded set |
| `(*Set).Resolve(doc, target)` | definition texts, or `*ResolveError` |
| `(*Set).Items()` | work items and outline rows |
| `(*Set).DeriveStatus(item, idx)` | status, DoD IDs, missing IDs |
| `(*Set).Build(itemID, budget)` | `*Pack` — nothing written |
| `itempack.Write(workspace, pack)` | `.dudes/items/<ID>` |
| `(*Set).Check()` | findings |

## 5. Non-functional requirements

| ID | Requirement |
|---|---|
| CP-NFR-01 | No network, no model call, no write outside `out` — the generator is deterministic software, by the strategy rule that what is deterministic is software |
| CP-NFR-02 | Reads UTF-8, accepts CRLF and LF, never modifies a spec file |
| CP-NFR-03 | A full instance set (~10 files, ~600 kB): check < 5 s, one pack < 1 s on a laptop |
| CP-NFR-04 | A pack never contains text that is not in a spec file, except headings, the generated line and budget pointers |
| CP-NFR-05 | Same inputs → byte-identical files except the date line and `manifest.generated` |

## 6. Acceptance tests (GIVEN / WHEN / THEN) → `itempack_test.go`

Test names below are the kit's; the Go tests carry the same names in CamelCase (`TestTableRowWithHeader`).

| ID | GIVEN / WHEN / THEN | Test |
|---|---|---|
| CP-AC-01 | GIVEN a table-row ID, WHEN resolved, THEN the row comes with its header and without neighbouring rows | `test_table_row_with_header` |
| CP-AC-02 | GIVEN a range over table rows, WHEN resolved, THEN one header, all rows | `test_range_merges_rows_under_one_header` |
| CP-AC-03 | GIVEN a heading ID or `§n`, WHEN resolved, THEN the section ends at the next equal-level heading | `test_heading_section_until_next_equal_heading`, `test_section_reference` |
| CP-AC-04 | GIVEN bold list items / bold paragraphs, WHEN resolved, THEN exactly that item / paragraph | `test_bold_list_item_and_subrule_range`, `test_bold_paragraph_block` |
| CP-AC-05 | GIVEN a table ID, WHEN resolved, THEN `CREATE TABLE` and its triggers are included | `test_table_includes_ddl_and_triggers` |
| CP-AC-06 | GIVEN an API ID, WHEN resolved, THEN index row and schema section | `test_api_includes_index_row_and_schema` |
| CP-AC-07 | GIVEN a heading-like line inside a code fence, WHEN resolving, THEN it is ignored | `test_code_fence_is_not_structure` |
| CP-AC-08 | GIVEN an undefined ID or unknown document, WHEN resolving, THEN an error, no pack | `test_unresolved_reference_fails_loudly` |
| CP-AC-09 | GIVEN a ready item, WHEN packed, THEN five files and a schema-valid manifest with release and expanded DoD | `test_pack_has_five_files_and_manifest_contract`, `test_write_pack_creates_folder` |
| CP-AC-10 | GIVEN a pack, THEN TASK is verbatim and CONTEXT is grouped in profile order | `test_task_is_verbatim_and_context_is_grouped` |
| CP-AC-11 | GIVEN the profile's rules sections, THEN RULES contains exactly those | `test_rules_come_from_configured_sections` |
| CP-AC-12 | GIVEN a budget smaller than TASK + RULES, WHEN packed, THEN references are shortened, TASK is whole, `over_budget` is true | `test_budget_trims_largest_first_and_flags_overrun` |
| CP-AC-13 | GIVEN an outline row, WHEN packing it, THEN an error | `test_outline_items_are_not_packable` |
| CP-AC-14 | GIVEN QR marks ⬜ / all 🟢 / one 🔴 / a missing row, THEN status open / done / blocked / untracked | `Status.*` |
| CP-AC-15 | GIVEN the clean fixture, WHEN checked, THEN zero findings; GIVEN one seeded violation per rule, THEN that rule's code is reported | `Consistency.*` (C1–C8, incl. an SC double definition) |
| CP-AC-16 | GIVEN IDs owned by SC and a section in PS, WHEN resolved, THEN they come from those documents; GIVEN overlapping owner patterns, THEN the narrow exception listed first wins (`TC-SEC-13` → SC, `TC-SEC-01` → PRD) | `Owners.test_sc_*`, `test_ps_section_resolves`, `test_narrow_owner_exception_wins` |
| CP-AC-17 | GIVEN a document listed in the profile but absent on disk, WHEN checking, THEN no finding; WHEN a reference points into it, THEN C1 | `Owners.test_listed_but_absent_document_is_tolerated_until_referenced` |

**Conformance:** a DUDES implementation is conforming when CP-AC-01..17 hold against the fixtures in `testdata/demo/` and when it reports 0 findings on the sample spec set (`examples/sample/`) with every work item packable (`TestSampleSpecsConform`). At port time the largest item of the first instance's live set had 37 references.

## 7. Out of scope (context-pack/1)

Writing or editing spec files; generating work items (a scaffold may *propose* a draft, never write it); semantic trimming (e.g. only the needed rows of a large table — candidate for context-pack/2); any model call; secrets, repository code or chat history in a pack (the strategy rule: never in context).
