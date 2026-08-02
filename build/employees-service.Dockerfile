FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/employee/employees/ ./services/employee/employees/
WORKDIR /app/services/employee/employees
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOMAXPROCS=2 go mod tidy && GOMAXPROCS=2 CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o /employees-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /employees-service /employees-service
EXPOSE 50066 8099
ENV EMPLOYEES_GRPC_PORT=50066
ENV EMPLOYEES_HTTP_PORT=8099
ENTRYPOINT ["/employees-service"]
