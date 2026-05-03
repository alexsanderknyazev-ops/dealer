package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/dealer/dealer/services/brands/internal/domain"
	"github.com/dealer/dealer/services/brands/internal/jwt"
	"github.com/dealer/dealer/services/brands/internal/service"
)

const (
	testJWTSecret          = "s"
	testBearerUserID       = "u"
	testBearerEmail        = "e"
	httpAuthBearerSpace    = "Bearer "
	testBrandNameFromGet   = "B"
	testBrandNameDefault   = "x"
)

type mockBrand struct {
	list      []*domain.Brand
	total     int32
	listErr   error
	createErr error
	getErr    error
	nf        string
}

func (m *mockBrand) Create(_ context.Context, name string) (*domain.Brand, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	now := time.Now().UTC()
	return &domain.Brand{ID: uuid.New(), Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

func (m *mockBrand) Get(_ context.Context, id string) (*domain.Brand, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.nf != "" && id == m.nf {
		return nil, service.ErrNotFound
	}
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	return &domain.Brand{ID: uid, Name: testBrandNameFromGet, CreatedAt: now, UpdatedAt: now}, nil
}

func (m *mockBrand) List(_ context.Context, _, _ int32, _ string) ([]*domain.Brand, int32, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.list, m.total, nil
}

func (m *mockBrand) Update(_ context.Context, id string, name *string) (*domain.Brand, error) {
	if m.nf != "" && id == m.nf {
		return nil, service.ErrNotFound
	}
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	b := &domain.Brand{ID: uid, Name: testBrandNameDefault, CreatedAt: now, UpdatedAt: now}
	if name != nil {
		b.Name = *name
	}
	return b, nil
}

func (m *mockBrand) Delete(_ context.Context, id string) error {
	if m.nf != "" && id == m.nf {
		return service.ErrNotFound
	}
	return nil
}

func tok(secret string) string {
	cl := &jwt.Claims{UserID: testBearerUserID, Email: testBearerEmail, RegisteredClaims: jwtlib.RegisteredClaims{
		ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
	}}
	s, _ := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, cl).SignedString([]byte(secret))
	return httpAuthBearerSpace + s
}

func TestBrandsHTTP(t *testing.T) {
	h := NewHandler(&mockBrand{list: []*domain.Brand{}, total: 0}, testJWTSecret)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	t.Run("options", func(t *testing.T) {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodOptions, pathAPIBrands, nil))
		if w.Code != http.StatusNoContent {
			t.Fatal(w.Code)
		}
	})
	t.Run("unauth", func(t *testing.T) {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, pathAPIBrands, nil))
		if w.Code != http.StatusUnauthorized {
			t.Fatal(w.Code)
		}
	})
	t.Run("list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, pathAPIBrands, nil)
		req.Header.Set("Authorization", tok(testJWTSecret))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code, w.Body.String())
		}
	})
	t.Run("create_bad", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, pathAPIBrands, bytes.NewReader([]byte("{}")))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", tok(testJWTSecret))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatal(w.Code)
		}
	})
	t.Run("create_ok", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"name": "  X  "})
		req := httptest.NewRequest(http.MethodPost, pathAPIBrands, bytes.NewReader(body))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", tok(testJWTSecret))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code, w.Body.String())
		}
	})
	t.Run("list_err", func(t *testing.T) {
		h2 := NewHandler(&mockBrand{listErr: errors.New("db")}, testJWTSecret)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodGet, pathAPIBrands, nil)
		req.Header.Set("Authorization", tok(testJWTSecret))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatal(w.Code)
		}
	})
	nf := "00000000-0000-0000-0000-000000000077"
	t.Run("get_nf", func(t *testing.T) {
		h2 := NewHandler(&mockBrand{nf: nf}, testJWTSecret)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodGet, pathBrandByID(nf), nil)
		req.Header.Set("Authorization", tok(testJWTSecret))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatal(w.Code)
		}
	})
	id := uuid.New().String()
	t.Run("get_ok", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, pathBrandByID(id), nil)
		req.Header.Set("Authorization", tok(testJWTSecret))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code)
		}
	})
	t.Run("update_nf", func(t *testing.T) {
		h2 := NewHandler(&mockBrand{nf: nf}, testJWTSecret)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		body, _ := json.Marshal(map[string]string{"name": "Z"})
		req := httptest.NewRequest(http.MethodPut, pathBrandByID(nf), bytes.NewReader(body))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", tok(testJWTSecret))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatal(w.Code)
		}
	})
	t.Run("delete_nf", func(t *testing.T) {
		h2 := NewHandler(&mockBrand{nf: nf}, testJWTSecret)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodDelete, pathBrandByID(nf), nil)
		req.Header.Set("Authorization", tok(testJWTSecret))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatal(w.Code)
		}
	})
	t.Run("create_svc_err", func(t *testing.T) {
		h2 := NewHandler(&mockBrand{createErr: errors.New("db")}, testJWTSecret)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		body, _ := json.Marshal(map[string]string{"name": "A"})
		req := httptest.NewRequest(http.MethodPost, pathAPIBrands, bytes.NewReader(body))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", tok(testJWTSecret))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatal(w.Code)
		}
	})
	t.Run("update_ok_delete_ok", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"name": "Z"})
		req := httptest.NewRequest(http.MethodPut, pathBrandByID(id), bytes.NewReader(body))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", tok(testJWTSecret))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code)
		}
		req2 := httptest.NewRequest(http.MethodDelete, pathBrandByID(id), nil)
		req2.Header.Set("Authorization", tok(testJWTSecret))
		w2 := httptest.NewRecorder()
		mux.ServeHTTP(w2, req2)
		if w2.Code != http.StatusNoContent {
			t.Fatal(w2.Code)
		}
	})
	t.Run("get_internal", func(t *testing.T) {
		h2 := NewHandler(&mockBrand{getErr: errors.New("db")}, testJWTSecret)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodGet, pathBrandByID(uuid.New().String()), nil)
		req.Header.Set("Authorization", tok(testJWTSecret))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatal(w.Code)
		}
	})
}
