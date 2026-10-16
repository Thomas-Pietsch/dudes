---
name: codebase-design
description: Keep the change inside the existing architecture and give new code deep modules at real seams. Use when shaping TA-aligned code.
---

# Codebase design

1. **Read the TA** (or the contract section that points at it). Match the existing packages, names and the glossary's terms.
2. **Extend, don't add a framework.** Prefer a small extension of the current structure over a new framework inside the product.
3. **Deep modules:** a lot of behaviour behind a small interface. The interface is everything a caller must know: signature, invariants, error modes, configuration. Callers and tests cross the same interface. If a test must reach past it, the module has the wrong shape.
4. **The deletion test.** Delete the module in your head. If its complexity reappears in its callers, it earns its keep; if nothing is lost, it was a pass-through. Inline it.
5. **A seam needs two adapters.** Introduce an interface only where something really varies (the real database and a test double, two providers). One adapter is a hypothetical seam: skip it.
6. **Testable by shape:** accept dependencies instead of creating them; return results instead of mutating the caller's state.
7. **Boundaries:** workspace containment, orchestrator-owned git, no secrets in artifacts. Do not rename instance spec files. Do not add a second PRD.
8. **A product decision stops you.** If the design needs one, stop and ask the human, as `rules/DUDES.md` §16 asks about a requirement. Do not pick a default.

Skills teach discipline. They do not choose the next phase.
