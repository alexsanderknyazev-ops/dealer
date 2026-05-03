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
	testJWTSecret          = "s"
	testBearerUserID       = "u"
	testBearerEmail        = "e"
	httpAuthBearerSpace    = "Bearer "
	testVehicleVIN         = "v"
	testVehicleMakeShort   = "m"
	testVehicleMakeGet     = "mk"
	testVehicleModelGet    = "md"
	testVehicleYear        = int32(1)
	testVehicleStatus      = "a"
	testVehDefaultStatus    = "available"
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

func TestVehiclesHTTP(t *testing.T) {
	h := NewHandler(&mockVehicle{}, testJWTSecret)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	t.Run("options", func(t *testing.T) {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodOptions, pathAPIVehicles, nil))
		if w.Code != http.StatusNoContent {
			t.Fatal(w.Code)
		}
	})
	t.Run("unauth", func(t *testing.T) {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, pathAPIVehicles, nil))
		if w.Code != http.StatusUnauthorized {
			t.Fatal(w.Code)
		}
	})
	t.Run("list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, pathAPIVehicles, nil)
		req.Header.Set("Authorization", bearerVeh(testJWTSecret))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code, w.Body.String())
		}
	})
	t.Run("create_no_vin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, pathAPIVehicles, bytes.NewReader([]byte("{}")))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", bearerVeh(testJWTSecret))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatal(w.Code)
		}
	})
	t.Run("create_ok", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{"vin": "VIN99", "make": "M", "model": "X", "year": 2020})
		req := httptest.NewRequest(http.MethodPost, pathAPIVehicles, bytes.NewReader(body))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", bearerVeh(testJWTSecret))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code, w.Body.String())
		}
	})
	t.Run("list_err", func(t *testing.T) {
		h2 := NewHandler(&mockVehicle{listErr: errors.New("db")}, testJWTSecret)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodGet, pathAPIVehicles, nil)
		req.Header.Set("Authorization", bearerVeh(testJWTSecret))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatal(w.Code)
		}
	})
	nf := "00000000-0000-0000-0000-000000000099"
	t.Run("get_nf", func(t *testing.T) {
		h2 := NewHandler(&mockVehicle{nf: nf}, testJWTSecret)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodGet, pathVehicleByID(nf), nil)
		req.Header.Set("Authorization", bearerVeh(testJWTSecret))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatal(w.Code)
		}
	})
	id := uuid.New().String()
	t.Run("get_ok", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, pathVehicleByID(id), nil)
		req.Header.Set("Authorization", bearerVeh(testJWTSecret))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code)
		}
	})
	t.Run("update_nf", func(t *testing.T) {
		h2 := NewHandler(&mockVehicle{nf: nf}, testJWTSecret)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		body, _ := json.Marshal(map[string]string{"make": "Z"})
		req := httptest.NewRequest(http.MethodPut, pathVehicleByID(nf), bytes.NewReader(body))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", bearerVeh(testJWTSecret))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatal(w.Code)
		}
	})
	t.Run("update_put_delete", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"make": "Z"})
		req := httptest.NewRequest(http.MethodPut, pathVehicleByID(id), bytes.NewReader(body))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", bearerVeh(testJWTSecret))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code)
		}
		req2 := httptest.NewRequest(http.MethodDelete, pathVehicleByID(id), nil)
		req2.Header.Set("Authorization", bearerVeh(testJWTSecret))
		w2 := httptest.NewRecorder()
		mux.ServeHTTP(w2, req2)
		if w2.Code != http.StatusNoContent {
			t.Fatal(w2.Code)
		}
	})
	t.Run("create_err", func(t *testing.T) {
		h2 := NewHandler(&mockVehicle{createErr: errors.New("db")}, testJWTSecret)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		body, _ := json.Marshal(map[string]any{"vin": "V2", "make": "M"})
		req := httptest.NewRequest(http.MethodPost, pathAPIVehicles, bytes.NewReader(body))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", bearerVeh(testJWTSecret))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatal(w.Code)
		}
	})
	t.Run("get_internal", func(t *testing.T) {
		h2 := NewHandler(&mockVehicle{getErr: errors.New("db")}, testJWTSecret)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodGet, pathVehicleByID(uuid.New().String()), nil)
		req.Header.Set("Authorization", bearerVeh(testJWTSecret))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatal(w.Code)
		}
	})
	t.Run("update_bad_json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, pathVehicleByID(id), bytes.NewReader([]byte("not-json")))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", bearerVeh(testJWTSecret))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatal(w.Code)
		}
	})
	t.Run("list_query_brand", func(t *testing.T) {
		bid := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, pathAPIVehicles+"?brand_id="+bid+"&limit=5&offset=0", nil)
		req.Header.Set("Authorization", bearerVeh(testJWTSecret))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code)
		}
	})
	t.Run("delete_nf", func(t *testing.T) {
		h2 := NewHandler(&mockVehicle{nf: nf}, testJWTSecret)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodDelete, pathVehicleByID(nf), nil)
		req.Header.Set("Authorization", bearerVeh(testJWTSecret))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatal(w.Code)
		}
	})
}
