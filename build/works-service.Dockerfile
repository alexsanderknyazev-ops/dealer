FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/employee/works/ ./services/employee/works/
WORKDIR /app/services/employee/works
RUN go mod tidy && go build -o /works-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /works-service /works-service
EXPOSE 50065 8098
ENV WORKS_GRPC_PORT=50065
ENV WORKS_HTTP_PORT=8098
ENTRYPOINT ["/works-service"]
