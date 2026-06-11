package config

import "testing"

const (
	testPostgresDSNOverride = "postgres://x"
	testJWTSecretOverride   = "j"
	testHTTPPortOverride    = 7771
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("PARTS_GRPC_PORT", "")
	c := Load()
	if c.GRPCPort != 50055 || c.HTTPPort != 8084 {
		t.Fatal(c)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("PARTS_GRPC_PORT", "bad")
	c := Load()
	if c.GRPCPort != 50055 {
		t.Fatal(c.GRPCPort)
	}
}

func TestLoad_Custom(t *testing.T) {
	t.Setenv("PARTS_HTTP_PORT", "7771")
	t.Setenv("POSTGRES_DSN", testPostgresDSNOverride)
	t.Setenv("JWT_SECRET", testJWTSecretOverride)
	c := Load()
	if c.HTTPPort != testHTTPPortOverride || c.PostgresDSN != testPostgresDSNOverride || c.JWTSecret != testJWTSecretOverride {
		t.Fatal(c)
	}
}
