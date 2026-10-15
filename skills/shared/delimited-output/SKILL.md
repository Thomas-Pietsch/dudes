---
name: delimited-output
description: Emit artifacts between DUDES delimiters. Use when writing Markdown or code the Go engine must extract.
---

# Delimited output

The DUDES engine parses worker replies. Wrap the entire artifact between these exact lines:

```
===BEGIN ARTIFACT===
…content…
===END ARTIFACT===
```

Do not put commentary outside the delimiters. Do not invent a competing file format. Skills teach discipline; they do not choose the next phase.
