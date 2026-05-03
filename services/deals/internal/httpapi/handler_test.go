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
	return &domain.Deal{ID: uid, CustomerID: cid, VehicleID: vid, Amount: "1", Stage: "d", CreatedAt: now, UpdatedAt: now}, nil
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
	d := &domain.Deal{ID: uid, CustomerID: cid, VehicleID: vid, Amount: "1", Stage: "d", CreatedAt: now, UpdatedAt: now}
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
	cl := &jwt.Claims{UserID: "u", Email: "e", RegisteredClaims: jwtlib.RegisteredClaims{
		ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
	}}
	s, _ := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, cl).SignedString([]byte(secret))
	return "Bearer " + s
}

func TestDealsHTTP(t *testing.T) {
	sec := "s"
	h := NewHandler(&mockDeal{}, sec)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	t.Run("options", func(t *testing.T) {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodOptions, pathAPIDeals, nil))
		if w.Code != http.StatusNoContent {
			t.Fatal(w.Code)
		}
	})
	t.Run("unauth", func(t *testing.T) {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, pathAPIDeals, nil))
		if w.Code != http.StatusUnauthorized {
			t.Fatal(w.Code)
		}
	})
	t.Run("list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, pathAPIDeals, nil)
		req.Header.Set("Authorization", bearerDeal(sec))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code)
		}
	})
	t.Run("create_missing_ids", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, pathAPIDeals, bytes.NewReader([]byte("{}")))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", bearerDeal(sec))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatal(w.Code)
		}
	})
	cid, vid := uuid.New().String(), uuid.New().String()
	t.Run("create_ok", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"customer_id": cid, "vehicle_id": vid, "amount": "10"})
		req := httptest.NewRequest(http.MethodPost, pathAPIDeals, bytes.NewReader(body))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", bearerDeal(sec))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code, w.Body.String())
		}
	})
	t.Run("list_err", func(t *testing.T) {
		h2 := NewHandler(&mockDeal{listErr: errors.New("db")}, sec)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodGet, pathAPIDeals, nil)
		req.Header.Set("Authorization", bearerDeal(sec))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatal(w.Code)
		}
	})
	nf := "00000000-0000-0000-0000-000000000088"
	t.Run("get_nf", func(t *testing.T) {
		h2 := NewHandler(&mockDeal{nf: nf}, sec)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodGet, pathDealByID(nf), nil)
		req.Header.Set("Authorization", bearerDeal(sec))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatal(w.Code)
		}
	})
	id := uuid.New().String()
	t.Run("put_del", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"amount": "99"})
		req := httptest.NewRequest(http.MethodPut, pathDealByID(id), bytes.NewReader(body))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", bearerDeal(sec))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatal(w.Code)
		}
		req2 := httptest.NewRequest(http.MethodDelete, pathDealByID(id), nil)
		req2.Header.Set("Authorization", bearerDeal(sec))
		w2 := httptest.NewRecorder()
		mux.ServeHTTP(w2, req2)
		if w2.Code != http.StatusNoContent {
			t.Fatal(w2.Code)
		}
	})
	t.Run("create_err", func(t *testing.T) {
		h2 := NewHandler(&mockDeal{createErr: errors.New("db")}, sec)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		cid, vid := uuid.New().String(), uuid.New().String()
		body, _ := json.Marshal(map[string]string{"customer_id": cid, "vehicle_id": vid})
		req := httptest.NewRequest(http.MethodPost, pathAPIDeals, bytes.NewReader(body))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", bearerDeal(sec))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400 got %d", w.Code)
		}
	})
	t.Run("get_internal", func(t *testing.T) {
		h2 := NewHandler(&mockDeal{getErr: errors.New("db")}, sec)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodGet, pathDealByID(uuid.New().String()), nil)
		req.Header.Set("Authorization", bearerDeal(sec))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatal(w.Code)
		}
	})
	t.Run("put_bad_json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, pathDealByID(id), bytes.NewReader([]byte("x")))
		setRequestJSONContentType(req)
		req.Header.Set("Authorization", bearerDeal(sec))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatal(w.Code)
		}
	})
	t.Run("delete_nf", func(t *testing.T) {
		h2 := NewHandler(&mockDeal{nf: nf}, sec)
		m2 := http.NewServeMux()
		h2.RegisterRoutes(m2)
		req := httptest.NewRequest(http.MethodDelete, pathDealByID(nf), nil)
		req.Header.Set("Authorization", bearerDeal(sec))
		w := httptest.NewRecorder()
		m2.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatal(w.Code)
		}
	})
}
