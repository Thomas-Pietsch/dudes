# Agent Review Report Rules

Append **immediately after** the `✅ Task Complete` block — same response, never separate.

## Format

```
---
### 🔍 Review Report

| Role | ✅ Went Well | ⚠️ Went Wrong / Improve |
|------|-------------|--------------------------|
| 🔹 <Role> | <concise points separated by ` · `> | <concise points separated by ` · `> |
```

## Rules

| Rule | Detail |
|------|--------|
| **One row per role** | Each role fills only its own row |
| **Honest, not diplomatic** | List real problems — no "everything was fine" padding |
| **Domain-focused** | Review concerns specific to your discipline (consult your DoD in `AGENTS.md`) |
| **Concise** | Max 3 points per cell; separate with ` · ` |
| **No blame** | Frame as process or tooling problems |
| **Row completeness** | Every file-type domain touched must have a role row. Single-role report is only valid if all edited files fall within that role's domain. |

## Focus

Each role reviews: iteration count, first-attempt correctness, efficiency gaps. Consult your **Workflow**, **Rules**, and **Definition of Done** in the active `AGENTS.md`.

## 🚫 Report Dedup

Never repeat the same information across Task Complete and Review Report. Performance metrics in Task Complete, quality reflection here.

## References
- Requires: `REPORT.md`, `DUDES.md` · Required by: none
