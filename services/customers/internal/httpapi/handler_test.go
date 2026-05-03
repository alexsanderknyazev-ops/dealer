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

	"github.com/dealer/dealer/customers-service/internal/domain"
	"github.com/dealer/dealer/customers-service/internal/jwt"
	"github.com/dealer/dealer/customers-service/internal/service"
)

type mockCustomerAPI struct {
	list       []*domain.Customer
	total      int32
	listErr    error
	createErr  error
	getErr     error
	updateErr  error
	deleteErr  error
	notFoundID string
}

func (m *mockCustomerAPI) Create(_ context.Context, name, email, phone, customerType, inn, address, notes string) (*domain.Customer, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	now := time.Now().UTC()
	return &domain.Customer{
		ID: uuid.New(), Name: name, Email: email, Phone: phone, CustomerType: customerType,
		INN: inn, Address: address, Notes: notes, CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (m *mockCustomerAPI) Get(_ context.Context, id string) (*domain.Customer, error) {
	if m.notFoundID != "" && id == m.notFoundID {
		return nil, service.ErrNotFound
	}
	if m.getErr != nil {
		return nil, m.getErr
	}
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	return &domain.Customer{ID: uid, Name: "N", Email: "e@e", CreatedAt: now, UpdatedAt: now}, nil
}

func (m *mockCustomerAPI) List(_ context.Context, _, _ int32, _ string) ([]*domain.Customer, int32, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.list, m.total, nil
}

func (m *mockCustomerAPI) Update(_ context.Context, id string, name, _, _, _, _, _, _ *string) (*domain.Customer, error) {
	if m.notFoundID != "" && id == m.notFoundID {
		return nil, service.ErrNotFound
	}
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	n := ""
	if name != nil {
		n = *name
	}
	return &domain.Customer{ID: uid, Name: n, Email: "e", CreatedAt: now, UpdatedAt: now}, nil
}

func (m *mockCustomerAPI) Delete(_ context.Context, id string) error {
	if m.notFoundID != "" && id == m.notFoundID {
		return service.ErrNotFound
	}
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return nil
}

func bearer(secret string) string {
	claims := &jwt.Claims{
		UserID: "u",
		Email:  "e@e",
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	s, _ := tok.SignedString([]byte(secret))
	return "Bearer " + s
}

func TestHandler_Options(t *testing.T) {
	h := NewHandler(&mockCustomerAPI{}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodOptions, "/api/customers", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("code %d", w.Code)
	}
}

func TestHandler_List_Unauthorized(t *testing.T) {
	h := NewHandler(&mockCustomerAPI{}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, "/api/customers", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code %d", w.Code)
	}
}

func TestHandler_List_ServiceError(t *testing.T) {
	h := NewHandler(&mockCustomerAPI{listErr: errors.New("db")}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, "/api/customers", nil)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code %d", w.Code)
	}
}

func TestHandler_List_OK(t *testing.T) {
	h := NewHandler(&mockCustomerAPI{list: []*domain.Customer{}, total: 0}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, "/api/customers?limit=5&offset=0", nil)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("code %d body %s", w.Code, w.Body.String())
	}
}

func TestHandler_Create_BadBody(t *testing.T) {
	h := NewHandler(&mockCustomerAPI{}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodPost, "/api/customers", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code %d", w.Code)
	}
}

func TestHandler_Create_NameRequired(t *testing.T) {
	h := NewHandler(&mockCustomerAPI{}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	body, _ := json.Marshal(map[string]string{"email": "a@b.c"})
	req := httptest.NewRequest(http.MethodPost, "/api/customers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code %d", w.Code)
	}
}

func TestHandler_Get_NotFound(t *testing.T) {
	nf := "00000000-0000-0000-0000-000000000001"
	h := NewHandler(&mockCustomerAPI{notFoundID: nf}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, "/api/customers/"+nf, nil)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code %d", w.Code)
	}
}

func TestHandler_Get_InternalErr(t *testing.T) {
	id := uuid.New().String()
	h := NewHandler(&mockCustomerAPI{getErr: errors.New("db")}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, "/api/customers/"+id, nil)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code %d", w.Code)
	}
}

func TestHandler_Get_OK(t *testing.T) {
	id := uuid.New().String()
	h := NewHandler(&mockCustomerAPI{}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, "/api/customers/"+id, nil)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("code %d %s", w.Code, w.Body.String())
	}
}

func TestHandler_Create_ServiceError(t *testing.T) {
	h := NewHandler(&mockCustomerAPI{createErr: errors.New("db")}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	body, _ := json.Marshal(map[string]string{"name": "A", "email": "a@b.c"})
	req := httptest.NewRequest(http.MethodPost, "/api/customers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code %d", w.Code)
	}
}

func TestHandler_Create_OK(t *testing.T) {
	h := NewHandler(&mockCustomerAPI{}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	body, _ := json.Marshal(map[string]string{"name": "A", "email": "a@b.c"})
	req := httptest.NewRequest(http.MethodPost, "/api/customers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("code %d", w.Code)
	}
}

func TestHandler_Update_NotFound(t *testing.T) {
	nf := "00000000-0000-0000-0000-000000000002"
	h := NewHandler(&mockCustomerAPI{notFoundID: nf}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	body, _ := json.Marshal(map[string]string{"name": "Z"})
	req := httptest.NewRequest(http.MethodPut, "/api/customers/"+nf, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code %d", w.Code)
	}
}

func TestHandler_Update_BadBody(t *testing.T) {
	id := uuid.New().String()
	h := NewHandler(&mockCustomerAPI{}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodPut, "/api/customers/"+id, bytes.NewReader([]byte("x")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code %d", w.Code)
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	nf := "00000000-0000-0000-0000-000000000003"
	h := NewHandler(&mockCustomerAPI{notFoundID: nf}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodDelete, "/api/customers/"+nf, nil)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code %d", w.Code)
	}
}

func TestHandler_Update_Delete_NoContent(t *testing.T) {
	id := uuid.New().String()
	h := NewHandler(&mockCustomerAPI{}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	body, _ := json.Marshal(map[string]string{"name": "Z"})
	req := httptest.NewRequest(http.MethodPut, "/api/customers/"+id, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("put code %d", w.Code)
	}
	req2 := httptest.NewRequest(http.MethodDelete, "/api/customers/"+id, nil)
	req2.Header.Set("Authorization", bearer("sec"))
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNoContent {
		t.Fatalf("del code %d", w2.Code)
	}
}

