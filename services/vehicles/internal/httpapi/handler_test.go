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

	"github.com/dealer/dealer/services/vehicles/internal/domain"
	"github.com/dealer/dealer/services/vehicles/internal/jwt"
	"github.com/dealer/dealer/services/vehicles/internal/service"
)

const (
	testJWTSecret           = "s"
	testBearerUserID        = "u"
	testBearerEmail         = "e"
	httpAuthBearerSpace     = "Bearer "
	headerAuthorization     = "Authorization"
	testVehicleVIN          = "v"
	testVehicleMakeShort    = "m"
	testVehicleMakeGet      = "mk"
	testVehicleModelGet     = "md"
	testVehicleYear         = int32(1)
	testVehicleStatus       = "a"
	testVehDefaultStatus    = "available"
	testVehicleNotFoundUUID = "00000000-0000-0000-0000-000000000099"
)

func testVehicleFromCreateInput(in service.CreateVehicleInput) *domain.Vehicle {
	now := time.Now().UTC()
	st := in.Status
	if st == "" {
		st = testVehDefaultStatus
	}
	return &domain.Vehicle{
		ID: uuid.New(), VIN: in.VIN, Make: in.Make, Model: in.Model, Year: in.Year, MileageKm: in.MileageKm, Price: in.Price, Status: st,
		Color: in.Color, Notes: in.Notes, BrandID: in.BrandID, DealerPointID: in.DealerPointID, LegalEntityID: in.LegalEntityID, WarehouseID: in.WarehouseID,
		CreatedAt: now, UpdatedAt: now,
	}
}

type mockVehicle struct {
	listErr   error
	createErr error
	getErr    error
	nf        string
}

func (m *mockVehicle) Create(_ context.Context, in service.CreateVehicleInput) (*domain.Vehicle, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return testVehicleFromCreateInput(in), nil
}

func (m *mockVehicle) Get(_ context.Context, id string) (*domain.Vehicle, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.nf != "" && id == m.nf {
		return nil, service.ErrNotFound
	}
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	return &domain.Vehicle{ID: uid, VIN: testVehicleVIN, Make: testVehicleMakeGet, Model: testVehicleModelGet, Year: testVehicleYear, Status: testVehicleStatus, CreatedAt: now, UpdatedAt: now}, nil
}

func (m *mockVehicle) List(_ context.Context, _ domain.VehicleListFilter) ([]*domain.Vehicle, int32, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return []*domain.Vehicle{}, 0, nil
}

func (m *mockVehicle) Update(_ context.Context, id string, in service.UpdateVehicleInput) (*domain.Vehicle, error) {
	if m.nf != "" && id == m.nf {
		return nil, service.ErrNotFound
	}
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	v := &domain.Vehicle{ID: uid, VIN: testVehicleVIN, Make: testVehicleMakeShort, Model: testVehicleMakeShort, Year: testVehicleYear, Status: testVehicleStatus, CreatedAt: now, UpdatedAt: now}
	if in.Make != nil {
		v.Make = *in.Make
	}
	return v, nil
}

func (m *mockVehicle) Delete(_ context.Context, id string) error {
	if m.nf != "" && id == m.nf {
		return service.ErrNotFound
	}
	return nil
}

func bearerVeh(secret string) string {
	cl := &jwt.Claims{UserID: testBearerUserID, Email: testBearerEmail, RegisteredClaims: jwtlib.RegisteredClaims{
		ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
	}}
	s, _ := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, cl).SignedString([]byte(secret))
	return httpAuthBearerSpace + s
}

func vehiclesServeMux(m *mockVehicle) *http.ServeMux {
	h := NewHandler(m, testJWTSecret)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func vehiclesWantCode(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatal(w.Code)
	}
}

func vehiclesWantCodeBody(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatal(w.Code, w.Body.String())
	}
}

func testVehiclesHTTPStepOptions(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodOptions, pathAPIVehicles, nil))
	vehiclesWantCode(t, w, http.StatusNoContent)
}

func testVehiclesHTTPStepUnauth(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, pathAPIVehicles, nil))
	vehiclesWantCode(t, w, http.StatusUnauthorized)
}

