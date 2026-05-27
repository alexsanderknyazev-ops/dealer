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
#копируем services/dealerpoints/ в рабочую директорию
COPY services/dealerpoints/ ./services/dealerpoints/
#устанавливаем рабочую директорию
WORKDIR /app/services/dealerpoints
#строим проект
RUN go build -o /dealer-points-service .

#берем alpine 3.19
FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
#устанавливаем ca-certificates
RUN apk --no-cache add ca-certificates
#копируем dealer-points-service в рабочую директорию
COPY --from=builder /dealer-points-service /dealer-points-service
#устанавливаем порты
EXPOSE 50057 8086
#устанавливаем переменные окружения
ENV DEALER_POINTS_GRPC_PORT=50057
ENV DEALER_POINTS_HTTP_PORT=8086
#устанавливаем entrypoint
ENTRYPOINT ["/dealer-points-service"]
