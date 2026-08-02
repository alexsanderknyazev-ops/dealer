FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/client/auth/ ./services/client/auth/
WORKDIR /app/services/client/auth
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOMAXPROCS=2 go mod tidy && GOMAXPROCS=2 CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o /client-auth-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /client-auth-service /client-auth-service
EXPOSE 50059 8088
ENV CLIENT_AUTH_GRPC_PORT=50059
ENV CLIENT_AUTH_HTTP_PORT=8088
ENTRYPOINT ["/client-auth-service"]
