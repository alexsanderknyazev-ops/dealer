package httpapi

import (
	"io"
	"net/http"
	"path"
	"strings"
)

const (
	headerContentType = "Content-Type"
	mimeHTMLUTF8      = "text/html; charset=utf-8"
)

// SPAFileServer отдаёт статику; для путей без файла отдаёт index.html (SPA routing).
func SPAFileServer(root http.FileSystem) http.Handler {
	fs := http.FileServer(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveSPAOrStatic(root, fs, w, r)
	})
}

func serveSPAOrStatic(root http.FileSystem, fs http.Handler, w http.ResponseWriter, r *http.Request) {
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
	if tryServeExistingFile(root, fs, w, r, name) {
		return
	}
	serveIndex(root, w, r)
}

func tryServeExistingFile(root http.FileSystem, fs http.Handler, w http.ResponseWriter, r *http.Request, name string) bool {
	f, err := root.Open(name)
	if err != nil {
		return false
	}
	_ = f.Close()
	fs.ServeHTTP(w, r)
	return true
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
	w.Header().Set(headerContentType, mimeHTMLUTF8)
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, f)
}
