FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/statistics/client/ ./services/statistics/client/
WORKDIR /app/services/statistics/client
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOMAXPROCS=2 go mod tidy && GOMAXPROCS=2 CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o /client-statistics-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /client-statistics-service /client-statistics-service
EXPOSE 50062 8095
ENV CLIENT_STATISTICS_GRPC_PORT=50062
ENV CLIENT_STATISTICS_HTTP_PORT=8095
ENTRYPOINT ["/client-statistics-service"]
