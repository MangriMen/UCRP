# UCRP backend — multi-stage build.
# Context: apps/api (корень Go-модуля). Dockerfile: apps/api/docker/Dockerfile.
#
#   docker build -f docker/Dockerfile --build-arg APP=api    -t ucrp-api:local .
#   docker build -f docker/Dockerfile --build-arg APP=worker -t ucrp-worker:local .

# ---------- build ----------
FROM golang:1.24-alpine AS build

RUN apk add --no-cache ca-certificates git

WORKDIR /src

# Слой зависимостей кэшируется, пока не изменится go.mod / go.sum
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# APP=api | worker — какую точку входа собираем
ARG APP=api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags='-s -w' -o /out/server ./cmd/${APP}

# ---------- runtime ----------
FROM alpine:3.21 AS runtime

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S ucrp && adduser -S -G ucrp -u 10001 ucrp

WORKDIR /app

# Один бинарь /app/server; содержимое зависит от ARG APP при сборке образа
COPY --from=build /out/server /app/server
COPY migrations /app/migrations

ENV UCRP_HTTP_ADDR=:8080 \
    UCRP_DATABASE_URL=sqlite:///data/ucrp.db \
    UCRP_UPLOAD_DIR=/data/uploads \
    UCRP_LOG_LEVEL=info \
    TZ=UTC

USER ucrp
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- "http://127.0.0.1:8080/healthz" || exit 1

ENTRYPOINT ["/app/server"]
