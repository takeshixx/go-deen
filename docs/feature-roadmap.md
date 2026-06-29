# deen feature roadmap

This is a local backlog of GUI, web UI, packaging, and plugin ideas. It was
refreshed on 2026-06-29 after the README, web UI, Examples, URL routing, and VS
Code extension updates.

## Completed recently

### GUI and web UI

- Saved chains / presets: import and export pipeline JSON, including
  source-optional recipe links.
- Shareable web URLs: support hash routes for Home, Examples, Plugins, About,
  chain recipes, example links, plugin links, and search links.
- Example links: Examples can be searched, deep-linked, expanded lazily, and
  copied with per-example Copy link buttons.
- Plugin links: plugin catalog search and per-plugin Copy link buttons.
- Web drag-and-drop file import on the source card, preserving filename-aware
  result downloads.
- Lazy web examples: initial Examples tab render builds title-level cards only;
  details and previews load on demand.
- Auto-detect transform suggestions for likely next steps.
- Split output viewers for text, hex, strings, image previews, metadata, and
  structured previews.
- Step reorder, duplicate, disable, and delete controls in both GUI and web UI.
- Undo / redo for source, step, direction, option, and manual output edits.
- Side-by-side compare for pipeline data.
- Typed plugin option controls for booleans, selects, numeric defaults, secret
  fields, and multiline jq filters.
- Input/output metadata, including byte length, UTF-8 status, line count,
  entropy, checksums, and compression ratio.
- Export equivalent command line for a chain.
- Plugin catalog examples with sample input/output and references.
- VS Code extension refresh: stdin-based `deen` execution, configurable binary,
  safer output handling, current command coverage, and Node 22 build docs.
- Browser route regression test for web routes, direct example/plugin links,
  copy-link behavior, QR preview, and source file drag-and-drop.
- Browser regression expansion for legacy `#chain=...` recipe links, clipboard
  fallback behavior, filename-aware result downloads, and narrow viewport
  routing.
- CLI/core tests for `deen chain` input precedence, saved source handling,
  newline behavior, and step error reporting.
- Option metadata regression tests for secret fields, select choices, and
  numeric controls.
- Option metadata audit for plugin-aware secret handling, JWT/certificate/LZW
  select controls, multiline JSON options, and catalog-wide renderability
  checks.
- Step move controls now disable impossible boundary moves in the GUI and web
  UI.
- Web step headers group editing actions, style disabled controls clearly, and
  keep step actions contained on narrow/mobile layouts.
- Web drag-and-drop rejects directories and multi-file drops with inline
  feedback, preserves the current source after invalid drops, and shows a busy
  state while files are read.
- Detect next now proposes bounded multi-step decode chains, including layered
  URL/Base64/compression/format workflows, with confidence and preview text.
- Agent-friendly CLI JSON: `deen inspect` and `deen detect` expose capped
  metadata, previews, hashes, structured previews, and decode suggestions for
  Claude Code/MCP-style integrations.
- Initial MCP stdio server: `deen mcp serve` exposes inspect, detect_next,
  transform, run_chain, read_result_range, list_plugins, and search_plugins
  tools.
- MCP resources and prompts expose plugin/example catalogs, session result
  previews, and common agent workflows.

### Plugins

Implemented plugin families and tools:

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
- Entropy / byte frequency analyzer.
- Magic / file-type detector.
- Binary structure inspector for ELF, PE, and Mach-O executables.
- QR encode / decode.

## Open priorities

### 1. Agentic Tooling / MCP Integration

Make deen a local data workbench for coding agents such as Claude Code.

Scope:

- Expand `deen mcp serve` beyond the initial stdio tools as needed by real
  Claude Code workflows.
- Keep schema-stable structured content and short human summaries for every
  tool.
- Keep improving capped previews plus in-session result references for large
  outputs.
- Publish Claude Code setup docs and, later, a plugin bundle with skills.
- Keep all agent tooling local-only by default: no network, no shell execution,
  no writes unless explicitly requested.

### 2. Automated Detect Next / Decode Assistant Follow-Up

Detect next now finds useful multi-step decode paths, not only immediate
single-step hints. Continue hardening it for security research and
reverse-engineering workflows.

Scope:

- Expand candidate chains beyond the current bounded URL/Base64/hex/HTML/PEM/
  Unicode/gzip/zlib search.
- Improve confidence scoring with stronger format-specific signals and branch
  ranking.
- Consider an explicit "Apply best chain" action if user testing shows it is
  more ergonomic than the current ranked suggestion list.
- Keep analysis local and bounded: cap bytes inspected, avoid network lookups,
  and stop expensive speculative branches early.
- Preserve the current simple suggestions as quick actions when the user only
  wants one transform.

### 3. Batch Mode

Apply a saved chain to multiple files from the CLI and, later, from the GUI/web
UI.

Initial CLI scope:

- Accept a chain JSON file and one or more input files.
- Write outputs to a destination directory or predictable filename suffix.
- Preserve binary-safe output.
- Report per-file errors without stopping the entire batch unless requested.

### 4. Batch Mode Tests and CLI/Core Coverage

`internal/core` now has direct `deen chain` coverage. Keep expanding it around
batch mode and subcommand dispatch as new CLI behavior lands.

Scope:

- Table tests for CLI flag parsing and subcommand dispatch.
- Batch-mode tests before/with the batch implementation.

### 5. Browser/UI Regression Coverage

The Playwright regression covers web routing, legacy chain links,
example/plugin copy links, clipboard fallback behavior, QR preview, source
drag-and-drop, filename-aware downloads, and narrow viewport routing. Keep this
as a growing smoke suite for user-facing web behavior.

Scope:

- Large-file busy state.
- Broader narrow/mobile layout coverage for editing controls.
- Route coverage for future web features as they are added.

### 6. Option Metadata Completeness

Scope:

- Continue extending metadata regression tests as new plugins/options are added.
- Consider debounce for high-cost text option edits in GUI/web UI.

### 7. Step Editing UX Polish

Step reorder, duplicate, disable, and delete exist in the GUI and web UI, and
impossible boundary moves are disabled. Web step headers now keep editing
actions grouped and contained on narrow/mobile layouts. Keep this as polish only
if user testing shows friction.

Scope:

- Consider drag handles or keyboard shortcuts for reorder.
- Keep undo/redo behavior covered around each action.

### 8. Diff Plugin

Add a plugin for textual and/or binary diff workflows.

Open questions:

- Whether diff should be a normal plugin, a compare-view enhancement, or both.
- How to feed two inputs in CLI, GUI, and web UI consistently.

### 9. Web Drag-and-Drop Polish

Basic source-file drag-and-drop exists, and invalid directory or multi-file drops
now produce inline feedback without replacing the current source. Remaining work
is future batch-mode behavior.

Scope:

- Decide how multiple dropped files should behave after batch mode exists.

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
