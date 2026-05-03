package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("DEALER_POINTS_GRPC_PORT", "")
	c := Load()
	if c.GRPCPort != 50057 || c.HTTPPort != 8086 {
		t.Fatal(c)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("DEALER_POINTS_GRPC_PORT", "bad")
	c := Load()
	if c.GRPCPort != 50057 {
		t.Fatal(c.GRPCPort)
	}
}

func TestLoad_Custom(t *testing.T) {
	t.Setenv("DEALER_POINTS_HTTP_PORT", "7772")
	t.Setenv("POSTGRES_DSN", "postgres://x")
	t.Setenv("JWT_SECRET", "j")
	c := Load()
	if c.HTTPPort != 7772 || c.PostgresDSN != "postgres://x" || c.JWTSecret != "j" {
		t.Fatal(c)
	}
}
