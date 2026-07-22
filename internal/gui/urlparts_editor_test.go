//go:build gui

package gui

import (
	"bytes"
	"strings"
	"testing"

	fynetest "fyne.io/fyne/v2/test"

	"github.com/takeshixx/deen/pkg/formatters"
)

const guiURLPartsTestURL = "https://login-update.example.invalid/account/verify.php?campaign=Q3&redirect=https%3A%2F%2Fportal.example.org%2Fsignin&campaign=retry#continue"

func newURLPartsScenario(t *testing.T) *DeenGUI {
	t.Helper()
	dg := newVisualScenarioGUI(t, appearanceLight)
	dg.pipe.SetSource([]byte(guiURLPartsTestURL))
	dg.pipe.AddStep("urlparts", false)
	dg.pipe.AddStep("urlparts", true)
	dg.selectedStage = 0
	dg.stepOutputView = ""
	dg.rebuild()
	return dg
}

func TestURLPartsStepUsesStructuredEditor(t *testing.T) {
	dg := newURLPartsScenario(t)
	card := dg.cards[0]
	if card == nil || card.urlParts == nil || card.urlPartsTab == nil {
		t.Fatal("URL Parts step did not create its structured editor")
	}
	if selected := card.viewer.Selected(); selected == nil || selected.Text != "URL Parts" {
		t.Fatalf("selected output tab = %v, want URL Parts", selected)
	}
	editor := card.urlParts
	if editor.hostnameEntry.Text != "login-update.example.invalid" || editor.schemeEntry.Text != "https" {
		t.Fatalf("authority fields scheme=%q hostname=%q", editor.schemeEntry.Text, editor.hostnameEntry.Text)
	}
	if len(editor.pathEntries) != 2 || editor.pathEntries[0].Text != "account" || editor.pathEntries[1].Text != "verify.php" {
		t.Fatalf("path fields = %#v", editor.doc.PathSegments)
	}
	if len(editor.queryKeyEntries) != 3 || editor.queryKeyEntries[1].Text != "redirect" {
		t.Fatalf("query fields = %#v", editor.doc.Query)
	}
	if editor.rebuiltEntry.Text != guiURLPartsTestURL || editor.copyURLButton.Disabled() {
		t.Fatalf("rebuilt preview=%q copyDisabled=%v", editor.rebuiltEntry.Text, editor.copyURLButton.Disabled())
	}
}

func TestURLPartsEditorShowsValidationErrorsInPreview(t *testing.T) {
	dg := newURLPartsScenario(t)
	editor := dg.cards[0].urlParts
	editor.portEntry.SetText("not-a-port")
	if editor.rebuiltEntry.Text != "" || !editor.copyURLButton.Disabled() {
		t.Fatalf("invalid URL preview=%q copyDisabled=%v", editor.rebuiltEntry.Text, editor.copyURLButton.Disabled())
	}
	if !strings.Contains(editor.message.Text, "invalid URL port") {
		t.Fatalf("validation message = %q", editor.message.Text)
	}
}

func TestURLPartsEditorRebuildDoesNotLeakWorkControls(t *testing.T) {
	dg := newURLPartsScenario(t)
	editor := dg.cards[0].urlParts
	want := len(dg.workControls)
	editor.rawCheck.SetChecked(true)
	editor.rawCheck.SetChecked(false)
	if got := len(dg.workControls); got != want {
		t.Fatalf("work controls after row rebuilds = %d, want %d", got, want)
	}
}

func TestURLPartsStructuredEditsUpdateJSONAndDownstreamURL(t *testing.T) {
	dg := newURLPartsScenario(t)
	editor := dg.cards[0].urlParts

	editor.hostnameEntry.SetText("review.invalid")
	editor.pathEntries[1].SetText("checked")
	editor.queryValueEntries[1].SetText("https://safe.example.org/result")
	editor.fragmentEntry.SetText("reviewed")

	result := string(dg.pipe.Result())
	want := "https://review.invalid/account/checked?campaign=Q3&redirect=https%3A%2F%2Fsafe.example.org%2Fresult&campaign=retry#reviewed"
	if result != want {
		t.Fatalf("downstream URL = %q, want %q", result, want)
	}
	doc, err := formatters.DecodeURLPartsJSON(bytes.NewReader(dg.pipe.Output(0)))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Hostname != "review.invalid" || doc.PathSegments[1] != "checked" || doc.Query[1].Value != "https://safe.example.org/result" {
		t.Fatalf("structured JSON was not updated: %#v", doc)
	}
}

