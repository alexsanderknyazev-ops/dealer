FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/vehicles/ ./services/vehicles/
WORKDIR /app/services/vehicles
RUN go build -o /vehicles-service .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /vehicles-service /vehicles-service
EXPOSE 50053 8082
ENV VEHICLES_GRPC_PORT=50053
ENV VEHICLES_HTTP_PORT=8082
ENTRYPOINT ["/vehicles-service"]
