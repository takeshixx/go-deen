# deen ![Build & Test](https://github.com/takeshixx/go-deen/workflows/Build%20&%20Test/badge.svg?branch=master)

`deen` is a tool for encoding, decoding, hashing, compressing and formatting
data. It ships as a single static binary with no runtime dependencies and runs
on Linux, Windows and macOS. The same plugins are exposed through several
interfaces: a command-line interface, a desktop GUI built with
[Fyne](https://github.com/fyne-io/fyne), a
[WebAssembly](https://webassembly.org) web interface, and a
[Visual Studio Code extension](extras/vscode-deen).

It is a Go reimplementation of the original Python/PyQt5
[deen](https://github.com/takeshixx/deen).

## Plugins

`deen` provides 60+ plugins across five categories. List them at runtime with
`deen -l` (or `deen -lj` for JSON).

| Category | Plugins |
| --- | --- |
| **codecs** | base32, base64, base85, hex, url, html, unicode, strconv, pem, quoted-printable, rot13 |
| **compressions** | flate, gzip, zlib, bzip2, lzma, lzma2, lzw, brotli, zstd |
| **hashs** | sha1, sha2 (224/256/384/512, 512/224, 512/256), sha3 (224/256/384/512), md4, md5, ripemd160, blake2s/2b/2x, blake3, bcrypt, scrypt, hmac, adler32, crc32/crc32c/crc32k, crc64/crc64-ecma, fnv (32/64/128 and a-variants) |
| **formatters** | json, xml, json2xml, jwt, jwk, jq, protobuf, saml, timestamp |
| **misc** | certPrinter, certCloner |

## Install

```bash
make            # build ./bin/deen (CLI)
make gui        # build with the desktop GUI (-tags gui)
```

## Usage

Run a plugin by name to process data; prefix the name with `.` to reverse the
operation. Input is read from the remaining arguments, stdin, or a file
(`-file`).

```bash
$ deen base64 test                  # encode
dGVzdA==
$ deen .base64 dGVzdA==             # decode
test
$ printf secret | deen sha256       # one-way
2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
```

One-way plugins (such as hashes) return an error if called with a `.` prefix.

### Output

Output is written raw, without a trailing newline, so binary data pipes between
plugins without corruption:

```bash
$ printf test | deen gzip | deen .gzip
test
```

Pass `-N` to append a newline for terminal use (accepted before or after the
plugin name). Decoders ignore surrounding whitespace, and base64 decoding
auto-detects the standard, raw, URL and raw-URL alphabets unless `-strict`,
`-url` or `-raw` is given.

### Listing and help

```bash
$ deen -l                # list plugins by category
$ deen -lj               # list as JSON
$ deen base64 -h         # plugin-specific options
```

## GUI

`make gui` produces a binary that launches the GUI when started without
arguments. On Linux the GUI build needs the X development headers:

```bash
sudo apt install xorg-dev    # Debian/Ubuntu
```

## WebAssembly

deen compiles to WebAssembly and serves the browser interface itself — no
external web server needed. `make web` builds the wasm, embeds it into a
self-contained binary and serves it on `http://127.0.0.1:9090`:

```bash
make web
```

To do it by hand, build with the `webembed` tag and run `serve`:

```bash
go build -tags webembed -o deen ./cmd/deen
deen serve --port 9090
```

`deen serve` can also serve any directory (`--root`) and supports TLS
(`--tls-cert`/`--tls-key`), HTTP basic auth (`--auth-user`/`--auth-pass`) and
request logging (`--log`). The experimental UI itself lives on the `web` branch.

## Writing plugins

Each plugin is a `*types.DeenPlugin` registered in
[`internal/plugins/plugins.go`](internal/plugins/plugins.go). It implements a
single contract:

```go
type DeenPlugin struct {
    Name        string
    Aliases     []string
    Category    string
    Description string

    RegisterFlags func(*flag.FlagSet)                              // optional flags
    Process       func(io.Reader, io.Writer, *flag.FlagSet) error  // forward
    Unprocess     func(io.Reader, io.Writer, *flag.FlagSet) error  // reverse; nil = one-way
}
```

`Process` and `Unprocess` stream from an `io.Reader` to an `io.Writer` and must
return on the first error. A `nil` `Unprocess` marks a one-way plugin. See
[`examples/example_plugin.go`](examples/example_plugin.go) for an annotated
reference, and [`pkg/hashs`](pkg/hashs) for the factory used to build families
of similar plugins with minimal boilerplate.

## Background

`deen` is distributed as a single static binary, avoiding the dependency and
packaging overhead of the original Python/PyQt5 version. Plugins stream their
input, so arbitrarily large inputs are processed with constant memory:

```bash
$ time cat disk.vmdk | deen sha256      # 18 GB input
003b375eeba7e56ae8e1aa03eb2ac6741023478c41fdc077619f144b114e0d02
```

The GUI is built with Fyne and requires no system GUI toolkit at runtime.
