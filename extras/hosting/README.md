# Hosting the deen web interface

The deen web interface runs **entirely in the browser** as WebAssembly — the
server only delivers static files (`index.html`, `main.js`, `style.css`,
`wasm_exec.js`, `deen.wasm`). No user data is processed server-side, so the
deployment is just a hardened static file host.

Build the assets:

```bash
make web-assets   # writes the wasm + wasm_exec.js into internal/web/assets
```

Then copy `internal/web/assets/*` to your web root.

## Options

- **nginx** — see [`nginx.conf`](nginx.conf). Static files behind nginx with a
  strict CSP, HSTS and the correct `application/wasm` type.
- **Caddy** — see [`Caddyfile`](Caddyfile). Automatic HTTPS, compression and
  wasm MIME with minimal config.
- **Built-in server** — `deen serve` can host the embedded UI itself
  (`make web` builds a self-contained binary). Run it behind a proxy with
  [`deen-serve.service`](deen-serve.service), or expose it directly with
  `--tls-cert`/`--tls-key`. It already sets a strict CSP (`--csp` to change),
  `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`,
  `Permissions-Policy`, HSTS (when serving TLS), and serves pre-compressed
  `.br`/`.gz` siblings when present.

## Notes

- Always serve over HTTPS. The built-in basic auth (`--auth-user`/`--auth-pass`)
  is a single shared credential with no rate limiting — fine as a personal gate,
  not for real multi-user auth (use the proxy for that).
- `deen.wasm` is large but compresses well (~15 MB → ~4 MB); enable compression
  (gzip/brotli/zstd) at the proxy or pre-compress the file.
