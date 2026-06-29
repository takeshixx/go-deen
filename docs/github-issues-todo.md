# GitHub Issues TODO Plan

Reviewed on 2026-06-29 from `takeshixx/go-deen` open GitHub issues.

## Open Issues

### Issue #23: Lazy load examples

URL: https://github.com/takeshixx/go-deen/issues/23

Problem: the web Examples tab should initially load only the example titles.
Details, source data, output previews, and expensive example execution should be
loaded only after an example is clicked. This should reduce the initial render
cost of the Examples tab.

Current context:

- `internal/webui/webui.go` already defers source/output preview generation
  until a `<details>` card is opened.
- The current first render still builds full cards for every example, including
  description, chain pills, metadata text, action buttons, and toggle handlers.
- `internal/webui/perf_benchmark_test.go` already has example tab benchmarks
  that can be adjusted to protect the lazy-load behavior.
- `internal/gui/gui.go` has a separate desktop Examples tab. The issue text
  names the Examples tab generally, but the load-time concern maps most directly
  to the web UI.

Implementation plan:

- [x] Split web example rendering into a lightweight title row/card and a lazy
  detail builder.
- [x] Store only `pipeline.Example` references and minimal DOM nodes during the
  initial `renderExamples` pass.
- [x] On first open/click, populate description, chain summary, source summary,
  expected-result text, load button, and preview container.
- [x] Keep output preview execution behind a second lazy boundary, so merely
  expanding the example does not compute `pipeline.ExampleResult` unless the
  preview is requested or the existing expanded-content design requires it.
- [x] Preserve search over name, description, source metadata, and chain terms
  by continuing to use `pipeline.ExampleMatches` against the in-memory example
  structs, not pre-rendered DOM.
- [x] Update web example benchmarks to separate initial title-list rendering
  from lazy detail rendering and result-preview rendering.
- [x] Manually verify the Examples tab search, expand/collapse, Load example,
  QR image preview, and no-results state.

Acceptance criteria:

- Opening the Examples tab builds only title-level UI for each example.
- Example details and previews are populated once, on demand.
- Search behavior and example loading remain unchanged from a user's point of
  view.
- Web UI benchmark coverage reflects the new lazy-load boundaries.

### Issue #24: Update README.md

URL: https://github.com/takeshixx/go-deen/issues/24

Problem: the README needs documentation for chain examples, the web example
workflow, and copy-link behavior.

Requested items:

- Add chain examples.
- Add `make web` example.
- Add copy link in web example.

Current context:

- The README already has CLI, GUI, WebAssembly, and plugin-writing sections.
- `internal/pipeline.CommandLine` can generate shell pipelines equivalent to a
  chain, and the web UI exposes command/export/share actions.
- `internal/webui/webui.go` implements `copyShareLink` using
  `ExportJSONWithoutSource`, so share links intentionally omit source input.
- `make web` is already mentioned, but it can be expanded with a concrete local
  usage example and share-link flow.

Implementation plan:

- [x] Add a "Chains" or "Transform chains" subsection after basic CLI usage.
- [x] Include concrete chain examples, including shell pipes and the
  `deen chain` subcommand for saved Web/GUI chain JSON.
- [x] Expand the WebAssembly section with a complete local run flow:
  `make web`, open `http://127.0.0.1:9090`, load an example, edit the chain,
  and download or copy the result.
- [x] Document "Copy link" behavior clearly: it shares the chain recipe in the
  URL hash and does not include source input.
- [x] Mention exported chain JSON files and imported chain links if appropriate,
  keeping the README concise.
- [x] Ensure examples use current plugin names (`base64`, `.base64`, `gzip`,
  `.gzip`, `json`, `url`) and options that are covered by tests.
- [x] Run markdown/link review manually and `go test ./...` if README command
  examples are adjusted against executable behavior.

Acceptance criteria:

- A new user can understand how to compose and reuse chains from the README.
- The local web workflow is documented with `make web`.
- Copy-link privacy semantics are explicit: source data is not embedded.

### Issue #25: Update VS Code extension

URL: https://github.com/takeshixx/go-deen/issues/25

Problem: the VS Code extension is stale and should be updated. The issue body is
empty, so scope needs to be inferred from the extension's current state and kept
small enough for one focused pass unless the maintainer clarifies otherwise.

Current context:

- Extension code lives in `extras/vscode-deen`.
- `src/extension.ts` calls `child.exec("godeen "+plugin+" "+content)`, which is
  shell-injection prone, breaks on large/special input, and uses the old
  `godeen` binary name instead of `deen`.
- The webview title is still "Cat Coding" and renders output without escaping.
- The command list is manually duplicated between `extension.ts` and
  `package.json`; there are typos and mismatches such as `sha3-244Hash`,
  `SHA3512`, and a duplicate `deen.sha1Hash` entry for RIPEMD160.
- The plugin set is far behind the main application, which now has 60+ plugins.
- Build tooling targets old container defaults (`node:14-alpine3.14`) while
  package dependencies are modern.

Implementation plan:

- [x] Decide the minimal supported extension model: command-palette transforms
  over selected text/full document, using an installed `deen` binary.
- [x] Add a setting for the binary path, defaulting to `deen`, and remove the
  hard-coded `godeen`.
- [x] Replace `child.exec` with `spawn`/`execFile` and pass selected content via
  stdin. This avoids shell quoting bugs and supports larger or binary-ish input
  better.
- [x] Show errors through `vscode.window.showErrorMessage` and render stderr or
  exit status in a controlled output channel/webview instead of throwing from an
  async callback.
- [x] Escape webview output or switch to a VS Code output document/channel for
  plain text. If a webview remains, set a narrow content security policy.
- [x] Consolidate command metadata into one typed table in `extension.ts` and
  generate/register commands from it, or keep package metadata hand-written but
  add a test/lint check for mismatches.
- [x] Add more current plugins in a staged
  way: first codecs/formatters/compressions/hash basics, then newer misc and
  crypto tools where option support is clear.
- [x] Fix existing command typos and command metadata mismatches.
- [x] Add a small transform helper test if the extension test harness is
  available; otherwise cover command metadata with TypeScript compile/lint.
- [x] Refresh extension README build instructions and the Makefile image to a
  current Node LTS version.

Acceptance criteria:

- Extension commands invoke `deen` safely through stdin.
- Existing contributed commands compile, register, and map to the intended
  plugin names.
- User-facing output and errors are handled without unsafe HTML or thrown async
  exceptions.
- Extension docs and build tooling match the updated implementation.

## Recommended Order

1. Fix issue #24 first. It is documentation-only, low risk, and will clarify the
   public behavior of chains and web share links before UI or extension changes.
2. Fix issue #23 second. It is localized to the web Examples tab and has
   existing benchmark coverage that can be adapted.
3. Fix issue #25 third. It has the largest scope and the issue body is empty, so
   it should be implemented as a conservative modernization rather than a full
   extension rewrite.

## Verification Checklist

After the implementation work, run:

```sh
go test ./...
GOOS=js GOARCH=wasm go test ./internal/webui
cd extras/vscode-deen && npm run compile && npm run lint
```

Manual smoke tests:

- Start the web UI with `make web` and verify the Examples tab initially renders
  quickly with title-level content only.
- Expand several examples, preview their input/output data, and load them into
  the main chain.
- Copy a web share link and confirm it restores the chain but not the source
  input.
- In VS Code, run a few commands against selected text and whole-document input,
  including Base64 encode/decode, URL encode/decode, JSON format, and at least
  one hash.
