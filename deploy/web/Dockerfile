FROM oven/bun:1.4.2 AS deps
WORKDIR /app
COPY package.json bun.lock turbo.jsonc ./
COPY apps/api/package.json apps/api/package.json
COPY apps/web/package.json apps/web/package.json
COPY packages/api-docs/package.json packages/api-docs/package.json
RUN bun install --frozen-lockfile --ignore-scripts

FROM deps AS builder
COPY tsconfig.base.json ./
COPY apps/web apps/web
ARG VERSION=dev
ARG BUILD_TIME=unknown
ENV VITE_APP_VERSION=${VERSION} VITE_BUILD_TIME=${BUILD_TIME}
RUN bun run --filter=@ucrp/web build

FROM deps AS docs
COPY docs/openapi.json docs/openapi.json
COPY docs/openapi docs/openapi
COPY packages/api-docs packages/api-docs
RUN bun run --filter=@ucrp/api-docs build

FROM nginx:alpine AS runner
COPY deploy/web/docs-auth.sh /docker-entrypoint.d/40-docs-auth.sh
COPY deploy/web/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/apps/web/dist /usr/share/nginx/html
COPY --from=docs /app/packages/api-docs/dist /usr/share/nginx/docs
RUN chmod +x /docker-entrypoint.d/40-docs-auth.sh
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
