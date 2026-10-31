# @ucrp/api-docs

Swagger UI for the OpenAPI contract. The source of truth is `packages/openapi`; the bundle is generated and never committed.

```bash
bun run docs
```

Builds `dist/` and serves it at <http://localhost:8080>. Production serves the same `dist/` from the `web` image at `/docs/`.
