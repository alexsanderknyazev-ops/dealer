FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY api/ ./api/
COPY services/employee/appointments/ ./services/employee/appointments/
WORKDIR /app/services/employee/appointments
RUN go mod tidy && go build -o /appointments-service .

FROM alpine:3.19
ARG SERVICE_VERSION=dev
ENV SERVICE_VERSION=${SERVICE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=builder /appointments-service /appointments-service
EXPOSE 50067 8101
ENV APPOINTMENTS_GRPC_PORT=50067
ENV APPOINTMENTS_HTTP_PORT=8101
ENTRYPOINT ["/appointments-service"]
