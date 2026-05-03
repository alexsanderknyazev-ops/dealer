package config

import (
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("AUTH_GRPC_PORT", "")
	c := Load()
	if c.GRPCPort != 50051 || c.HTTPPort != 8080 {
		t.Fatal(c)
	}
	if c.AccessTTL != 15*time.Minute || c.RefreshTTL != 168*time.Hour {
		t.Fatal(c.AccessTTL, c.RefreshTTL)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("AUTH_GRPC_PORT", "nope")
	c := Load()
	if c.GRPCPort != 50051 {
		t.Fatal(c.GRPCPort)
	}
}

func TestLoad_CustomURLsAndRedis(t *testing.T) {
	t.Setenv("AUTH_HTTP_PORT", "9090")
	t.Setenv("STATIC_DIR", "/var/www")
	t.Setenv("CUSTOMERS_SERVICE_URL", "http://c:8081")
	t.Setenv("REDIS_ADDR", "10.0.0.1:6380")
	t.Setenv("REDIS_PASSWORD", "p")
	t.Setenv("REDIS_DB", "2")
	t.Setenv("JWT_ACCESS_TTL", "1h")
	t.Setenv("JWT_REFRESH_TTL", "24h")
	c := Load()
	if c.HTTPPort != 9090 || c.StaticDir != "/var/www" || c.CustomersServiceURL != "http://c:8081" {
		t.Fatal(c)
	}
	if c.RedisAddr != "10.0.0.1:6380" || c.RedisPass != "p" || c.RedisDB != 2 {
		t.Fatal(c.RedisAddr, c.RedisPass, c.RedisDB)
	}
	if c.AccessTTL != time.Hour || c.RefreshTTL != 24*time.Hour {
		t.Fatal(c.AccessTTL, c.RefreshTTL)
	}
}
