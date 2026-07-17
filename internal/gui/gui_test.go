//go:build gui

package gui

import (
	"errors"
	"image/color"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	fynetest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/takeshixx/deen/internal/pipeline"
	"github.com/takeshixx/deen/internal/plugins"
)

func TestGUISafeSuggestionPreviewEscapesControls(t *testing.T) {
	got := guiSafeSuggestionPreview("ok\x00\n\x1b[31m")
	want := `ok\x00` + "\n" + `\x1b[31m`
	if got != want {
		t.Fatalf("guiSafeSuggestionPreview() = %q, want %q", got, want)
	}
}

func TestNormalizeAppearance(t *testing.T) {
	tests := map[string]appearanceMode{
		"system":  appearanceSystem,
		"dark":    appearanceDark,
		"light":   appearanceLight,
		"":        appearanceSystem,
		"unknown": appearanceSystem,
	}
	for input, want := range tests {
		if got := normalizeAppearance(input); got != want {
			t.Errorf("normalizeAppearance(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNormalizeStepEditorSplit(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  float64
	}{
		{name: "default", value: defaultStepEditorSplit, want: defaultStepEditorSplit},
		{name: "lower bound", value: -1, want: minStepEditorSplit},
		{name: "upper bound", value: 2, want: maxStepEditorSplit},
		{name: "NaN", value: math.NaN(), want: defaultStepEditorSplit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeStepEditorSplit(tt.value); got != tt.want {
				t.Fatalf("normalizeStepEditorSplit(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestWorkspacePersistenceNormalization(t *testing.T) {
	if got := normalizeSourcePane("Inspector"); got != "Inspector" {
		t.Fatalf("normalizeSourcePane(Inspector) = %q", got)
	}
	if got := normalizeSourcePane("invalid"); got != defaultSourcePane {
		t.Fatalf("normalizeSourcePane(invalid) = %q, want %q", got, defaultSourcePane)
	}
	if got := normalizeAddSplit(-1); got != minAddSplit {
		t.Fatalf("normalizeAddSplit(-1) = %v, want %v", got, minAddSplit)
	}
	if got := normalizeAddSplit(2); got != maxAddSplit {
		t.Fatalf("normalizeAddSplit(2) = %v, want %v", got, maxAddSplit)
	}
	if got := normalizeAddSplit(math.NaN()); got != defaultAddSplit {
		t.Fatalf("normalizeAddSplit(NaN) = %v, want %v", got, defaultAddSplit)
	}
	if got := normalizeAddCategory("not a category"); got != allTransformerCategories {
		t.Fatalf("normalizeAddCategory(invalid) = %q, want %q", got, allTransformerCategories)
	}
	if got := normalizeCompareSplit(-1); got != minCompareSplit {
		t.Fatalf("normalizeCompareSplit(-1) = %v, want %v", got, minCompareSplit)
	}
	if got := normalizeCompareSplit(2); got != maxCompareSplit {
		t.Fatalf("normalizeCompareSplit(2) = %v, want %v", got, maxCompareSplit)
	}
	if got := normalizeCompareSplit(math.NaN()); got != defaultCompareSplit {
		t.Fatalf("normalizeCompareSplit(NaN) = %v, want %v", got, defaultCompareSplit)
	}
	if got := normalizeCompactSplit(-1); got != minCompactSplit {
		t.Fatalf("normalizeCompactSplit(-1) = %v, want %v", got, minCompactSplit)
	}
	if got := normalizeCompactSplit(2); got != maxCompactSplit {
		t.Fatalf("normalizeCompactSplit(2) = %v, want %v", got, maxCompactSplit)
	}
	if got := normalizeCompactSplit(math.NaN()); got != defaultCompactSplit {
		t.Fatalf("normalizeCompactSplit(NaN) = %v, want %v", got, defaultCompactSplit)
	}
	for input, want := range map[int]int{-1: 0, 0: 0, 2: 2, 4: 4, 5: 0} {
		if got := normalizeWorkspace(input); got != want {
			t.Errorf("normalizeWorkspace(%d) = %d, want %d", input, got, want)
		}
	}
	if got := normalizeWindowDimension(400, defaultWindowWidth, minimumWindowWidth); got != defaultWindowWidth {
		t.Fatalf("undersized window dimension = %v, want fallback %v", got, defaultWindowWidth)
	}
}

func TestResponsiveSplitKeepsIndependentOrientations(t *testing.T) {
	a := fynetest.NewTempApp(t)
	dg := &DeenGUI{app: a}
	split := container.NewHSplit(widget.NewLabel("Leading"), widget.NewLabel("Trailing"))
	controller := newResponsiveSplit(
		dg,
		split,
		"test.horizontal",
		"test.compact",
		0.42,
		defaultCompactSplit,
		normalizeAddSplit,
		compactWorkspaceSplit,
	)
	controller.host.Resize(fyne.NewSize(700, 600))
	if split.Horizontal || split.Offset != defaultCompactSplit {
		t.Fatalf("compact split horizontal=%v offset=%v", split.Horizontal, split.Offset)
	}
	split.Offset = 0.64
	if horizontal := controller.remember(); horizontal != 0.42 {
		t.Fatalf("remembering compact split returned horizontal offset %v, want 0.42", horizontal)
	}
	if got := a.Preferences().Float("test.horizontal"); got != 0.42 {
		t.Fatalf("compact drag overwrote horizontal preference with %v", got)
	}
	if got := a.Preferences().Float("test.compact"); got != 0.64 {
		t.Fatalf("remembered compact preference = %v, want 0.64", got)
	}

	controller.host.Resize(fyne.NewSize(800, 600))
	if !split.Horizontal || split.Offset != 0.42 {
		t.Fatalf("restored horizontal split horizontal=%v offset=%v", split.Horizontal, split.Offset)
	}
	split.Offset = 0.57
	controller.remember()
	controller.host.Resize(fyne.NewSize(700, 600))
	if split.Horizontal || split.Offset != 0.64 {
		t.Fatalf("restored compact split horizontal=%v offset=%v, want vertical/0.64", split.Horizontal, split.Offset)
	}
	if got := a.Preferences().Float("test.horizontal"); got != 0.57 {
		t.Fatalf("remembered horizontal preference = %v, want 0.57", got)
	}

	controller.host.Resize(fyne.NewSize(700, 500))
	if !split.Horizontal || split.Offset != 0.57 {
		t.Fatal("short workspaces should retain the horizontal layout to preserve vertical room")
	}
	restoredSplit := container.NewHSplit(widget.NewLabel("Leading"), widget.NewLabel("Trailing"))
	restored := newResponsiveSplit(dg, restoredSplit, "test.horizontal", "test.compact", 0.42, defaultCompactSplit, normalizeAddSplit, compactWorkspaceSplit)
	restored.host.Resize(fyne.NewSize(700, 600))
	if restoredSplit.Horizontal || restoredSplit.Offset != 0.64 {
		t.Fatalf("restored compact split horizontal=%v offset=%v", restoredSplit.Horizontal, restoredSplit.Offset)
	}
	restored.host.Resize(fyne.NewSize(800, 600))
	if !restoredSplit.Horizontal || restoredSplit.Offset != 0.57 {
		t.Fatalf("restored wide split horizontal=%v offset=%v", restoredSplit.Horizontal, restoredSplit.Offset)
	}
}

func TestPipelineEditorSplitFollowsCompactStageMode(t *testing.T) {
	dg := newVisualScenarioGUI(t, appearanceDark)
	dg.pipe.SetSource([]byte("responsive split"))
	dg.pipe.AddStep("base64", false)
	dg.selectedStage = 0
	dg.rebuild()
	dg.applicationShell.Resize(fyne.NewSize(1200, 800))
	card := dg.cards[0]
	if card == nil || !card.editorSplit.Horizontal {
		t.Fatal("wide focused-step editor should use a horizontal split")
	}
	card.editorSplit.Offset = 0.56
	dg.rememberStepEditorSplit()

	dg.applicationShell.Resize(fyne.NewSize(700, 800))
	if !dg.compactStages || card.editorSplit.Horizontal || card.editorSplit.Offset != defaultCompactSplit {
		t.Fatalf("compact editor stages=%v horizontal=%v offset=%v", dg.compactStages, card.editorSplit.Horizontal, card.editorSplit.Offset)
	}
	card.editorSplit.Offset = 0.63
	dg.rememberStepEditorSplit()
	preferences := dg.app.Preferences()
	if preferences.Float(stepEditorSplitPreferenceKey) != 0.56 || preferences.Float(stepEditorCompactPreferenceKey) != 0.63 {
		t.Fatalf("step split preferences horizontal=%v compact=%v", preferences.Float(stepEditorSplitPreferenceKey), preferences.Float(stepEditorCompactPreferenceKey))
	}

	dg.applicationShell.Resize(fyne.NewSize(1200, 800))
	if !card.editorSplit.Horizontal || card.editorSplit.Offset != 0.56 {
		t.Fatalf("restored wide editor horizontal=%v offset=%v", card.editorSplit.Horizontal, card.editorSplit.Offset)
	}
	dg.selectPipelineStage(pipelineStageInput)
	dg.applicationShell.Resize(fyne.NewSize(700, 800))
	if dg.sourceWorkspace == nil || len(dg.sourceWorkspace.Items) != 2 || dg.sourceWorkspace.Selected().Text != "Editor" {
		t.Fatal("Input should retain its exclusive Editor/Inspector views at compact widths")
	}
}

func TestResponsiveWorkspaceSplitScopes(t *testing.T) {
	dg := newVisualScenarioGUI(t, appearanceLight)
	dg.applicationShell.Resize(fyne.NewSize(1200, 800))
	dg.selectedStage = pipelineStageAdd
	dg.rebuild()
	if dg.compactStages || dg.addCatalog == nil || !dg.addCatalog.split.Horizontal {
		t.Fatal("wide Add Transformer should retain its horizontal master/detail layout")
	}
	dg.applicationShell.Resize(fyne.NewSize(700, 800))
	if !dg.compactStages || dg.addCatalog.split.Horizontal {
		t.Fatal("compact Add Transformer should stack catalog and details")
	}

	dg.selectTab(1)
	dg.workflowLibrary.splitController.host.Resize(fyne.NewSize(700, 600))
	if dg.workflowLibrary.split.Horizontal {
		t.Fatal("narrow, tall Workflows should use a vertical master/detail layout")
	}
	dg.workflowLibrary.splitController.host.Resize(fyne.NewSize(700, 500))
	if !dg.workflowLibrary.split.Horizontal {
		t.Fatal("short Workflows should retain its horizontal master/detail layout")
	}

	dg.selectTab(2)
	dg.browserCatalog.splitController.host.Resize(fyne.NewSize(700, 600))
	if dg.browserCatalog.split.Horizontal {
		t.Fatal("narrow, tall Transformers should use a vertical master/detail layout")
	}
	dg.selectTab(4)
	dg.compareWorkspace.splitController.host.Resize(fyne.NewSize(700, 600))
	if dg.compareWorkspace.split.Horizontal {
		t.Fatal("narrow, tall Compare should stack its data panes")
	}
}

func TestAdaptiveNavigationDrawersPreservePreferences(t *testing.T) {
	dg := newVisualScenarioGUI(t, appearanceLight)
	dg.pipe.SetSource([]byte("adaptive layout"))
	dg.pipe.AddStep("base64", false)
	dg.rebuild()
	dg.applicationShell.Resize(fyne.NewSize(1200, 700))
	dg.pipelineShell.Resize(fyne.NewSize(900, 620))
	preferences := dg.app.Preferences()
	workspace := dg.applicationShell.Objects[0]

	if dg.compactSidebar || dg.sidebarCollapsed || !dg.navigationPanel.Visible() || workspace.Position().X <= collapsedSidebarWidth {
		t.Fatalf("wide sidebar state compact=%v collapsed=%v visible=%v workspaceX=%v", dg.compactSidebar, dg.sidebarCollapsed, dg.navigationPanel.Visible(), workspace.Position().X)
	}
	dg.window.Canvas().Focus(dg.tabButtons[1])
	dg.applicationShell.Resize(fyne.NewSize(800, 700))
	if !dg.compactSidebar || !dg.sidebarCollapsed || !dg.navigationPanel.Visible() || workspace.Position().X <= collapsedSidebarWidth || dg.window.Canvas().Focused() != nil {
		t.Fatalf("compact sidebar state compact=%v collapsed=%v visible=%v workspaceX=%v focused=%T", dg.compactSidebar, dg.sidebarCollapsed, dg.navigationPanel.Visible(), workspace.Position().X, dg.window.Canvas().Focused())
	}
	if !dg.tabButtons[0].iconOnly || dg.tabButtons[0].label.Visible() || dg.tabButtons[0].AccessibilityLabel() != "Pipeline" {
		t.Fatal("collapsed sidebar destination should retain its icon and accessible label while hiding visible text")
	}
	if !preferences.BoolWithFallback(sidebarPreferenceKey, true) {
		t.Fatal("automatic sidebar collapse overwrote the wide-window preference")
	}
	dg.toggleSidebar()
	if !dg.sidebarDrawerOpen || dg.sidebarCollapsed || !dg.navigationPanel.Visible() || workspace.Position().X <= collapsedSidebarWidth || dg.window.Canvas().Focused() != dg.tabButtons[0] {
		t.Fatalf("compact sidebar drawer open=%v collapsed=%v visible=%v workspace=%v@%v", dg.sidebarDrawerOpen, dg.sidebarCollapsed, dg.navigationPanel.Visible(), workspace.Size(), workspace.Position())
	}
	if !preferences.BoolWithFallback(sidebarPreferenceKey, true) {
		t.Fatal("opening the compact sidebar drawer changed its persistent preference")
	}
	dg.selectTab(1)
	if dg.sidebarDrawerOpen || !dg.sidebarCollapsed || !dg.navigationPanel.Visible() || dg.window.Canvas().Focused() != nil {
		t.Fatal("selecting a workspace should return the compact sidebar to its icon rail")
	}
	dg.applicationShell.Resize(fyne.NewSize(1200, 700))
	if dg.compactSidebar || dg.sidebarCollapsed || !dg.navigationPanel.Visible() {
		t.Fatal("growing the window should restore the preferred sidebar state")
	}
	dg.toggleSidebar()
	if dg.sidebarOpen || !dg.sidebarCollapsed || !dg.navigationPanel.Visible() || preferences.BoolWithFallback(sidebarPreferenceKey, true) {
		t.Fatal("wide sidebar toggle should collapse to a persistent icon rail")
	}
	dg.applicationShell.Resize(fyne.NewSize(800, 700))
	dg.toggleSidebar()
	if !dg.sidebarDrawerOpen || dg.sidebarCollapsed || !dg.navigationPanel.Visible() || preferences.BoolWithFallback(sidebarPreferenceKey, true) {
		t.Fatal("compact drawer should open without changing a saved closed preference")
	}
	dg.applicationShell.Resize(fyne.NewSize(1200, 700))
	if dg.sidebarDrawerOpen || !dg.navigationPanel.Visible() || !dg.sidebarCollapsed || dg.sidebarOpen {
		t.Fatal("growing the window should restore a saved icon-only sidebar")
	}

	dg.selectTab(0)
	dg.applicationShell.Resize(fyne.NewSize(1200, 700))
	if dg.compactStages || !dg.pipelineNavigator.Visible() {
		t.Fatal("wide pipeline should show the preferred stage navigator")
	}
	dg.window.Canvas().Focus(dg.pipelineStageButtons[0])
	dg.applicationShell.Resize(fyne.NewSize(700, 700))
	if !dg.compactStages || dg.pipelineNavigator.Visible() || dg.window.Canvas().Focused() != nil {
		t.Fatalf("compact stages state compact=%v visible=%v focused=%T", dg.compactStages, dg.pipelineNavigator.Visible(), dg.window.Canvas().Focused())
	}
	if !preferences.BoolWithFallback(pipelineNavPreferenceKey, true) {
		t.Fatal("automatic stage collapse overwrote the wide-window preference")
	}
	dg.togglePipelineNavigator()
	if !dg.stagesDrawerOpen || !dg.pipelineNavigator.Visible() || dg.window.Canvas().Focused() != dg.pipelineStageButtons[0] {
		t.Fatal("Stages should open the compact pipeline navigator as a drawer")
	}
	dg.selectPipelineStage(pipelineStageInput)
	if dg.stagesDrawerOpen || dg.pipelineNavigator.Visible() || dg.window.Canvas().Focused() != nil {
		t.Fatal("selecting the current stage should still close the compact stages drawer")
	}
	dg.applicationShell.Resize(fyne.NewSize(1200, 700))
	if dg.compactStages || !dg.pipelineNavigator.Visible() {
		t.Fatal("growing the pipeline should restore the preferred stage navigator state")
	}
	dg.togglePipelineNavigator()
	if dg.pipelineNavOpen || dg.pipelineNavigator.Visible() || preferences.BoolWithFallback(pipelineNavPreferenceKey, true) {
		t.Fatal("wide Stages toggle should hide and persist normally")
	}
	dg.applicationShell.Resize(fyne.NewSize(700, 700))
	dg.togglePipelineNavigator()
	if !dg.stagesDrawerOpen || !dg.pipelineNavigator.Visible() || preferences.BoolWithFallback(pipelineNavPreferenceKey, true) {
		t.Fatal("compact stages drawer should not change a saved closed preference")
	}
	dg.applicationShell.Resize(fyne.NewSize(1200, 700))
	if dg.stagesDrawerOpen || dg.pipelineNavigator.Visible() || dg.pipelineNavOpen {
		t.Fatal("growing the window should restore a saved closed stage navigator")
	}
}

func TestCompareFormattingAndDifferenceSummary(t *testing.T) {
	if got := normalizeCompareMode("unknown"); got != compareModeText {
		t.Fatalf("normalizeCompareMode(unknown) = %q, want Text", got)
	}
	if got := formatCompareData([]byte("hi"), compareModeText); got != "hi" {
		t.Fatalf("text comparison view = %q", got)
	}
	if got := formatCompareData([]byte("hi"), compareModeBase64); got != "aGk=" {
		t.Fatalf("base64 comparison view = %q", got)
	}
	if got := formatCompareData([]byte("hi"), compareModeHex); !strings.Contains(got, "68 69") {
		t.Fatalf("hex comparison view = %q, want byte values", got)
	}

	equal := compareBytes([]byte("same"), []byte("same"))
	if !equal.Equal || equal.FirstDifference != -1 || equal.Summary() != "Identical · 4 B" {
		t.Fatalf("equal comparison = %#v summary=%q", equal, equal.Summary())
	}
	different := compareBytes([]byte("abc"), []byte("axc"))
	if different.Equal || different.FirstDifference != 1 || !strings.Contains(different.Summary(), "offset 1") {
		t.Fatalf("different comparison = %#v summary=%q", different, different.Summary())
	}
	prefix := compareBytes([]byte("abc"), []byte("abcd"))
	if prefix.Equal || prefix.FirstDifference != 3 {
		t.Fatalf("prefix comparison = %#v, want first difference at length boundary", prefix)
	}
}

func TestCompareWorkspaceTracksPipelineAndPersistsState(t *testing.T) {
	empty := newVisualScenarioGUI(t, appearanceLight)
	empty.showCompare()
	empty.pipe.SetSource([]byte("input"))
	empty.pipe.AddStep("base64", false)
	empty.rebuild()
	if empty.compareWorkspace.leftIndex != 0 || empty.compareWorkspace.rightIndex != 1 {
		t.Fatalf("empty-pipeline comparison became %d/%d after first step, want input/final", empty.compareWorkspace.leftIndex, empty.compareWorkspace.rightIndex)
	}

	dg := newVisualScenarioGUI(t, appearanceLight)
	dg.pipe.SetSource([]byte("deen"))
	dg.pipe.AddStep("base64", false)
	dg.rebuild()
	dg.showCompare()
	workspace := dg.compareWorkspace
	if dg.activeTab != 4 || workspace == nil {
		t.Fatal("Compare command should open the Compare workspace")
	}
	if len(workspace.points) != 2 || workspace.leftIndex != 0 || workspace.rightIndex != 1 {
		t.Fatalf("initial comparison points=%d selection=%d/%d, want 2 and input/final", len(workspace.points), workspace.leftIndex, workspace.rightIndex)
	}
	if workspace.points[1].Label != "Step 1 · base64 · encode" {
		t.Fatalf("step comparison label = %q", workspace.points[1].Label)
	}
	if workspace.difference.Importance != widget.WarningImportance || !strings.Contains(workspace.difference.Text, "Different") {
		t.Fatalf("initial comparison status = %q importance=%v", workspace.difference.Text, workspace.difference.Importance)
	}

	workspace.swap()
	if workspace.leftIndex != 1 || workspace.rightIndex != 0 {
		t.Fatalf("swapped selection = %d/%d", workspace.leftIndex, workspace.rightIndex)
	}
	preferences := dg.app.Preferences()
	if preferences.Int(compareLeftPreferenceKey) != 1 || preferences.Int(compareRightPreferenceKey) != 0 {
		t.Fatal("swapped comparison points were not persisted")
	}
	workspace.modeSelect.SetSelected(compareModeBase64)
	if preferences.String(compareModePreferenceKey) != compareModeBase64 {
		t.Fatalf("persisted comparison mode = %q", preferences.String(compareModePreferenceKey))
	}
	if workspace.rightBody.Text != "ZGVlbg==" {
		t.Fatalf("right comparison view = %q, want Base64-formatted input", workspace.rightBody.Text)
	}
	// The in-memory test window returns a fresh clipboard on every access, so
	// exercise the action here while formatCompareData verifies its payload.
	workspace.copyView(false)
	if dg.workStatus.Text != "Right comparison view copied" || dg.workStatus.Importance != widget.SuccessImportance {
		t.Fatalf("comparison copy feedback = %q importance=%v", dg.workStatus.Text, dg.workStatus.Importance)
	}

	// A selection on the previous final output follows the new final output.
	dg.pipe.AddStep("hex", false)
	dg.rebuild()
	if len(workspace.points) != 3 || workspace.leftIndex != 2 || workspace.rightIndex != 0 {
		t.Fatalf("refreshed comparison points=%d selection=%d/%d, want 3 and final/input", len(workspace.points), workspace.leftIndex, workspace.rightIndex)
	}
	workspace.rightSelect.SetSelectedIndex(2)
	if !compareBytes(compareData(workspace.points, workspace.leftIndex), compareData(workspace.points, workspace.rightIndex)).Equal || workspace.difference.Importance != widget.SuccessImportance {
		t.Fatal("selecting the same pipeline point should show an identical success status")
	}

	workspace.split.Offset = 0.61
	dg.rememberWorkspaceState()
	if preferences.Float(compareSplitPreferenceKey) != 0.61 {
		t.Fatalf("persisted comparison split = %v", preferences.Float(compareSplitPreferenceKey))
	}
	dg.working = true
	dg.setWorking("Testing", true)
	if !workspace.leftSelect.Disabled() || !workspace.modeSelect.Disabled() {
		t.Fatal("comparison controls should disable during pipeline work")
	}
	dg.working = false
	dg.setWorking("", false)
	if workspace.leftSelect.Disabled() || workspace.modeSelect.Disabled() {
		t.Fatal("comparison controls should re-enable after pipeline work")
	}

	// Recreate the cached tab to verify a fresh workspace restores every
	// persisted comparison choice instead of only storing it.
	dg.selectTab(0)
	dg.compareWorkspace = nil
	dg.tabViews[4] = nil
	dg.selectTab(4)
	restored := dg.compareWorkspace
	if restored == nil || restored.leftIndex != 2 || restored.rightIndex != 2 {
		t.Fatalf("restored comparison selection = %v", restored)
	}
	if restored.modeSelect.Selected != compareModeBase64 || restored.split.Offset != 0.61 {
		t.Fatalf("restored comparison mode/split = %q/%v", restored.modeSelect.Selected, restored.split.Offset)
	}
}

func TestRememberWorkspaceState(t *testing.T) {
	a := fynetest.NewTempApp(t)
	window := a.NewWindow("test")
	window.Resize(fyne.NewSize(940, 680))
	dg := &DeenGUI{
		app:           a,
		window:        window,
		pipe:          pipeline.New(),
		activeTab:     2,
		selectedStage: pipelineStageAdd,
	}
	dg.rememberWorkspaceState()
	prefs := a.Preferences()
	if prefs.Int(workspacePreferenceKey) != 2 || prefs.Int(selectedStagePreferenceKey) != pipelineStageAdd {
		t.Fatalf("remembered workspace/stage = %d/%d", prefs.Int(workspacePreferenceKey), prefs.Int(selectedStagePreferenceKey))
	}
	if prefs.Float(windowWidthPreferenceKey) != 940 || prefs.Float(windowHeightPreferenceKey) != 680 {
		t.Fatalf("remembered window size = %v×%v", prefs.Float(windowWidthPreferenceKey), prefs.Float(windowHeightPreferenceKey))
	}
}

func TestTransformerCatalogFiltering(t *testing.T) {
	all := filterTransformerCatalog("", allTransformerCategories, nil, nil)
	if len(all) == 0 {
		t.Fatal("unfiltered transformer catalog is empty")
	}
	matches := filterTransformerCatalog("b64", allTransformerCategories, nil, nil)
	if len(matches) == 0 || matches[0].Name != "base64" {
		t.Fatalf("b64 catalog matches = %#v, want base64 first", matches)
	}
	hashes := filterTransformerCatalog("", "Hashes", nil, nil)
	if len(hashes) == 0 {
		t.Fatal("hash category filter returned no transformers")
	}
	for _, info := range hashes {
		if info.Category != "hashs" {
			t.Fatalf("hash filter included %q from %q", info.Name, info.Category)
		}
	}
	favorites := filterTransformerCatalog("", favoriteTransformerCategory, []string{"hex", "base64"}, nil)
	if len(favorites) != 2 || favorites[0].Name != "hex" || favorites[1].Name != "base64" {
		t.Fatalf("favorite filter order = %#v, want hex then base64", favorites)
	}
	recentSearch := filterTransformerCatalog("base", recentTransformerCategory, nil, []string{"hex", "base64"})
	if len(recentSearch) != 1 || recentSearch[0].Name != "base64" {
		t.Fatalf("recent search = %#v, want base64", recentSearch)
	}
}

func TestTransformerHistorySanitizesAndPrioritizes(t *testing.T) {
	got := sanitizeTransformerHistory([]string{"base64", "missing", "hex", "base64", "url"}, 2)
	if len(got) != 2 || got[0] != "base64" || got[1] != "hex" {
		t.Fatalf("sanitized history = %#v, want [base64 hex]", got)
	}
	got = prependTransformer(got, "hex", maxRecentTransformers)
	if len(got) != 2 || got[0] != "hex" || got[1] != "base64" {
		t.Fatalf("reprioritized history = %#v, want [hex base64]", got)
	}
}

func TestAddTransformerWorkspaceSelection(t *testing.T) {
	a := fynetest.NewTempApp(t)
	dg := &DeenGUI{app: a, window: a.NewWindow("test"), pipe: pipeline.New()}
	content := dg.newAddSlot()
	catalog := dg.addCatalog
	if content == nil || catalog == nil || catalog.search == nil || catalog.category == nil || catalog.results == nil {
		t.Fatal("add-transformer workspace did not initialize its catalog controls")
	}
	if catalog.selected != 0 || len(catalog.matches) == 0 || len(catalog.detail.Objects) == 0 {
		t.Fatalf("initial catalog selection=%d matches=%d details=%d", catalog.selected, len(catalog.matches), len(catalog.detail.Objects))
	}
	if catalog.split == nil || catalog.split.Offset != defaultAddSplit {
		t.Fatalf("add catalog split = %v, want default %v", catalog.split, defaultAddSplit)
	}
	if catalog.resultSummary == nil || catalog.resultSummary.Text != transformerResultSummary(len(catalog.matches)) {
		t.Fatalf("result summary = %v, want count for %d matches", catalog.resultSummary, len(catalog.matches))
	}
	firstSelection := catalog.selected
	catalog.search.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
	if catalog.selected != firstSelection+1 {
		t.Fatalf("Down from search selected %d, want %d", catalog.selected, firstSelection+1)
	}
	catalog.search.TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
	if catalog.selected != firstSelection {
		t.Fatalf("Up from search selected %d, want %d", catalog.selected, firstSelection)
	}
	catalog.search.TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
	if catalog.selected != firstSelection {
		t.Fatalf("Up crossed first result: selected %d, want %d", catalog.selected, firstSelection)
	}
	catalog.search.SetText("base64")
	foundBase64 := false
	for _, info := range catalog.matches {
		foundBase64 = foundBase64 || info.Name == "base64"
	}
	if !foundBase64 {
		t.Fatalf("base64 search matches = %#v", catalog.matches)
	}
	catalog.category.SetSelected("Hashes")
	if len(catalog.matches) != 0 || catalog.selected != -1 {
		t.Fatalf("conflicting filters matches=%d selection=%d, want empty", len(catalog.matches), catalog.selected)
	}

	catalog.search.SetText("")
	catalog.category.SetSelected(allTransformerCategories)
	base64 := catalog.matches[0]
	for _, info := range catalog.matches {
		if info.Name == "base64" {
			base64 = info
			break
		}
	}
	catalog.renderDetail(base64)
	workControlCount := len(catalog.controls)
	catalog.renderDetail(base64)
	if len(catalog.controls) != workControlCount {
		t.Fatalf("detail rerender retained stale work controls: %d became %d", workControlCount, len(catalog.controls))
	}
	dg.toggleFavoriteTransformer(base64)
	if !containsTransformer(dg.favoriteTransformers, "base64") {
		t.Fatal("favorite toggle did not retain base64")
	}
	if got := a.Preferences().StringList(favoriteTransformersPreferenceKey); len(got) != 1 || got[0] != "base64" {
		t.Fatalf("persisted favorites = %#v, want [base64]", got)
	}
	catalog.category.SetSelected(favoriteTransformerCategory)
	if len(catalog.matches) != 1 || catalog.matches[0].Name != "base64" {
		t.Fatalf("favorite catalog = %#v, want base64", catalog.matches)
	}
	if got := a.Preferences().String(addCategoryPreferenceKey); got != favoriteTransformerCategory {
		t.Fatalf("persisted Add category = %q, want Favorites", got)
	}
	dg.recordRecentTransformer("hex")
	dg.recordRecentTransformer("base64")
	dg.recordRecentTransformer("hex")
	catalog.category.SetSelected(recentTransformerCategory)
	if len(catalog.matches) != 2 || catalog.matches[0].Name != "hex" || catalog.matches[1].Name != "base64" {
		t.Fatalf("recent catalog = %#v, want hex then base64", catalog.matches)
	}
	if got := a.Preferences().StringList(recentTransformersPreferenceKey); len(got) != 2 || got[0] != "hex" || got[1] != "base64" {
		t.Fatalf("persisted recents = %#v, want [hex base64]", got)
	}
	catalog.split.Offset = 0.57
	dg.rememberAddSplit()
	if a.Preferences().Float(addSplitPreferenceKey) != 0.57 {
		t.Fatalf("remembered add split = %v, want 0.57", a.Preferences().Float(addSplitPreferenceKey))
	}
}

func TestTransformerCatalogEmptyMessages(t *testing.T) {
	if got := emptyTransformerCatalogMessage("", favoriteTransformerCategory); !strings.Contains(got, "No favorites") {
		t.Fatalf("favorite empty message = %q", got)
	}
	if got := emptyTransformerCatalogMessage("", recentTransformerCategory); !strings.Contains(got, "No recent") {
		t.Fatalf("recent empty message = %q", got)
	}
	if got := emptyTransformerCatalogMessage("base64", favoriteTransformerCategory); !strings.Contains(got, "No transformers match") {
		t.Fatalf("filtered empty message = %q", got)
	}
	if got := transformerResultSummary(1); got != "1 transformer" {
		t.Fatalf("single result summary = %q", got)
	}
}

func TestTransformerWorkspaceUsesSharedCatalogState(t *testing.T) {
	dg := newVisualScenarioGUI(t, appearanceLight)
	dg.newAddSlot()
	content := dg.pluginsTab()
	addCatalog := dg.addCatalog
	browser := dg.browserCatalog
	if content == nil || addCatalog == nil || browser == nil || browser == addCatalog {
		t.Fatal("Transformers workspace did not create an independent reusable catalog view")
	}
	if len(browser.matches) != len(plugins.UICatalog()) || browser.selected != 0 || len(browser.detail.Objects) == 0 {
		t.Fatalf("browser matches=%d selected=%d details=%d", len(browser.matches), browser.selected, len(browser.detail.Objects))
	}

	browser.category.SetSelected("Hashes")
	if got := dg.app.Preferences().String(browserCategoryPreferenceKey); got != "Hashes" {
		t.Fatalf("persisted browser category = %q, want Hashes", got)
	}
	browser.split.Offset = 0.59
	dg.rememberWorkspaceState()
	if got := dg.app.Preferences().Float(browserSplitPreferenceKey); got != 0.59 {
		t.Fatalf("persisted browser split = %v, want 0.59", got)
	}

	addCatalog.category.SetSelected(favoriteTransformerCategory)
	browser.search.SetText("")
	browser.category.SetSelected(favoriteTransformerCategory)
	var base64 plugins.UIPluginInfo
	for _, info := range plugins.UICatalog() {
		if info.Name == "base64" {
			base64 = info
			break
		}
	}
	if base64.Name == "" {
		t.Fatal("base64 is missing from the transformer catalog")
	}
	dg.toggleFavoriteTransformer(base64)
	for name, catalog := range map[string]*transformerCatalog{"Add": addCatalog, "Browse": browser} {
		if len(catalog.matches) != 1 || catalog.matches[0].Name != "base64" {
			t.Fatalf("%s favorite view = %#v, want base64", name, catalog.matches)
		}
	}

	dg.setWorking("Testing", true)
	if !addCatalog.search.Disabled() || !browser.search.Disabled() {
		t.Fatal("all live catalog controls should disable during background work")
	}
	dg.setWorking("", false)
	if addCatalog.search.Disabled() || browser.search.Disabled() {
		t.Fatal("catalog controls should re-enable after background work")
	}
}

func TestWorkflowLibraryFilteringNavigationAndPersistence(t *testing.T) {
	dg := newVisualScenarioGUI(t, appearanceLight)
	dg.selectTab(1)
	library := dg.workflowLibrary
	if library == nil || library.search == nil || library.filter == nil || library.results == nil || len(library.detail.Objects) == 0 {
		t.Fatal("Workflows workspace did not initialize its master/detail controls")
	}
	wantCount := len(pipeline.BuiltinExamples()) + len(pipeline.BuiltinPresets())
	if len(library.items) != wantCount || len(library.matches) != wantCount || library.selected != 0 {
		t.Fatalf("workflow items=%d matches=%d selected=%d, want %d/%d/0", len(library.items), len(library.matches), library.selected, wantCount, wantCount)
	}
	if workflowStepCountLabel(1) != "1 step" || workflowStepCountLabel(2) != "2 steps" {
		t.Fatal("workflow step count labels should use correct singular and plural forms")
	}

	library.filter.SetSelected(presetWorkflowsFilter)
	if got := dg.app.Preferences().String(workflowFilterPreferenceKey); got != presetWorkflowsFilter {
		t.Fatalf("persisted workflow filter = %q, want Presets", got)
	}
	for _, item := range library.matches {
		if item.kind != workflowPreset {
			t.Fatalf("preset filter included %q kind %q", item.name, item.kind)
		}
	}
	library.search.SetText("JWT")
	if len(library.matches) != 1 || library.matches[0].name != "Decode JWT" {
		t.Fatalf("preset JWT search = %#v", library.matches)
	}
	library.search.SetText("")
	if len(library.matches) < 2 {
		t.Fatal("preset filter should expose multiple workflows for keyboard navigation")
	}
	first := library.selected
	library.search.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
	if library.selected != first+1 {
		t.Fatalf("workflow Down selected %d, want %d", library.selected, first+1)
	}

	library.split.Offset = 0.58
	dg.rememberWorkspaceState()
	if got := dg.app.Preferences().Float(workflowSplitPreferenceKey); got != 0.58 {
		t.Fatalf("persisted workflow split = %v, want 0.58", got)
	}
	dg.selectTab(0)
	dg.showPresets()
	if dg.activeTab != 1 || dg.workflowLibrary.filter.Selected != presetWorkflowsFilter || dg.window.Canvas().Focused() != dg.workflowLibrary.search {
		t.Fatal("Presets command should open and focus the filtered Workflows workspace")
	}
	dg.pipe.SetSource([]byte("fresh input"))
	dg.selectTab(0)
	dg.selectTab(1)
	wantInputSummary := pipeline.DataMetadata(dg.pipe.Source(), 0).Summary()
	if library.currentInputSummary == nil || library.currentInputSummary.Text != wantInputSummary {
		t.Fatalf("refreshed preset input summary = %v, want %q", library.currentInputSummary, wantInputSummary)
	}

	dg.working = true
	dg.setWorking("Testing", true)
	if !library.search.Disabled() || !library.filter.Disabled() {
		t.Fatal("workflow controls should disable during background work")
	}
	dg.working = false
	dg.setWorking("", false)
	if library.search.Disabled() || library.filter.Disabled() {
		t.Fatal("workflow controls should re-enable after background work")
	}
}

func TestWorkflowApplicationMakesInputSemanticsExplicit(t *testing.T) {
	dg := newVisualScenarioGUI(t, appearanceDark)
	dg.pipe.SetSource([]byte("dGVzdA=="))
	dg.pipe.AddStep("hex", false)
	dg.sourceName = "current.txt"
	dg.rebuild()

	wait := func(label string, completed <-chan struct{}) {
		t.Helper()
		select {
		case <-completed:
		case <-time.After(3 * time.Second):
			t.Fatalf("%s did not finish", label)
		}
	}

	presetItem := workflowItem{
		kind:   workflowPreset,
		name:   "Decode current Base64",
		detail: "test preset",
		preset: pipeline.Preset{
			Name:  "Decode current Base64",
			Steps: []pipeline.PresetStep{{Plugin: "base64", Unprocess: true}},
		},
	}
	wait("preset application", dg.applyWorkflow(presetItem))
	if got := string(dg.pipe.Source()); got != "dGVzdA==" {
		t.Fatalf("preset replaced source with %q", got)
	}
	if got := string(dg.pipe.Result()); got != "test" || dg.sourceName != "current.txt" {
		t.Fatalf("preset result=%q sourceName=%q", got, dg.sourceName)
	}
	if got := dg.workStatus.Text; got != "Applied preset “Decode current Base64”" || dg.workStatus.Importance != widget.SuccessImportance {
		t.Fatalf("preset feedback = %q importance=%v", got, dg.workStatus.Importance)
	}

	exampleItem := workflowItem{
		kind:   workflowExample,
		name:   "Hex sample",
		detail: "test example",
		example: pipeline.Example{
			Name:   "Hex sample",
			Source: []byte("6869"),
			Steps:  []pipeline.PresetStep{{Plugin: "hex", Unprocess: true}},
		},
	}
	wait("example application", dg.applyWorkflow(exampleItem))
	if got := string(dg.pipe.Source()); got != "6869" {
		t.Fatalf("example source = %q, want bundled input", got)
	}
	if got := string(dg.pipe.Result()); got != "hi" || dg.sourceName != "" || dg.selectedStage != pipelineStageInput {
		t.Fatalf("example result=%q sourceName=%q selected=%d", got, dg.sourceName, dg.selectedStage)
	}
	if got := dg.workStatus.Text; got != "Loaded example “Hex sample”" || dg.workStatus.Importance != widget.SuccessImportance {
		t.Fatalf("example feedback = %q importance=%v", got, dg.workStatus.Importance)
	}

	dg.selectTab(1)
	library := dg.workflowLibrary
	library.items = []workflowItem{exampleItem}
	library.search.SetText("")
	library.filter.SetSelected(allWorkflowsFilter)
	library.refresh()
	preview := library.previewSlot
	if preview == nil {
		t.Fatal("example detail did not create its preview slot")
	}
	wait("example preview", library.previewExample(exampleItem, library.detailVersion, preview))
	if len(preview.Objects) != 2 {
		t.Fatalf("example preview object count = %d, want summary and data panels", len(preview.Objects))
	}
	if library.detailScroll == nil || library.detailScroll.Direction != container.ScrollVerticalOnly {
		t.Fatal("workflow details should use a vertical scroll viewport")
	}
	if library.previewInput == nil || library.previewOutput == nil {
		t.Fatal("example preview did not retain its input and output views")
	}
	shortEntry := multilineEntry(5)
	if library.previewInput.MinSize().Height <= shortEntry.MinSize().Height || library.previewOutput.MinSize().Height <= shortEntry.MinSize().Height {
		t.Fatalf("preview heights input=%v output=%v short=%v", library.previewInput.MinSize().Height, library.previewOutput.MinSize().Height, shortEntry.MinSize().Height)
	}
	previewPanels, ok := preview.Objects[1].(*fyne.Container)
	if !ok || len(previewPanels.Objects) != 2 {
		t.Fatalf("preview panels = %T with unexpected contents", preview.Objects[1])
	}
	library.detailScroll.Resize(fyne.NewSize(480, 280))
	library.detailScroll.ScrollToTop()
	library.detailScroll.ScrollToBottom()
	if library.detail.MinSize().Height <= library.detailScroll.Size().Height || library.detailScroll.Offset.Y <= 0 {
		t.Fatalf("detail scroll content=%v viewport=%v offset=%v", library.detail.MinSize(), library.detailScroll.Size(), library.detailScroll.Offset)
	}
	library.renderDetail(presetItem)
	if library.detailScroll.Offset != (fyne.Position{}) {
		t.Fatalf("selecting another workflow retained preview scroll offset %v", library.detailScroll.Offset)
	}

	before := dg.pipe.Len()
	wait("invalid workflow", dg.applyWorkflow(workflowItem{}))
	if dg.pipe.Len() != before {
		t.Fatal("invalid workflow mutated the pipeline")
	}
}

func TestInputWorkspaceSplitAndDynamicPreview(t *testing.T) {
	a := fynetest.NewTempApp(t)
	p := pipeline.New()
	p.SetSource([]byte(`{"ok":true}`))
	dg := &DeenGUI{
		app:           a,
		window:        a.NewWindow("test"),
		pipe:          p,
		sourcePane:    "Inspector",
		sourceView:    "Preview",
		workActivity:  widget.NewActivity(),
		workStatus:    widget.NewLabel(""),
		workIndicator: container.NewHBox(),
	}
	dg.newSourceCard()
	if dg.sourceWorkspace == nil || len(dg.sourceWorkspace.Items) != 2 || dg.sourceWorkspace.Selected().Text != "Inspector" {
		t.Fatal("input workspace should expose and restore exclusive Editor/Inspector views")
	}
	if dg.sourceWorkspace.Items[0].Content.Visible() || !dg.sourceWorkspace.Items[1].Content.Visible() {
		t.Fatal("only the selected Input view should be visible")
	}
	if dg.sourceViewer == nil || len(dg.sourceViewer.Items) != 4 || dg.sourceViewer.Selected().Text != "Preview" {
		t.Fatalf("source inspector tabs=%d selected=%q, want four tabs and Preview", len(dg.sourceViewer.Items), dg.sourceViewer.Selected().Text)
	}
	if dg.sourceEntry.Disabled() || !dg.sourceRaw.Disabled() || !dg.sourceHex.Disabled() || !dg.sourceStrings.Disabled() {
		t.Fatal("source editor/inspection editability is incorrect")
	}
	dg.sourceViewer.SelectIndex(0)
	if got := a.Preferences().String(sourceViewPreferenceKey); got != "Raw" {
		t.Fatalf("persisted source view = %q, want Raw", got)
	}
	dg.sourceWorkspace.SelectIndex(0)
	if got := a.Preferences().String(sourcePanePreferenceKey); got != "Editor" {
		t.Fatalf("persisted source pane = %q, want Editor", got)
	}
	if !dg.sourceWorkspace.Items[0].Content.Visible() || dg.sourceWorkspace.Items[1].Content.Visible() {
		t.Fatal("selecting Editor should hide Inspector")
	}
	dg.newSourceCard()
	if dg.sourceWorkspace.Selected().Text != "Editor" {
		t.Fatal("rebuilding Input should restore the selected outer view")
	}

	p.SetSource([]byte("plain text"))
	dg.refreshFrom(0)
	if dg.sourcePreview != nil || len(dg.sourceViewer.Items) != 3 {
		t.Fatal("structured preview tab should disappear for plain text")
	}
	p.SetSource([]byte(`{"again":true}`))
	dg.refreshFrom(0)
	if dg.sourcePreview == nil || len(dg.sourceViewer.Items) != 4 || dg.sourceViewer.Selected().Text != "Preview" {
		t.Fatal("structured preview tab should return and become selected")
	}

}

func TestTypingInputKeepsFocusAndUpdatesPipeline(t *testing.T) {
	dg := newVisualScenarioGUI(t, appearanceLight)
	dg.pipe.SetSource([]byte("ac"))
	dg.rebuild()
	entry := dg.sourceEntry
	if entry == nil {
		t.Fatal("Input Editor did not create its editable source")
	}

	entry.CursorColumn = 1
	dg.window.Canvas().Focus(entry)
	fynetest.Type(entry, "bd")
	fyne.DoAndWait(func() {})
	if got := entry.Text; got != "abdc" {
		t.Fatalf("typed source = %q, want %q", got, "abdc")
	}
	if got := string(dg.pipe.Source()); got != "abdc" {
		t.Fatalf("pipeline source = %q, want %q", got, "abdc")
	}
	if focused := dg.window.Canvas().Focused(); focused != entry {
		t.Fatalf("Input focus moved to %T after typing", focused)
	}
}

func TestTypingStepOutputKeepsFocusAndUpdatesPipeline(t *testing.T) {
	dg := newVisualScenarioGUI(t, appearanceDark)
	dg.pipe.SetSource([]byte("ac"))
	dg.pipe.AddStep("base64", false)
	dg.selectedStage = 0
	dg.rebuild()
	entry := dg.cards[0].body
	if entry == nil {
		t.Fatal("focused step did not create its editable output")
	}
	if entry.Text != "YWM=" {
		t.Fatalf("initial step output = %q, want %q", entry.Text, "YWM=")
	}

	entry.CursorColumn = 1
	dg.window.Canvas().Focus(entry)
	fynetest.Type(entry, "bd")
	fyne.DoAndWait(func() {})
	if got := entry.Text; got != "YbdWM=" {
		t.Fatalf("edited step output = %q, want %q", got, "YbdWM=")
	}
	if got := string(dg.pipe.Output(0)); got != "YbdWM=" {
		t.Fatalf("pipeline step output = %q, want %q", got, "YbdWM=")
	}
	if focused := dg.window.Canvas().Focused(); focused != entry {
		t.Fatalf("step output focus moved to %T after typing", focused)
	}
}

func TestTypingStepOptionKeepsFocusAndUpdatesPipeline(t *testing.T) {
	dg := newVisualScenarioGUI(t, appearanceLight)
	dg.pipe.SetSource([]byte("compress me"))
	dg.pipe.AddStep("gzip", false)
	dg.selectedStage = 0
	dg.rebuild()
	card := dg.cards[0]

	var findEditableEntry func(fyne.CanvasObject) *widget.Entry
	findEditableEntry = func(object fyne.CanvasObject) *widget.Entry {
		if entry, ok := object.(*widget.Entry); ok && !entry.Disabled() {
			return entry
		}
		if children, ok := object.(*fyne.Container); ok {
			for _, child := range children.Objects {
				if entry := findEditableEntry(child); entry != nil {
					return entry
				}
			}
		}
		return nil
	}
	entry := findEditableEntry(card.options)
	if entry == nil {
		t.Fatal("gzip options did not create an editable level control")
	}

	dg.window.Canvas().Focus(entry)
	fynetest.Type(entry, "5")
	fyne.DoAndWait(func() {})
	if got := dg.pipe.Steps()[0].Options["level"]; got != "5" {
		t.Fatalf("pipeline gzip level = %q, want %q", got, "5")
	}
	if focused := dg.window.Canvas().Focused(); focused != entry {
		t.Fatalf("step option focus moved to %T after typing", focused)
	}
}

func TestFileReaderLoadsIntoInputWorkspace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dropped.txt")
	if err := os.WriteFile(path, []byte("dropped input"), 0o600); err != nil {
		t.Fatal(err)
	}
	dg := newVisualScenarioGUI(t, appearanceDark)
	rc, err := storage.Reader(storage.NewFileURI(path))
	if err != nil {
		t.Fatal(err)
	}
	completed := dg.loadSourceReader(rc, "dropped.txt")
	select {
	case <-completed:
	case <-time.After(3 * time.Second):
		t.Fatal("dropped file processing did not finish")
	}
	if got := string(dg.pipe.Source()); got != "dropped input" {
		t.Fatalf("dropped source = %q", got)
	}
	if dg.sourceName != "dropped.txt" || dg.selectedStage != pipelineStageInput || dg.sourceEntry == nil {
		t.Fatalf("drop state name=%q stage=%d sourceEntry=%v", dg.sourceName, dg.selectedStage, dg.sourceEntry != nil)
	}
	if got := dg.workStatus.Text; got != "Loaded “dropped.txt”" || dg.workStatus.Importance != widget.SuccessImportance {
		t.Fatalf("file-load feedback = %q importance=%v", got, dg.workStatus.Importance)
	}
}

func TestOptionEntryValidation(t *testing.T) {
	find := func(plugin, name string) pipeline.Option {
		for _, option := range pipeline.PluginOptions(plugin) {
			if option.Name == name {
				return option
			}
		}
		t.Fatalf("missing option %s:%s", plugin, name)
		return pipeline.Option{}
	}
	validator := optionEntryValidator("gzip", find("gzip", "level"))
	if validator == nil || validator("9") != nil || validator("not-a-number") == nil || validator("") == nil {
		t.Fatal("numeric option validator did not accept integers and reject invalid values")
	}
	if validator := optionEntryValidator("xor", find("xor", "value")); validator != nil {
		t.Fatal("arithmetic values must continue accepting hex and character operands")
	}
	secret := find("jwt", "secret")
	if help := optionHelp(secret); help == nil {
		t.Fatal("secret option should include its security guidance")
	}
}

func TestAdversecThemeUsesRequestedSystemVariant(t *testing.T) {
	system := newSystemAdversecTheme()
	dark := color.NRGBAModel.Convert(system.Color(theme.ColorNameBackground, theme.VariantDark))
	light := color.NRGBAModel.Convert(system.Color(theme.ColorNameBackground, theme.VariantLight))
	if dark == light {
		t.Fatal("system appearance should expose distinct dark and light palettes")
	}

	fixedDark := newAdversecTheme(theme.VariantDark)
	got := color.NRGBAModel.Convert(fixedDark.Color(theme.ColorNameBackground, theme.VariantLight))
	if got != dark {
		t.Fatalf("fixed dark theme followed requested light variant: got %v, want %v", got, dark)
	}
}

func TestAdversecThemeModernSizes(t *testing.T) {
	th := newAdversecTheme(theme.VariantDark)
	tests := map[fyne.ThemeSizeName]float32{
		theme.SizeNameButtonRadius:    7,
		theme.SizeNameCardRadius:      10,
		theme.SizeNameDialogRadius:    12,
		theme.SizeNameInputRadius:     7,
		theme.SizeNameMenuRadius:      8,
		theme.SizeNameModalBlurRadius: 2.5,
		theme.SizeNamePopupRadius:     9,
		theme.SizeNameSelectionRadius: 5,
	}
	for name, want := range tests {
		if got := th.Size(name); got != want {
			t.Errorf("theme size %q = %v, want %v", name, got, want)
		}
	}
	if got := th.Size(theme.SizeNameText); got <= 0 {
		t.Fatalf("fallback text size = %v, want positive value", got)
	}
}

func TestStepSurfaceUsesHardwareShadowAndAccentRail(t *testing.T) {
	a := fynetest.NewTempApp(t)
	a.Settings().SetTheme(newAdversecTheme(theme.VariantDark))
	accentColor := color.NRGBA{R: 0x38, G: 0xd9, B: 0xc8, A: 0xff}
	surface := newStepSurface(widget.NewLabel("content"), accentColor, false)

	outer, ok := surface.(*fyne.Container)
	if !ok {
		t.Fatalf("step surface outer object = %T, want *fyne.Container", surface)
	}
	if len(outer.Objects) != 1 {
		t.Fatalf("step surface outer child count = %d, want 1", len(outer.Objects))
	}
	stack, ok := outer.Objects[0].(*fyne.Container)
	if !ok {
		t.Fatalf("step surface stack = %T, want *fyne.Container", outer.Objects[0])
	}
	if len(stack.Objects) != 2 {
		t.Fatalf("step surface stack child count = %d, want 2", len(stack.Objects))
	}
	background, ok := stack.Objects[0].(*canvas.Rectangle)
	if !ok {
		t.Fatalf("step surface background = %T, want *canvas.Rectangle", stack.Objects[0])
	}
	if background.Shadow.BlurRadius != 12 || background.Shadow.Variant != canvas.BoxShadow {
		t.Fatalf("step surface shadow = %+v", background.Shadow)
	}
	if background.CornerRadius != 10 {
		t.Fatalf("step surface corner radius = %v, want 10", background.CornerRadius)
	}

	body := stack.Objects[1].(*fyne.Container)
	var rail *canvas.Rectangle
	for _, object := range body.Objects {
		if rectangle, ok := object.(*canvas.Rectangle); ok {
			rail = rectangle
			break
		}
	}
	if rail == nil {
		t.Fatal("step surface has no accent rail rectangle")
	}
	if got := color.NRGBAModel.Convert(rail.FillColor); got != color.NRGBAModel.Convert(accentColor) {
		t.Fatalf("accent rail color = %v, want %v", got, accentColor)
	}
}

func TestNavTabAccessibility(t *testing.T) {
	a := fynetest.NewTempApp(t)
	a.Settings().SetTheme(newAdversecTheme(theme.VariantDark))
	taps := 0
	tab := newSidebarNavItem("Transformers", theme.SearchIcon(), func() { taps++ })
	if got := tab.AccessibilityLabel(); got != "Transformers" {
		t.Fatalf("AccessibilityLabel() = %q, want Transformers", got)
	}
	if got := tab.AccessibilityRole(); got != fyne.AccessibleRoleButton {
		t.Fatalf("AccessibilityRole() = %q, want button", got)
	}
	tab.setActive(true)
	if active := color.NRGBAModel.Convert(tab.background.FillColor).(color.NRGBA); active.A == 0 {
		t.Fatal("active sidebar destination should have a visible selection surface")
	}
	tab.setActive(false)
	if inactive := color.NRGBAModel.Convert(tab.background.FillColor).(color.NRGBA); inactive.A != 0 {
		t.Fatalf("inactive sidebar destination background alpha = %d, want 0", inactive.A)
	}
	tab.FocusGained()
	if !tab.focused || tab.background.StrokeWidth == 0 {
		t.Fatal("focused sidebar destination should expose a visible focus outline")
	}
	for _, key := range []fyne.KeyName{fyne.KeySpace, fyne.KeyReturn, fyne.KeyEnter} {
		tab.TypedKey(&fyne.KeyEvent{Name: key})
	}
	if taps != 3 {
		t.Fatalf("keyboard activation count = %d, want 3", taps)
	}
	tab.TypedKey(nil)
	tab.TypedKey(&fyne.KeyEvent{Name: fyne.KeyEscape})
	if taps != 3 {
		t.Fatalf("unhandled keys changed activation count to %d", taps)
	}
	tab.FocusLost()
	if tab.focused || tab.background.StrokeWidth != 0 {
		t.Fatal("sidebar destination retained its focus outline after losing focus")
	}
}

func TestMainMenuUsesDesktopNavigationAndShortcuts(t *testing.T) {
	a := fynetest.NewTempApp(t)
	dg := &DeenGUI{app: a, window: a.NewWindow("test"), pipe: pipeline.New(), sidebarOpen: true}
	menu := dg.mainMenu()
	if !dg.undoMenuItem.Disabled || !dg.redoMenuItem.Disabled || !dg.clearMenuItem.Disabled || dg.addMenuItem.Disabled {
		t.Fatal("fresh-pipeline native menu state is incorrect")
	}
	dg.pipe.SetSource([]byte("changed"))
	dg.refreshNativeMenuState()
	if dg.undoMenuItem.Disabled || dg.clearMenuItem.Disabled {
		t.Fatal("undo and clear should enable after a pipeline mutation")
	}
	dg.working = true
	dg.refreshNativeMenuState()
	if !dg.undoMenuItem.Disabled || !dg.addMenuItem.Disabled || !dg.clearMenuItem.Disabled {
		t.Fatal("mutating native menu commands should disable while work is running")
	}
	dg.working = false
	dg.refreshNativeMenuState()
	wantMenus := []string{"File", "Edit", "View", "Navigate", "Chain", "Workflow", "Appearance", "Help"}
	if len(menu.Items) != len(wantMenus) {
		t.Fatalf("menu count = %d, want %d", len(menu.Items), len(wantMenus))
	}
	for i, want := range wantMenus {
		if got := menu.Items[i].Label; got != want {
			t.Fatalf("menu %d = %q, want %q", i, got, want)
		}
	}
	for _, item := range []*fyne.MenuItem{
		menu.Items[0].Items[0], // Open
		menu.Items[0].Items[1], // Save
		menu.Items[1].Items[0], // Undo
		menu.Items[1].Items[1], // Redo
		menu.Items[2].Items[0], // Sidebar
		menu.Items[2].Items[1], // Pipeline navigator
		menu.Items[3].Items[0], // Pipeline
		menu.Items[3].Items[6], // Previous pipeline stage
		menu.Items[3].Items[7], // Next pipeline stage
		menu.Items[5].Items[0], // Add transformer
		menu.Items[5].Items[2], // Toggle selected step
		menu.Items[5].Items[3], // Move selected step up
		menu.Items[5].Items[4], // Move selected step down
		menu.Items[5].Items[5], // Duplicate selected step
	} {
		if item.Shortcut == nil {
			t.Fatalf("menu item %q should expose a desktop shortcut", item.Label)
		}
	}
	wantStepShortcuts := map[*fyne.MenuItem]struct {
		key      fyne.KeyName
		modifier fyne.KeyModifier
	}{
		dg.toggleStepMenuItem:    {fyne.KeyE, fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift},
		dg.moveStepUpMenuItem:    {fyne.KeyUp, fyne.KeyModifierShortcutDefault | fyne.KeyModifierAlt | fyne.KeyModifierShift},
		dg.moveStepDownMenuItem:  {fyne.KeyDown, fyne.KeyModifierShortcutDefault | fyne.KeyModifierAlt | fyne.KeyModifierShift},
		dg.duplicateStepMenuItem: {fyne.KeyD, fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift},
	}
	for item, want := range wantStepShortcuts {
		shortcut, ok := item.Shortcut.(*desktop.CustomShortcut)
		if !ok || shortcut.KeyName != want.key || shortcut.Modifier != want.modifier {
			t.Fatalf("shortcut for %q = %#v, want %v/%v", item.Label, item.Shortcut, want.key, want.modifier)
		}
	}
	seen := make(map[string]string)
	for _, submenu := range menu.Items {
		for _, item := range submenu.Items {
			if item.Shortcut == nil {
				continue
			}
			name := item.Shortcut.ShortcutName()
			if previous, exists := seen[name]; exists {
				t.Fatalf("shortcut %q is assigned to both %q and %q", name, previous, item.Label)
			}
			seen[name] = item.Label
		}
	}
}

func TestNativePipelineStageNavigation(t *testing.T) {
	dg := newVisualScenarioGUI(t, appearanceDark)
	dg.pipe.AddStep("base64", false)
	dg.pipe.AddStep("hex", false)
	dg.selectedStage = pipelineStageInput
	dg.rebuild()
	dg.mainMenu()
	if !dg.previousStageMenuItem.Disabled || dg.nextStageMenuItem.Disabled {
		t.Fatal("input boundary should disable previous and enable next stage")
	}
	if !dg.toggleStepMenuItem.Disabled || !dg.moveStepUpMenuItem.Disabled || !dg.moveStepDownMenuItem.Disabled || !dg.duplicateStepMenuItem.Disabled || !dg.removeStepMenuItem.Disabled {
		t.Fatal("selected-step menu actions should disable while Input is focused")
	}

	dg.navigatePipelineStage(1)
	if dg.selectedStage != 0 || dg.previousStageMenuItem.Disabled || dg.nextStageMenuItem.Disabled {
		t.Fatalf("after next stage selection=%d previousDisabled=%v nextDisabled=%v", dg.selectedStage, dg.previousStageMenuItem.Disabled, dg.nextStageMenuItem.Disabled)
	}
	if dg.toggleStepMenuItem.Disabled || !dg.moveStepUpMenuItem.Disabled || dg.moveStepDownMenuItem.Disabled || dg.duplicateStepMenuItem.Disabled || dg.removeStepMenuItem.Disabled {
		t.Fatal("first-step menu action state is incorrect")
	}
	dg.navigatePipelineStage(1)
	dg.navigatePipelineStage(1)
	if dg.selectedStage != pipelineStageAdd || dg.nextStageMenuItem.Disabled == false {
		t.Fatalf("final navigation selection=%d nextDisabled=%v, want Add/disabled", dg.selectedStage, dg.nextStageMenuItem.Disabled)
	}
	dg.navigatePipelineStage(-1)
	if dg.selectedStage != 1 {
		t.Fatalf("previous from Add selected stage %d, want final transformer", dg.selectedStage)
	}

	dg.selectTab(1)
	if !dg.previousStageMenuItem.Disabled || !dg.nextStageMenuItem.Disabled {
		t.Fatal("pipeline stage navigation should disable outside the Pipeline workspace")
	}
	if !dg.toggleStepMenuItem.Disabled || !dg.duplicateStepMenuItem.Disabled || !dg.removeStepMenuItem.Disabled {
		t.Fatal("selected-step actions should disable outside the Pipeline workspace")
	}
	dg.navigatePipelineStage(-1)
	if dg.selectedStage != 1 {
		t.Fatal("disabled stage navigation changed selection outside Pipeline")
	}
}

func TestReusableStepActionsPreserveSelectionAndBoundaries(t *testing.T) {
	dg := newVisualScenarioGUI(t, appearanceDark)
	dg.pipe.AddStep("base64", false)
	dg.pipe.AddStep("hex", false)
	dg.selectedStage = 0
	dg.rebuild()

	wait := func(label string, completed <-chan struct{}) {
		t.Helper()
		select {
		case <-completed:
		case <-time.After(3 * time.Second):
			t.Fatalf("%s did not finish", label)
		}
	}

	wait("toggle", dg.toggleStep(0))
	if !dg.pipe.Steps()[0].Disabled || dg.selectedStage != 0 {
		t.Fatalf("toggle disabled=%v selected=%d", dg.pipe.Steps()[0].Disabled, dg.selectedStage)
	}
	wait("duplicate", dg.duplicateStep(0))
	if dg.pipe.Len() != 3 || dg.selectedStage != 1 || dg.pipe.Steps()[1].Plugin != "base64" {
		t.Fatalf("duplicate len=%d selected=%d steps=%#v", dg.pipe.Len(), dg.selectedStage, dg.pipe.Steps())
	}
	wait("move", dg.moveStep(1, 1))
	if dg.selectedStage != 2 || dg.pipe.Steps()[2].Plugin != "base64" {
		t.Fatalf("move selected=%d steps=%#v", dg.selectedStage, dg.pipe.Steps())
	}
	wait("remove", dg.removeStep(2))
	if dg.pipe.Len() != 2 || dg.selectedStage != 1 {
		t.Fatalf("remove len=%d selected=%d", dg.pipe.Len(), dg.selectedStage)
	}

	before := dg.pipe.Len()
	wait("invalid move", dg.moveStep(0, -1))
	wait("invalid remove", dg.removeStep(99))
	if dg.pipe.Len() != before || dg.selectedStage != 1 {
		t.Fatal("invalid step actions mutated the pipeline or selection")
	}
}

func TestPipelineMasterDetailNavigation(t *testing.T) {
	a := fynetest.NewTempApp(t)
	a.Settings().SetTheme(newAdversecTheme(theme.VariantDark))
	p := pipeline.New()
	p.SetSource([]byte("hello"))
	p.AddStep("base64", false)
	p.AddStep("base64", true)
	dg := &DeenGUI{
		app:             a,
		window:          a.NewWindow("test"),
		pipe:            p,
		stepsBox:        container.NewVBox(),
		pipelineOutline: container.NewVBox(),
		selectedStage:   pipelineStageInput,
		pipelineNavOpen: true,
		stepEditorSplit: defaultStepEditorSplit,
		activeTab:       0,
	}
	dg.workspaceToolbar()
	if dg.sidebarButton.AccessibilityLabel() != "Sidebar" || dg.pipelineNavButton.AccessibilityLabel() != "Stages" {
		t.Fatal("navigator toolbar controls should expose semantic accessibility labels")
	}
	dg.homeTab()
	dg.rebuild()

	if got := len(dg.pipelineStageButtons); got != 4 {
		t.Fatalf("pipeline navigator destinations = %d, want input + 2 steps + add", got)
	}
	if !dg.pipelineStageButtons[0].active || dg.sourceEntry == nil {
		t.Fatal("input should be the initially focused pipeline stage")
	}
	if len(dg.stepsBox.Objects) != 1 {
		t.Fatalf("focused detail object count = %d, want 1", len(dg.stepsBox.Objects))
	}

	dg.selectPipelineStage(0)
	if dg.selectedStage != 0 || !dg.pipelineStageButtons[1].active {
		t.Fatalf("selected stage = %d, first step active=%v", dg.selectedStage, dg.pipelineStageButtons[1].active)
	}
	if dg.cards[0] == nil || dg.cards[1] != nil || dg.sourceEntry != nil {
		t.Fatal("detail editor should construct only the selected step")
	}
	if got := a.Preferences().Int(selectedStagePreferenceKey); got != 0 {
		t.Fatalf("persisted selected stage = %d, want 0", got)
	}
	card := dg.cards[0]
	wantActions := []string{"Collapse step", "Disable step", "Move step up", "Move step down", "Duplicate step", "Remove step"}
	if len(card.headerActions) != len(wantActions) {
		t.Fatalf("step header action count = %d, want %d", len(card.headerActions), len(wantActions))
	}
	for i, want := range wantActions {
		if got := card.headerActions[i].AccessibilityLabel(); got != want {
			t.Errorf("step header action %d accessibility label = %q, want %q", i, got, want)
		}
		if got := card.headerActions[i].AccessibilityRole(); got != fyne.AccessibleRoleButton {
			t.Errorf("step header action %d accessibility role = %q, want button", i, got)
		}
	}
	card.toggleCollapse()
	if got := card.collapse.AccessibilityLabel(); got != "Expand step" {
		t.Fatalf("collapsed disclosure accessibility label = %q, want Expand step", got)
	}
	card.toggleCollapse()
	if got := card.collapse.AccessibilityLabel(); got != "Collapse step" {
		t.Fatalf("expanded disclosure accessibility label = %q, want Collapse step", got)
	}
	if card.editorSplit == nil || !card.editorSplit.Horizontal {
		t.Fatal("focused step should use a horizontal configuration/output split")
	}
	if card.editorSplit.Offset != defaultStepEditorSplit {
		t.Fatalf("editor split offset = %v, want %v", card.editorSplit.Offset, defaultStepEditorSplit)
	}
	card.viewer.SelectIndex(1)
	if got := a.Preferences().String(stepOutputPreferenceKey); got != "Hex" {
		t.Fatalf("persisted step output view = %q, want Hex", got)
	}
	configuration, ok := card.editorSplit.Leading.(*widget.Card)
	if !ok {
		t.Fatalf("leading editor pane = %T, want *widget.Card", card.editorSplit.Leading)
	}
	if configuration.Title != "Configuration" {
		t.Fatalf("leading editor title = %q, want Configuration", configuration.Title)
	}
	output, ok := card.editorSplit.Trailing.(*widget.Card)
	if !ok {
		t.Fatalf("trailing editor pane = %T, want *widget.Card", card.editorSplit.Trailing)
	}
	if output.Title != "Output" {
		t.Fatalf("trailing editor title = %q, want Output", output.Title)
	}
	if got := card.detail.MinSize().Height; got != focusedStepEditorHeight {
		t.Fatalf("focused editor height = %v, want %v", got, focusedStepEditorHeight)
	}

	// Split has no public drag callback. A rebuild captures its live offset.
	card.editorSplit.Offset = 0.56

	dg.selectPipelineStage(pipelineStageAdd)
	if dg.selectedStage != pipelineStageAdd || !dg.pipelineStageButtons[len(dg.pipelineStageButtons)-1].active {
		t.Fatal("add-transformer destination should become the focused detail")
	}
	if dg.stepEditorSplit != 0.56 || a.Preferences().Float(stepEditorSplitPreferenceKey) != 0.56 {
		t.Fatalf("remembered editor split = %v / %v, want 0.56", dg.stepEditorSplit, a.Preferences().Float(stepEditorSplitPreferenceKey))
	}

	dg.togglePipelineNavigator()
	if dg.pipelineNavigator.Visible() || a.Preferences().BoolWithFallback(pipelineNavPreferenceKey, true) {
		t.Fatal("pipeline navigator should hide and persist its closed state")
	}
	dg.togglePipelineNavigator()
	if !dg.pipelineNavigator.Visible() || !a.Preferences().Bool(pipelineNavPreferenceKey) {
		t.Fatal("pipeline navigator should show and persist its open state")
	}
}

func TestPipelineSelectionClampsAfterStructuralChanges(t *testing.T) {
	p := pipeline.New()
	p.AddStep("base64", false)
	p.AddStep("hex", false)
	dg := &DeenGUI{pipe: p, selectedStage: 9}
	dg.clampSelectedStage()
	if dg.selectedStage != 1 {
		t.Fatalf("selection after shrink = %d, want final step 1", dg.selectedStage)
	}

	p.Clear()
	dg.clampSelectedStage()
	if dg.selectedStage != pipelineStageInput {
		t.Fatalf("selection after clear = %d, want input", dg.selectedStage)
	}

	dg.selectedStage = pipelineStageAdd
	dg.clampSelectedStage()
	if dg.selectedStage != pipelineStageAdd {
		t.Fatal("add-transformer destination should remain valid for an empty pipeline")
	}
}

func TestCommandBarTracksTabAndWorkState(t *testing.T) {
	a := fynetest.NewTempApp(t)
	a.Settings().SetTheme(newAdversecTheme(theme.VariantDark))
	dg := &DeenGUI{
		app:           a,
		window:        a.NewWindow("test"),
		pipe:          pipeline.New(),
		stepsBox:      container.NewVBox(),
		selectedStage: pipelineStageInput,
		activeTab:     -1,
		tabContent:    container.NewMax(),
		sidebarOpen:   true,
	}
	dg.navigationPanel = dg.navigationSidebar()
	dg.homeCommands = dg.homeMenuBar()
	dg.workspaceToolbar()
	wantCommandOrder := []string{"Open file", "Save result", "Copy result", "Add step", "More", "Undo", "Redo"}
	if len(dg.commandButtons) != len(wantCommandOrder) {
		t.Fatalf("command count = %d, want %d", len(dg.commandButtons), len(wantCommandOrder))
	}
	for i, want := range wantCommandOrder {
		if got := dg.commandButtons[i].Text; got != want {
			t.Errorf("command %d = %q, want %q", i, got, want)
		}
	}
	if dg.homeCommandScroll == nil {
		t.Fatal("Home commands should be wrapped in a horizontal scroller")
	}
	if got, content := dg.homeCommandScroll.MinSize().Width, dg.homeCommandScroll.Content.MinSize().Width; got >= content {
		t.Fatalf("command scroller minimum width = %v, want less than command content width %v", got, content)
	}
	editable := widget.NewEntry()
	readOnly := widget.NewEntry()
	readOnly.Disable()
	dg.registerWorkControl(editable)
	dg.registerWorkControl(readOnly)
	dg.workActivity = widget.NewActivity()
	dg.workActivity.Hide()
	dg.workStatus = widget.NewLabel("Ready")
	dg.workIndicator = container.NewHBox(dg.workActivity, dg.workStatus)
	dg.resultStatus = widget.NewLabel("")
	dg.tabViews[0] = widget.NewLabel("Home")
	dg.tabViews[1] = widget.NewLabel("Workflows")

	dg.selectTab(0)
	if !dg.homeCommands.Visible() {
		t.Fatal("Home commands should be visible on the Home tab")
	}
	if dg.workspaceTitle.Text != "Pipeline" || dg.workspaceTitle.Visible() || !dg.tabButtons[0].active {
		t.Fatalf("pipeline navigation state title=%q visible=%v active=%v", dg.workspaceTitle.Text, dg.workspaceTitle.Visible(), dg.tabButtons[0].active)
	}
	if !dg.undoCommand.Disabled() || !dg.redoCommand.Disabled() {
		t.Fatal("undo and redo should be disabled for a fresh pipeline")
	}

	dg.pipe.SetSource([]byte("changed"))
	dg.refreshCommandButtons()
	dg.refreshResultStatus()
	if dg.undoCommand.Disabled() {
		t.Fatal("undo should be enabled after a pipeline mutation")
	}
	if dg.resultStatus.Text != "0 steps  •  7 B result" {
		t.Fatalf("result status = %q", dg.resultStatus.Text)
	}

	dg.working = true
	dg.setWorking("Processing", true)
	if dg.workStatus.Text != "Processing…" || !dg.workActivity.Visible() {
		t.Fatalf("working status activity=%v text=%q", dg.workActivity.Visible(), dg.workStatus.Text)
	}
	for _, button := range dg.commandButtons {
		if !button.Disabled() {
			t.Fatalf("command button %q remained enabled while working", button.Text)
		}
	}
	if !editable.Disabled() || !readOnly.Disabled() {
		t.Fatal("workflow controls should be disabled while work is running")
	}

	dg.working = false
	dg.setWorking("", false)
	if !dg.workIndicator.Visible() || dg.workStatus.Text != "Ready" || dg.workActivity.Visible() {
		t.Fatalf("ready status container=%v activity=%v text=%q", dg.workIndicator.Visible(), dg.workActivity.Visible(), dg.workStatus.Text)
	}
	if editable.Disabled() || !readOnly.Disabled() {
		t.Fatal("workflow controls did not restore their prior disabled state")
	}
	dg.selectTab(1)
	if dg.homeCommands.Visible() {
		t.Fatal("Home commands should be hidden outside the Home tab")
	}
	if dg.workspaceTitle.Text != "Workflows" || !dg.workspaceTitle.Visible() || !dg.tabButtons[1].active {
		t.Fatalf("workflow navigation state title=%q visible=%v active=%v", dg.workspaceTitle.Text, dg.workspaceTitle.Visible(), dg.tabButtons[1].active)
	}
	if dg.pipelineNavButton.Visible() {
		t.Fatal("pipeline navigator toggle should hide outside the Pipeline workspace")
	}
	dg.selectTab(0)
	if !dg.pipelineNavButton.Visible() {
		t.Fatal("pipeline navigator toggle should show in the Pipeline workspace")
	}

	dg.toggleSidebar()
	if !dg.navigationPanel.Visible() || !dg.sidebarCollapsed || a.Preferences().BoolWithFallback(sidebarPreferenceKey, true) {
		t.Fatal("sidebar should collapse to icons and persist its compact state")
	}
	dg.toggleSidebar()
	if !dg.navigationPanel.Visible() || dg.sidebarCollapsed || !a.Preferences().Bool(sidebarPreferenceKey) {
		t.Fatal("sidebar should expand labels and persist its open state")
	}

	dg.applyAppearance(appearanceLight)
	if got := a.Preferences().String(appearancePreferenceKey); got != string(appearanceLight) {
		t.Fatalf("saved appearance = %q, want %q", got, appearanceLight)
	}
}

func TestAdaptiveCommandToolbarPrioritizesActions(t *testing.T) {
	dg := newVisualScenarioGUI(t, appearanceDark)
	layout := dg.commandLayout
	bar := dg.homeCommandBar
	if layout == nil || bar == nil || len(dg.commandButtons) != 7 {
		t.Fatal("adaptive command toolbar did not initialize")
	}
	wideWidth := layout.width(layout.objects)
	bar.Resize(fyne.NewSize(wideWidth+20, bar.MinSize().Height))
	if dg.commandBarCompact {
		t.Fatal("toolbar should expose all commands when its complete row fits")
	}
	for _, button := range dg.commandButtons {
		if !button.Visible() {
			t.Fatalf("wide command %q is hidden", button.Text)
		}
	}
	if !dg.overflowUndoMenuItem.Disabled || !dg.overflowRedoMenuItem.Disabled || dg.overflowCopyMenuItem.Disabled {
		t.Fatal("fresh-pipeline overflow command state is incorrect")
	}

	dg.window.Canvas().Focus(dg.undoCommand)
	compactWidth := layout.MinSize(nil).Width
	bar.Resize(fyne.NewSize(compactWidth, bar.MinSize().Height))
	if !dg.commandBarCompact || dg.window.Canvas().Focused() != nil {
		t.Fatalf("compact toolbar state compact=%v focused=%T", dg.commandBarCompact, dg.window.Canvas().Focused())
	}
	for _, index := range []int{2, 5, 6} {
		if dg.commandButtons[index].Visible() {
			t.Fatalf("secondary compact command %q remained visible", dg.commandButtons[index].Text)
		}
	}
	for _, index := range []int{0, 1, 3, 4} {
		button := dg.commandButtons[index]
		if !button.Visible() {
			t.Fatalf("primary compact command %q was hidden", button.Text)
		}
		if edge := button.Position().X + button.Size().Width; edge > bar.Size().Width+0.01 {
			t.Fatalf("primary compact command %q extends to %v beyond toolbar width %v", button.Text, edge, bar.Size().Width)
		}
	}

	dg.pipe.SetSource([]byte("changed"))
	dg.refreshCommandButtons()
	if dg.overflowUndoMenuItem.Disabled || !dg.overflowRedoMenuItem.Disabled {
		t.Fatal("overflow history state did not track the mutated pipeline")
	}
	dg.working = true
	dg.setWorking("Testing", true)
	if !dg.overflowUndoMenuItem.Disabled || !dg.overflowRedoMenuItem.Disabled || !dg.overflowCopyMenuItem.Disabled {
		t.Fatal("overflow commands should disable while pipeline work is running")
	}
	dg.working = false
	dg.setWorking("", false)

	bar.Resize(fyne.NewSize(wideWidth+20, bar.MinSize().Height))
	if dg.commandBarCompact {
		t.Fatal("growing the toolbar should restore the complete command row")
	}
	for _, button := range dg.commandButtons {
		if !button.Visible() {
			t.Fatalf("restored wide command %q is hidden", button.Text)
		}
	}
}

func TestSystemSettingsRefreshCustomSurfaces(t *testing.T) {
	a := fynetest.NewTempApp(t)
	a.Settings().SetTheme(newAdversecTheme(theme.VariantDark))
	tab := newSidebarNavItem("Pipeline", theme.HomeIcon(), nil)
	tab.setActive(true)
	dark := color.NRGBAModel.Convert(tab.label.Color)

	dg := &DeenGUI{
		app:        a,
		appearance: appearanceSystem,
		activeTab:  0,
		tabButtons: []*navTab{tab},
	}
	a.Settings().SetTheme(newAdversecTheme(theme.VariantLight))
	dg.handleSettingsChange()
	light := color.NRGBAModel.Convert(tab.label.Color)
	if light == dark {
		t.Fatal("custom navigation color did not refresh after a system settings change")
	}

	dg.working = true
	dg.handleSettingsChange()
	if !dg.themeRefreshPending {
		t.Fatal("settings refresh should be deferred while pipeline work is running")
	}
}

type trackingCloser struct {
	closed bool
}

func (c *trackingCloser) Close() error {
	c.closed = true
	return nil
}

func TestCloseIfWorkingRejectsStaleDialogResources(t *testing.T) {
	dg := &DeenGUI{}
	idle := &trackingCloser{}
	if dg.closeIfWorking(idle) || idle.closed {
		t.Fatal("idle GUI should accept an open dialog resource")
	}

	dg.working = true
	busy := &trackingCloser{}
	if !dg.closeIfWorking(busy) || !busy.closed {
		t.Fatal("busy GUI should reject and close a stale dialog resource")
	}
}

type recordingURIWriteCloser struct {
	uri      fyne.URI
	data     []byte
	writeLen int
	writeErr error
	closeErr error
	closed   bool
}

func (wc *recordingURIWriteCloser) URI() fyne.URI { return wc.uri }

func (wc *recordingURIWriteCloser) Write(data []byte) (int, error) {
	if wc.writeErr != nil {
		return 0, wc.writeErr
	}
	if wc.writeLen > 0 && wc.writeLen < len(data) {
		wc.data = append(wc.data, data[:wc.writeLen]...)
		return wc.writeLen, nil
	}
	wc.data = append(wc.data, data...)
	return len(data), nil
}

func (wc *recordingURIWriteCloser) Close() error {
	wc.closed = true
	return wc.closeErr
}

func TestWriteAndCloseRequiresCompleteSave(t *testing.T) {
	wc := &recordingURIWriteCloser{}
	if err := writeAndClose(wc, []byte("saved")); err != nil || !wc.closed || string(wc.data) != "saved" {
		t.Fatalf("successful save err=%v closed=%v data=%q", err, wc.closed, wc.data)
	}

	writeFailure := errors.New("write failed")
	wc = &recordingURIWriteCloser{writeErr: writeFailure}
	if err := writeAndClose(wc, []byte("ignored")); !errors.Is(err, writeFailure) || !wc.closed {
		t.Fatalf("write failure err=%v closed=%v", err, wc.closed)
	}

	wc = &recordingURIWriteCloser{writeLen: 2}
	if err := writeAndClose(wc, []byte("short")); !errors.Is(err, io.ErrShortWrite) || !wc.closed {
		t.Fatalf("short write err=%v closed=%v data=%q", err, wc.closed, wc.data)
	}

	closeFailure := errors.New("close failed")
	wc = &recordingURIWriteCloser{closeErr: closeFailure}
	if err := writeAndClose(wc, []byte("written")); !errors.Is(err, closeFailure) || !wc.closed {
		t.Fatalf("close failure err=%v closed=%v", err, wc.closed)
	}
}

func TestFeedbackSubjectsStayCompact(t *testing.T) {
	if got := feedbackSubject("  report\nwith   spaces.txt  "); got != "report with spaces.txt" {
		t.Fatalf("normalized feedback subject = %q", got)
	}
	longName := strings.Repeat("界", 50)
	got := feedbackSubject(longName)
	if len([]rune(got)) != 40 || !strings.HasSuffix(got, "…") {
		t.Fatalf("compact feedback subject has %d runes and value %q", len([]rune(got)), got)
	}
	uri := storage.NewFileURI(filepath.Join(t.TempDir(), "result.txt"))
	if got := fileActionFeedback("Result saved", "Saved", uri); got != "Saved “result.txt”" {
		t.Fatalf("file feedback = %q", got)
	}
	if got := fileActionFeedback("Result saved", "Saved", nil); got != "Result saved" {
		t.Fatalf("fallback file feedback = %q", got)
	}
}

func TestActionFeedbackLifecycleAndCopyActions(t *testing.T) {
	dg := newVisualScenarioGUI(t, appearanceDark)
	dg.pipe.SetSource([]byte("copy me"))

	dg.copyResult()
	if got := dg.workStatus.Text; got != "Result copied" || dg.workStatus.Importance != widget.SuccessImportance || !dg.actionFeedbackActive {
		t.Fatalf("result-copy feedback = %q importance=%v active=%v", got, dg.workStatus.Importance, dg.actionFeedbackActive)
	}

	dg.pipe.AddStep("base64", false)
	dg.copyCommand()
	if got := dg.workStatus.Text; got != "Command copied" || dg.workStatus.Importance != widget.SuccessImportance {
		t.Fatalf("command-copy feedback = %q importance=%v", got, dg.workStatus.Importance)
	}

	dg.working = true
	dg.setWorking("Processing", true)
	if dg.actionFeedbackActive || dg.workStatus.Text != "Processing…" || dg.workStatus.Importance != widget.LowImportance {
		t.Fatalf("new work did not clear feedback: active=%v text=%q importance=%v", dg.actionFeedbackActive, dg.workStatus.Text, dg.workStatus.Importance)
	}

	// Completion callbacks run before setWorking(false); their success message
	// must survive the common cleanup while the activity indicator is hidden.
	dg.showActionFeedback("Pipeline updated")
	dg.working = false
	dg.setWorking("", false)
	if dg.workStatus.Text != "Pipeline updated" || dg.workStatus.Importance != widget.SuccessImportance || dg.workActivity.Visible() {
		t.Fatalf("completion feedback was not preserved: text=%q importance=%v activity=%v", dg.workStatus.Text, dg.workStatus.Importance, dg.workActivity.Visible())
	}

	dg.working = true
	dg.setWorking("Refreshing", true)
	dg.working = false
	dg.setWorking("", false)
	if dg.workStatus.Text != "Ready" || dg.workStatus.Importance != widget.LowImportance || dg.actionFeedbackActive {
		t.Fatalf("ordinary completion state text=%q importance=%v active=%v", dg.workStatus.Text, dg.workStatus.Importance, dg.actionFeedbackActive)
	}
}

func TestStepCardInitializationDoesNotStartPipelineWork(t *testing.T) {
	a := fynetest.NewTempApp(t)
	a.Settings().SetTheme(newAdversecTheme(theme.VariantDark))
	p := pipeline.New()
	p.SetSource([]byte("1700000000"))
	p.AddStep("timestamp", false) // includes a select option with a default value
	dg := &DeenGUI{
		app:           a,
		window:        a.NewWindow("test"),
		pipe:          p,
		stepsBox:      container.NewVBox(),
		selectedStage: pipelineStageInput,
		workActivity:  widget.NewActivity(),
		workStatus:    widget.NewLabel(""),
		workIndicator: container.NewHBox(),
	}

	dg.newStepCard(0)
	if dg.working {
		t.Fatal("initializing a select option should not invoke its change callback")
	}
}
