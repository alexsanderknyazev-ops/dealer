FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/employee/reviews/ ./services/employee/reviews/
WORKDIR /app/services/employee/reviews
RUN go mod tidy && go build -o /employee-reviews-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /employee-reviews-service /employee-reviews-service
EXPOSE 50063 8096
ENV EMPLOYEE_REVIEWS_GRPC_PORT=50063
ENV EMPLOYEE_REVIEWS_HTTP_PORT=8096
ENTRYPOINT ["/employee-reviews-service"]
