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

func assertHTTPStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("unexpected HTTP status: want %d got %d", want, w.Code)
	}
}

func assertHTTPStatusBody(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("unexpected HTTP status: want %d got %d body=%s", want, w.Code, w.Body.String())
	}
}

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

func (m *mockCustomerAPI) Create(_ context.Context, in service.CreateCustomerInput) (*domain.Customer, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	now := time.Now().UTC()
	return &domain.Customer{
		ID: uuid.New(), Name: in.Name, Email: in.Email, Phone: in.Phone, CustomerType: in.CustomerType,
		INN: in.INN, Address: in.Address, Notes: in.Notes, CreatedAt: now, UpdatedAt: now,
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

func (m *mockCustomerAPI) List(_ context.Context, _ domain.CustomerListParams) ([]*domain.Customer, int32, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.list, m.total, nil
}

func (m *mockCustomerAPI) Update(_ context.Context, id string, in service.UpdateCustomerInput) (*domain.Customer, error) {
	if m.notFoundID != "" && id == m.notFoundID {
		return nil, service.ErrNotFound
	}
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	uid, _ := uuid.Parse(id)
	now := time.Now().UTC()
	n := ""
	if in.Name != nil {
		n = *in.Name
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
	req := httptest.NewRequest(http.MethodOptions, pathAPICustomers, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assertHTTPStatus(t, w, http.StatusNoContent)
}

func TestHandler_List_Unauthorized(t *testing.T) {
	h := NewHandler(&mockCustomerAPI{}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, pathAPICustomers, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assertHTTPStatus(t, w, http.StatusUnauthorized)
}

func TestHandler_List_ServiceError(t *testing.T) {
	h := NewHandler(&mockCustomerAPI{listErr: errors.New("db")}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, pathAPICustomers, nil)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assertHTTPStatus(t, w, http.StatusInternalServerError)
}

func TestHandler_List_OK(t *testing.T) {
	h := NewHandler(&mockCustomerAPI{list: []*domain.Customer{}, total: 0}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, pathAPICustomers+"?limit=5&offset=0", nil)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assertHTTPStatusBody(t, w, http.StatusOK)
}

func TestHandler_Create_BadBody(t *testing.T) {
	h := NewHandler(&mockCustomerAPI{}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodPost, pathAPICustomers, bytes.NewReader([]byte("{")))
	setRequestJSONContentType(req)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assertHTTPStatus(t, w, http.StatusBadRequest)
}

func TestHandler_Create_NameRequired(t *testing.T) {
	h := NewHandler(&mockCustomerAPI{}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	body, _ := json.Marshal(map[string]string{"email": "a@b.c"})
	req := httptest.NewRequest(http.MethodPost, pathAPICustomers, bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assertHTTPStatus(t, w, http.StatusBadRequest)
}

func TestHandler_Get_NotFound(t *testing.T) {
	nf := "00000000-0000-0000-0000-000000000001"
	h := NewHandler(&mockCustomerAPI{notFoundID: nf}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, pathCustomerByID(nf), nil)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assertHTTPStatus(t, w, http.StatusNotFound)
}

func TestHandler_Get_InternalErr(t *testing.T) {
	id := uuid.New().String()
	h := NewHandler(&mockCustomerAPI{getErr: errors.New("db")}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, pathCustomerByID(id), nil)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assertHTTPStatus(t, w, http.StatusInternalServerError)
}

func TestHandler_Get_OK(t *testing.T) {
	id := uuid.New().String()
	h := NewHandler(&mockCustomerAPI{}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, pathCustomerByID(id), nil)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assertHTTPStatusBody(t, w, http.StatusOK)
}

func TestHandler_Create_ServiceError(t *testing.T) {
	h := NewHandler(&mockCustomerAPI{createErr: errors.New("db")}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	body, _ := json.Marshal(map[string]string{"name": "A", "email": "a@b.c"})
	req := httptest.NewRequest(http.MethodPost, pathAPICustomers, bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assertHTTPStatus(t, w, http.StatusInternalServerError)
}

func TestHandler_Create_OK(t *testing.T) {
	h := NewHandler(&mockCustomerAPI{}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	body, _ := json.Marshal(map[string]string{"name": "A", "email": "a@b.c"})
	req := httptest.NewRequest(http.MethodPost, pathAPICustomers, bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assertHTTPStatus(t, w, http.StatusOK)
}

func TestHandler_Update_NotFound(t *testing.T) {
	nf := "00000000-0000-0000-0000-000000000002"
	h := NewHandler(&mockCustomerAPI{notFoundID: nf}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	body, _ := json.Marshal(map[string]string{"name": "Z"})
	req := httptest.NewRequest(http.MethodPut, pathCustomerByID(nf), bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assertHTTPStatus(t, w, http.StatusNotFound)
}

func TestHandler_Update_BadBody(t *testing.T) {
	id := uuid.New().String()
	h := NewHandler(&mockCustomerAPI{}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodPut, pathCustomerByID(id), bytes.NewReader([]byte("x")))
	setRequestJSONContentType(req)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assertHTTPStatus(t, w, http.StatusBadRequest)
}

func TestHandler_Delete_NotFound(t *testing.T) {
	nf := "00000000-0000-0000-0000-000000000003"
	h := NewHandler(&mockCustomerAPI{notFoundID: nf}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodDelete, pathCustomerByID(nf), nil)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assertHTTPStatus(t, w, http.StatusNotFound)
}

func TestHandler_Update_Delete_NoContent(t *testing.T) {
	id := uuid.New().String()
	h := NewHandler(&mockCustomerAPI{}, "sec")
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	body, _ := json.Marshal(map[string]string{"name": "Z"})
	req := httptest.NewRequest(http.MethodPut, pathCustomerByID(id), bytes.NewReader(body))
	setRequestJSONContentType(req)
	req.Header.Set("Authorization", bearer("sec"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assertHTTPStatus(t, w, http.StatusOK)
	req2 := httptest.NewRequest(http.MethodDelete, pathCustomerByID(id), nil)
	req2.Header.Set("Authorization", bearer("sec"))
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	assertHTTPStatus(t, w2, http.StatusNoContent)
}
