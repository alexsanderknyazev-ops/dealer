package jwt

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

const (
	testHMACSecret       = "s3cr3t"
	testHMACWrongSecret  = "wrong"
	testHMACSignOnlyGood = "good"
	testUserID           = "u1"
	testUserEmail        = "a@b.c"
	testUserRole         = "sales"
	testValidateSecret   = "secret"
)

func TestValidate_Missing(t *testing.T) {
	_, _, _, err := Validate(testValidateSecret, "")
	if err == nil {
		t.Fatal("want err")
	}
}

func TestValidate_OK(t *testing.T) {
	claims := &Claims{
		UserID: testUserID,
		Email:  testUserEmail,
		Role:   testUserRole,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
	}
	tok := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(testHMACSecret))
	if err != nil {
		t.Fatal(err)
	}
	uid, email, role, err := Validate(testHMACSecret, s)
	if err != nil || uid != testUserID || email != testUserEmail || role != testUserRole {
		t.Fatalf("uid=%q email=%q role=%q err=%v", uid, email, role, err)
	}
}

func TestValidate_BadSecret(t *testing.T) {
	claims := &Claims{
		UserID: testUserID,
		Email:  testUserEmail,
		Role:   testUserRole,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	s, _ := tok.SignedString([]byte(testHMACSignOnlyGood))
	_, _, _, err := Validate(testHMACWrongSecret, s)
	if err == nil {
		t.Fatal("want err")
	}
}
