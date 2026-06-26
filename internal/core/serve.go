package core

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/takeshixx/deen/internal/server"
	"github.com/takeshixx/deen/internal/web"
	"github.com/takeshixx/deen/pkg/helpers"
)

// runServe handles the "serve" subcommand: it starts an HTTP server for the
// WebAssembly interface (or any directory). Returns a process exit code.
func runServe() int {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of serve:\n\n")
		fmt.Fprintf(os.Stderr, "Serve the deen WebAssembly interface over HTTP.\n\n")
		fs.PrintDefaults()
	}
	host := fs.String("host", "127.0.0.1", "listening host")
	port := fs.Int("port", 9090, "listening port")
	root := fs.String("root", "", "serve this directory instead of the embedded assets")
	tlsCert := fs.String("tls-cert", "", "TLS certificate file (enables HTTPS)")
	tlsKey := fs.String("tls-key", "", "TLS private key file")
	authUser := fs.String("auth-user", "", "require HTTP basic auth with this user")
	authPass := fs.String("auth-pass", "", "HTTP basic auth password (random if omitted)")
	logReq := fs.Bool("log", false, "log requests")
	fs.Parse(helpers.RemoveBeforeSubcommand(os.Args, "serve"))

	pass := *authPass
	if *authUser != "" && pass == "" {
		pass = randomPassword()
		log.Printf("deen: basic auth credentials %s:%s", *authUser, pass)
	}

	if *root == "" && web.FS() == nil {
		fmt.Fprintln(os.Stderr, "deen: this binary has no embedded web assets; pass --root <dir> or rebuild with -tags webembed (see 'make web')")
		return 1
	}

	err := server.Run(server.Config{
		Host:     *host,
		Port:     *port,
		Root:     *root,
		Assets:   web.FS(),
		TLSCert:  *tlsCert,
		TLSKey:   *tlsKey,
		AuthUser: *authUser,
		AuthPass: pass,
		Log:      *logReq,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "deen:", err)
		return 1
	}
	return 0
}

// randomPassword returns a short random password for basic auth.
func randomPassword() string {
	b := make([]byte, 9)
	if _, err := rand.Read(b); err != nil {
		return "deen-changeme"
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
