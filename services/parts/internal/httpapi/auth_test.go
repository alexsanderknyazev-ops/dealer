package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/dealer/dealer/services/parts/internal/jwt"
)

const (
	testPartsJWTSecret = "parts-secret"
	testPartsUserID    = "user-1"
	testPartsEmail     = "parts@dealer.local"
	testPartsPath      = "/api/parts"
)

func partsBearer(role string) string {
	claims := &jwt.Claims{
		UserID: testPartsUserID,
		Email:  testPartsEmail,
		Role:   role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	signed, _ := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString([]byte(testPartsJWTSecret))
	return "Bearer " + signed
}

func TestPartsAuthMiddleware_WriteRequiresRole(t *testing.T) {
	h := &Handler{jwtSecret: testPartsJWTSecret}
	next := h.auth(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	req := httptest.NewRequest(http.MethodPost, testPartsPath, nil)
	req.Header.Set("Authorization", partsBearer("viewer"))
	w := httptest.NewRecorder()
	next(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want %d got %d", http.StatusForbidden, w.Code)
	}
}

func TestPartsAuthMiddleware_WriteAllowedRole(t *testing.T) {
	h := &Handler{jwtSecret: testPartsJWTSecret}
	next := h.auth(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	req := httptest.NewRequest(http.MethodPost, testPartsPath, nil)
	req.Header.Set("Authorization", partsBearer("parts_manager"))
	w := httptest.NewRecorder()
	next(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want %d got %d", http.StatusOK, w.Code)
	}
}

func TestPartsRoutes_WriteForbiddenForViewer(t *testing.T) {
	h := NewHandler(nil, testPartsJWTSecret)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: testPartsPath},
		{method: http.MethodPut, path: "/api/parts/00000000-0000-0000-0000-000000000001"},
		{method: http.MethodDelete, path: "/api/parts/00000000-0000-0000-0000-000000000001"},
		{method: http.MethodPost, path: "/api/parts/folders"},
		{method: http.MethodPut, path: "/api/parts/folders/00000000-0000-0000-0000-000000000001"},
		{method: http.MethodDelete, path: "/api/parts/folders/00000000-0000-0000-0000-000000000001"},
	}
	for _, tc := range tests {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		req.Header.Set("Authorization", partsBearer("viewer"))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("%s %s: want %d got %d", tc.method, tc.path, http.StatusForbidden, w.Code)
		}
	}
}
