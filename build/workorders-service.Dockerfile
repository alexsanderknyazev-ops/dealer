FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/employee/workorders/ ./services/employee/workorders/
WORKDIR /app/services/employee/workorders
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOMAXPROCS=2 go mod tidy && GOMAXPROCS=2 CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o /workorders-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /workorders-service /workorders-service
EXPOSE 50064 8097
ENV WORKORDERS_GRPC_PORT=50064
ENV WORKORDERS_HTTP_PORT=8097
ENTRYPOINT ["/workorders-service"]
