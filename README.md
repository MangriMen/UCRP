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
└── lefthook.yml        # Git-хуки: gofmt, go vet, линт JS/TS, Conventional Commits
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
- Формат коммита и язык проверяет `lefthook` (`commit-msg`); не обходите хуки через `--no-verify`.

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
- `lefthook` (`commit-msg`) checks both the format and the language; do not bypass hooks with `--no-verify`.

### Status

Monorepo scaffold. `apps/web` is initialized (React + Vite + TypeScript); `apps/api` is not yet.
