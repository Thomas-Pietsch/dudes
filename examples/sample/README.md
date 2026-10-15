# Sample product: Bookshelf

A made-up product in the reference format, so the engine has a spec set to read: a problem statement, a roadmap, a PRD, a TA, a UX document, a backlog of two work items and a QR, with the profile, the spec-set registry and one execution contract that go with them. It describes no real product. Nothing is built: every QR row is open.

The tests of the framework load it (`internal/spec`, `internal/specset`, `internal/itempack`, `internal/contract`, `internal/skill`), so the format stays true.

## Try it

No model, no network. From the root of this repository, after `make build config` (the binary is `bin/dudes`, `bin/dudes.exe` on Windows; `dudes` below is that binary). Set `"product": "bookshelf"` in the new `dudes.json`:

```text
mkdir -p .dudes
cp examples/sample/profile.json examples/sample/specset.json .dudes/
dudes item check               # C1–C8: 0 findings
dudes item board               # status per release, derived from the QR
dudes item pack BK-001         # .dudes/items/BK-001/{RULES,TASK,CONTEXT,PACK}.md + manifest.json
dudes item contract BK-001     # contract.json projected from the item
```

`specset.json` names the roots as paths relative to the workspace: here `examples/sample/lead` and `examples/sample/pm` inside this repository. In an instance they are the checkouts of its lead and spec repos; adjust them if yours live elsewhere (absolute paths work). The engine **reads** those files; it never renames them and never commits in those repos.

## The reference format

One owner per fact, `[[DOC:ID]]` references in the backlog, the status computed from the QR and never stored:

- `profile.json` declares the ID families (`AC-SHELF-nn` in the PRD, `API-nn`, `TBL-nn` and `ALG-nn.n` in the TA, `UXB-nn` and `W-MC-nn` in the UX, `BK-nnn` in the backlog), the backlog layout and the QR marks. DOC keys are the spec-set entry IDs.
- The PRD holds behaviour as `SHALL` lines with GIVEN/WHEN/THEN scenarios (F37) and the criteria the work items cite. It holds no route table and no DDL.
- The TA owns routes, tables and algorithms. A work item reaches them only by reference; a quoted UI string needs a `[[UX:…]]` anchor on its line.
- A business hypothesis stays field evidence (`kind: field`) and is not a CI gate.

## The execution contract

`contract.json` is what `dudes item contract BK-001` projects (sources, the three criteria as `behavior` rows) plus what a role adds: a field row, constraints, the item's writable scope and the tasks. It covers the service, slice 1 of 2; the form (FR-3) is a later slice, because an item's UI part stays with the role-play (`ROADMAP.md` lane `E1`).

- `author-tests`: the stub writes one failing test file per criterion.
- `test-add`: red first. `expect: fail`, `writable` limited to its test files, `approval: after`: the run pauses with the tests written. Read them, then `dudes decide test-add`, or `dudes redo test-add "<what is wrong>"`.
- `impl-add`: the code. The tests are outside its `writable`, so it cannot weaken them. `approval: after` as well.
- `qa-review`: the QR from requirement, test and run evidence; the field row stays open.

To run it, use a scratch git repository that holds a committed copy of `examples/sample/` (`dudes freeze` reads committed content), with the three JSON files in its `.dudes/`:

```text
dudes contract                 # show and validate
dudes freeze                   # hash every registered input with its git revision
dudes init -prompt "BK-001"    # item-bound: starts at test_authoring
dudes step -stub               # the first task, with the stub worker and no model
```

## Instance map

Named roots are the product layout. Dispatch goes to the root that owns the file.

| Root | Owns (this product) | An instance adds |
|------|---------------------|------------------|
| `lead` | PS, ROADMAP | BM |
| `pm` | PRD, TA, UX, BACKLOG, QR | SC |

There is no `COMPANY.md`. A vanilla clone with no checkouts skips dispatch.

## Manual spec-to-mockup handoff

[The copy-ready German briefing](mockup-brief.md) demonstrates the [`user-centered-design` skill's spec-to-mockup mode](../../skills/spec/user-centered-design/SKILL.md). Copy only its fenced prompt into a suitable design agent after checking the sharing boundary. It is self-contained; no DUDES command or image-generation API is needed.

It is an exploratory snapshot of the add-a-book flow, not a product spec. Three visual concepts are proposed, not three approved modes. The provenance and gap tables explain scope limits, missing references and open decisions. Regenerate it from the current sources before using it for implementation.
