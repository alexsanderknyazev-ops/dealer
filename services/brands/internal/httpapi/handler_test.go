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
	testJWTSecret         = "s"
	testBearerUserID      = "u"
	testBearerEmail       = "e"
	httpAuthBearerSpace   = "Bearer "
	headerAuthorization   = "Authorization"
	testBrandNameFromGet  = "B"
	testBrandNameDefault  = "x"
	testBrandNotFoundUUID = "00000000-0000-0000-0000-000000000077"
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

func brandsServeMux(m *mockBrand) *http.ServeMux {
	h := NewHandler(m, testJWTSecret)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func brandsWantCode(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatal(w.Code)
	}
}

func brandsWantCodeBody(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatal(w.Code, w.Body.String())
	}
}

func testBrandsHTTPStepOptions(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodOptions, pathAPIBrands, nil))
	brandsWantCode(t, w, http.StatusNoContent)
}

func testBrandsHTTPStepUnauth(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, pathAPIBrands, nil))
	brandsWantCode(t, w, http.StatusUnauthorized)
}

func testBrandsHTTPStepList(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, pathAPIBrands, nil)
	req.Header.Set(headerAuthorization, tok(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	brandsWantCodeBody(t, w, http.StatusOK)
}

func testBrandsHTTPStepCreateBad(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, pathAPIBrands, bytes.NewReader([]byte("{}")))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, tok(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	brandsWantCode(t, w, http.StatusBadRequest)
}

func testBrandsHTTPStepCreateOK(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"name": "  X  "})
	req := httptest.NewRequest(http.MethodPost, pathAPIBrands, bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, tok(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	brandsWantCodeBody(t, w, http.StatusOK)
}

func testBrandsHTTPStepListErr(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, pathAPIBrands, nil)
	req.Header.Set(headerAuthorization, tok(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	brandsWantCode(t, w, http.StatusInternalServerError)
}

func testBrandsHTTPStepGetNF(t *testing.T, nf string) {
	t.Helper()
	mux := brandsServeMux(&mockBrand{nf: nf})
	req := httptest.NewRequest(http.MethodGet, pathBrandByID(nf), nil)
	req.Header.Set(headerAuthorization, tok(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	brandsWantCode(t, w, http.StatusNotFound)
}

func testBrandsHTTPStepGetOK(t *testing.T, mux *http.ServeMux, id string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, pathBrandByID(id), nil)
	req.Header.Set(headerAuthorization, tok(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	brandsWantCode(t, w, http.StatusOK)
}

func testBrandsHTTPStepUpdateNF(t *testing.T, nf string) {
	t.Helper()
	mux := brandsServeMux(&mockBrand{nf: nf})
	body, _ := json.Marshal(map[string]string{"name": "Z"})
	req := httptest.NewRequest(http.MethodPut, pathBrandByID(nf), bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, tok(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	brandsWantCode(t, w, http.StatusNotFound)
}

func testBrandsHTTPStepDeleteNF(t *testing.T, nf string) {
	t.Helper()
	mux := brandsServeMux(&mockBrand{nf: nf})
	req := httptest.NewRequest(http.MethodDelete, pathBrandByID(nf), nil)
	req.Header.Set(headerAuthorization, tok(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	brandsWantCode(t, w, http.StatusNotFound)
}

func testBrandsHTTPStepCreateSvcErr(t *testing.T) {
	t.Helper()
	mux := brandsServeMux(&mockBrand{createErr: errors.New("db")})
	body, _ := json.Marshal(map[string]string{"name": "A"})
	req := httptest.NewRequest(http.MethodPost, pathAPIBrands, bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, tok(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	brandsWantCode(t, w, http.StatusInternalServerError)
}

func testBrandsHTTPStepUpdateOKDeleteOK(t *testing.T, mux *http.ServeMux, id string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"name": "Z"})
	req := httptest.NewRequest(http.MethodPut, pathBrandByID(id), bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, tok(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	brandsWantCode(t, w, http.StatusOK)
	req2 := httptest.NewRequest(http.MethodDelete, pathBrandByID(id), nil)
	req2.Header.Set(headerAuthorization, tok(testJWTSecret))
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	brandsWantCode(t, w2, http.StatusNoContent)
}

func testBrandsHTTPStepGetInternal(t *testing.T) {
	t.Helper()
	mux := brandsServeMux(&mockBrand{getErr: errors.New("db")})
	req := httptest.NewRequest(http.MethodGet, pathBrandByID(uuid.New().String()), nil)
	req.Header.Set(headerAuthorization, tok(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	brandsWantCode(t, w, http.StatusInternalServerError)
}

func TestBrandsHTTP_listAndCreate(t *testing.T) {
	mux := brandsServeMux(&mockBrand{list: []*domain.Brand{}, total: 0})
	testBrandsHTTPStepOptions(t, mux)
	testBrandsHTTPStepUnauth(t, mux)
	testBrandsHTTPStepList(t, mux)
	testBrandsHTTPStepCreateBad(t, mux)
	testBrandsHTTPStepCreateOK(t, mux)
}

func TestBrandsHTTP_listInternalError(t *testing.T) {
	testBrandsHTTPStepListErr(t, brandsServeMux(&mockBrand{listErr: errors.New("db")}))
}

func TestBrandsHTTP_notFoundAndMutations(t *testing.T) {
	mux := brandsServeMux(&mockBrand{list: []*domain.Brand{}, total: 0})
	nf := testBrandNotFoundUUID
	testBrandsHTTPStepGetNF(t, nf)
	id := uuid.New().String()
	testBrandsHTTPStepGetOK(t, mux, id)
	testBrandsHTTPStepUpdateNF(t, nf)
	testBrandsHTTPStepDeleteNF(t, nf)
	testBrandsHTTPStepCreateSvcErr(t)
	testBrandsHTTPStepUpdateOKDeleteOK(t, mux, id)
	testBrandsHTTPStepGetInternal(t)
}
