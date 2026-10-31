package presentation

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

// NewFrontendHandler serves the built frontend in dir (ADR-0018). A file that
// exists is served as it is. A path without a file extension is a client-side
// route, so it gets index.html and survives a reload; a missing file with an
// extension, such as an old asset, is 404.
func NewFrontendHandler(dir string) http.Handler {
	files := os.DirFS(dir)
	fileServer := http.FileServerFS(files)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		name := strings.TrimPrefix(path.Clean(req.URL.Path), "/")
		if name != "" && name != "index.html" {
			if info, err := fs.Stat(files, name); err == nil && !info.IsDir() {
				fileServer.ServeHTTP(w, req)
				return
			}
			if path.Ext(name) != "" {
				http.NotFound(w, req)
				return
			}
		}
		// index.html names the current hashed assets, so it is always checked again.
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFileFS(w, req, files, "index.html")
	})
}
