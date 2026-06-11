package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/dealer/dealer/pkg/grpclient"
	"github.com/dealer/dealer/services/deals/internal/domain"
	"github.com/dealer/dealer/services/deals/internal/jwt"
	"github.com/dealer/dealer/services/deals/internal/service"
)

type Handler struct {
	svc       service.DealAPI
	jwtSecret string
}

func NewHandler(svc service.DealAPI, jwtSecret string) *Handler {
	return &Handler{svc: svc, jwtSecret: jwtSecret}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodGet+" "+pathAPIDeals, h.cors(h.auth(h.handleList)))
	mux.HandleFunc(http.MethodPost+" "+pathAPIDeals, h.cors(h.auth(h.handleCreate)))
	mux.HandleFunc(http.MethodGet+" "+pathAPIDeals+"/{id}", h.cors(h.auth(h.handleGet)))
	mux.HandleFunc(http.MethodPut+" "+pathAPIDeals+"/{id}", h.cors(h.auth(h.handleUpdate)))
	mux.HandleFunc(http.MethodDelete+" "+pathAPIDeals+"/{id}", h.cors(h.auth(h.handleDelete)))
	mux.HandleFunc(http.MethodOptions+" "+pathAPIDeals, h.cors(nil))
}

func (h *Handler) cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", headerCORSAllowHeader)
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
		userID, _, role, err := jwt.Validate(h.jwtSecret, token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if isWriteMethod(r.Method) && !hasAllowedRole(role, "admin", "manager", "sales") {
			log.Printf(
				"rbac deny: service=deals user_id=%s role=%s endpoint=%s action=%s reason=%s",
				userID, role, r.URL.Path, actionFromMethod(r.Method), "insufficient_role",
			)
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		}
		next(w, r)
	}
}

func isWriteMethod(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete
}

func actionFromMethod(method string) string {
	switch method {
	case http.MethodPost:
		return "create"
	case http.MethodPut:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return "read"
	}
}

func requestContext(r *http.Request) context.Context {
	if auth := r.Header.Get("Authorization"); auth != "" {
		return grpclient.WithAuthorization(r.Context(), auth)
	}
	return r.Context()
}

func hasAllowedRole(role string, allowed ...string) bool {
	for _, r := range allowed {
		if role == r {
			return true
		}
	}
	return false
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
	stage := r.URL.Query().Get("stage")
	customerID := r.URL.Query().Get("customer_id")

	list, total, err := h.svc.List(requestContext(r), int32(limit), int32(offset), stage, customerID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	deals := make([]map[string]any, len(list))
	for i, d := range list {
		deals[i] = dealToMap(d)
	}
	writeJSON(w, http.StatusOK, map[string]any{"deals": deals, "total": total})
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID string `json:"customer_id"`
		VehicleID  string `json:"vehicle_id"`
		Amount     string `json:"amount"`
		Stage      string `json:"stage"`
		AssignedTo string `json:"assigned_to"`
		Notes      string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if req.CustomerID == "" || req.VehicleID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "customer_id and vehicle_id required"})
		return
	}
	d, err := h.svc.Create(requestContext(r), service.CreateDealInput{
		CustomerID: req.CustomerID, VehicleID: req.VehicleID, Amount: req.Amount, Stage: req.Stage, AssignedTo: req.AssignedTo, Notes: req.Notes,
	})
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) || errors.Is(err, service.ErrVehicleNotFound) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, dealToMap(d))
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	d, err := h.svc.Get(requestContext(r), id)
	if err != nil {
		if err == service.ErrNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, dealToMap(d))
}

func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	var req struct {
		CustomerID *string `json:"customer_id"`
		VehicleID  *string `json:"vehicle_id"`
		Amount     *string `json:"amount"`
		Stage      *string `json:"stage"`
		AssignedTo *string `json:"assigned_to"`
		Notes      *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	d, err := h.svc.Update(requestContext(r), id, service.UpdateDealInput{
		CustomerID: req.CustomerID, VehicleID: req.VehicleID, Amount: req.Amount,
		Stage: req.Stage, AssignedTo: req.AssignedTo, Notes: req.Notes,
	})
	if err != nil {
		if err == service.ErrNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, dealToMap(d))
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	if err := h.svc.Delete(requestContext(r), id); err != nil {
		if err == service.ErrNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func dealToMap(d *domain.Deal) map[string]any {
	if d == nil {
		return nil
	}
	assignedTo := ""
	if d.AssignedTo != nil {
		assignedTo = d.AssignedTo.String()
	}
	return map[string]any{
		"id":          d.ID.String(),
		"customer_id": d.CustomerID.String(),
		"vehicle_id":  d.VehicleID.String(),
		"amount":      d.Amount,
		"stage":       d.Stage,
		"assigned_to": assignedTo,
		"notes":       d.Notes,
		"created_at":  d.CreatedAt.Unix(),
		"updated_at":  d.UpdatedAt.Unix(),
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set(headerContentType, mimeApplicationJSON)
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
