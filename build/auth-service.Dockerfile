# Сборка из корня репозитория: docker build -f build/auth-service.Dockerfile .
#берем node 20 alpine
FROM node:20-alpine AS frontend 
#устанавливаем рабочую директорию
WORKDIR /app/frontend/auth 
#копируем frontend/auth/ в рабочую директорию
COPY frontend/auth/ ./
#устанавливаем зависимости и строим проект
RUN npm install && npm run build 
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
#копируем services/auth/ в рабочую директорию
COPY services/auth/ ./services/auth/ 
#устанавливаем рабочую директорию
WORKDIR /app/services/auth
#строим проект
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /auth-service . \
  && CGO_ENABLED=0 go build -ldflags="-w -s" -o /seed-admin ./cmd/seed-admin
#берем alpine 3.19
FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
#устанавливаем ca-certificates
RUN apk --no-cache add ca-certificates
#копируем auth-service в рабочую директорию
COPY --from=builder /auth-service /auth-service
#копируем seed-admin в рабочую директорию
COPY --from=builder /seed-admin /seed-admin
#копируем frontend/auth/dist в рабочую директорию
COPY --from=frontend /app/frontend/auth/dist /app/web
#устанавливаем порты
EXPOSE 50051 8080
#устанавливаем переменные окружения
ENV STATIC_DIR=/app/web
ENV AUTH_HTTP_PORT=8080
#устанавливаем entrypoint
ENTRYPOINT ["/auth-service"]
