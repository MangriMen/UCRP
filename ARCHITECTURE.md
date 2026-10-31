# UCRP Architecture Guidelines

Rules for code structure, dependency direction and module boundaries. Mandatory for people and AI agents.

Product requirements, design diagrams and the OpenAPI contract live in `docs/` (see `docs/README.md`). This document describes **how the code is organized**, not **what the system does**.

---

## 1. Overview

```text
apps/web (React SPA)  ──REST /api/v1──▶  apps/api (Go)
                                            ├── cmd/api      HTTP server
                                            ├── cmd/worker   workers and scheduler
                                            ├── internal/    domain and use cases
                                            └── migrations/  goose
```

Monorepo on Bun workspaces and Turborepo. There is a single Go module for the repository (`github.com/MangriMen/UCRP`), located in `apps/api`.

---

## 2. Backend: layers and dependency direction

Dependencies point strictly inward:

```text
handler → service → domain ← repository
```

| Layer        | Responsibility                                                                        | May import                                             | Must not import                                             |
| ------------ | ------------------------------------------------------------------------------------- | ------------------------------------------------------ | ----------------------------------------------------------- |
| `domain`     | Entities, value objects, invariants, errors, **interfaces** of repositories and ports | standard library only                                  | everything else                                             |
| `service`    | Use cases, transaction boundaries, coordination                                       | `domain`                                               | `handler`, `repository` (only the interfaces from `domain`) |
| `repository` | SQL implementations of `domain` interfaces                                            | `domain`, `repository/sqlstore`                        | `service`, `handler`                                        |
| `handler`    | HTTP: routing, request parsing, status codes, DTOs                                    | `service`, `domain`, `platform/httpx`, `platform/auth` | `repository`                                                |
| `cmd/*`      | Composition: config, DI, graceful shutdown                                            | everything                                             | —                                                           |

Capability packages (`ingestion`, `identity`, `segments`, `recommend`, `export`) depend only on `domain` and `platform/*`. They never call each other directly; they interact through `service` or the job queue.

### Target tree of `apps/api`

```text
apps/api/
├── cmd/
│   ├── api/            # HTTP entry point
│   └── worker/         # workers: job queue, scheduler
├── internal/
│   ├── domain/         # entities, value objects, errors, port interfaces
│   ├── access/         # permission policy: roles, grants, permissions
│   ├── handler/        # HTTP layer, /api/v1 router, dto/
│   ├── service/        # use cases
│   ├── repository/     # SQL implementations, sqlstore/ (connection, UnitOfWork)
│   ├── ingestion/      # import pipeline, connector/ (csv, crm, pos, api)
│   ├── identity/       # key normalization, resolver, merge policy
│   ├── segments/       # rule compiler, evaluator
│   ├── recommend/      # scoring and explanations
│   ├── export/         # export channels and formats
│   └── platform/       # auth, httpx, queue, cache, storage, config, logging
├── migrations/         # goose
└── go.mod
```

---

## 3. Frontend: vertical slices

`apps/web/src` follows a light FSD (without extra layers):

- **`app/`**: composition: providers, router, layout, role-based menu, global styles.
- **`pages/`**: one folder per route; only composes features and entities.
- **`widgets/`**: created only when a block is reused on two or more pages.
- **`features/`**: user actions (`login`, `upload-import`, `confirm-merge`, `build-segment`, `create-export`, ...).
- **`entities/`**: business entities (`guest`, `segment`, `import-batch`, `user`, ...): types, API hooks, mocks, small UI blocks.
- **`shared/`**: technical slices without business logic: `api/` (fetch client, types generated from OpenAPI, MSW mocks), `ui/`, `lib/`, `config/`.

Inside a slice the segments `ui/ api/ model/ mocks/ lib/` are created as needed; the public API is `index.ts`.

### Rules

| Rule              | Description                                                                                    |
| ----------------- | ---------------------------------------------------------------------------------------------- |
| Single data entry | Components never call `fetch` directly, only hooks from `entities/*/api` or `features/*/api`   |
| Import direction  | `app → pages → widgets → features → entities → shared`. Reverse imports are forbidden (oxlint) |
| Slice isolation   | Slices of one layer do not import each other; shared code moves down a layer                   |
| Server state      | TanStack Query; a global store only for authentication                                         |
| Permission checks | The UI hides what is unavailable, but the deciding check is on the backend (403)               |

---

## 4. Shared packages (`packages/`)

| Package             | Purpose                                                    |
| ------------------- | ---------------------------------------------------------- |
| `packages/api-docs` | Swagger UI built from `docs/openapi.json` (`bun run docs`) |

A new package goes to `packages/` only if **two or more** applications use it. Otherwise the code lives inside the application.

---

## 5. Testing

| Level             | What is tested              | Tools                       |
| ----------------- | --------------------------- | --------------------------- |
| `domain` (Go)     | Invariants, pure functions  | `testing`, no DB or network |
| `service` (Go)    | Use cases with mocked ports | `testing`, `testify/mock`   |
| `repository` (Go) | SQL against a real database | `testing`, temporary DB     |
| `handler` (Go)    | HTTP contracts              | `httptest`                  |
| Frontend          | Components and hooks        | Vitest, Testing Library     |

Test names describe the expected behavior and the condition: `TestMerge_RejectsConflictingNames_WhenPhoneMatches`.

---

## 6. Comments in code

1. **Do not leave comments in code.** The code must read on its own. Remove "what the line does" narration, TODO noise and placeholder phrases.
2. Public godoc only when an exported signature is invalid without it; prefer none. `doc.go` in a skeleton is a bare `package name`.
3. Style and simplification follow [`.agents/skills/code-simplification/SKILL.md`](.agents/skills/code-simplification/SKILL.md): clarity over cleverness, behavior preserved, no slop.

---

## 7. Rules for AI agents

1. **Do not bloat the tree.** Add a file only when an existing one became unreadable (over ~400 lines) or a new cohesive module appeared.
2. **Respect the dependency direction.** Importing `repository` from `service` or `handler` is an error even if it compiles.
3. **Do not change visibility for tests.** Use tests in the same package.
4. **The API contract source of truth is `docs/openapi.json` and `docs/openapi/`.** When changing an endpoint, update the spec and regenerate the frontend types (`gen:api` in `apps/web`).
5. **Do not add a dependency** without checking that the standard library or an existing package does not already cover it.
6. **Code without comments and without slop**: `AGENTS.md` plus the `code-simplification` skill.
