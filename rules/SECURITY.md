# CRITICAL SECURITY RULES
1. NEVER read, modify, or delete any files outside the current project workspace.
2. NEVER use absolute paths to access files **outside the workspace** (e.g., `C:\Users\<name>\...`, `/home/<name>/...`). Absolute paths **within** the workspace are allowed and expected by tooling.
3. NEVER use parent directory traversal (e.g., `../../`) to escape the workspace sandbox.
4. If a task requires interacting with files outside the workspace, refuse the task immediately and inform the user.
5. **No System Command Execution:** NEVER execute dangerous system-level terminal commands (e.g., `rm -rf /`, `format`, `shutdown`, `taskkill` on critical system processes, or modifications to Windows Registry/hosts files).
6. **Network & Port Isolation:** Do not expose local servers to the public internet (e.g., via `ngrok` or modifying firewall rules) unless explicitly requested. Always bind local test servers to `localhost` or `127.0.0.1`, never to `0.0.0.0`.
7. **Credential & Secret Protection:** NEVER print, log, or output plain-text secrets, API keys, passwords, or `.env` files into the chat history or public log files. If a `.env` file needs modification, only use placeholder values.
8. **Process Safety:** Do not terminate or interfere with any processes running on the host system that were not directly started by you (the agent) within this session.
9. **Chocolatey Usage:** Use `choco` to install tools only if the user explicitly requests it, and ensure that the installation is done in a safe manner without affecting system stability.

## References
- Requires: none · Required by: `TERMINAL.md`