func testVehiclesHTTPStepList(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, pathAPIVehicles, nil)
	req.Header.Set(headerAuthorization, bearerVeh(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	vehiclesWantCodeBody(t, w, http.StatusOK)
}

func testVehiclesHTTPStepCreateNoVIN(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, pathAPIVehicles, bytes.NewReader([]byte("{}")))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, bearerVeh(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	vehiclesWantCode(t, w, http.StatusBadRequest)
}

func testVehiclesHTTPStepCreateOK(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"vin": "VIN99", "make": "M", "model": "X", "year": 2020})
	req := httptest.NewRequest(http.MethodPost, pathAPIVehicles, bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, bearerVeh(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	vehiclesWantCodeBody(t, w, http.StatusOK)
}

func testVehiclesHTTPStepListErr(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, pathAPIVehicles, nil)
	req.Header.Set(headerAuthorization, bearerVeh(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	vehiclesWantCode(t, w, http.StatusInternalServerError)
}

func testVehiclesHTTPStepGetNF(t *testing.T, nf string) {
	t.Helper()
	mux := vehiclesServeMux(&mockVehicle{nf: nf})
	req := httptest.NewRequest(http.MethodGet, pathVehicleByID(nf), nil)
	req.Header.Set(headerAuthorization, bearerVeh(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	vehiclesWantCode(t, w, http.StatusNotFound)
}

func testVehiclesHTTPStepGetOK(t *testing.T, mux *http.ServeMux, id string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, pathVehicleByID(id), nil)
	req.Header.Set(headerAuthorization, bearerVeh(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	vehiclesWantCode(t, w, http.StatusOK)
}

func testVehiclesHTTPStepUpdateNF(t *testing.T, nf string) {
	t.Helper()
	mux := vehiclesServeMux(&mockVehicle{nf: nf})
	body, _ := json.Marshal(map[string]string{"make": "Z"})
	req := httptest.NewRequest(http.MethodPut, pathVehicleByID(nf), bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, bearerVeh(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	vehiclesWantCode(t, w, http.StatusNotFound)
}

func testVehiclesHTTPStepUpdatePutDelete(t *testing.T, mux *http.ServeMux, id string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"make": "Z"})
	req := httptest.NewRequest(http.MethodPut, pathVehicleByID(id), bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, bearerVeh(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	vehiclesWantCode(t, w, http.StatusOK)
	req2 := httptest.NewRequest(http.MethodDelete, pathVehicleByID(id), nil)
	req2.Header.Set(headerAuthorization, bearerVeh(testJWTSecret))
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	vehiclesWantCode(t, w2, http.StatusNoContent)
}

func testVehiclesHTTPStepCreateErr(t *testing.T) {
	t.Helper()
	mux := vehiclesServeMux(&mockVehicle{createErr: errors.New("db")})
	body, _ := json.Marshal(map[string]any{"vin": "V2", "make": "M"})
	req := httptest.NewRequest(http.MethodPost, pathAPIVehicles, bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, bearerVeh(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	vehiclesWantCode(t, w, http.StatusInternalServerError)
}

func testVehiclesHTTPStepGetInternal(t *testing.T) {
	t.Helper()
	mux := vehiclesServeMux(&mockVehicle{getErr: errors.New("db")})
	req := httptest.NewRequest(http.MethodGet, pathVehicleByID(uuid.New().String()), nil)
	req.Header.Set(headerAuthorization, bearerVeh(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	vehiclesWantCode(t, w, http.StatusInternalServerError)
}

func testVehiclesHTTPStepUpdateBadJSON(t *testing.T, mux *http.ServeMux, id string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, pathVehicleByID(id), bytes.NewReader([]byte("not-json")))
	setRequestJSONContentType(req)
	req.Header.Set(headerAuthorization, bearerVeh(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	vehiclesWantCode(t, w, http.StatusBadRequest)
}

func testVehiclesHTTPStepListQueryBrand(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	bid := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, pathAPIVehicles+"?brand_id="+bid+"&limit=5&offset=0", nil)
	req.Header.Set(headerAuthorization, bearerVeh(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	vehiclesWantCode(t, w, http.StatusOK)
}

func testVehiclesHTTPStepDeleteNF(t *testing.T, nf string) {
	t.Helper()
	mux := vehiclesServeMux(&mockVehicle{nf: nf})
	req := httptest.NewRequest(http.MethodDelete, pathVehicleByID(nf), nil)
	req.Header.Set(headerAuthorization, bearerVeh(testJWTSecret))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	vehiclesWantCode(t, w, http.StatusNotFound)
}

func TestVehiclesHTTP(t *testing.T) {
	mux := vehiclesServeMux(&mockVehicle{})
	testVehiclesHTTPStepOptions(t, mux)
	testVehiclesHTTPStepUnauth(t, mux)
	testVehiclesHTTPStepList(t, mux)
	testVehiclesHTTPStepCreateNoVIN(t, mux)
	testVehiclesHTTPStepCreateOK(t, mux)
	testVehiclesHTTPStepListErr(t, vehiclesServeMux(&mockVehicle{listErr: errors.New("db")}))
	nf := testVehicleNotFoundUUID
	testVehiclesHTTPStepGetNF(t, nf)
	id := uuid.New().String()
	testVehiclesHTTPStepGetOK(t, mux, id)
	testVehiclesHTTPStepUpdateNF(t, nf)
	testVehiclesHTTPStepUpdatePutDelete(t, mux, id)
	testVehiclesHTTPStepCreateErr(t)
	testVehiclesHTTPStepGetInternal(t)
	testVehiclesHTTPStepUpdateBadJSON(t, mux, id)
	testVehiclesHTTPStepListQueryBrand(t, mux)
	testVehiclesHTTPStepDeleteNF(t, nf)
}
