package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/dealer/dealer/services/dealerpoints/internal/domain"
	"github.com/dealer/dealer/services/dealerpoints/internal/jwt"
	"github.com/dealer/dealer/services/dealerpoints/internal/service"
)

type Handler struct {
	svc       *service.DealerPointsService
	jwtSecret string
}

func NewHandler(svc *service.DealerPointsService, jwtSecret string) *Handler {
	return &Handler{svc: svc, jwtSecret: jwtSecret}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Dealer points
	mux.HandleFunc(http.MethodGet+" "+pathAPIDealerPoints, h.cors(h.auth(h.handleListDealerPoints)))
	mux.HandleFunc(http.MethodPost+" "+pathAPIDealerPoints, h.cors(h.auth(h.handleCreateDealerPoint)))
	mux.HandleFunc(http.MethodGet+" "+pathAPIDealerPoints+"/{id}", h.cors(h.auth(h.handleGetDealerPoint)))
	mux.HandleFunc(http.MethodPut+" "+pathAPIDealerPoints+"/{id}", h.cors(h.auth(h.handleUpdateDealerPoint)))
	mux.HandleFunc(http.MethodDelete+" "+pathAPIDealerPoints+"/{id}", h.cors(h.auth(h.handleDeleteDealerPoint)))
	// Legal entities
	mux.HandleFunc(http.MethodGet+" "+pathAPILegalEntities, h.cors(h.auth(h.handleListLegalEntities)))
	mux.HandleFunc(http.MethodPost+" "+pathAPILegalEntities, h.cors(h.auth(h.handleCreateLegalEntity)))
	mux.HandleFunc(http.MethodGet+" "+pathAPILegalEntities+"/{id}", h.cors(h.auth(h.handleGetLegalEntity)))
	mux.HandleFunc(http.MethodPut+" "+pathAPILegalEntities+"/{id}", h.cors(h.auth(h.handleUpdateLegalEntity)))
	mux.HandleFunc(http.MethodDelete+" "+pathAPILegalEntities+"/{id}", h.cors(h.auth(h.handleDeleteLegalEntity)))
	mux.HandleFunc(http.MethodGet+" "+pathAPIDealerPoints+"/{id}/legal-entities", h.cors(h.auth(h.handleListLegalEntitiesByDealerPoint)))
	mux.HandleFunc(http.MethodPost+" "+pathAPIDealerPoints+"/{id}/legal-entities", h.cors(h.auth(h.handleLinkLegalEntity)))
	mux.HandleFunc(http.MethodDelete+" "+pathAPIDealerPoints+"/{dpId}/legal-entities/{leId}", h.cors(h.auth(h.handleUnlinkLegalEntity)))
	// Warehouses
	mux.HandleFunc(http.MethodGet+" "+pathAPIWarehouses, h.cors(h.auth(h.handleListWarehouses)))
	mux.HandleFunc(http.MethodPost+" "+pathAPIWarehouses, h.cors(h.auth(h.handleCreateWarehouse)))
	mux.HandleFunc(http.MethodGet+" "+pathAPIWarehouses+"/{id}", h.cors(h.auth(h.handleGetWarehouse)))
	mux.HandleFunc(http.MethodPut+" "+pathAPIWarehouses+"/{id}", h.cors(h.auth(h.handleUpdateWarehouse)))
	mux.HandleFunc(http.MethodDelete+" "+pathAPIWarehouses+"/{id}", h.cors(h.auth(h.handleDeleteWarehouse)))
	mux.HandleFunc(http.MethodOptions+" "+pathAPIDealerPoints, h.cors(nil))
	mux.HandleFunc(http.MethodOptions+" "+pathAPILegalEntities, h.cors(nil))
	mux.HandleFunc(http.MethodOptions+" "+pathAPIWarehouses, h.cors(nil))
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

func limitOffset(r *http.Request) (limit, offset int) {
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func writeErr(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrDealerPointNotFound) || errors.Is(err, service.ErrLegalEntityNotFound) || errors.Is(err, service.ErrWarehouseNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}

// Dealer points
func (h *Handler) handleListDealerPoints(w http.ResponseWriter, r *http.Request) {
	limit, offset := limitOffset(r)
	list, total, err := h.svc.ListDealerPoints(r.Context(), int32(limit), int32(offset), r.URL.Query().Get("search"))
	if err != nil {
		writeErr(w, err)
		return
	}
	out := make([]map[string]any, len(list))
	for i, d := range list {
		out[i] = dealerPointToMap(d)
	}
	writeJSON(w, http.StatusOK, map[string]any{"dealer_points": out, "total": total})
}

func (h *Handler) handleCreateDealerPoint(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	d, err := h.svc.CreateDealerPoint(r.Context(), strings.TrimSpace(req.Name), strings.TrimSpace(req.Address))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dealerPointToMap(d))
}

func (h *Handler) handleGetDealerPoint(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	d, err := h.svc.GetDealerPoint(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dealerPointToMap(d))
}

func (h *Handler) handleUpdateDealerPoint(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	var req struct {
		Name    *string `json:"name"`
		Address *string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	d, err := h.svc.UpdateDealerPoint(r.Context(), id, req.Name, req.Address)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dealerPointToMap(d))
}

func (h *Handler) handleDeleteDealerPoint(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	if err := h.svc.DeleteDealerPoint(r.Context(), id); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Legal entities
func (h *Handler) handleListLegalEntities(w http.ResponseWriter, r *http.Request) {
	limit, offset := limitOffset(r)
	list, total, err := h.svc.ListLegalEntities(r.Context(), int32(limit), int32(offset), r.URL.Query().Get("search"))
	if err != nil {
		writeErr(w, err)
		return
	}
	out := make([]map[string]any, len(list))
	for i, e := range list {
		out[i] = legalEntityToMap(e)
	}
	writeJSON(w, http.StatusOK, map[string]any{"legal_entities": out, "total": total})
}

func (h *Handler) handleCreateLegalEntity(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		INN     string `json:"inn"`
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	e, err := h.svc.CreateLegalEntity(r.Context(), strings.TrimSpace(req.Name), strings.TrimSpace(req.INN), strings.TrimSpace(req.Address))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, legalEntityToMap(e))
}

func (h *Handler) handleGetLegalEntity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	e, err := h.svc.GetLegalEntity(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, legalEntityToMap(e))
}

