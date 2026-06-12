FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY services/scheduler/ ./services/scheduler/
WORKDIR /app/services/scheduler
RUN go mod tidy && go build -o /scheduler-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /scheduler-service /scheduler-service
EXPOSE 8100
ENV SCHEDULER_HTTP_PORT=8100
ENTRYPOINT ["/scheduler-service"]
