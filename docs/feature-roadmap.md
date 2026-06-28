# deen feature roadmap

This is a local backlog of GUI, web UI, and plugin ideas captured during the
June 2026 codebase review.

## GUI and web UI

- Saved chains / presets: save and load source-optional pipeline JSON.
- Shareable web URLs: encode chain configuration in the URL hash.
- Auto-detect transform suggestions: score likely next steps such as base64,
  hex, URL encoding, JWT, gzip, PEM, JSON, XML, and common compression magic.
- Binary-safe input/output modes: text, hex, base64, and raw file handling.
- Split output viewer: text, hex dump, byte stats, and structured previews.
- Step reorder / duplicate / disable controls.
- Undo / redo for source, step, direction, option, and manual output edits.
- Side-by-side compare for input vs output or any two step outputs.
- Plugin search palette over names, aliases, descriptions, and use cases.
- Better option controls: numeric controls, select boxes, and secret fields.
- Input/output metadata: byte length, UTF-8 validity, line count, entropy,
  checksums, and compression ratio.
- Web drag-and-drop files with original filename-aware downloads.
- Export equivalent command line for a chain.
- Batch mode for applying a chain to multiple files.
- Plugin catalog examples with sample input/output and security notes.

## Plugin ideas

- AES encrypt/decrypt: GCM, CBC, CTR.
- ChaCha20-Poly1305 encrypt/decrypt.
- RSA / ECDSA / Ed25519 sign and verify.
- JWK / JWKS formatter and PEM conversion helpers.
- ASN.1 DER parser.
- SAML decoder / formatter.
- UUID tools.
- Timestamp tools.
- DNS wire format / DNS name codec.
- Protobuf wire decoder.
- MessagePack formatter.
- CBOR formatter.
- YAML formatter.
- TOML formatter.
- CSV / TSV tools.
- Regex extract / replace.
- Diff plugin.
- Entropy / byte frequency analyzer.
- Magic / file-type detector.
- QR encode / decode.

## Near-term implementation picks

- Shared undo / redo in `internal/pipeline`.
- Binary-safe output display modes in desktop GUI and web UI.
- One-way protobuf wire decoder plugin.

## Packaging and distribution

### Homebrew

Status: keep local for now. Test the current beta before publishing through
Homebrew.

Plan:

- Start with a personal tap, likely `takeshixx/homebrew-tap`, so users can run
  `brew install takeshixx/tap/deen`.
- After the tap is installed once with `brew tap takeshixx/tap`, users can run
  `brew install deen`.
- Target the CLI first. A GUI build should be a separate cask later, likely
  `brew install --cask deen`, once we have signed/notarized macOS app assets.
- Use a stable, non-beta release before submitting to `homebrew-core`.
- Add `Formula/deen.rb` with:
  - `desc`, `homepage`, `license`, release `url`, and `sha256`.
  - `depends_on "go" => :build`.
  - `go build -mod=readonly -trimpath` for `./cmd/deen`.
  - A small `test do` block that checks `deen -version` and a simple transform.
- Validate formula updates with:
  - `brew audit --strict --online deen`
  - `brew install --build-from-source deen`
  - `brew test deen`
- Maintenance per release:
  - Cut and verify the GitHub release.
  - Update formula URL/version and tarball SHA256.
  - Run local Homebrew audit/install/test.
  - Push the tap update.
  - Optionally add tap CI and bottles once releases are frequent enough.

Before implementation:

- Fix or account for the release workflow/binary version behavior that produced
  `v3.4.0-beta-master` during the beta release workflow.
- Decide whether the first Homebrew publication should use the next stable
  release, for example `v3.4.0`, instead of the current beta.
