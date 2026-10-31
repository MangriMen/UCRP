# apps/

Приложения монорепозитория. Каталог — отдельный workspace-пакет со своим `package.json`. У `api/` он тонкий: только `dev` (`go run`), чтобы `bun run dev` поднимал и бэкенд. Сборка, линт и тесты Go идут отдельно: `go build/vet/test` и golangci-lint в `apps/api`, в CI — job `go`.

| Каталог | Стек | Назначение |
|---|---|---|
| `api/` | Go | HTTP API (`cmd/api`), воркеры (`cmd/worker`), миграции goose |
| `web/` | React + Vite + TypeScript | SPA: импорт, профили, сегменты, рекомендации, выгрузка, дашборды |

Оба приложения пока не инициализированы. Правила структуры — в [../docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md).
