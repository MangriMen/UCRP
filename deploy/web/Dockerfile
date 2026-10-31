FROM oven/bun:1.4 AS deps
WORKDIR /app
COPY package.json bun.lock turbo.jsonc ./
COPY apps/api/package.json apps/api/package.json
COPY apps/web/package.json apps/web/package.json
RUN bun install --frozen-lockfile --ignore-scripts

FROM deps AS builder
COPY tsconfig.base.json ./
COPY apps/web apps/web
RUN bun run --filter=@ucrp/web build

FROM nginx:alpine AS runner
COPY deploy/web/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/apps/web/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
