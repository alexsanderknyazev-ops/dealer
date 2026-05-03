package jwt

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

func TestValidate(t *testing.T) {
	_, _, err := Validate("s", "")
	if err == nil {
		t.Fatal("want err")
	}
	claims := &Claims{UserID: "1", Email: "e@e", RegisteredClaims: jwtlib.RegisteredClaims{
		ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
	}}
	tok, _ := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString([]byte("sec"))
	u, e, err := Validate("sec", tok)
	if err != nil || u != "1" || e != "e@e" {
		t.Fatal(err)
	}
}
