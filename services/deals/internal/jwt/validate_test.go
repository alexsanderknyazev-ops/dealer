package jwt

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

const (
	testJWTArgShort   = "s"
	testJWTSignedWith = "sec"
	testJWTWrongKey   = "wrong"
	testClaimUserID   = "1"
	testClaimEmail    = "e@e"
)

func TestValidate(t *testing.T) {
	_, _, err := Validate(testJWTArgShort, "")
	if err == nil {
		t.Fatal("want err")
	}
	claims := &Claims{UserID: testClaimUserID, Email: testClaimEmail, RegisteredClaims: jwtlib.RegisteredClaims{
		ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
	}}
	tok, _ := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString([]byte(testJWTSignedWith))
	u, e, err := Validate(testJWTSignedWith, tok)
	if err != nil || u != testClaimUserID || e != testClaimEmail {
		t.Fatal(err)
	}
	bad, _ := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString([]byte(testJWTWrongKey))
	if _, _, err := Validate(testJWTSignedWith, bad); err == nil {
		t.Fatal("want err")
	}
}
