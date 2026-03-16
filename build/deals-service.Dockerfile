FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/deals/ ./services/deals/
WORKDIR /app/services/deals
RUN go build -o /deals-service .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /deals-service /deals-service
EXPOSE 50054 8083
ENV DEALS_GRPC_PORT=50054
ENV DEALS_HTTP_PORT=8083
ENTRYPOINT ["/deals-service"]
