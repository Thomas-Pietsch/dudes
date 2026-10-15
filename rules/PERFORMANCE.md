# UNIVERSAL TOKEN & CONTEXT OPTIMIZATION RULES

## 1. Output Token Efficiency (Stop Conversational Waste)
- **Zero Politeness:** Never write introductory or concluding phrases (e.g., "Sure, I can help", "Here is the modified code", "Let me know if you need anything else"). Start your response immediately with the technical output or tool call.
- **No Pre-Explanations:** Do not explain what you are about to do before invoking a tool (like writing a file or running a command). Just invoke the tool immediately.
- **Post-Explanation Minimalism:** If a task is successful, output zero explanations of the code syntax unless the user explicitly asked a question with a question mark (?).

## 2. Input Context & Memory Efficiency (Prevent Context Bloat)
- **Targeted File Modification:** When modifying code, never rewrite or output unchanged lines or sections of a file in the chat panel. Use precise search-and-replace blocks or unified diffs.
- **No Redundant Inspecting:** Do not run exploratory terminal commands (like `ls`, `dir`, `pwd`, `cat`) multiple times within the same session. Keep track of the directory structure and file contents in your active context memory.

## 3. Autonomous Execution Efficiency
- **Immediate Tool Invocation:** If a prompt requires code execution or terminal interaction, trigger the appropriate tool in your very first token. Do not output text before a tool call unless strictly necessary.
- **Strict Loop Prevention:** If a terminal command, build step, or test suite fails twice with the exact same error log, stop immediately. Do not attempt a third autonomous fix with the same approach. Present the error log to the user and ask for strategic guidance.
- **Tool-Call Budget:** Limit to **≤5 tool calls per response**. Large chains risk output truncation mid-tool-call, which corrupts the conversation history with orphaned `tool_result` errors and crashes the session with an unrecoverable API 400. If a task needs more than 5 calls, split across multiple response turns.

## References
- Requires: `TERMINAL.md` · Required by: none
