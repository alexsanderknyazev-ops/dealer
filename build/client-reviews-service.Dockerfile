FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/client/reviews/ ./services/client/reviews/
WORKDIR /app/services/client/reviews
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOMAXPROCS=2 go mod tidy && GOMAXPROCS=2 CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o /client-reviews-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /client-reviews-service /client-reviews-service
EXPOSE 50060 8089
ENV CLIENT_REVIEWS_GRPC_PORT=50060
ENV CLIENT_REVIEWS_HTTP_PORT=8089
ENTRYPOINT ["/client-reviews-service"]
