FROM golang:1.26-alpine AS build
RUN apk add --no-cache ca-certificates git
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api .
ARG APP=api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags='-s -w' -o /out/server ./cmd/${APP}

FROM alpine:3.21 AS runner
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/server /app/server
RUN mkdir -p /data/uploads
ENV UCRP_HTTP_ADDR=:8080 \
    UCRP_UPLOAD_DIR=/data/uploads \
    UCRP_LOG_LEVEL=info \
    TZ=UTC
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/app/server"]
