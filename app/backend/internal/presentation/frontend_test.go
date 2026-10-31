package presentation

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestFrontendDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o750); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"index.html":       "<html>app</html>",
		"favicon.svg":      "<svg/>",
		"assets/app-1.js":  "console.log(1)",
		"assets/app-1.css": "body{}",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestFrontendHandler(t *testing.T) {
	handler := NewFrontendHandler(newTestFrontendDir(t))

	tests := []struct {
		name        string
		path        string
		want        int
		wantBody    string
		wantNoCache bool
	}{
		{name: "root", path: "/", want: http.StatusOK, wantBody: "<html>app</html>", wantNoCache: true},
		{name: "client-side route", path: "/races/123/running-order", want: http.StatusOK, wantBody: "<html>app</html>", wantNoCache: true},
		{name: "client-side route with a query", path: "/auth/complete?code=abc", want: http.StatusOK, wantBody: "<html>app</html>", wantNoCache: true},
		{name: "asset", path: "/assets/app-1.js", want: http.StatusOK, wantBody: "console.log(1)"},
		{name: "file at the root", path: "/favicon.svg", want: http.StatusOK, wantBody: "<svg/>"},
		{name: "missing asset", path: "/assets/app-0.js", want: http.StatusNotFound},
		{name: "directory", path: "/assets", want: http.StatusOK, wantBody: "<html>app</html>", wantNoCache: true},
		// net/http rejects a path with "..", so nothing outside dir is served.
		{name: "path escaping the directory", path: "/../../etc/passwd", want: http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, testCase.path, nil))

			if rec.Code != testCase.want {
				t.Fatalf("status = %d, want %d", rec.Code, testCase.want)
			}
			if testCase.wantBody != "" && strings.TrimSpace(rec.Body.String()) != testCase.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), testCase.wantBody)
			}
			if got := rec.Header().Get("Cache-Control") == "no-cache"; got != testCase.wantNoCache {
				t.Errorf("Cache-Control no-cache = %v, want %v", got, testCase.wantNoCache)
			}
		})
	}
}
