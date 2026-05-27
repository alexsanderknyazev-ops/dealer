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
#копируем services/vehicles/ в рабочую директорию
COPY services/vehicles/ ./services/vehicles/
#устанавливаем рабочую директорию
WORKDIR /app/services/vehicles
#строим проект
RUN go build -o /vehicles-service .
#берем alpine 3.19
FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
#устанавливаем ca-certificates
RUN apk --no-cache add ca-certificates
#копируем vehicles-service в рабочую директорию
COPY --from=builder /vehicles-service /vehicles-service
#устанавливаем порты
EXPOSE 50053 8082
#устанавливаем переменные окружения
ENV VEHICLES_GRPC_PORT=50053
ENV VEHICLES_HTTP_PORT=8082
#устанавливаем entrypoint
ENTRYPOINT ["/vehicles-service"]
