FROM golang:1.26-alpine AS build

RUN apk add --no-cache ca-certificates git

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

ARG APP=api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags='-s -w' -o /out/server ./cmd/${APP}

FROM alpine:3.21 AS runtime

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=build /out/server /app/server
COPY migrations /app/migrations

RUN mkdir -p /data /data/uploads

ENV UCRP_HTTP_ADDR=:8080 \
    UCRP_DATABASE_URL=sqlite:///data/ucrp.db \
    UCRP_UPLOAD_DIR=/data/uploads \
    UCRP_LOG_LEVEL=info \
    TZ=UTC

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- "http://127.0.0.1:8080/healthz" || exit 1

ENTRYPOINT ["/app/server"]
