# GitHub Issues TODO Plan

Reviewed on 2026-06-28 from `takeshixx/go-deen` open GitHub issues.

## Open Issues

### Issue #22: Add load animation

URL: https://github.com/takeshixx/go-deen/issues/22

Problem: when large files or expensive pipelines are processed, the UI appears
to hang because pipeline recomputation runs synchronously without visible busy
feedback.

Likely areas:

- `internal/webui/webui.go`: `refreshOutputs`, `renderOutput`, source/file load
  handlers, and option/change callbacks that call pipeline recomputation.
- `internal/web/assets/style.css`: busy indicator, disabled/working visual
  states, and reduced-motion handling.
- `internal/gui/gui.go` and `internal/gui/step.go`: source changes and
  `refreshFrom`/step refresh paths in the Fyne UI.
- `internal/pipeline/pipeline.go`: keep core transform behavior unchanged unless
  UI integration requires a small progress/busy hook.

TODO:

- [ ] Reproduce with a large input file and a costly transform chain in both web
  UI and desktop GUI.
- [ ] Identify every UI action that can trigger synchronous recomputation:
  source edits, file imports, step add/remove/reorder/duplicate, plugin changes,
  decode toggles, option edits, output edits, undo/redo, preset/example loads,
  and chain imports.
- [x] Add a web UI busy state that appears before recomputation starts and is
  cleared after rendering finishes.
- [x] Yield to the browser event loop before expensive web recomputation so the
  animation can paint.
- [x] Add desktop GUI busy feedback for source edits, file loads, refreshes, and
  rebuilds.
- [ ] Temporarily disable or debounce controls that can stack recomputation while
  work is in progress.
- [x] Respect `prefers-reduced-motion` in the web UI.
- [ ] Verify that errors still render after a failed transform and the busy state
  always clears.
- [ ] Add tests where practical for any shared busy-state/debounce helpers.

Acceptance criteria:

- A large file no longer leaves the UI looking idle or frozen during processing.
- The busy indicator is visible for long operations and does not flash
  distractingly for tiny operations.
- Existing outputs, previews, metadata, errors, undo/redo, and downloads keep
  working.

### Issue #21: Display output image in QR example

URL: https://github.com/takeshixx/go-deen/issues/21

Problem: the QR example should show its generated output image, not just text or
binary-looking data.

Current context:

- QR encoding already produces PNG bytes in `pkg/formatters/qr.go`.
- Desktop GUI shows an Image tab for QR encode steps via
  `internal/gui/step.go`.
- Web UI shows an Image tab for QR encode steps via `internal/webui/webui.go`.
- The plugin catalog example currently describes the expected output as text in
  `internal/plugins/catalog_ui.go`.

TODO:

- [x] Reproduce the QR example flow from the plugin catalog/examples in the web
  UI and desktop GUI.
- [x] Confirm whether the issue is in the example loading flow, the selected
  output tab, or the catalog/example metadata.
- [x] Make QR encode examples open or highlight the Image tab when loaded, if
  the pipeline output is image-capable.
- [x] Ensure generated QR PNG bytes are previewed without requiring a manual tab
  switch when the example is specifically image-oriented.
- [x] Keep the Text/Hex/Preview tabs available for users who need the raw bytes.
- [ ] Add or update a test around QR example metadata/loading if the example
  flow has testable state.
- [ ] Manually verify QR encode and QR decode still work in CLI, desktop GUI,
  and web UI.

Acceptance criteria:

- Loading the QR encode example displays the generated QR image in the UI.
- The output image is visible in both web and desktop UIs, or any intentional
  platform difference is documented.
- QR round-trip behavior remains covered by `pkg/formatters` tests.

## Recommended Order

1. Fix issue #21 first. The image-preview plumbing mostly exists, so this should
   be a smaller UI/example-state fix.
2. Fix issue #22 second. It may require broader UI orchestration and careful
   manual testing with large inputs.
3. After both fixes, run:

```sh
go test ./...
GOOS=js GOARCH=wasm go test ./internal/webui
```

4. Manually smoke test:
   - web QR encode example image preview
   - desktop QR encode example image preview
   - large file import with visible busy feedback
   - error-producing transform while busy feedback is active
