package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/dealer/dealer/auth-service/internal/service"
)

// Handler — HTTP API для браузера (регистрация, логин, refresh, logout, me).
type Handler struct {
	svc *service.AuthService
}

// NewHandler создаёт HTTP-обработчик.
func NewHandler(svc *service.AuthService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes вешает маршруты на mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodPost+" "+pathAPIRegister, h.cors(h.handleRegister))
	mux.HandleFunc(http.MethodPost+" "+pathAPILogin, h.cors(h.handleLogin))
	mux.HandleFunc(http.MethodPost+" "+pathAPIRefresh, h.cors(h.handleRefresh))
	mux.HandleFunc(http.MethodPost+" "+pathAPILogout, h.cors(h.handleLogout))
	mux.HandleFunc(http.MethodGet+" "+pathAPIMe, h.cors(h.handleMe))
	// OPTIONS для CORS preflight только для auth-путей (общий OPTIONS /api/ конфликтует с прокси /api/customers и /api/vehicles в Go 1.22)
	mux.HandleFunc(http.MethodOptions+" "+pathAPIRegister, h.cors(nil))
	mux.HandleFunc(http.MethodOptions+" "+pathAPILogin, h.cors(nil))
	mux.HandleFunc(http.MethodOptions+" "+pathAPIRefresh, h.cors(nil))
	mux.HandleFunc(http.MethodOptions+" "+pathAPILogout, h.cors(nil))
	mux.HandleFunc(http.MethodOptions+" "+pathAPIMe, h.cors(nil))
}

func (h *Handler) cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Phone    string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	user, access, refresh, expiresAt, err := h.svc.Register(r.Context(), req.Email, req.Password, req.Name, req.Phone)
	if err != nil {
		if err == service.ErrUserExists {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "user with this email already exists"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":       user.ID.String(),
		"email":         user.Email,
		"access_token":  access,
		"refresh_token": refresh,
		"expires_at":    expiresAt.Unix(),
	})
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	user, access, refresh, expiresAt, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrBadCredentials {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":       user.ID.String(),
		"email":         user.Email,
		"access_token":  access,
		"refresh_token": refresh,
		"expires_at":    expiresAt.Unix(),
	})
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "refresh_token required"})
		return
	}
	access, refresh, expiresAt, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if err == service.ErrInvalidToken {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired refresh token"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  access,
		"refresh_token": refresh,
		"expires_at":    expiresAt.Unix(),
	})
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	_ = h.svc.Logout(r.Context(), req.RefreshToken)
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"valid": false})
		return
	}
	userID, email, valid := h.svc.Validate(r.Context(), token)
	if !valid {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"valid": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id": userID,
		"email":   email,
		"valid":   true,
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// SPAFileServer отдаёт статику; для путей без файла отдаёт index.html (SPA routing).
func SPAFileServer(root http.FileSystem) http.Handler {
	fs := http.FileServer(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		clean := path.Clean(r.URL.Path)
		if clean == "/" || clean == "" {
			serveIndex(root, w, r)
			return
		}
		name := strings.TrimPrefix(clean, "/")
		if f, err := root.Open(name); err == nil {
			f.Close()
			fs.ServeHTTP(w, r)
			return
		}
		// Нет такого файла — отдаём index.html для клиентского роутинга (SPA)
		serveIndex(root, w, r)
	})
}

// serveIndex отдаёт index.html без редиректов (обход поведения FileServer для корня).
func serveIndex(root http.FileSystem, w http.ResponseWriter, _ *http.Request) {
	f, err := root.Open("index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusNotFound)
		return
	}
	defer f.Close()
	if stat, err := f.Stat(); err != nil || stat.IsDir() {
		http.Error(w, "index.html not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, f)
}
