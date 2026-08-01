//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	tc "github.com/dealer/dealer/pkg/testcontainers"
)

var pg *tc.Postgres

func TestMain(m *testing.M) {
	ctx := context.Background()
	if !tc.DockerAvailable() {
		fmt.Println("skip integration tests: docker unavailable")
		os.Exit(0)
	}
	var err error
	pg, err = tc.StartPostgres(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "start postgres:", err)
		os.Exit(1)
	}
	code := m.Run()
	_ = pg.Close(ctx)
	os.Exit(code)
}

func requirePostgres(t *testing.T) *tc.Postgres {
	t.Helper()
	tc.SkipIfNoDocker(t)
	return pg
}
