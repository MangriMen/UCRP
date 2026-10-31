# Contributing

Technical setup and workflow. Architecture rules are in [ARCHITECTURE.md](ARCHITECTURE.md), agent rules in [AGENTS.md](AGENTS.md).

## Requirements

- Go (version in `apps/api/go.mod`)
- Node.js >= 24
- Bun (version in `package.json`)
- golangci-lint (optional, for local Go lint)
- On Windows: Git for Windows (Git Bash); hooks and scripts are bash

## Setup

```sh
bun install            # workspace dependencies and git hooks
cp .env.example .env   # local configuration
bun run dev            # run all apps through turbo
```

Other commands: `bun run build`, `bun run lint`, `bun run test`, `bun run format`, `bun run clean`. Backend, web and demo-data commands are in `apps/api/README.md` and `apps/web/README.md`.

## Repository layout

```text
apps/api/           Go HTTP API and workers
apps/web/           React + Vite + TypeScript SPA
packages/           openapi (API contract), api-docs (Swagger UI)
deploy/             Dockerfiles and compose files
docs/               requirements and design
archive/            finished planning artifacts
scripts/            go-check.sh, run-on-tree.sh used by hooks and CI
```

## Commits and pull requests

- Commits follow [Conventional Commits](https://www.conventionalcommits.org/); the message is English, one line, no body.
- PR titles and bodies are English, see [`.agents/skills/pr-description/SKILL.md`](.agents/skills/pr-description/SKILL.md).
- `lefthook` (`commit-msg`) checks the format and the language. Do not bypass hooks with `--no-verify`; if a hook fails, fix the cause or ask.

## Checks

- Go checks (`gofmt`, `go vet`, `golangci-lint`, tests, build) are defined in `scripts/go-check.sh` and used by both `lefthook` and CI. Locally: `scripts/go-check.sh` (everything) or `scripts/go-check.sh lint`.
- Hooks check a snapshot, not the working tree (`scripts/run-on-tree.sh`): pre-commit checks the staged index, pre-push checks `HEAD`. Unstaged and untracked files are not checked and the working tree is never touched.
- Pre-commit: Go format, vet and lint, `oxfmt --check` on staged files, web lint. Pre-push: Go tests and build, web tests and build (about 1.5 minutes). `oxfmt` only checks; fix with `bun run format`.
- `node_modules` is linked into the snapshot (symlinks on Linux and macOS, junctions on Windows), no separate install is needed.
- CI runs the frontend job (no Go required) and the `go` job separately; see `.github/workflows/ci.yml`.

## Contract changes

`packages/openapi/` is the source of truth. After changing it, run `bun run gen:api` in `apps/web` and update `docs/design/endpoint-registry.md`.
