package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func tempRoot(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>deen</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "deen.wasm"), []byte("\x00asm"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestHandlerServesFiles(t *testing.T) {
	h, err := Handler(Config{Root: tempRoot(t)})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("index.html: got %d", rec.Code)
	}
	if rec.Body.String() != "<html>deen</html>" {
		t.Errorf("unexpected body: %q", rec.Body.String())
	}
}

func TestHandlerWasmContentType(t *testing.T) {
	h, err := Handler(Config{Root: tempRoot(t)})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/deen.wasm", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/wasm" {
		t.Errorf("wasm content-type = %q, want application/wasm", ct)
	}
}

func TestHandlerBasicAuth(t *testing.T) {
	h, err := Handler(Config{Root: tempRoot(t), AuthUser: "op", AuthPass: "secret"})
	if err != nil {
		t.Fatal(err)
	}

	// No credentials -> 401.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("no creds: got %d, want 401", rec.Code)
	}

	// Correct credentials -> 200.
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("op", "secret")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("good creds: got %d, want 200", rec.Code)
	}

	// Wrong password -> 401.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("op", "wrong")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("bad creds: got %d, want 401", rec.Code)
	}
}

func TestHandlerNoAssets(t *testing.T) {
	if _, err := Handler(Config{}); err == nil {
		t.Error("expected an error when neither Root nor Assets is set")
	}
}
