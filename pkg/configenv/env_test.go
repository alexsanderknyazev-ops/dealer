package configenv

import (
	"testing"
	"time"
)

func TestString(t *testing.T) {
	t.Setenv("CFG_TEST_STR", "")
	if String("CFG_TEST_STR", "def") != "def" {
		t.Fatal()
	}
	t.Setenv("CFG_TEST_STR", "val")
	if String("CFG_TEST_STR", "def") != "val" {
		t.Fatal()
	}
}

func TestInt(t *testing.T) {
	t.Setenv("CFG_TEST_INT", "bad")
	if Int("CFG_TEST_INT", 42) != 42 {
		t.Fatal()
	}
	t.Setenv("CFG_TEST_INT", "7")
	if Int("CFG_TEST_INT", 42) != 7 {
		t.Fatal()
	}
}

func TestDuration(t *testing.T) {
	t.Setenv("CFG_TEST_DUR", "1h")
	if Duration("CFG_TEST_DUR", time.Minute) != time.Hour {
		t.Fatal()
	}
}

func TestSplitCSV(t *testing.T) {
	got := SplitCSV(" a, b ,,c ")
	if len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("%v", got)
	}
}

func TestLoadPostgresJWT(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://x")
	t.Setenv("JWT_SECRET", "sec")
	pj := LoadPostgresJWT()
	if pj.PostgresDSN != "postgres://x" || pj.JWTSecret != "sec" {
		t.Fatalf("%+v", pj)
	}
}
