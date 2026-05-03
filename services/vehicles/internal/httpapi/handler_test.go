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

type mockVehicle struct {
	listErr   error
	createErr error
	getErr    error
	nf        string
}

func (m *mockVehicle) Create(_ context.Context, vin, make, model string, year int32, mileageKm int64, price, status, color, notes string, brandID, dealerPointID, legalEntityID, warehouseID *uuid.UUID) (*domain.Vehicle, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	now := time.Now().UTC()
	return &domain.Vehicle{
		ID: uuid.New(), VIN: vin, Make: make, Model: model, Year: year, MileageKm: mileageKm, Price: price, Status: status,
		Color: color, Notes: notes, BrandID: brandID, DealerPointID: dealerPointID, LegalEntityID: legalEntityID, WarehouseID: warehouseID,
		CreatedAt: now, UpdatedAt: now,
	}, nil
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
	return &domain.Vehicle{ID: uid, VIN: "v", Make: "mk", Model: "md", Year: 1, Status: "a", CreatedAt: now, UpdatedAt: now}, nil
}

func (m *mockVehicle) List(_ context.Context, _, _ int32, _, _ string, _, _, _, _ *uuid.UUID) ([]*domain.Vehicle, int32, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return []*domain.Vehicle{}, 0, nil
}

func (m *mockVehicle) Update(_ context.Context, id string, vin, make, model *string, year *int32, mileageKm *int64, price, status, color, notes *string, brandID *uuid.UUID, clearBrand bool, dealerPointID, legalEntityID, warehouseID *uuid.UUID, clearDealerPoint, clearLegalEntity, clearWarehouse bool) (*domain.Vehicle, error) {
	if m.nf != "" && id == m.nf {
		return nil, service.ErrNotFound
	}
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	v := &domain.Vehicle{ID: uid, VIN: "v", Make: "m", Model: "m", Year: 1, Status: "a", CreatedAt: now, UpdatedAt: now}
	if make != nil {
		v.Make = *make
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
	cl := &jwt.Claims{UserID: "u", Email: "e", RegisteredClaims: jwtlib.RegisteredClaims{
		ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
	}}
	s, _ := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, cl).SignedString([]byte(secret))
	return "Bearer " + s
}

func TestVehiclesHTTP(t *testing.T) {
	sec := "s"
	h := NewHandler(&mockVehicle{}, sec)
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
		req.Header.Set("Authorization", bearerVeh(sec))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code, w.Body.String())
		}
	})
	t.Run("create_no_vin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, pathAPIVehicles, bytes.NewReader([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", bearerVeh(sec))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatal(w.Code)
		}
	})
	t.Run("create_ok", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{"vin": "VIN99", "make": "M", "model": "X", "year": 2020})
		req := httptest.NewRequest(http.MethodPost, pathAPIVehicles, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", bearerVeh(sec))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code, w.Body.String())
		}
	})
	t.Run("list_err", func(t *testing.T) {
		h2 := NewHandler(&mockVehicle{listErr: errors.New("db")}, sec)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodGet, pathAPIVehicles, nil)
		req.Header.Set("Authorization", bearerVeh(sec))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatal(w.Code)
		}
	})
	nf := "00000000-0000-0000-0000-000000000099"
	t.Run("get_nf", func(t *testing.T) {
		h2 := NewHandler(&mockVehicle{nf: nf}, sec)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodGet, pathAPIVehicles+"/"+nf, nil)
		req.Header.Set("Authorization", bearerVeh(sec))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatal(w.Code)
		}
	})
	id := uuid.New().String()
	t.Run("get_ok", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, pathAPIVehicles+"/"+id, nil)
		req.Header.Set("Authorization", bearerVeh(sec))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code)
		}
	})
	t.Run("update_nf", func(t *testing.T) {
		h2 := NewHandler(&mockVehicle{nf: nf}, sec)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		body, _ := json.Marshal(map[string]string{"make": "Z"})
		req := httptest.NewRequest(http.MethodPut, pathAPIVehicles+"/"+nf, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", bearerVeh(sec))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatal(w.Code)
		}
	})
	t.Run("update_put_delete", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"make": "Z"})
		req := httptest.NewRequest(http.MethodPut, pathAPIVehicles+"/"+id, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", bearerVeh(sec))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code)
		}
		req2 := httptest.NewRequest(http.MethodDelete, pathAPIVehicles+"/"+id, nil)
		req2.Header.Set("Authorization", bearerVeh(sec))
		w2 := httptest.NewRecorder()
		mux.ServeHTTP(w2, req2)
		if w2.Code != http.StatusNoContent {
			t.Fatal(w2.Code)
		}
	})
	t.Run("create_err", func(t *testing.T) {
		h2 := NewHandler(&mockVehicle{createErr: errors.New("db")}, sec)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		body, _ := json.Marshal(map[string]any{"vin": "V2", "make": "M"})
		req := httptest.NewRequest(http.MethodPost, pathAPIVehicles, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", bearerVeh(sec))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatal(w.Code)
		}
	})
	t.Run("get_internal", func(t *testing.T) {
		h2 := NewHandler(&mockVehicle{getErr: errors.New("db")}, sec)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodGet, pathAPIVehicles+"/"+uuid.New().String(), nil)
		req.Header.Set("Authorization", bearerVeh(sec))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatal(w.Code)
		}
	})
	t.Run("update_bad_json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, pathAPIVehicles+"/"+id, bytes.NewReader([]byte("not-json")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", bearerVeh(sec))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatal(w.Code)
		}
	})
	t.Run("list_query_brand", func(t *testing.T) {
		bid := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, pathAPIVehicles+"?brand_id="+bid+"&limit=5&offset=0", nil)
		req.Header.Set("Authorization", bearerVeh(sec))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code)
		}
	})
	t.Run("delete_nf", func(t *testing.T) {
		h2 := NewHandler(&mockVehicle{nf: nf}, sec)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodDelete, pathAPIVehicles+"/"+nf, nil)
		req.Header.Set("Authorization", bearerVeh(sec))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatal(w.Code)
		}
	})
}
