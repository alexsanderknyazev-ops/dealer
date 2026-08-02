#берем golang 1.22 alpine
FROM golang:1.22-alpine AS builder
#устанавливаем рабочую директорию
WORKDIR /app
#копируем go.mod и go.sum в рабочую директорию
COPY go.mod go.sum ./
#копируем pkg/ в рабочую директорию
COPY pkg/ ./pkg/
#копируем api/ в рабочую директорию
COPY api/ ./api/
#копируем services/brands/ в рабочую директорию
COPY services/employee/brands/ ./services/employee/brands/
#устанавливаем рабочую директорию
WORKDIR /app/services/employee/brands
#строим проект
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOMAXPROCS=2 go mod tidy && GOMAXPROCS=2 CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o /brands-service .

#берем alpine 3.19
FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
#устанавливаем ca-certificates
RUN apk --no-cache add ca-certificates
#копируем brands-service в рабочую директорию
COPY --from=builder /brands-service /brands-service
#устанавливаем порты
EXPOSE 50056 8085
#устанавливаем переменные окружения
ENV BRANDS_GRPC_PORT=50056
ENV BRANDS_HTTP_PORT=8085
#устанавливаем entrypoint
ENTRYPOINT ["/brands-service"]
