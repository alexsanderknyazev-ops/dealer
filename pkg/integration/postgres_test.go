//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/dealer/dealer/pkg/postgres"
)

func TestMigrations_CreateSchemasAndTables(t *testing.T) {
	inst := requirePostgres(t)
	ctx := context.Background()

	schemas := []string{
		"auth", "customers", "vehicles", "deals", "parts", "brands",
		"dealerpoints", "clients", "clientauth", "reviews",
		"employee_statistics", "client_statistics", "employee_reviews",
		"workorders", "works", "employees", "appointments",
	}
	for _, s := range schemas {
		var n int
		if err := inst.Pool.QueryRow(ctx,
			`SELECT count(*) FROM information_schema.schemata WHERE schema_name = $1`, s,
		).Scan(&n); err != nil {
			t.Fatalf("query schema %q: %v", s, err)
		}
		if n != 1 {
			t.Errorf("schema %q not created by migrations", s)
		}
	}

	tables := [][2]string{
		{"auth", "users"},
		{"customers", "customers"},
		{"vehicles", "vehicles"},
		{"deals", "deals"},
		{"parts", "parts"},
		{"parts", "part_stock"},
		{"parts", "stock_movements"},
		{"parts", "movement_documents"},
		{"parts", "supplier_orders"},
		{"parts", "customer_orders"},
		{"brands", "brands"},
		{"dealerpoints", "dealer_points"},
		{"clients", "clients"},
		{"clientauth", "users"},
		{"reviews", "reviews"},
		{"workorders", "work_orders"},
		{"works", "works"},
		{"employees", "employees"},
		{"appointments", "repair_appointments"},
	}
	for _, tt := range tables {
		var n int
		if err := inst.Pool.QueryRow(ctx,
			`SELECT count(*) FROM information_schema.tables WHERE table_schema = $1 AND table_name = $2`,
			tt[0], tt[1],
		).Scan(&n); err != nil {
			t.Fatalf("query table %s.%s: %v", tt[0], tt[1], err)
		}
		if n != 1 {
			t.Errorf("table %s.%s not created by migrations", tt[0], tt[1])
		}
	}
}

func TestPostgres_NewPoolSearchPath(t *testing.T) {
	inst := requirePostgres(t)
	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, inst.DSN, "auth")
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&n); err != nil {
		t.Fatalf("query auth.users via search_path: %v", err)
	}

	plain, err := postgres.NewPool(ctx, inst.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer plain.Close()

	_, err = plain.Exec(ctx, `SELECT count(*) FROM users`)
	if err == nil {
		t.Error("expected error for unqualified users outside auth schema")
	}
}
