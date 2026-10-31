# UCRP — Unified Customer Recommendation Platform

Единая рекомендательная клиентская платформа для торговых центров.

## Содержание

- [Русский](#русский)
- [English](#english)

---

## Русский

### О продукте

Торговый центр — это 80–150 независимых бизнесов под одной крышей. Каждый арендатор знает только своих клиентов, управляющая компания (УК) знает всё про аренду и трафик, но ничего про конкретного человека.

**UCRP** собирает данные арендаторов (CRM/PMS, касса/POS, API, Excel/CSV) в единую модель событий, строит по ним единый профиль гостя и превращает его в действие: приглашение в место, где человек ещё не был.

Три функции в порядке ценности:

1. **Единый профиль** — один человек, одна запись, сколько бы систем арендаторов он ни прошёл.
2. **История гостя** — что человек делал (визит, покупка, бронь), где, когда и на какую сумму.
3. **Рекомендация следующего шага** — «был в A и B, не был в C → пригласить в C», сразу с выгрузкой в канал рассылки.

### Структура репозитория

```text
.
├── apps/
│   ├── api/            # Go: HTTP API + воркеры (cmd/api, cmd/worker, internal/, migrations/)
│   └── web/            # React + Vite + TypeScript (SPA)
├── packages/           # Общие TS-пакеты (типы API, UI-кит, конфиги)
├── docs/               # Проектная документация: аналитика, ER, C4, DFD, контракты
├── .github/workflows/  # CI: frontend (bun/turbo) и backend (go)
├── turbo.jsonc         # Пайплайн задач монорепозитория
└── lefthook.yml        # Git-хуки: Go-проверки, линт JS/TS, Conventional Commits
```

Подробнее о границах модулей и правилах зависимостей — в [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

### Стек

| Слой        | Технологии                                                                   |
| ----------- | ---------------------------------------------------------------------------- |
| Backend     | Go, слои `handler → service → repository`, REST, воркеры + очередь заданий   |
| Frontend    | React, Vite, TypeScript, TanStack Query/Table, React Hook Form, React Router |
| БД          | SQLite на старте, переход на PostgreSQL без смены модели; миграции goose     |
| Инструменты | Bun workspaces, Turborepo, lefthook, golangci-lint                           |

### Требования

- **Go** 1.24+
- **Node.js** >= 24
- **Bun** 1.4.2
- **golangci-lint** (опционально, для локального линта Go)

### Быстрый старт

```sh
bun install          # зависимости монорепозитория и git-хуки
cp .env.example .env # локальная конфигурация
bun run dev          # запуск всех приложений через turbo
```

Полезные команды:

```sh
bun run build        # сборка всех пакетов
bun run lint         # линт всех пакетов
bun run test         # тесты всех пакетов
bun run clean        # очистка артефактов сборки
```

### Соглашения

- Коммиты — [Conventional Commits](https://www.conventionalcommits.org/), сообщение на английском, одна строка без тела.
- Заголовок и описание PR — на английском, см. [`.agents/skills/pr-description/SKILL.md`](.agents/skills/pr-description/SKILL.md).
- Формат коммита и язык проверяет `lefthook` (`commit-msg`); не обходите хуки через `--no-verify`. Если хук падает, остановитесь и разберитесь (или спросите), а не пропускайте проверку.
- Go-проверки (`gofmt`, `go vet`, `golangci-lint`, тесты, сборка) описаны в `scripts/go-check.sh`; их же вызывают `lefthook` и CI. Локально: `scripts/go-check.sh` (всё) или `scripts/go-check.sh lint`.
- Хуки проверяют снимок, а не рабочее дерево (`scripts/run-on-tree.sh`): pre-commit — то, что в индексе (staged), pre-push — `HEAD`. Незастейдженные и неотслеживаемые файлы не проверяются, рабочее дерево не меняется (в отличие от `git stash`). Это касается Go и `apps/web` (линт в pre-commit; тесты и сборка в pre-push, занимают около полутора минут). Форматирование `oxfmt` в pre-commit только проверяет staged-файлы и ничего не правит; исправить — `bun run format`. `node_modules` подключаются в снимок ссылками (симлинки на Linux/macOS, junction на Windows), отдельная установка не нужна. Скрипты — bash: на Windows нужен Git for Windows (Git Bash), хуки запускаются из него.

### Документация

Проектная документация лежит в [`docs/`](docs/). Точка входа — [`docs/README.md`](docs/README.md).

### Статус

Каркас монорепозитория. `apps/web` инициализирован (React + Vite + TypeScript), `apps/api` ещё нет.

---

## English

### About

**UCRP** is a unified customer recommendation platform for shopping malls. It ingests tenant data (CRM/PMS, POS, API, Excel/CSV) into a single event model, builds a unified guest profile, and turns it into action: an invitation to a place the person has never visited.

### Repository layout

```text
apps/api    # Go HTTP API + workers
apps/web    # React + Vite + TypeScript SPA
packages/   # Shared TypeScript packages
docs/       # Project documentation (analysis, ER, C4, DFD, API contracts)
```

### Stack

Go backend (`handler → service → repository`, workers + job queue), React + Vite + TypeScript frontend, SQLite with a migration path to PostgreSQL, goose migrations. Tooling: Bun workspaces, Turborepo, lefthook, golangci-lint.

### Getting started

```sh
bun install
cp .env.example .env
bun run dev
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for module boundaries and dependency rules.

### Conventions

- Commits follow [Conventional Commits](https://www.conventionalcommits.org/); the message is English, one line, no body.
- PR titles and bodies are English, see [`.agents/skills/pr-description/SKILL.md`](.agents/skills/pr-description/SKILL.md).
- `lefthook` (`commit-msg`) checks both the format and the language; do not bypass hooks with `--no-verify`; if a hook fails, stop and fix it or ask, do not skip the check.
- Go checks (`gofmt`, `go vet`, `golangci-lint`, tests, build) are defined in `scripts/go-check.sh`, used by both `lefthook` and CI. Locally: `scripts/go-check.sh` (everything) or `scripts/go-check.sh lint`.
- Hooks check a snapshot, not the working tree (`scripts/run-on-tree.sh`): pre-commit checks the staged index, pre-push checks `HEAD`. Unstaged and untracked files are not checked and the working tree is never touched (unlike `git stash`). This covers Go and `apps/web` (lint on pre-commit; tests and build on pre-push, about 1.5 minutes). `oxfmt` in pre-commit only checks staged files and never rewrites them; fix with `bun run format`. `node_modules` is linked into the snapshot (symlinks on Linux/macOS, junctions on Windows), no separate install is needed. The scripts are bash: on Windows Git for Windows (Git Bash) is required, and hooks run from it.

### Status

Monorepo scaffold. `apps/web` is initialized (React + Vite + TypeScript); `apps/api` is not yet.