func (h *Handler) handleUpdateLegalEntity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	var req struct {
		Name    *string `json:"name"`
		INN     *string `json:"inn"`
		Address *string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	e, err := h.svc.UpdateLegalEntity(r.Context(), id, req.Name, req.INN, req.Address)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, legalEntityToMap(e))
}

func (h *Handler) handleDeleteLegalEntity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	if err := h.svc.DeleteLegalEntity(r.Context(), id); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleListLegalEntitiesByDealerPoint(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	limit, offset := limitOffset(r)
	list, total, err := h.svc.ListLegalEntitiesByDealerPoint(r.Context(), id, int32(limit), int32(offset))
	if err != nil {
		writeErr(w, err)
		return
	}
	out := make([]map[string]any, len(list))
	for i, e := range list {
		out[i] = legalEntityToMap(e)
	}
	writeJSON(w, http.StatusOK, map[string]any{"legal_entities": out, "total": total})
}

func (h *Handler) handleLinkLegalEntity(w http.ResponseWriter, r *http.Request) {
	dpID := r.PathValue("id")
	if dpID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	var req struct {
		LegalEntityID string `json:"legal_entity_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if req.LegalEntityID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "legal_entity_id required"})
		return
	}
	if err := h.svc.LinkLegalEntityToDealerPoint(r.Context(), dpID, req.LegalEntityID); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleUnlinkLegalEntity(w http.ResponseWriter, r *http.Request) {
	dpID := r.PathValue("dpId")
	leID := r.PathValue("leId")
	if dpID == "" || leID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "dpId and leId required"})
		return
	}
	if err := h.svc.UnlinkLegalEntityFromDealerPoint(r.Context(), dpID, leID); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Warehouses
func (h *Handler) handleListWarehouses(w http.ResponseWriter, r *http.Request) {
	limit, offset := limitOffset(r)
	q := r.URL.Query()
	list, total, err := h.svc.ListWarehouses(r.Context(), int32(limit), int32(offset), q.Get("dealer_point_id"), q.Get("legal_entity_id"), q.Get("type"))
	if err != nil {
		writeErr(w, err)
		return
	}
	out := make([]map[string]any, len(list))
	for i, wh := range list {
		out[i] = warehouseToMap(wh)
	}
	writeJSON(w, http.StatusOK, map[string]any{"warehouses": out, "total": total})
}

func (h *Handler) handleCreateWarehouse(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DealerPointID string `json:"dealer_point_id"`
		LegalEntityID string `json:"legal_entity_id"`
		Type          string `json:"type"`
		Name          string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if req.DealerPointID == "" || req.LegalEntityID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "dealer_point_id and legal_entity_id required"})
		return
	}
	wh, err := h.svc.CreateWarehouse(r.Context(), req.DealerPointID, req.LegalEntityID, req.Type, strings.TrimSpace(req.Name))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, warehouseToMap(wh))
}

func (h *Handler) handleGetWarehouse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	wh, err := h.svc.GetWarehouse(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, warehouseToMap(wh))
}

func (h *Handler) handleUpdateWarehouse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	var req struct {
		Name *string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	wh, err := h.svc.UpdateWarehouse(r.Context(), id, req.Name)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, warehouseToMap(wh))
}

func (h *Handler) handleDeleteWarehouse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	if err := h.svc.DeleteWarehouse(r.Context(), id); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func dealerPointToMap(d *domain.DealerPoint) map[string]any {
	if d == nil {
		return nil
	}
	return map[string]any{
		"id":         d.ID.String(),
		"name":       d.Name,
		"address":    d.Address,
		"created_at": d.CreatedAt.Unix(),
		"updated_at": d.UpdatedAt.Unix(),
	}
}

func legalEntityToMap(e *domain.LegalEntity) map[string]any {
	if e == nil {
		return nil
	}
	return map[string]any{
		"id":         e.ID.String(),
		"name":       e.Name,
		"inn":        e.INN,
		"address":    e.Address,
		"created_at": e.CreatedAt.Unix(),
		"updated_at": e.UpdatedAt.Unix(),
	}
}

func warehouseToMap(w *domain.Warehouse) map[string]any {
	if w == nil {
		return nil
	}
	return map[string]any{
		"id":              w.ID.String(),
		"dealer_point_id": w.DealerPointID.String(),
		"legal_entity_id": w.LegalEntityID.String(),
		"type":            w.Type,
		"name":            w.Name,
		"created_at":      w.CreatedAt.Unix(),
		"updated_at":      w.UpdatedAt.Unix(),
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