func TestURLPartsEditorAddsDynamicRows(t *testing.T) {
	dg := newURLPartsScenario(t)
	editor := dg.cards[0].urlParts

	fynetest.Tap(editor.addPathButton)
	editor.pathEntries[len(editor.pathEntries)-1].SetText("extra")
	fynetest.Tap(editor.addQueryButton)
	last := len(editor.queryKeyEntries) - 1
	editor.queryKeyEntries[last].SetText("review")
	editor.queryValueEntries[last].SetText("passed")

	result := string(dg.pipe.Result())
	if !strings.Contains(result, "/account/verify.php/extra?") {
		t.Fatalf("added path segment missing from %q", result)
	}
	if !strings.Contains(result, "&review=passed") {
		t.Fatalf("added query parameter missing from %q", result)
	}
	if len(editor.doc.PathSegments) != 3 || len(editor.doc.Query) != 4 {
		t.Fatalf("dynamic row counts path=%d query=%d", len(editor.doc.PathSegments), len(editor.doc.Query))
	}
}

func TestURLPartsRawJSONEditsRefreshStructuredFields(t *testing.T) {
	dg := newURLPartsScenario(t)
	card := dg.cards[0]
	doc, err := formatters.DecodeURLPartsJSON(bytes.NewReader(dg.pipe.Output(0)))
	if err != nil {
		t.Fatal(err)
	}
	doc.Fragment = "from-raw-json"
	var raw bytes.Buffer
	if err := formatters.EncodeURLPartsJSON(&raw, doc); err != nil {
		t.Fatal(err)
	}

	card.body.SetText(raw.String())
	if card.urlParts.doc == nil || card.urlParts.fragmentEntry.Text != "from-raw-json" {
		t.Fatalf("structured fragment did not refresh: doc=%v entry=%q", card.urlParts.doc != nil, card.urlParts.fragmentEntry.Text)
	}
	if got := string(dg.pipe.Result()); !strings.HasSuffix(got, "#from-raw-json") {
		t.Fatalf("downstream URL did not refresh from raw JSON: %q", got)
	}

	card.body.SetText("{")
	if card.urlParts.doc != nil || !strings.Contains(card.urlParts.message.Text, "Structured editor unavailable") {
		t.Fatal("invalid raw JSON should put the structured editor into its error state")
	}
}

func TestURLPartsReverseStepDoesNotShowStructuredEditor(t *testing.T) {
	dg := newURLPartsScenario(t)
	dg.selectedStage = 1
	dg.rebuild()
	card := dg.cards[1]
	if card == nil {
		t.Fatal("reverse URL Parts card was not built")
	}
	if card.urlParts != nil || card.urlPartsTab != nil {
		t.Fatal("JSON-to-URL direction should use the ordinary output viewer")
	}
	if !dg.pipe.Steps()[1].Unprocess {
		t.Fatal("test setup lost the reverse URL Parts direction")
	}
}

func TestURLPartsEditorPreservesPipelineUndo(t *testing.T) {
	dg := newURLPartsScenario(t)
	dg.cards[0].urlParts.hostnameEntry.SetText("undo.invalid")
	if !dg.pipe.CanUndo() {
		t.Fatal("structured edit did not create an undo snapshot")
	}
	if !dg.pipe.Undo() {
		t.Fatal("undo failed")
	}
	if got := string(dg.pipe.Result()); got != guiURLPartsTestURL {
		t.Fatalf("undo result = %q, want original URL", got)
	}
	dg.refreshFrom(0)
	if dg.cards[0].urlParts.hostnameEntry.Text != "login-update.example.invalid" {
		t.Fatalf("structured editor did not refresh after undo: %q", dg.cards[0].urlParts.hostnameEntry.Text)
	}
}
