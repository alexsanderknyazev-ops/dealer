FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/gateway/client-public/ ./services/gateway/client-public/
WORKDIR /app/services/gateway/client-public
RUN go mod tidy && go build -o /client-public-gateway-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /client-public-gateway-service /client-public-gateway-service
EXPOSE 8091
ENV CLIENT_PUBLIC_GATEWAY_HTTP_PORT=8091
ENTRYPOINT ["/client-public-gateway-service"]
