# Agent rules

Short rules for any AI agent working in this repo. Humans: see `README.md` and `docs/`.

## Don't loop

- Same command or edit failed twice with the same error → stop. Do not retry a third time. Report the error and what you tried, then ask.
- Never re-run a passing check "to be sure". One run per change is enough.
- Don't redo work that is already in the tree or in `git log`. Read the current state first.
- If a task needs more than ~10 tool calls without visible progress, stop and summarize where you are.
- Don't fix unrelated problems you notice along the way. Mention them in the final message.
- Don't revert and re-apply your own edits back and forth. Decide, make the change once.

## Ask first

- Do not initialize `apps/web` or `apps/api` (Vite, `go mod init`, frameworks, libraries) without the user's explicit go-ahead.
- Open questions are in `docs/00-open-questions.md`. Don't silently pick answers; ask.
- Don't add dependencies, change CI, hooks, or tooling versions unasked.
- Never commit, push, or open a PR unless asked.
- Commits (only when the user asks): message in English, single line, no body/description, author = the user (`git commit` without agent attribution tricks).

## Repo facts

- Bun workspaces + Turborepo. Use `bun`, not npm/yarn/pnpm.
- Go module lives in `apps/api` (not the repo root). Run Go commands from there.
- Go layers: handler → service → domain ← repository. Domain imports nothing from the others.
- CI: the frontend job must not require Go; Go lint/build/test run in the separate `go` job.
- Turbo runs in strict env mode: new env vars must be listed in `turbo.jsonc` (`passThroughEnv`).
- Commits: Conventional Commits (enforced by lefthook). Don't bypass hooks with `--no-verify`.
- Commit messages and PR titles/bodies: **English only**. The `commit-message-language` lefthook job rejects non-ASCII commit subjects; PR text rules are in `.agents/skills/pr-description/SKILL.md`.
- Skills live in `.agents/skills/` (`.claude/skills` is a symlink to it).

## Code style (anti-neuroslop)

- **No code comments.** Do not leave explanatory comments, TODOs-as-noise, «scaffold only», or step narration. The code must read without them. Package doc files (`doc.go`) are bare `package x` only.
- **Write via [code-simplification](.agents/skills/code-simplification/SKILL.md).** Clarity over cleverness; preserve behavior; no wrappers-for-wrappers, no speculative abstractions, no renaming for taste.
- Prefer guard clauses over deep nesting; named helpers over repeated conditionals; full words in names (`userID`, not `uid2`).
- Do not batch unrelated “cleanup” into feature changes. Stay in scope.

## Finishing

- Say what was actually run and its result. If something wasn't run or failed, say so.
- Keep the final message short: what changed, what's unresolved.
