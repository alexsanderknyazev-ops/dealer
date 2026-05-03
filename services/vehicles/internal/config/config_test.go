package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("VEHICLES_GRPC_PORT", "")
	c := Load()
	if c.GRPCPort != 50053 || c.HTTPPort != 8082 {
		t.Fatal(c)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("VEHICLES_GRPC_PORT", "bad")
	c := Load()
	if c.GRPCPort != 50053 {
		t.Fatal(c.GRPCPort)
	}
}

func TestLoad_Custom(t *testing.T) {
	t.Setenv("VEHICLES_HTTP_PORT", "7777")
	t.Setenv("POSTGRES_DSN", "postgres://x")
	t.Setenv("JWT_SECRET", "j")
	c := Load()
	if c.HTTPPort != 7777 || c.PostgresDSN != "postgres://x" || c.JWTSecret != "j" {
		t.Fatal(c)
	}
}
