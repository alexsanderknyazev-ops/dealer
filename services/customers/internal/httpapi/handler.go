package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/dealer/dealer/customers-service/internal/domain"
	"github.com/dealer/dealer/customers-service/internal/jwt"
	"github.com/dealer/dealer/customers-service/internal/service"
)

type Handler struct {
	svc       *service.CustomerService
	jwtSecret string
}

func NewHandler(svc *service.CustomerService, jwtSecret string) *Handler {
	return &Handler{svc: svc, jwtSecret: jwtSecret}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/customers", h.cors(h.auth(h.handleList)))
	mux.HandleFunc("POST /api/customers", h.cors(h.auth(h.handleCreate)))
	mux.HandleFunc("GET /api/customers/{id}", h.cors(h.auth(h.handleGet)))
	mux.HandleFunc("PUT /api/customers/{id}", h.cors(h.auth(h.handleUpdate)))
	mux.HandleFunc("DELETE /api/customers/{id}", h.cors(h.auth(h.handleDelete)))
	mux.HandleFunc("OPTIONS /api/customers", h.cors(nil))
}

func (h *Handler) cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if next != nil {
			next(w, r)
		}
	}
}

func (h *Handler) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if _, _, err := jwt.Validate(h.jwtSecret, token); err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		next(w, r)
	}
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}
	search := r.URL.Query().Get("search")

	list, total, err := h.svc.List(r.Context(), int32(limit), int32(offset), search)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	customers := make([]map[string]any, len(list))
	for i, c := range list {
		customers[i] = customerToMap(c)
	}
	writeJSON(w, http.StatusOK, map[string]any{"customers": customers, "total": total})
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string `json:"name"`
		Email        string `json:"email"`
		Phone        string `json:"phone"`
		CustomerType string `json:"customer_type"`
		INN          string `json:"inn"`
		Address      string `json:"address"`
		Notes        string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name required"})
		return
	}
	c, err := h.svc.Create(r.Context(), req.Name, req.Email, req.Phone, req.CustomerType, req.INN, req.Address, req.Notes)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, customerToMap(c))
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	c, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, customerToMap(c))
}

func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	var req struct {
		Name         *string `json:"name"`
		Email        *string `json:"email"`
		Phone        *string `json:"phone"`
		CustomerType *string `json:"customer_type"`
		INN          *string `json:"inn"`
		Address      *string `json:"address"`
		Notes        *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	c, err := h.svc.Update(r.Context(), id, req.Name, req.Email, req.Phone, req.CustomerType, req.INN, req.Address, req.Notes)
	if err != nil {
		if err == service.ErrNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, customerToMap(c))
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		if err == service.ErrNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func customerToMap(c *domain.Customer) map[string]any {
	if c == nil {
		return nil
	}
	return map[string]any{
		"id":            c.ID.String(),
		"name":          c.Name,
		"email":         c.Email,
		"phone":         c.Phone,
		"customer_type": c.CustomerType,
		"inn":           c.INN,
		"address":       c.Address,
		"notes":         c.Notes,
		"created_at":    c.CreatedAt.Unix(),
		"updated_at":    c.UpdatedAt.Unix(),
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
