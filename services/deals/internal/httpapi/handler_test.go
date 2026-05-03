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

	"github.com/dealer/dealer/services/deals/internal/domain"
	"github.com/dealer/dealer/services/deals/internal/jwt"
	"github.com/dealer/dealer/services/deals/internal/service"
)

const (
	testJWTSecret        = "s"
	testBearerUserID     = "u"
	testBearerEmail      = "e"
	httpAuthBearerSpace  = "Bearer "
	headerAuthorization  = "Authorization"
	testDealAmountStub   = "1"
	testDealStageStub    = "d"
	testDealNotFoundUUID = "00000000-0000-0000-0000-000000000088"
)

type mockDeal struct {
	listErr   error
	createErr error
	getErr    error
	nf        string
}

func (m *mockDeal) Create(_ context.Context, in service.CreateDealInput) (*domain.Deal, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	now := time.Now().UTC()
	cid, _ := uuid.Parse(in.CustomerID)
	vid, _ := uuid.Parse(in.VehicleID)
	return &domain.Deal{ID: uuid.New(), CustomerID: cid, VehicleID: vid, Amount: in.Amount, Stage: in.Stage, Notes: in.Notes, CreatedAt: now, UpdatedAt: now}, nil
}

func (m *mockDeal) Get(_ context.Context, id string) (*domain.Deal, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.nf != "" && id == m.nf {
		return nil, service.ErrNotFound
	}
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	cid, vid := uuid.New(), uuid.New()
	return &domain.Deal{ID: uid, CustomerID: cid, VehicleID: vid, Amount: testDealAmountStub, Stage: testDealStageStub, CreatedAt: now, UpdatedAt: now}, nil
}

func (m *mockDeal) List(_ context.Context, _, _ int32, _, _ string) ([]*domain.Deal, int32, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return []*domain.Deal{}, 0, nil
}

func (m *mockDeal) Update(_ context.Context, id string, in service.UpdateDealInput) (*domain.Deal, error) {
	if m.nf != "" && id == m.nf {
		return nil, service.ErrNotFound
	}
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	cid, vid := uuid.New(), uuid.New()
	d := &domain.Deal{ID: uid, CustomerID: cid, VehicleID: vid, Amount: testDealAmountStub, Stage: testDealStageStub, CreatedAt: now, UpdatedAt: now}
	if in.Amount != nil {
		d.Amount = *in.Amount
	}
	return d, nil
}

func (m *mockDeal) Delete(_ context.Context, id string) error {
	if m.nf != "" && id == m.nf {
		return service.ErrNotFound
	}
	return nil
}

func bearerDeal(secret string) string {
	cl := &jwt.Claims{UserID: testBearerUserID, Email: testBearerEmail, RegisteredClaims: jwtlib.RegisteredClaims{
		ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
	}}
	s, _ := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, cl).SignedString([]byte(secret))
	return httpAuthBearerSpace + s
}

func dealsServeMux(m *mockDeal) *http.ServeMux {
	h := NewHandler(m, testJWTSecret)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func dealsWantCode(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatal(w.Code)
	}
}

func dealsWantCodeBody(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatal(w.Code, w.Body.String())
	}
}

func testDealsHTTPStepOptions(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodOptions, pathAPIDeals, nil))
	dealsWantCode(t, w, http.StatusNoContent)
}

func testDealsHTTPStepUnauth(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, pathAPIDeals, nil))
	dealsWantCode(t, w, http.StatusUnauthorized)
}

