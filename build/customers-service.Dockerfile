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
#копируем services/customers/ в рабочую директорию
COPY services/customers/ ./services/customers/
#устанавливаем рабочую директорию
WORKDIR /app/services/customers
#строим проект
RUN go build -o /customers-service .

#берем alpine 3.19
FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
#устанавливаем ca-certificates
RUN apk --no-cache add ca-certificates
#копируем customers-service в рабочую директорию
COPY --from=builder /customers-service /customers-service
#устанавливаем порты
EXPOSE 50052 8081
#устанавливаем переменные окружения
ENV CUSTOMERS_GRPC_PORT=50052
ENV CUSTOMERS_HTTP_PORT=8081
#устанавливаем entrypoint
ENTRYPOINT ["/customers-service"]
