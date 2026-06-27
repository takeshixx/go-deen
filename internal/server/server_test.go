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

func TestSecurityHeaders(t *testing.T) {
	h, err := Handler(Config{Root: tempRoot(t), CSP: DefaultCSP, TLSCert: "cert.pem"})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	hdr := rec.Header()
	if hdr.Get("Content-Security-Policy") != DefaultCSP {
		t.Errorf("CSP = %q", hdr.Get("Content-Security-Policy"))
	}
	if hdr.Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing nosniff")
	}
	if hdr.Get("X-Frame-Options") != "DENY" {
		t.Error("missing X-Frame-Options DENY")
	}
	if hdr.Get("Strict-Transport-Security") == "" {
		t.Error("HSTS should be set when serving TLS")
	}
}

func TestNoHSTSWithoutTLS(t *testing.T) {
	h, _ := Handler(Config{Root: tempRoot(t), CSP: DefaultCSP})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("Strict-Transport-Security") != "" {
		t.Error("HSTS must not be set without TLS")
	}
}

func TestPrecompressedServing(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("PLAIN"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.js.gz"), []byte("GZIPPED"), 0o644); err != nil {
		t.Fatal(err)
	}
	h, _ := Handler(Config{Root: dir})

	// Client accepts gzip -> serves the .gz with Content-Encoding.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected gzip encoding, got %q", rec.Header().Get("Content-Encoding"))
	}
	if rec.Body.String() != "GZIPPED" {
		t.Errorf("expected precompressed body, got %q", rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct == "" || ct == "application/gzip" {
		t.Errorf("Content-Type should reflect the original extension, got %q", ct)
	}

	// No Accept-Encoding -> serves the plain file.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/app.js", nil))
	if rec.Header().Get("Content-Encoding") != "" || rec.Body.String() != "PLAIN" {
		t.Errorf("expected plain file, got encoding=%q body=%q", rec.Header().Get("Content-Encoding"), rec.Body.String())
	}
}
