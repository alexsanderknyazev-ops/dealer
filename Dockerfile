# Образ по умолчанию из корня репозитория: auth-service + frontend/auth.
# Сборка: docker build -t auth-service:local .
FROM node:20-alpine AS frontend
WORKDIR /app/frontend/auth
COPY frontend/auth/ ./
RUN npm install && npm run build

FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/auth/ ./services/auth/

WORKDIR /app/services/auth
RUN go build -o /auth-service .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /auth-service /auth-service
COPY --from=frontend /app/frontend/auth/dist /app/web
EXPOSE 50051 8080
ENV STATIC_DIR=/app/web
ENV AUTH_HTTP_PORT=8080
ENTRYPOINT ["/auth-service"]
