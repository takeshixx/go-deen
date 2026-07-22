# deen [![Build & Test](https://github.com/takeshixx/go-deen/actions/workflows/deen-ci.yml/badge.svg?branch=master)](https://github.com/takeshixx/go-deen/actions/workflows/deen-ci.yml)

`deen` is a tool for encoding, decoding, hashing, compressing and formatting
data. It ships as a single static binary with no runtime dependencies and runs
on Linux, Windows and macOS. The same plugins are exposed through several
interfaces: a command-line interface, a desktop GUI built with
[Fyne](https://github.com/fyne-io/fyne), a
[WebAssembly](https://webassembly.org) web interface available at
[deen.adversec.com](https://deen.adversec.com), and a
[Visual Studio Code extension](extras/vscode-deen).

The WebAssembly interface processes data entirely in the browser. Sensitive
input stays client-side instead of being sent to a backend service. See
[WebAssembly](#webassembly) for instructions to run the web interface locally.

It is a Go reimplementation of the original Python/PyQt5
[deen](https://github.com/takeshixx/deen).

## Install

### Homebrew

Install the `deen` CLI on macOS or Linux from the official deen tap:

```bash
brew install takeshixx/tap/deen
```

Verify the installation with `deen --version`. To use the shorter formula name
for future installs and upgrades, add the tap first:

```bash
brew tap takeshixx/tap
brew install deen
```

### Arch Linux (AUR)

Install the CLI and desktop GUI from the
[`deen-git`](https://aur.archlinux.org/packages/deen-git) AUR package with an
AUR helper such as `yay`:

```bash
yay -S deen-git
```

Or build and install the package manually:

```bash
git clone https://aur.archlinux.org/deen-git.git
cd deen-git
makepkg -si
```

### Build from source

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

### Transform chains

Plugins can be composed into repeatable transform chains. In the CLI, a chain is
just a shell pipeline:

```bash
$ printf '%s' '%7B%22role%22%3A%22admin%22%7D' | deen .url | deen json
{
    "role": "admin"
}

$ printf H4sIAAAAAAAA/ypJLS4BBAAA//8Mfn/YBAAAAA== | deen .base64 | deen .gzip
test
```

The GUI and web UI provide the same model with editable steps, previews,
examples, undo/redo, and import/export for chain JSON files. Chain exports are
useful for saving an analysis workflow or sharing it with someone who will
provide their own input data.

Saved Web/GUI chains can also be run from the CLI with the `chain` subcommand.
For example, this source-free chain recipe URL-decodes input and formats the
result as JSON:

```json
{
  "version": 1,
  "steps": [
    {
      "plugin": "url",
      "unprocess": true
    },
    {
      "plugin": "json"
    }
  ]
}
```

Save it as `decode-url-json.json`, then run it against stdin:

```bash
$ printf '%s' '%7B%22role%22%3A%22admin%22%7D' | deen chain -stdin decode-url-json.json
{
    "role": "admin"
}
```

Use `-input-file`, `-stdin`, or trailing text to override a source saved in the
chain. Without an override, `deen chain` uses the source embedded in the JSON.

### Advanced CLI examples

Plugin-specific flags follow the plugin name; use `deen <plugin> -h` to discover
them. Files can be read directly, including binary input:

```bash
deen sha256 -file disk.img -N
deen .base64 -file payload.b64 > payload.bin
deen certPrinter -file server.pem
```

Extract and decode a PowerShell `-EncodedCommand` from a process log:

```bash
$ printf '%s' 'powershell -EncodedCommand VwByAGkAdABlAC0ASABvAHMAdAAgAGgAaQA=' \
  | deen regex -re '(?i)-encodedcommand\s+([A-Za-z0-9+/=]+)' -group 1 \
  | deen .base64 \
  | deen .utf16le
Write-Host hi
```

Plugin options remain attached to each step in a shell pipeline. This example
performs an authenticated AES-GCM round trip; the fixed nonce is only for a
reproducible demonstration—use a unique nonce for each real encryption:

```bash
$ printf sensitive \
  | deen aes -key 000102030405060708090a0b0c0d0e0f -nonce 000102030405060708090a0b -aad case=42 \
  | deen .aes -key 000102030405060708090a0b0c0d0e0f -nonce 000102030405060708090a0b -aad case=42
sensitive
```

See [real-world examples](docs/real-world-examples.md) for longer CloudFront,
Kubernetes, PowerShell, Protobuf, and JWKS/X.509 analysis chains.
The [URL Parts schema](docs/urlparts.md) documents loss-aware URL inspection,
editing, reconstruction, local indicators, nested redirect extraction,
defanging, and jq usage.

### Listing and help

```bash
$ deen -l                # list plugins by category
$ deen -lj               # list as JSON
$ deen base64 -h         # plugin-specific options
```

### Agent-friendly inspection

`inspect` and `detect` emit structured JSON for agents, scripts and future MCP
integrations. Output is capped by default so large or binary inputs do not flood
an agent context.

```bash
$ deen inspect '{"ok":true}'
$ printf '%s' '%7B%22ok%22%3Atrue%7D' | deen detect
```

`inspect` returns metadata, SHA-256, MIME sniffing, a safe preview, structured
preview text when available, and likely next transforms. `detect` focuses on
ranked one-step and multi-step decode suggestions that can be turned into deen
chains.

### MCP server

`deen mcp serve` runs a local stdio MCP server for coding agents. It exposes
tools for inspection, detection, single transforms, saved-chain execution,
bounded result-range reads, and plugin discovery. It also exposes MCP resources
for plugin/example catalogs and prompts for common triage, decode, binary
inspection, and chain-explanation workflows. The server is local-only: it does
not perform network access, shell execution, or writes.

Example Claude Code project configuration:

```json
{
  "mcpServers": {
    "deen": {
      "type": "stdio",
      "command": "deen",
      "args": ["mcp", "serve"]
    }
  }
}
```

## Plugins

`deen` provides 60+ plugins across six categories. List them at runtime with
`deen -l` (or `deen -lj` for JSON).

| Category | Plugins |
| --- | --- |
| **codecs** | base32, base64, base85, hex, url, html, unicode, strconv, pem, quoted-printable, rot13 |
| **compressions** | flate, gzip, zlib, bzip2, lzma, lzma2, lzw, brotli, zstd |
| **hashs** | sha1, sha2 (224/256/384/512, 512/224, 512/256), sha3 (224/256/384/512), md4, md5, ripemd160, blake2s/2b/2x, blake3, bcrypt, scrypt, hmac, adler32, crc32/crc32c/crc32k, crc64/crc64-ecma, fnv (32/64/128 and a-variants) |
| **formatters** | json, xml, json2xml, toml, urlparts, jwt, jwk, jq, protobuf, msgpack, cbor, yaml, csv/tsv, qr, saml, timestamp |
| **misc** | asn1, dns, uuid, entropy, magic, regex, aes, chacha20poly1305, sign/verify, certPrinter, certCloner |
| **arithmetic** | xor, add, sub, not |

Recent utility plugins add structured binary and security workflows:

- `msgpack`, `cbor`, `protobuf`, `asn1`, `dns`, `magic` and `qr` inspect or decode common binary payloads.
- `yaml`, `toml`, `csv`/`tsv`, `regex`, `uuid` and `entropy` cover day-to-day data cleanup and inspection.
- `aes`, `chacha20poly1305` and `sign` support encryption, decryption, signing and verification. Binary keys, nonces and signatures can be supplied as hex or Base64; AES-GCM supports configurable tag lengths and an explicit unsafe verification bypass for research, and AES-CBC supports PKCS#7 or unpadded block data.

## GUI

`make gui` produces a binary that launches the GUI when started without
arguments. The Fyne 2.8 desktop build supports Windows 10 and later, macOS
10.15 and later, and Linux desktops using X11 or Wayland. Fyne selects the
available Linux window protocol at runtime.

Forward `urlparts` steps include a structured URL editor for modifying the
authority, path segments, ordered query parameters, and fragment while keeping
the JSON output and downstream steps synchronized.

On Debian, Ubuntu, and Raspberry Pi OS, install the graphics development
headers before building:

```bash
sudo apt install gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev
```

## WebAssembly

deen compiles to WebAssembly and serves the browser interface itself — no
external web server needed. `make web` builds the wasm, embeds it into a
self-contained binary and serves it on `http://127.0.0.1:9090`:

```bash
make web
```

Open `http://127.0.0.1:9090`, paste or load input data, then add transforms or
start from one of the built-in examples. The web UI can download results, export
and import chain JSON, and copy a share link for the current chain. Share links
include the transform recipe in the URL hash, but intentionally do not include
the source input. Forward `urlparts` steps provide the same structured URL editor
as the desktop GUI.

To do it by hand, build with the `webembed` tag and run `serve`:

```bash
go build -tags webembed -o deen ./cmd/deen
deen serve --port 9090
```

`deen serve` can also serve any directory (`--root`) and supports TLS
(`--tls-cert`/`--tls-key`), HTTP basic auth (`--auth-user`/`--auth-pass`) and
request logging (`--log`).

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
