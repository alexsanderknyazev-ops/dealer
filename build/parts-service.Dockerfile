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
#копируем services/parts/ в рабочую директорию
COPY services/parts/ ./services/parts/
#устанавливаем рабочую директорию
WORKDIR /app/services/parts
#строим проект
RUN go build -o /parts-service .
#берем alpine 3.19
FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
#устанавливаем ca-certificates
RUN apk --no-cache add ca-certificates
#копируем parts-service в рабочую директорию
COPY --from=builder /parts-service /parts-service
#устанавливаем порты
EXPOSE 50055 8084
#устанавливаем переменные окружения
ENV PARTS_GRPC_PORT=50055
ENV PARTS_HTTP_PORT=8084
#устанавливаем entrypoint
ENTRYPOINT ["/parts-service"]
