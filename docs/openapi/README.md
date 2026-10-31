# OpenAPI specification

Entry point — `../openapi.json`. It is an index: `info`, `servers`, `tags`, `security`, and a `$ref`
map to external files. Every file is ≤ 300 lines.

## Validate and bundle

```powershell
# validate (all external $refs resolve)
npx @redocly/cli lint docs/openapi.json

# bundle into a single file
npx @redocly/cli bundle docs/openapi.json --output docs/openapi.bundled.json

# preview the docs
npx @redocly/cli preview-docs docs/openapi.json
```

## Layout

| Path                                                            | Contents                                                                                                                                                          |
| --------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `paths/*.json`                                                  | Operations grouped by topic (auth, users, reference data, imports, guests, identity, segments, campaigns, exports, analytics, audit, access, invitations, system) |
| `components/schemas/*.json`                                     | Models grouped by topic; shared types (`ErrorResponse`, `PageMeta`, `UserRole`) are in `common.json`                                                              |
| `components/parameters.json`, `responses.json`, `security.json` | Shared parameters, error responses, `bearerAuth` and `cookieAuth`                                                                                                 |

Operations carry `x-roles` and `x-uc` extensions that link them to roles and use cases. `docs/design/endpoint-registry.md` complements the spec with entities and queue jobs.

## Editing rules

1. The spec is the source of truth: change it first, then the code, then `docs/design/endpoint-registry.md`.
2. Put a schema in the matching topic file and add a `$ref` line to `../openapi.json`.
3. Cross-file references are relative, with a JSON Pointer: `./common.json#/PageMeta`.
4. Always run `npx @redocly/cli lint docs/openapi.json` after changes.
