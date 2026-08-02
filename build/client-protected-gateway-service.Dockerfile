FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/gateway/client-protected/ ./services/gateway/client-protected/
WORKDIR /app/services/gateway/client-protected
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOMAXPROCS=2 go mod tidy && GOMAXPROCS=2 CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o /client-protected-gateway-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /client-protected-gateway-service /client-protected-gateway-service
EXPOSE 8093
ENV CLIENT_PROTECTED_GATEWAY_HTTP_PORT=8093
ENTRYPOINT ["/client-protected-gateway-service"]
