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
- Commits (only when asked): Conventional Commits, English, single line, no body, no agent attribution.

## Repo facts

- Bun workspaces + Turborepo. Use `bun`, not npm/yarn/pnpm.
- Go module is `apps/api`; run Go commands from there.
- Go layers: handler → service → domain ← repository. Domain imports nothing from the others.
- Go checks (`gofmt`, `go vet`, `golangci-lint`, `go test -race`, `go build`) live in `scripts/go-check.go`; lefthook and CI call it. Change checks there only. Hooks check a snapshot (`scripts/run-on-tree.ts`): pre-commit the staged index, pre-push `HEAD`; the working tree is never touched. `golangci-lint` is resolved via `PATH` or `$(go env GOPATH)/bin`.
- CI: the frontend job must not require Go; Go checks run in the separate `go` job.
- Turbo strict env mode: new env vars go into `turbo.jsonc` (`passThroughEnv`).
- Commit messages and PR titles/bodies: English only. The `commit-message-language` hook rejects non-ASCII subjects; PR text rules in `.agents/skills/pr-description/SKILL.md`.
- Skills live in `.agents/skills/` (`.claude/skills` is a symlink).
- Architecture: `ARCHITECTURE.md`. Docs layout: `docs/README.md`.

## Code style

- No code comments. No TODOs-as-noise, no step narration. `doc.go` is a bare `package x`.
- Follow [code-simplification](.agents/skills/code-simplification/SKILL.md): clarity over cleverness, behavior preserved, no speculative abstractions.
- Guard clauses over deep nesting; named helpers over repeated conditionals; full words in names (`userID`, not `uid2`).
- No unrelated cleanup inside a feature change.
- `database/sql`: always close `*sql.Rows` — `defer func() { _ = rows.Close() }()` right after `QueryContext`/`Query`, including early returns. `QueryRowContext` needs none.

## Finishing

- State what was run and its result; say what was not run or failed.
- Final message: what changed, what is unresolved.
