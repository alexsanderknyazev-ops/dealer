package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"

	"github.com/dealer/dealer/auth-service/internal/domain"
	"github.com/dealer/dealer/auth-service/internal/service"
)

type httpUserFake struct {
	byEmail map[string]*domain.User
	byID    map[uuid.UUID]*domain.User
}

func (f *httpUserFake) Create(_ context.Context, u *domain.User) error {
	if f.byEmail == nil {
		f.byEmail = make(map[string]*domain.User)
	}
	if f.byID == nil {
		f.byID = make(map[uuid.UUID]*domain.User)
	}
	cp := *u
	f.byEmail[u.Email] = &cp
	f.byID[u.ID] = &cp
	return nil
}

func (f *httpUserFake) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return u, nil
}

func (f *httpUserFake) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return u, nil
}

func testAuthHTTP(t *testing.T) (*http.ServeMux, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := &httpUserFake{byEmail: map[string]*domain.User{}, byID: map[uuid.UUID]*domain.User{}}
	svc := service.NewAuthService(repo, rdb, nil, service.AuthConfig{
		JWTSecret: "hs", AccessTTL: time.Hour, RefreshTTL: time.Hour, RefreshPrefix: "rt:",
	})
	h := NewHandler(svc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	cleanup := func() {
		_ = rdb.Close()
		mr.Close()
	}
	return mux, cleanup
}

func TestAuthHTTP_RegisterLoginMe(t *testing.T) {
	mux, cleanup := testAuthHTTP(t)
	defer cleanup()

	t.Run("options_register", func(t *testing.T) {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodOptions, pathAPIRegister, nil))
		if w.Code != http.StatusNoContent {
			t.Fatal(w.Code)
		}
	})
	t.Run("register_bad_json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, pathAPIRegister, bytes.NewReader([]byte("{")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatal(w.Code)
		}
	})
	body, _ := json.Marshal(map[string]string{"email": "h@http.test", "password": "password123", "name": "H"})
	req := httptest.NewRequest(http.MethodPost, pathAPIRegister, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code, w.Body.String())
	}
	var reg map[string]any
	_ = json.NewDecoder(w.Body).Decode(&reg)
	at := reg["access_token"].(string)

	req2 := httptest.NewRequest(http.MethodGet, pathAPIMe, nil)
	req2.Header.Set("Authorization", "Bearer "+at)
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatal(w2.Code, w2.Body.String())
	}

	req3 := httptest.NewRequest(http.MethodGet, pathAPIMe, nil)
	w3 := httptest.NewRecorder()
	mux.ServeHTTP(w3, req3)
	if w3.Code != http.StatusUnauthorized {
		t.Fatal(w3.Code)
	}

	lb, _ := json.Marshal(map[string]string{"email": "h@http.test", "password": "wrong"})
	req4 := httptest.NewRequest(http.MethodPost, pathAPILogin, bytes.NewReader(lb))
	req4.Header.Set("Content-Type", "application/json")
	w4 := httptest.NewRecorder()
	mux.ServeHTTP(w4, req4)
	if w4.Code != http.StatusUnauthorized {
		t.Fatal(w4.Code)
	}
}

func TestAuthHTTP_RegisterConflict(t *testing.T) {
	mux, cleanup := testAuthHTTP(t)
	defer cleanup()
	regBody := func(email string) []byte {
		b, _ := json.Marshal(map[string]string{"email": email, "password": "password123", "name": "N"})
		return b
	}
	req := httptest.NewRequest(http.MethodPost, pathAPIRegister, bytes.NewReader(regBody("conf@x")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	req2 := httptest.NewRequest(http.MethodPost, pathAPIRegister, bytes.NewReader(regBody("conf@x")))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Fatalf("want 409 got %d", w2.Code)
	}
}

func TestAuthHTTP_Refresh(t *testing.T) {
	mux, cleanup := testAuthHTTP(t)
	defer cleanup()
	body, _ := json.Marshal(map[string]string{"email": "r@http.test", "password": "password123", "name": "R"})
	req := httptest.NewRequest(http.MethodPost, pathAPIRegister, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	var reg map[string]any
	_ = json.NewDecoder(w.Body).Decode(&reg)
	rt := reg["refresh_token"].(string)

	rb, _ := json.Marshal(map[string]string{"refresh_token": ""})
	req2 := httptest.NewRequest(http.MethodPost, pathAPIRefresh, bytes.NewReader(rb))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	if w2.Code != http.StatusBadRequest {
		t.Fatal(w2.Code)
	}

	rb2, _ := json.Marshal(map[string]string{"refresh_token": "invalid"})
	req3 := httptest.NewRequest(http.MethodPost, pathAPIRefresh, bytes.NewReader(rb2))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	mux.ServeHTTP(w3, req3)
	if w3.Code != http.StatusUnauthorized {
		t.Fatal(w3.Code)
	}

	rb3, _ := json.Marshal(map[string]string{"refresh_token": rt})
	req4 := httptest.NewRequest(http.MethodPost, pathAPIRefresh, bytes.NewReader(rb3))
	req4.Header.Set("Content-Type", "application/json")
	w4 := httptest.NewRecorder()
	mux.ServeHTTP(w4, req4)
	if w4.Code != http.StatusOK {
		t.Fatal(w4.Code, w4.Body.String())
	}
}

func TestSPAFileServer(t *testing.T) {
	fs := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html></html>")},
		"app.js":     &fstest.MapFile{Data: []byte("//x")},
	}
	h := SPAFileServer(http.FS(fs))

	t.Run("index", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		if w.Code != http.StatusOK {
			t.Fatal(w.Code)
		}
		b, _ := io.ReadAll(w.Body)
		if string(b) != "<html></html>" {
			t.Fatal(string(b))
		}
	})
	t.Run("static_file", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/app.js", nil))
		if w.Code != http.StatusOK {
			t.Fatal(w.Code)
		}
	})
	t.Run("spa_fallback", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/unknown-route", nil))
		if w.Code != http.StatusOK {
			t.Fatal(w.Code)
		}
	})
	t.Run("method_not_allowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/", nil))
		if w.Code != http.StatusMethodNotAllowed {
			t.Fatal(w.Code)
		}
	})
}
