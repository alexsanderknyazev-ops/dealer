FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/parts/ ./services/parts/
WORKDIR /app/services/parts
RUN go build -o /parts-service .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /parts-service /parts-service
EXPOSE 50055 8084
ENV PARTS_GRPC_PORT=50055
ENV PARTS_HTTP_PORT=8084
ENTRYPOINT ["/parts-service"]
