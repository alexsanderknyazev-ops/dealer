FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY services/errors-ingest/ ./services/errors-ingest/
WORKDIR /app/services/errors-ingest
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOMAXPROCS=2 go mod tidy && GOMAXPROCS=2 CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o /errors-ingest-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /errors-ingest-service /errors-ingest-service
EXPOSE 8092
ENV ERRORS_INGEST_HTTP_PORT=8092
ENTRYPOINT ["/errors-ingest-service"]
