# Claude project notes

See README.md for full documentation — tech stack, how to build, how to run, module layout, and folder structure all live there.

## Memory

At the start of every conversation, read `.claude/project-memory/MEMORY.md` to load project context from previous conversations. Coding conventions, agent delegation rules, Temporal-specific guardrails, the runbook for adding a new pattern, and the NATS event architecture all live there.

Use the `project-memory` skill (from the skillbox plugin) proactively — without being asked — whenever the conversation reveals decisions, references, or corrective feedback worth persisting across conversations.

**Important:** Always use the `project-memory` skill to persist information. Never use the built-in auto-memory system (`~/.claude/projects/.../memory/`) for project decisions or context — it is local and not shared.
