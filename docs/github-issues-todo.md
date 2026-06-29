# GitHub Issues TODO Plan

Reviewed on 2026-06-29 from `takeshixx/go-deen` open GitHub issues.
Implementation status refreshed after commits:

- `2382bcf` Update docs, lazy examples, and VS Code extension
- `82f33dd` Add web URL routing and copy links
- local follow-ups through Binary Inspector and Detect next planning

## Completed Issues

### Issue #23: Lazy load examples

URL: https://github.com/takeshixx/go-deen/issues/23

Implemented:

- [x] Split web example rendering into lightweight title cards plus lazy detail
  rendering.
- [x] Store only `pipeline.Example` references and minimal DOM nodes during the
  initial `renderExamples` pass.
- [x] Populate description, chain summary, source summary, expected-result text,
  load button, preview button, and copy-link button only after first open.
- [x] Keep output preview execution behind a second lazy boundary.
- [x] Preserve search with `pipeline.ExampleMatches`.
- [x] Update web example benchmarks for initial title filtering, detail
  preparation, and result preview work.
- [x] Manually verify Examples search, expand/collapse, Load example, QR image
  preview, no-results state, and example Copy link behavior.

Acceptance criteria:

- Opening the Examples tab builds only title-level UI for each example.
- Example details and previews are populated once, on demand.
- Search behavior and example loading remain unchanged from a user's point of
  view.
- Example links such as `#examples/qr-payload-fixture` are copyable and
  directly open the target example.

### Issue #24: Update README.md

URL: https://github.com/takeshixx/go-deen/issues/24

Implemented:

- [x] Add a "Transform chains" subsection after basic CLI usage.
- [x] Include concrete chain examples, including shell pipes and the
  `deen chain` subcommand for saved Web/GUI chain JSON.
- [x] Expand the WebAssembly section with a local `make web` flow.
- [x] Document Copy link behavior for web chain recipe links.
- [x] Mention exported/imported chain JSON files.
- [x] Ensure examples use current plugin names and verified commands.

Acceptance criteria:

- A new user can understand how to compose and reuse chains from the README.
- The local web workflow is documented with `make web`.
- Copy-link privacy semantics are explicit: source data is not embedded.

### Issue #25: Update VS Code extension

URL: https://github.com/takeshixx/go-deen/issues/25

Implemented:

- [x] Define the minimal extension model: command-palette transforms over
  selected text/full document, using an installed `deen` binary.
- [x] Add `deen.binaryPath`, defaulting to `deen`.
- [x] Replace shell-based `child.exec` with stdin-based `spawn`.
- [x] Show errors through `vscode.window.showErrorMessage` and a `deen` output
  channel.
- [x] Replace unsafe webview rendering with VS Code text documents.
- [x] Consolidate command registration around one typed command table.
- [x] Fix command typos and command metadata mismatches.
- [x] Add broader current command coverage for low-risk codecs, formatters,
  compressions, checksums, entropy, and magic.
- [x] Cover command metadata with TypeScript compile/lint plus a consistency
  check.
- [x] Refresh extension README build instructions and the Makefile image to
  Node 22.

Acceptance criteria:

- Extension commands invoke `deen` safely through stdin.
- Existing contributed commands compile, register, and map to the intended
  plugin names.
- User-facing output and errors are handled without unsafe HTML or thrown async
  exceptions.
- Extension docs and build tooling match the updated implementation.

## Completed Follow-Up: Web URL Routing

Implemented after issues #23-#25:

- [x] Hash routes for `#home`, `#examples`, `#plugins`, and `#about`.
- [x] Backwards-compatible `#chain=...` recipe links.
- [x] Example search and direct example routes, such as
  `#examples?search=qr` and `#examples/qr-payload-fixture`.
- [x] Plugin search and direct plugin routes, such as
  `#plugins?search=base64` and `#plugins/base64`.
- [x] Copy link buttons for example and plugin cards.
- [x] Browser back/forward support for route changes.

## Completed Follow-Up: Coverage and UX Hardening

Implemented after the URL routing work:

- [x] Add a Playwright browser regression for web routes, Examples search,
  direct example links, QR preview, example/plugin Copy link buttons, direct
  plugin links, and source file drag-and-drop.
- [x] Expand the browser regression for backwards-compatible `#chain=...`
  recipe links, Clipboard API fallback behavior, filename-aware result
  downloads, and narrow viewport routes.
- [x] Add `npm --prefix extras/web-perf test` plus `make test-web-browser` for
  the browser smoke suite.
- [x] Refactor `deen chain` execution behind an injectable test helper.
- [x] Add core tests for `deen chain` input precedence, stdin override, saved
  source handling, newline behavior, missing chain files, and step errors.
- [x] Expand option metadata tests for security-sensitive secret fields, select
  choices, and numeric controls.
- [x] Audit and improve option metadata for plugin-aware secret handling,
  JWT/certificate/LZW select controls, multiline JSON options, and
  catalog-wide renderability.
- [x] Disable impossible step move buttons at chain boundaries in GUI and web
  UI.
- [x] Polish web step editing headers by grouping actions, styling disabled
  controls, and adding a narrow/mobile regression check.
- [x] Polish web drag-and-drop by rejecting directories and multi-file drops
  with inline feedback, preserving the current source after invalid drops, and
  showing a file-read busy state.
- [x] Add a Binary Inspector plugin for ELF, PE, and Mach-O structure summaries.
- [x] Surface executable inputs in Detect next with a Binary Inspector
  suggestion.
- [x] Add bounded multi-step Detect next suggestions with confidence, reasons,
  previews, and one-click chain application.

## Remaining Local TODOs By Priority

1. Automated Detect next follow-up: broaden candidate chains, improve
   confidence ranking, and consider an explicit "Apply best chain" control if
   user testing shows ranked chain suggestions are not enough.
2. Browser/UI regression expansion for broader narrow/mobile editing layouts and
   future routed web features.
3. Keep option metadata tests extended as new plugin flags are added.
4. Optional step editing polish such as drag handles or keyboard shortcuts if
   user testing shows friction.
5. Diff plugin or compare-view enhancement for two-input workflows.
6. Batch mode for applying a saved chain to multiple files, if it becomes
   relevant again.
7. Batch-mode tests and any remaining CLI/core subcommand dispatch coverage if
   batch mode is resumed.
8. Homebrew packaging plan and release workflow cleanup.

## Verification Checklist

For the completed work above, the verification commands were:

```sh
go test ./...
PATH=$(go env GOROOT)/lib/wasm:$PATH GOOS=js GOARCH=wasm go test ./internal/webui
npm --prefix extras/web-perf test
cd extras/vscode-deen && npm run compile && npm run lint
```

Manual smoke tests covered:

- Web Examples lazy loading, search, QR preview, and example loading.
- Web chain share links without source input.
- Web routes for Examples, Plugins, About, example cards, and plugin cards.
- VS Code command metadata consistency for all contributed commands.
