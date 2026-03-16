FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/dealerpoints/ ./services/dealerpoints/
WORKDIR /app/services/dealerpoints
RUN go build -o /dealer-points-service .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /dealer-points-service /dealer-points-service
EXPOSE 50057 8086
ENV DEALER_POINTS_GRPC_PORT=50057
ENV DEALER_POINTS_HTTP_PORT=8086
ENTRYPOINT ["/dealer-points-service"]
