package jwt

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

func TestValidate_Missing(t *testing.T) {
	_, _, err := Validate("secret", "")
	if err == nil {
		t.Fatal("want err")
	}
}

func TestValidate_OK(t *testing.T) {
	claims := &Claims{
		UserID: "u1",
		Email:  "a@b.c",
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
	}
	tok := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte("s3cr3t"))
	if err != nil {
		t.Fatal(err)
	}
	uid, email, err := Validate("s3cr3t", s)
	if err != nil || uid != "u1" || email != "a@b.c" {
		t.Fatalf("uid=%q email=%q err=%v", uid, email, err)
	}
}

func TestValidate_BadSecret(t *testing.T) {
	claims := &Claims{
		UserID: "u1",
		Email:  "a@b.c",
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	s, _ := tok.SignedString([]byte("good"))
	_, _, err := Validate("wrong", s)
	if err == nil {
		t.Fatal("want err")
	}
}
