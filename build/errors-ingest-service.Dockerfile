FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY services/errors-ingest/ ./services/errors-ingest/
WORKDIR /app/services/errors-ingest
RUN go mod tidy && go build -o /errors-ingest-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /errors-ingest-service /errors-ingest-service
EXPOSE 8092
ENV ERRORS_INGEST_HTTP_PORT=8092
ENTRYPOINT ["/errors-ingest-service"]
