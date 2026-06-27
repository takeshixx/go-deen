// Package server implements the static HTTP server behind `deen serve`, used
// to serve the WebAssembly interface (or any directory).
package server

import (
	"crypto/subtle"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

// DefaultCSP is a strict Content-Security-Policy suitable for the deen web UI.
// 'wasm-unsafe-eval' is required to compile WebAssembly; everything else is
// restricted to same-origin with no inline scripts or styles.
const DefaultCSP = "default-src 'none'; script-src 'self' 'wasm-unsafe-eval'; " +
	"style-src 'self'; connect-src 'self'; img-src 'self' data:; " +
	"base-uri 'none'; form-action 'none'; frame-ancestors 'none'"

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
	CSP      string // Content-Security-Policy; empty disables the header
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

	var h http.Handler = assetHandler(fsys)
	h = securityHeaders(h, cfg.CSP, cfg.TLSCert != "")
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

// assetHandler serves files, preferring a pre-compressed sibling (.br/.gz) when
// the client accepts it, and ensures the correct wasm media type.
func assetHandler(fsys fs.FS) http.Handler {
	fileServer := wasmContentType(http.FileServerFS(fsys))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (r.Method == http.MethodGet || r.Method == http.MethodHead) &&
			servePrecompressed(w, r, fsys) {
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

// servePrecompressed serves <path>.br or <path>.gz if present and accepted,
// returning true when it did.
func servePrecompressed(w http.ResponseWriter, r *http.Request, fsys fs.FS) bool {
	accept := r.Header.Get("Accept-Encoding")
	type enc struct{ name, ext string }
	var candidates []enc
	if strings.Contains(accept, "br") {
		candidates = append(candidates, enc{"br", ".br"})
	}
	if strings.Contains(accept, "gzip") {
		candidates = append(candidates, enc{"gzip", ".gz"})
	}
	if len(candidates) == 0 {
		return false
	}

	clean := path.Clean(r.URL.Path)
	if strings.HasSuffix(clean, "/") {
		return false
	}
	name := strings.TrimPrefix(clean, "/")

	for _, c := range candidates {
		f, err := fsys.Open(name + c.ext)
		if err != nil {
			continue
		}
		rs, ok := f.(io.ReadSeeker)
		if !ok {
			f.Close()
			continue
		}
		info, err := f.Stat()
		if err != nil {
			f.Close()
			continue
		}
		defer f.Close()

		ctype := mime.TypeByExtension(path.Ext(name))
		if ctype == "" {
			ctype = "application/octet-stream"
		}
		w.Header().Set("Content-Type", ctype)
		w.Header().Set("Content-Encoding", c.name)
		w.Header().Set("Vary", "Accept-Encoding")
		http.ServeContent(w, r, name, info.ModTime(), rs)
		return true
	}
	return false
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

// securityHeaders adds hardening headers. HSTS is only sent when serving TLS.
func securityHeaders(next http.Handler, csp string, tls bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		if csp != "" {
			h.Set("Content-Security-Policy", csp)
		}
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		if tls {
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
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
