# Agent Reporting Rules

At the end of every completed task, the closing agent appends:

```
---
### ✅ Task Complete — 🔷 <Role>

| Step | Role | Performance |
|------|------|-------------|
| <what was done> | 🔹 <Role> | ~X min / ~Y tokens |
```

**Rules:**

| Rule | Detail |
|------|--------|
| **One row per role** | Each role fills in only its own row — never pre-fill downstream |
| **Last agent closes** | The final role writes the full report |
| **Time** | Wall-clock estimate for that role's turns |
| **Tokens** | Input + output combined for that role's turns |

## References
- Requires: `DUDES.md` · Required by: `REVIEW.md`


