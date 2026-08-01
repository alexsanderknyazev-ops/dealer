//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dealer/dealer/pkg/dbschema"
	"github.com/dealer/dealer/pkg/postgres"
	tc "github.com/dealer/dealer/pkg/testcontainers"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()
	if !tc.DockerAvailable() {
		fmt.Println("skip integration tests: docker unavailable")
		os.Exit(0)
	}
	pg, err := tc.StartPostgres(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "start postgres:", err)
		os.Exit(1)
	}
	pool, err := postgres.NewPool(ctx, pg.DSN, dbschema.EmployeeReviews, dbschema.Public)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create pool:", err)
		os.Exit(1)
	}
	testPool = pool
	code := m.Run()
	pool.Close()
	_ = pg.Close(ctx)
	os.Exit(code)
}
