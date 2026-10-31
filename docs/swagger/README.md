# Swagger UI — UCRP API

Static Swagger UI build for the spec in `../openapi.json` (multi-file structure in `../openapi/`).

## Open

Open `index.html` in a browser — it works offline, no server or internet required.

If the browser blocks `fetch` of a local file (some Chrome versions), start a static server:

```powershell
# from the repository root
npx --yes serve docs/swagger
# or
python -m http.server 8080 --directory docs/swagger
```

Then open <http://localhost:3000> (serve) or <http://localhost:8080>.

## Rebuild after spec changes

`openapi.json` here is a **generated** bundle. Edit files in `../openapi/`, then:

```powershell
npx @redocly/cli bundle docs/openapi.json --output docs/swagger/openapi.json
```

## Contents

| File | Purpose |
|---|---|
| `index.html` | Swagger UI page with settings (sorting, filter, `persistAuthorization`) |
| `openapi.json` | Bundled spec: 90 operations, 113 schemas, 0 external `$ref` |
| `swagger-ui.css` | Swagger UI 5.34.1 styles |
| `swagger-ui-bundle.js` | Swagger UI core |
| `swagger-ui-standalone-preset.js` | `StandaloneLayout` preset (top bar, server selector) |
| `oauth2-redirect.html` | OAuth2 redirect (for a future OAuth2 migration) |
| `favicon-*.png`, `LICENSE`, `NOTICE` | Assets and Apache-2.0 license from `swagger-ui-dist` |

## Authorizing in the UI

1. `POST /api/v1/auth/login` — run it in the **Auth** section and copy `accessToken`.
2. **Authorize** button → paste the token into `bearerAuth` (without the `Bearer` prefix).
3. The token survives page reloads (`persistAuthorization: true`).

The refresh token (`cookieAuth`) is not used in the UI: it lives in an httpOnly cookie and is not accessible from JavaScript.

## Limitations

* `POST /api/v1/imports` — multipart; pick the file manually in the UI.
* `GET /api/v1/exports/{id}/download` and `GET /api/v1/imports/{id}/rejected-rows?format=csv` return `text/csv` — the UI shows the text instead of downloading a file.
* `POST /api/v1/exports` requires the `Idempotency-Key` header — set it in the parameter field.