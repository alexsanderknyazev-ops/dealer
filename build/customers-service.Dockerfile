FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/customers/ ./services/customers/
WORKDIR /app/services/customers
RUN go build -o /customers-service .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /customers-service /customers-service
EXPOSE 50052 8081
ENV CUSTOMERS_GRPC_PORT=50052
ENV CUSTOMERS_HTTP_PORT=8081
ENTRYPOINT ["/customers-service"]
