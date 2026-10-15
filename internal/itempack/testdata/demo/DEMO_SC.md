# DEMO SC

## 2. Threats

| ID | STRIDE | Threat | Measure |
|---|---|---|---|
| T-DM-01 | Tampering | Author of a note rewritten later | `author_id` trigger (TA TBL-01) |

## 14. Attribution tests

| ID | Test | Priority |
|---|---|---|
| TC-SEC-ATTR-01 | Direct SQL update of `author_id` aborts | P0 |
| TC-SEC-13 | Profile file with unknown key refuses start (SC extends the PRD base TC-SEC-01..12) | P0 |
