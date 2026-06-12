module github.com/dealer/dealer/services/appointments

go 1.22

replace github.com/dealer/dealer => ../../..

require (
	github.com/dealer/dealer v0.0.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.2
	google.golang.org/grpc v1.64.1
)
