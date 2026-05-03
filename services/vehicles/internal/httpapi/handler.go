package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/dealer/dealer/services/vehicles/internal/domain"
	"github.com/dealer/dealer/services/vehicles/internal/jwt"
	"github.com/dealer/dealer/services/vehicles/internal/service"
	"github.com/google/uuid"
)

type Handler struct {
	svc       service.VehicleAPI
	jwtSecret string
}

func NewHandler(svc service.VehicleAPI, jwtSecret string) *Handler {
	return &Handler{svc: svc, jwtSecret: jwtSecret}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodGet+" "+pathAPIVehicles, h.cors(h.auth(h.handleList)))
	mux.HandleFunc(http.MethodPost+" "+pathAPIVehicles, h.cors(h.auth(h.handleCreate)))
	mux.HandleFunc(http.MethodGet+" "+pathAPIVehicles+"/{id}", h.cors(h.auth(h.handleGet)))
	mux.HandleFunc(http.MethodPut+" "+pathAPIVehicles+"/{id}", h.cors(h.auth(h.handleUpdate)))
	mux.HandleFunc(http.MethodDelete+" "+pathAPIVehicles+"/{id}", h.cors(h.auth(h.handleDelete)))
	mux.HandleFunc(http.MethodOptions+" "+pathAPIVehicles, h.cors(nil))
}

func (h *Handler) cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", headerCORSAllowHeaders)
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
	statusFilter := r.URL.Query().Get("status")
	parseUUIDQuery := func(name string) *uuid.UUID {
		s := r.URL.Query().Get(name)
		if s == "" {
			return nil
		}
		uid, err := uuid.Parse(s)
		if err != nil {
			return nil
		}
		return &uid
	}
	brandID := parseUUIDQuery("brand_id")
	dealerPointID := parseUUIDQuery("dealer_point_id")
	legalEntityID := parseUUIDQuery("legal_entity_id")
	warehouseID := parseUUIDQuery("warehouse_id")

	list, total, err := h.svc.List(r.Context(), domain.VehicleListFilter{
		Limit: int32(limit), Offset: int32(offset), Search: search, StatusFilter: statusFilter,
		BrandID: brandID, DealerPointID: dealerPointID, LegalEntityID: legalEntityID, WarehouseID: warehouseID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	vehicles := make([]map[string]any, len(list))
	for i, v := range list {
		vehicles[i] = vehicleToMap(v)
	}
	writeJSON(w, http.StatusOK, map[string]any{"vehicles": vehicles, "total": total})
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		VIN           string  `json:"vin"`
		Make          string  `json:"make"`
		Model         string  `json:"model"`
		Year          int32   `json:"year"`
		MileageKm     int64   `json:"mileage_km"`
		Price         string  `json:"price"`
		Status        string  `json:"status"`
		Color         string  `json:"color"`
		Notes         string  `json:"notes"`
		BrandID       *string `json:"brand_id"`
		DealerPointID *string `json:"dealer_point_id"`
		LegalEntityID *string `json:"legal_entity_id"`
		WarehouseID   *string `json:"warehouse_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if req.VIN == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "vin required"})
		return
	}
	parseOpt := func(s *string) *uuid.UUID {
		if s == nil || *s == "" {
			return nil
		}
		uid, err := uuid.Parse(*s)
		if err != nil {
			return nil
		}
		return &uid
	}
	v, err := h.svc.Create(r.Context(), service.CreateVehicleInput{
		VIN: req.VIN, Make: req.Make, Model: req.Model, Year: req.Year, MileageKm: req.MileageKm,
		Price: req.Price, Status: req.Status, Color: req.Color, Notes: req.Notes,
		BrandID: parseOpt(req.BrandID), DealerPointID: parseOpt(req.DealerPointID),
		LegalEntityID: parseOpt(req.LegalEntityID), WarehouseID: parseOpt(req.WarehouseID),
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, vehicleToMap(v))
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	v, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, vehicleToMap(v))
}

func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	var req struct {
		VIN           *string `json:"vin"`
		Make          *string `json:"make"`
		Model         *string `json:"model"`
		Year          *int32  `json:"year"`
		MileageKm     *int64  `json:"mileage_km"`
		Price         *string `json:"price"`
		Status        *string `json:"status"`
		Color         *string `json:"color"`
		Notes         *string `json:"notes"`
		BrandID       *string `json:"brand_id"`
		DealerPointID *string `json:"dealer_point_id"`
		LegalEntityID *string `json:"legal_entity_id"`
		WarehouseID   *string `json:"warehouse_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	parseOptClear := func(s *string) (*uuid.UUID, bool) {
		if s == nil {
			return nil, false
		}
		if *s == "" {
			return nil, true
		}
		uid, err := uuid.Parse(*s)
		if err != nil {
			return nil, false
		}
		return &uid, false
	}
	brandID, clearBrand := parseOptClear(req.BrandID)
	dpID, clearDP := parseOptClear(req.DealerPointID)
	leID, clearLE := parseOptClear(req.LegalEntityID)
	whID, clearWH := parseOptClear(req.WarehouseID)
	v, err := h.svc.Update(r.Context(), id, service.UpdateVehicleInput{
		VIN: req.VIN, Make: req.Make, Model: req.Model, Year: req.Year, MileageKm: req.MileageKm,
		Price: req.Price, Status: req.Status, Color: req.Color, Notes: req.Notes,
		BrandID: brandID, ClearBrand: clearBrand, DealerPointID: dpID, LegalEntityID: leID, WarehouseID: whID,
		ClearDealerPoint: clearDP, ClearLegalEntity: clearLE, ClearWarehouse: clearWH,
	})
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, vehicleToMap(v))
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

func vehicleToMap(v *domain.Vehicle) map[string]any {
	if v == nil {
		return nil
	}
	m := map[string]any{
		"id":         v.ID.String(),
		"vin":        v.VIN,
		"make":       v.Make,
		"model":      v.Model,
		"year":       v.Year,
		"mileage_km": v.MileageKm,
		"price":      v.Price,
		"status":     v.Status,
		"color":      v.Color,
		"notes":      v.Notes,
		"created_at": v.CreatedAt.Unix(),
		"updated_at": v.UpdatedAt.Unix(),
	}
	if v.BrandID != nil {
		m["brand_id"] = v.BrandID.String()
	} else {
		m["brand_id"] = nil
	}
	if v.DealerPointID != nil {
		m["dealer_point_id"] = v.DealerPointID.String()
	} else {
		m["dealer_point_id"] = nil
	}
	if v.LegalEntityID != nil {
		m["legal_entity_id"] = v.LegalEntityID.String()
	} else {
		m["legal_entity_id"] = nil
	}
	if v.WarehouseID != nil {
		m["warehouse_id"] = v.WarehouseID.String()
	} else {
		m["warehouse_id"] = nil
	}
	return m
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set(headerContentType, mimeApplicationJSON)
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
