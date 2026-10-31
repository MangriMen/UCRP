# apps/web

React + TypeScript SPA built with Vite, Tailwind, shadcn/ui and TanStack Query. Structure follows a light Feature-Sliced Design (`app → pages → features → entities → shared`); the rules are in [ARCHITECTURE.md](../../ARCHITECTURE.md).

```sh
bun run dev        # dev server (proxy settings in vite.config.ts)
bun run dev:mock   # dev server with MSW mocks, no backend needed
bun run test       # Vitest
bun run lint       # Oxlint
bun run build      # type-check and production build
bun run gen:api    # regenerate src/shared/api/schema.d.ts from docs/openapi.json
```

API types are generated from the OpenAPI contract; after changing `packages/openapi/`, run `gen:api` and commit the result.
