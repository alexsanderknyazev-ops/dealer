FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/brands/ ./services/brands/
WORKDIR /app/services/brands
RUN go build -o /brands-service .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /brands-service /brands-service
EXPOSE 50056 8085
ENV BRANDS_GRPC_PORT=50056
ENV BRANDS_HTTP_PORT=8085
ENTRYPOINT ["/brands-service"]
