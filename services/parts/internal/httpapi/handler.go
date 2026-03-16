package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/dealer/dealer/services/parts/internal/domain"
	"github.com/dealer/dealer/services/parts/internal/jwt"
	"github.com/dealer/dealer/services/parts/internal/service"
)

type Handler struct {
	svc       *service.PartService
	jwtSecret string
}

func NewHandler(svc *service.PartService, jwtSecret string) *Handler {
	return &Handler{svc: svc, jwtSecret: jwtSecret}
}

func parseUUIDOpt(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &u
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Folders (more specific paths first)
	mux.HandleFunc("GET /api/parts/folders", h.cors(h.auth(h.handleListFolders)))
	mux.HandleFunc("POST /api/parts/folders", h.cors(h.auth(h.handleCreateFolder)))
	mux.HandleFunc("GET /api/parts/folders/{id}", h.cors(h.auth(h.handleGetFolder)))
	mux.HandleFunc("PUT /api/parts/folders/{id}", h.cors(h.auth(h.handleUpdateFolder)))
	mux.HandleFunc("DELETE /api/parts/folders/{id}", h.cors(h.auth(h.handleDeleteFolder)))
	mux.HandleFunc("OPTIONS /api/parts/folders", h.cors(nil))
	// Parts
	mux.HandleFunc("GET /api/parts", h.cors(h.auth(h.handleList)))
	mux.HandleFunc("POST /api/parts", h.cors(h.auth(h.handleCreate)))
	mux.HandleFunc("GET /api/parts/{id}", h.cors(h.auth(h.handleGet)))
	mux.HandleFunc("PUT /api/parts/{id}", h.cors(h.auth(h.handleUpdate)))
	mux.HandleFunc("DELETE /api/parts/{id}", h.cors(h.auth(h.handleDelete)))
	mux.HandleFunc("OPTIONS /api/parts", h.cors(nil))
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
	categoryFilter := r.URL.Query().Get("category")
	folderID := parseUUIDOpt(r.URL.Query().Get("folder_id"))
	brandID := parseUUIDOpt(r.URL.Query().Get("brand_id"))
	dealerPointID := parseUUIDOpt(r.URL.Query().Get("dealer_point_id"))
	legalEntityID := parseUUIDOpt(r.URL.Query().Get("legal_entity_id"))
	warehouseID := parseUUIDOpt(r.URL.Query().Get("warehouse_id"))

	list, total, err := h.svc.List(r.Context(), int32(limit), int32(offset), search, categoryFilter, folderID, brandID, dealerPointID, legalEntityID, warehouseID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	parts := make([]map[string]any, len(list))
	for i, p := range list {
		parts[i] = partToMap(p)
	}
	writeJSON(w, http.StatusOK, map[string]any{"parts": parts, "total": total})
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SKU           string `json:"sku"`
		Name          string `json:"name"`
		Category      string `json:"category"`
		FolderID      string `json:"folder_id"`
		BrandID       string `json:"brand_id"`
		DealerPointID string `json:"dealer_point_id"`
		LegalEntityID string `json:"legal_entity_id"`
		WarehouseID   string `json:"warehouse_id"`
		Quantity      int32  `json:"quantity"`
		Unit          string `json:"unit"`
		Price         string `json:"price"`
		Location      string `json:"location"`
		Notes         string `json:"notes"`
		Stock         []struct {
			WarehouseID string `json:"warehouse_id"`
			Quantity    int32  `json:"quantity"`
		} `json:"stock"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if req.SKU == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sku required"})
		return
	}
	var initialStock []service.StockRow
	if len(req.Stock) > 0 {
		for _, row := range req.Stock {
			if w := parseUUIDOpt(row.WarehouseID); w != nil {
				initialStock = append(initialStock, service.StockRow{WarehouseID: *w, Quantity: row.Quantity})
			}
		}
	}
	p, err := h.svc.Create(r.Context(), req.SKU, req.Name, req.Category, parseUUIDOpt(req.FolderID), parseUUIDOpt(req.BrandID), parseUUIDOpt(req.DealerPointID), parseUUIDOpt(req.LegalEntityID), parseUUIDOpt(req.WarehouseID), req.Quantity, req.Unit, req.Price, req.Location, req.Notes, initialStock)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	out := partToMap(p)
	if list, _ := h.svc.ListStock(r.Context(), p.ID.String()); len(list) > 0 {
		out["stock"] = stockToList(list)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	p, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	out := partToMap(p)
	if list, _ := h.svc.ListStock(r.Context(), id); len(list) > 0 {
		out["stock"] = stockToList(list)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	var req struct {
		SKU           *string `json:"sku"`
		Name          *string `json:"name"`
		Category      *string `json:"category"`
		FolderID      *string `json:"folder_id"`
		BrandID       *string `json:"brand_id"`
		DealerPointID *string `json:"dealer_point_id"`
		LegalEntityID *string `json:"legal_entity_id"`
		WarehouseID   *string `json:"warehouse_id"`
		Quantity      *int32  `json:"quantity"`
		Unit          *string `json:"unit"`
		Price         *string `json:"price"`
		Location      *string `json:"location"`
		Notes         *string `json:"notes"`
		Stock         *[]struct {
			WarehouseID string `json:"warehouse_id"`
			Quantity    int32  `json:"quantity"`
		} `json:"stock"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if req.Stock != nil {
		var rows []service.StockRow
		for _, row := range *req.Stock {
			if w := parseUUIDOpt(row.WarehouseID); w != nil && row.Quantity >= 0 {
				rows = append(rows, service.StockRow{WarehouseID: *w, Quantity: row.Quantity})
			}
		}
		if err := h.svc.ReplaceStock(r.Context(), id, rows); err != nil && err != service.ErrNotFound {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	p, err := h.svc.Update(r.Context(), id, req.SKU, req.Name, req.Category, req.FolderID, req.BrandID, req.DealerPointID, req.LegalEntityID, req.WarehouseID, req.Quantity, req.Unit, req.Price, req.Location, req.Notes)
	if err != nil {
		if err == service.ErrNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	out := partToMap(p)
	if list, _ := h.svc.ListStock(r.Context(), id); len(list) > 0 {
		out["stock"] = stockToList(list)
	}
	writeJSON(w, http.StatusOK, out)
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

func (h *Handler) handleListFolders(w http.ResponseWriter, r *http.Request) {
	parentID := parseUUIDOpt(r.URL.Query().Get("parent_id"))
	list, err := h.svc.ListFolders(r.Context(), parentID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	folders := make([]map[string]any, len(list))
	for i, f := range list {
		folders[i] = folderToMap(f)
	}
	writeJSON(w, http.StatusOK, map[string]any{"folders": folders})
}

func (h *Handler) handleCreateFolder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		ParentID string `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name required"})
		return
	}
	f, err := h.svc.CreateFolder(r.Context(), strings.TrimSpace(req.Name), parseUUIDOpt(req.ParentID))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, folderToMap(f))
}

func (h *Handler) handleGetFolder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	f, err := h.svc.GetFolder(r.Context(), id)
	if err != nil {
		if err == service.ErrFolderNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, folderToMap(f))
}

func (h *Handler) handleUpdateFolder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	var req struct {
		Name     *string `json:"name"`
		ParentID *string `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	f, err := h.svc.UpdateFolder(r.Context(), id, req.Name, req.ParentID)
	if err != nil {
		if err == service.ErrFolderNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, folderToMap(f))
}

func (h *Handler) handleDeleteFolder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}
	if err := h.svc.DeleteFolder(r.Context(), id); err != nil {
		if err == service.ErrFolderNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func stockToList(list []*domain.PartStock) []map[string]any {
	out := make([]map[string]any, len(list))
	for i, s := range list {
		out[i] = map[string]any{
			"warehouse_id": s.WarehouseID.String(),
			"quantity":     s.Quantity,
		}
	}
	return out
}

func folderToMap(f *domain.PartFolder) map[string]any {
	if f == nil {
		return nil
	}
	parentID := ""
	if f.ParentID != nil {
		parentID = f.ParentID.String()
	}
	return map[string]any{
		"id":         f.ID.String(),
		"name":      f.Name,
		"parent_id": parentID,
		"created_at": f.CreatedAt.Unix(),
		"updated_at": f.UpdatedAt.Unix(),
	}
}

func partToMap(p *domain.Part) map[string]any {
	if p == nil {
		return nil
	}
	folderID := ""
	if p.FolderID != nil {
		folderID = p.FolderID.String()
	}
	brandID := ""
	if p.BrandID != nil {
		brandID = p.BrandID.String()
	}
	dpID, leID, whID := "", "", ""
	if p.DealerPointID != nil {
		dpID = p.DealerPointID.String()
	}
	if p.LegalEntityID != nil {
		leID = p.LegalEntityID.String()
	}
	if p.WarehouseID != nil {
		whID = p.WarehouseID.String()
	}
	return map[string]any{
		"id":               p.ID.String(),
		"sku":              p.SKU,
		"name":             p.Name,
		"category":         p.Category,
		"folder_id":        folderID,
		"brand_id":         brandID,
		"dealer_point_id":  dpID,
		"legal_entity_id":  leID,
		"warehouse_id":     whID,
		"quantity":         p.Quantity,
		"unit":       p.Unit,
		"price":      p.Price,
		"location":   p.Location,
		"notes":      p.Notes,
		"created_at": p.CreatedAt.Unix(),
		"updated_at": p.UpdatedAt.Unix(),
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
