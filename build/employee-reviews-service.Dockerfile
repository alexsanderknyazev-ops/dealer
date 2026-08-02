FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/employee/reviews/ ./services/employee/reviews/
WORKDIR /app/services/employee/reviews
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOMAXPROCS=2 go mod tidy && GOMAXPROCS=2 CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o /employee-reviews-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /employee-reviews-service /employee-reviews-service
EXPOSE 50063 8096
ENV EMPLOYEE_REVIEWS_GRPC_PORT=50063
ENV EMPLOYEE_REVIEWS_HTTP_PORT=8096
ENTRYPOINT ["/employee-reviews-service"]