func testDealsHTTPStepList(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, pathAPIDeals, nil)
	req.Header.Set(headerAuthorization, bearerDeal(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	dealsWantCode(t, w, http.StatusOK)
}

func testDealsHTTPStepCreateMissingIDs(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, pathAPIDeals, bytes.NewReader([]byte("{}")))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, bearerDeal(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	dealsWantCode(t, w, http.StatusBadRequest)
}

func testDealsHTTPStepCreateOK(t *testing.T, mux *http.ServeMux, cid, vid string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"customer_id": cid, "vehicle_id": vid, "amount": "10"})
	req := httptest.NewRequest(http.MethodPost, pathAPIDeals, bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, bearerDeal(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	dealsWantCodeBody(t, w, http.StatusOK)
}

func testDealsHTTPStepListErr(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, pathAPIDeals, nil)
	req.Header.Set(headerAuthorization, bearerDeal(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	dealsWantCode(t, w, http.StatusInternalServerError)
}

func testDealsHTTPStepGetNF(t *testing.T, nf string) {
	t.Helper()
	mux := dealsServeMux(&mockDeal{nf: nf})
	req := httptest.NewRequest(http.MethodGet, pathDealByID(nf), nil)
	req.Header.Set(headerAuthorization, bearerDeal(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	dealsWantCode(t, w, http.StatusNotFound)
}

func testDealsHTTPStepPutDel(t *testing.T, mux *http.ServeMux, id string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"amount": "99"})
	req := httptest.NewRequest(http.MethodPut, pathDealByID(id), bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, bearerDeal(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	dealsWantCode(t, w, http.StatusOK)
	req2 := httptest.NewRequest(http.MethodDelete, pathDealByID(id), nil)
	req2.Header.Set(headerAuthorization, bearerDeal(testJWTSecret))
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	dealsWantCode(t, w2, http.StatusNoContent)
}

func testDealsHTTPStepCreateErr(t *testing.T) {
	t.Helper()
	mux := dealsServeMux(&mockDeal{createErr: errors.New("db")})
	cid, vid := uuid.New().String(), uuid.New().String()
	body, _ := json.Marshal(map[string]string{"customer_id": cid, "vehicle_id": vid})
	req := httptest.NewRequest(http.MethodPost, pathAPIDeals, bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, bearerDeal(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	dealsWantCode(t, w, http.StatusBadRequest)
}

func testDealsHTTPStepGetInternal(t *testing.T) {
	t.Helper()
	mux := dealsServeMux(&mockDeal{getErr: errors.New("db")})
	req := httptest.NewRequest(http.MethodGet, pathDealByID(uuid.New().String()), nil)
	req.Header.Set(headerAuthorization, bearerDeal(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	dealsWantCode(t, w, http.StatusInternalServerError)
}

func testDealsHTTPStepPutBadJSON(t *testing.T, mux *http.ServeMux, id string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, pathDealByID(id), bytes.NewReader([]byte("x")))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, bearerDeal(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	dealsWantCode(t, w, http.StatusBadRequest)
}

func testDealsHTTPStepDeleteNF(t *testing.T, nf string) {
	t.Helper()
	mux := dealsServeMux(&mockDeal{nf: nf})
	req := httptest.NewRequest(http.MethodDelete, pathDealByID(nf), nil)
	req.Header.Set(headerAuthorization, bearerDeal(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	dealsWantCode(t, w, http.StatusNotFound)
}

func TestDealsHTTP(t *testing.T) {
	mux := dealsServeMux(&mockDeal{})
	testDealsHTTPStepOptions(t, mux)
	testDealsHTTPStepUnauth(t, mux)
	testDealsHTTPStepList(t, mux)
	testDealsHTTPStepCreateMissingIDs(t, mux)
	cid, vid := uuid.New().String(), uuid.New().String()
	testDealsHTTPStepCreateOK(t, mux, cid, vid)
	testDealsHTTPStepListErr(t, dealsServeMux(&mockDeal{listErr: errors.New("db")}))
	nf := testDealNotFoundUUID
	testDealsHTTPStepGetNF(t, nf)
	id := uuid.New().String()
	testDealsHTTPStepPutDel(t, mux, id)
	testDealsHTTPStepCreateErr(t)
	testDealsHTTPStepGetInternal(t)
	testDealsHTTPStepPutBadJSON(t, mux, id)
	testDealsHTTPStepDeleteNF(t, nf)
}
