FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/statistics/employee/ ./services/statistics/employee/
WORKDIR /app/services/statistics/employee
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOMAXPROCS=2 go mod tidy && GOMAXPROCS=2 CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o /employee-statistics-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /employee-statistics-service /employee-statistics-service
EXPOSE 50061 8094
ENV EMPLOYEE_STATISTICS_GRPC_PORT=50061
ENV EMPLOYEE_STATISTICS_HTTP_PORT=8094
ENTRYPOINT ["/employee-statistics-service"]
