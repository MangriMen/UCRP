# Agent rules

Repo-wide rules for AI agents. Humans: `README.md`, `docs/`.

## Don't loop

- Same command or edit failed twice with the same error → stop, report, ask.
- Never re-run a passing check. One run per change.
- Read the current tree and `git log` before redoing work.
- No visible progress after ~10 tool calls → stop and summarize.
- Don't fix unrelated problems; list them in the final message.
- Don't revert and re-apply your own edits. Decide once.

## Ask first

- Open questions: `docs/requirements/00-open-questions.md`. Don't pick answers silently.
- No new dependencies, CI, hook or tooling changes unasked.
- Never commit, push or open a PR unless asked.
- Never bypass hooks (`--no-verify`, `LEFTHOOK=0`). A failing hook → stop and ask, even if the breakage is not yours.
- Commits (only when asked): Conventional Commits, English, single line, no body, no agent attribution. Details: `CONTRIBUTING.md`.

## Required reading

Open these files with the read tool before the matching work. Do not skip them or rely on this summary.

- Any code change: `ARCHITECTURE.md` (layers, slices, testing, comments).
- Commits, PRs, checks, hooks, CI, API contract changes: `CONTRIBUTING.md`.
- Writing a PR title or body: `.agents/skills/pr-description/SKILL.md`.
- Writing or refactoring code: `.agents/skills/code-simplification/SKILL.md`.
- Docs layout and where new documents go: `docs/README.md`.

## Repo facts

- Bun workspaces + Turborepo. Use `bun`, not npm/yarn/pnpm.
- Go module is `apps/api`; run Go commands from there.
- Turbo strict env mode: new env vars go into `turbo.jsonc` (`passThroughEnv`).
- Skills live in `.agents/skills/` (`.claude/skills` is a symlink).
- Go checks live in `scripts/go-check.go` (used by lefthook and CI); change checks only there. The frontend CI job must not require Go.

## Code style

- No code comments. No TODOs-as-noise, no step narration. `doc.go` is a bare `package x`.
- Follow the code-simplification skill: clarity over cleverness, behavior preserved, no wrappers-for-wrappers, no speculative abstractions, no renaming for taste, follow project conventions.
- Guard clauses over deep nesting; named helpers over repeated conditionals; full words in names (`userID`, not `uid2`).
- No unrelated cleanup inside a feature change.
- `database/sql`: always close `*sql.Rows` — `defer func() { _ = rows.Close() }()` right after `QueryContext`/`Query`, including early returns. `QueryRowContext` needs none.

## Finishing

- State what was run and its result; say what was not run or failed.
- Final message: what changed, what is unresolved.
