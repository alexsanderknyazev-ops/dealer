FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/client/registration/ ./services/client/registration/
WORKDIR /app/services/client/registration
RUN go mod tidy && go build -o /client-registration-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /client-registration-service /client-registration-service
EXPOSE 50058 8087
ENV CLIENT_REGISTRATION_GRPC_PORT=50058
ENV CLIENT_REGISTRATION_HTTP_PORT=8087
ENTRYPOINT ["/client-registration-service"]
