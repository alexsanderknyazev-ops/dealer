package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

const (
	testSPAIndexHTML = "<html></html>"
	testSPAJS        = "//x"
)

func TestSPAFileServer(t *testing.T) {
	fs := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(testSPAIndexHTML)},
		"app.js":     &fstest.MapFile{Data: []byte(testSPAJS)},
	}
	h := SPAFileServer(http.FS(fs))

	t.Run("index", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		if w.Code != http.StatusOK {
			t.Fatal(w.Code)
		}
		b, _ := io.ReadAll(w.Body)
		if string(b) != testSPAIndexHTML {
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
