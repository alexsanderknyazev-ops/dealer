package config

import "testing"

const (
	testPostgresDSNOverride = "postgres://x"
	testJWTSecretOverride   = "j"
	testHTTPPortOverride    = 8888
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("DEALS_GRPC_PORT", "")
	c := Load()
	if c.GRPCPort != 50054 || c.HTTPPort != 8083 {
		t.Fatal(c)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("DEALS_GRPC_PORT", "bad")
	c := Load()
	if c.GRPCPort != 50054 {
		t.Fatal(c.GRPCPort)
	}
}

func TestLoad_Custom(t *testing.T) {
	t.Setenv("DEALS_HTTP_PORT", "8888")
	t.Setenv("POSTGRES_DSN", testPostgresDSNOverride)
	t.Setenv("JWT_SECRET", testJWTSecretOverride)
	c := Load()
	if c.HTTPPort != testHTTPPortOverride || c.PostgresDSN != testPostgresDSNOverride || c.JWTSecret != testJWTSecretOverride {
		t.Fatal(c)
	}
}
