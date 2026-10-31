FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
RUN apk add --no-cache ca-certificates git
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY apps/api .
ARG APP=api
ARG TARGETARCH
ARG VERSION=dev
ARG BUILD_TIME=unknown
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w -X github.com/MangriMen/UCRP/internal/platform/buildinfo.Version=${VERSION} -X github.com/MangriMen/UCRP/internal/platform/buildinfo.BuildTime=${BUILD_TIME}" -o /out/server ./cmd/${APP}

FROM alpine:3.21 AS runner
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/server /app/server
RUN addgroup -S -g 10001 ucrp \
    && adduser -S -u 10001 -G ucrp ucrp \
    && mkdir -p /data/uploads \
    && chown -R ucrp:ucrp /data
USER ucrp
ENV UCRP_HTTP_ADDR=:8080 \
    UCRP_UPLOAD_DIR=/data/uploads \
    UCRP_LOG_LEVEL=info \
    TZ=UTC
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/app/server"]
