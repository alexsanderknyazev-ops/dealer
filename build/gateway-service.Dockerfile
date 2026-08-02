FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/gateway/employee/ ./services/gateway/employee/
WORKDIR /app/services/gateway/employee
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOMAXPROCS=2 go mod tidy && GOMAXPROCS=2 CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o /gateway-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /gateway-service /gateway-service
EXPOSE 8090
ENV GATEWAY_HTTP_PORT=8090
ENTRYPOINT ["/gateway-service"]
