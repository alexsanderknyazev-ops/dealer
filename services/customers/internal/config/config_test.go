package config

import (
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("CUSTOMERS_GRPC_PORT", "")
	t.Setenv("CUSTOMERS_HTTP_PORT", "")
	t.Setenv("POSTGRES_DSN", "")
	t.Setenv("JWT_SECRET", "")
	c := Load()
	if c.GRPCPort != 50052 || c.HTTPPort != 8081 {
		t.Fatalf("ports grpc=%d http=%d", c.GRPCPort, c.HTTPPort)
	}
	if c.PostgresDSN == "" || c.JWTSecret == "" {
		t.Fatal("empty defaults")
	}
}

func TestLoad_CustomEnv(t *testing.T) {
	t.Setenv("CUSTOMERS_GRPC_PORT", "60000")
	t.Setenv("CUSTOMERS_HTTP_PORT", "9001")
	t.Setenv("POSTGRES_DSN", "postgres://x")
	t.Setenv("JWT_SECRET", "abc")
	c := Load()
	if c.GRPCPort != 60000 || c.HTTPPort != 9001 || c.PostgresDSN != "postgres://x" || c.JWTSecret != "abc" {
		t.Fatalf("%+v", c)
	}
}

func TestLoad_InvalidIntFallsBack(t *testing.T) {
	t.Setenv("CUSTOMERS_GRPC_PORT", "notint")
	c := Load()
	if c.GRPCPort != 50052 {
		t.Fatalf("got %d", c.GRPCPort)
	}
}
