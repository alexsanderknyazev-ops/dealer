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
#копируем services/deals/ в рабочую директорию
COPY services/deals/ ./services/deals/
#устанавливаем рабочую директорию
WORKDIR /app/services/deals
#строим проект
RUN go build -o /deals-service .
#берем alpine 3.19
FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
#устанавливаем ca-certificates
RUN apk --no-cache add ca-certificates
#копируем deals-service в рабочую директорию
COPY --from=builder /deals-service /deals-service
#устанавливаем порты
EXPOSE 50054 8083
#устанавливаем переменные окружения
ENV DEALS_GRPC_PORT=50054
ENV DEALS_HTTP_PORT=8083
#устанавливаем entrypoint
ENTRYPOINT ["/deals-service"]
