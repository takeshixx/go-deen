// Package server implements the static HTTP server behind `deen serve`, used
// to serve the WebAssembly interface (or any directory).
package server

import (
	"crypto/subtle"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Config configures the deen web server.
type Config struct {
	Host     string
	Port     int
	Root     string // directory to serve; overrides Assets when non-empty
	Assets   fs.FS  // embedded assets, used when Root is empty
	TLSCert  string
	TLSKey   string
	AuthUser string // when non-empty, HTTP basic auth is required
	AuthPass string
	Log      bool
}

// Handler builds the HTTP handler for the given configuration. It is separate
// from Run so it can be exercised in tests.
func Handler(cfg Config) (http.Handler, error) {
	var fsys fs.FS
	switch {
	case cfg.Root != "":
		fsys = os.DirFS(cfg.Root)
	case cfg.Assets != nil:
		fsys = cfg.Assets
	default:
		return nil, fmt.Errorf("no web assets embedded in this build; pass --root <dir> (or build with -tags webembed)")
	}

	var h http.Handler = wasmContentType(http.FileServerFS(fsys))
	if cfg.AuthUser != "" {
		h = basicAuth(h, cfg.AuthUser, cfg.AuthPass)
	}
	if cfg.Log {
		h = logRequests(h)
	}
	return h, nil
}

// Run builds the handler and serves until an error occurs.
func Run(cfg Config) error {
	handler, err := Handler(cfg)
	if err != nil {
		return err
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	scheme := "http"
	if cfg.TLSCert != "" {
		scheme = "https"
	}
	log.Printf("deen: serving %s://%s", scheme, addr)
	if cfg.TLSCert != "" {
		return srv.ListenAndServeTLS(cfg.TLSCert, cfg.TLSKey)
	}
	return srv.ListenAndServe()
}

// wasmContentType ensures .wasm responses carry the correct media type, which
// browsers require for WebAssembly.instantiateStreaming.
func wasmContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("Content-Type", "application/wasm")
		}
		next.ServeHTTP(w, r)
	})
}

// basicAuth enforces HTTP basic authentication using constant-time comparison.
func basicAuth(next http.Handler, user, pass string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(u), []byte(user)) != 1 ||
			subtle.ConstantTimeCompare([]byte(p), []byte(pass)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="deen"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// logRequests writes a concise access log line per request.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
