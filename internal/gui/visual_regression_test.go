//go:build gui

package gui

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fynetest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/takeshixx/deen/internal/pipeline"
)

func TestVisualRegressionScenariosRender(t *testing.T) {
	scenarios := []struct {
		name       string
		size       fyne.Size
		appearance appearanceMode
		setup      func(*DeenGUI)
		afterBuild func(*DeenGUI)
		assert     func(*testing.T, *DeenGUI)
	}{
		{
			name:       "add-catalog-dark",
			size:       fyne.NewSize(1180, 800),
			appearance: appearanceDark,
			setup: func(dg *DeenGUI) {
				dg.pipe.SetSource([]byte(`%7B%22role%22%3A%22admin%22%7D`))
				dg.selectedStage = pipelineStageAdd
			},
			assert: func(t *testing.T, dg *DeenGUI) {
				if dg.addCatalog == nil || dg.addCatalog.search == nil || dg.addCatalog.results == nil || len(dg.addCatalog.detail.Objects) == 0 {
					t.Fatal("add catalog scene is missing its focused controls")
				}
			},
		},
		{
			name:       "input-inspector-light",
			size:       fyne.NewSize(1180, 800),
			appearance: appearanceLight,
			setup: func(dg *DeenGUI) {
				dg.pipe.SetSource([]byte(`{"name":"deen","enabled":true}`))
				dg.sourcePane = "Inspector"
				dg.sourceView = "Preview"
				dg.selectedStage = pipelineStageInput
			},
			assert: func(t *testing.T, dg *DeenGUI) {
				if dg.sourceWorkspace == nil || dg.sourceWorkspace.Selected().Text != "Inspector" || dg.sourceViewer == nil || dg.sourcePreview == nil {
					t.Fatal("input inspector scene is missing editor, tabs, or structured preview")
				}
				if dg.commandBarCompact {
					t.Fatal("default-width input scene should show the complete command toolbar")
				}
			},
		},
		{
			name:       "action-feedback-dark",
			size:       fyne.NewSize(1180, 800),
			appearance: appearanceDark,
			setup: func(dg *DeenGUI) {
				dg.pipe.SetSource([]byte("non-modal feedback"))
			},
			afterBuild: func(dg *DeenGUI) {
				dg.showActionFeedback("Result copied")
			},
			assert: func(t *testing.T, dg *DeenGUI) {
				if dg.workStatus.Text != "Result copied" || dg.workStatus.Importance != widget.SuccessImportance || !dg.workIndicator.Visible() {
					t.Fatalf("feedback scene text=%q importance=%v visible=%v", dg.workStatus.Text, dg.workStatus.Importance, dg.workIndicator.Visible())
				}
			},
		},
		{
			name:       "collapsed-sidebar-editor-dark",
			size:       fyne.NewSize(1180, 800),
			appearance: appearanceDark,
			setup: func(dg *DeenGUI) {
				dg.sidebarOpen = false
				dg.sourcePane = "Editor"
				dg.pipe.SetSource([]byte("focused editor"))
			},
			assert: func(t *testing.T, dg *DeenGUI) {
				if !dg.sidebarCollapsed || !dg.navigationPanel.Visible() || !dg.tabButtons[0].iconOnly {
					t.Fatalf("collapsed rail state collapsed=%v visible=%v iconOnly=%v", dg.sidebarCollapsed, dg.navigationPanel.Visible(), dg.tabButtons[0].iconOnly)
				}
				if dg.sourceWorkspace == nil || dg.sourceWorkspace.Selected().Text != "Editor" || dg.sourceWorkspace.Items[1].Content.Visible() {
					t.Fatal("editor scene should show only the Input Editor view")
				}
				if dg.commandBarCompact {
					t.Fatal("icon rail should leave enough room for the complete top command row")
				}
			},
		},
		{
			name:       "transformer-browser-light",
			size:       fyne.NewSize(1180, 800),
			appearance: appearanceLight,
			setup: func(dg *DeenGUI) {
				dg.selectTab(2)
			},
			assert: func(t *testing.T, dg *DeenGUI) {
				if dg.browserCatalog == nil || dg.browserCatalog.search == nil || len(dg.browserCatalog.detail.Objects) == 0 {
					t.Fatal("transformer browser scene is missing its shared catalog controls")
				}
			},
		},
		{
			name:       "workflow-library-dark",
			size:       fyne.NewSize(1180, 800),
			appearance: appearanceDark,
			setup: func(dg *DeenGUI) {
				dg.selectTab(1)
			},
			assert: func(t *testing.T, dg *DeenGUI) {
				if dg.workflowLibrary == nil || dg.workflowLibrary.search == nil || len(dg.workflowLibrary.detail.Objects) == 0 {
					t.Fatal("workflow library scene is missing its master/detail controls")
				}
			},
		},
		{
			name:       "workflow-preview-scroll-light",
			size:       fyne.NewSize(1180, 800),
			appearance: appearanceLight,
			setup: func(dg *DeenGUI) {
				dg.selectTab(1)
			},
			afterBuild: func(dg *DeenGUI) {
				library := dg.workflowLibrary
				for index, candidate := range library.matches {
					if candidate.kind == workflowExample {
						library.results.Select(widget.ListItemID(index))
						break
					}
				}
				item, ok := library.selectedItem()
				if !ok || item.kind != workflowExample || library.previewSlot == nil {
					return
				}
				<-library.previewExample(item, library.detailVersion, library.previewSlot)
			},
			assert: func(t *testing.T, dg *DeenGUI) {
				library := dg.workflowLibrary
				if library.previewInput == nil || library.previewOutput == nil || library.detailScroll == nil {
					t.Fatal("workflow preview scene is missing its data views or detail scroller")
				}
				if library.previewInput.MinSize().Height < 250 || library.previewOutput.MinSize().Height < 250 {
					t.Fatalf("workflow preview heights input=%v output=%v", library.previewInput.MinSize().Height, library.previewOutput.MinSize().Height)
				}
				if library.detail.MinSize().Height <= library.detailScroll.Size().Height || library.detailScroll.Offset.Y <= 0 {
					t.Fatalf("workflow preview scroll content=%v viewport=%v offset=%v", library.detail.MinSize(), library.detailScroll.Size(), library.detailScroll.Offset)
				}
			},
		},
		{
			name:       "keyboard-focus-sidebar-light",
			size:       fyne.NewSize(1180, 800),
			appearance: appearanceLight,
			setup: func(dg *DeenGUI) {
				dg.selectTab(1)
			},
			afterBuild: func(dg *DeenGUI) {
				dg.window.Canvas().Focus(dg.tabButtons[2])
			},
			assert: func(t *testing.T, dg *DeenGUI) {
				if !dg.tabButtons[1].active || !dg.tabButtons[2].focused || dg.tabButtons[2].background.StrokeWidth == 0 {
					t.Fatal("keyboard focus scene should distinguish the active and focused destinations")
				}
			},
		},
		{
			name:       "compare-workspace-dark",
			size:       fyne.NewSize(1180, 800),
			appearance: appearanceDark,
			setup: func(dg *DeenGUI) {
				dg.pipe.SetSource([]byte(`{"name":"deen","enabled":true}`))
				dg.pipe.AddStep("base64", false)
				dg.pipe.AddStep("hex", false)
				dg.selectTab(4)
			},
			assert: func(t *testing.T, dg *DeenGUI) {
				workspace := dg.compareWorkspace
				if workspace == nil || len(workspace.points) != 3 || workspace.split == nil || workspace.difference == nil {
					t.Fatal("compare scene is missing pipeline points, split view, or difference status")
				}
			},
		},
		{
			name:       "compact-stages-drawer-dark",
			size:       fyne.NewSize(720, 620),
			appearance: appearanceDark,
			setup: func(dg *DeenGUI) {
				dg.pipe.SetSource([]byte("responsive navigation"))
				dg.pipe.AddStep("base64", false)
				dg.pipe.AddStep("hex", false)
				dg.selectedStage = 1
			},
			afterBuild: func(dg *DeenGUI) {
				dg.togglePipelineNavigator()
			},
			assert: func(t *testing.T, dg *DeenGUI) {
				if !dg.compactSidebar || !dg.sidebarCollapsed || !dg.compactStages || !dg.navigationPanel.Visible() || !dg.pipelineNavigator.Visible() || !dg.stagesDrawerOpen {
					t.Fatalf("compact drawer state sidebar=%v collapsed=%v stages=%v globalVisible=%v stagesVisible=%v drawer=%v", dg.compactSidebar, dg.sidebarCollapsed, dg.compactStages, dg.navigationPanel.Visible(), dg.pipelineNavigator.Visible(), dg.stagesDrawerOpen)
				}
			},
		},
		{
			name:       "compact-vertical-editor-light",
			size:       fyne.NewSize(720, 800),
			appearance: appearanceLight,
			setup: func(dg *DeenGUI) {
				dg.pipe.SetSource([]byte(`{"layout":"responsive","value":42}`))
				dg.pipe.AddStep("base64", false)
				dg.selectedStage = 0
			},
			assert: func(t *testing.T, dg *DeenGUI) {
				card := dg.cards[0]
				if !dg.compactSidebar || !dg.compactStages || !dg.commandBarCompact || card == nil || card.editorSplit.Horizontal {
					t.Fatalf("compact vertical editor state sidebar=%v stages=%v card=%v horizontal=%v", dg.compactSidebar, dg.compactStages, card != nil, card != nil && card.editorSplit.Horizontal)
				}
				for _, index := range []int{2, 5, 6} {
					if dg.commandButtons[index].Visible() {
						t.Fatalf("compact visual command %q should be in overflow", dg.commandButtons[index].Text)
					}
				}
			},
		},
		{
			name:       "compact-vertical-workflows-dark",
			size:       fyne.NewSize(720, 800),
			appearance: appearanceDark,
			setup: func(dg *DeenGUI) {
				dg.selectTab(1)
			},
			assert: func(t *testing.T, dg *DeenGUI) {
				if !dg.compactSidebar || dg.workflowLibrary == nil || dg.workflowLibrary.split.Horizontal {
					t.Fatalf("compact workflow state sidebar=%v library=%v horizontal=%v", dg.compactSidebar, dg.workflowLibrary != nil, dg.workflowLibrary != nil && dg.workflowLibrary.split.Horizontal)
				}
			},
		},
		{
			name:       "step-error-compact-dark",
			size:       fyne.NewSize(620, 720),
			appearance: appearanceDark,
			setup: func(dg *DeenGUI) {
				dg.pipe.SetSource([]byte("not encrypted data"))
				dg.pipe.AddStepWithOptions("aes", true, map[string]string{"mode": "gcm", "key": "00"})
				dg.selectedStage = 0
				dg.sidebarOpen = false
				dg.pipelineNavOpen = false
			},
			assert: func(t *testing.T, dg *DeenGUI) {
				card := dg.cards[0]
				if card == nil || card.editorSplit == nil || card.status == nil || !card.status.Visible() {
					t.Fatal("compact error scene is missing its focused editor or visible error status")
				}
				if len(card.options.Objects) < 2 {
					t.Fatal("AES scene should render grouped option sections")
				}
			},
		},
		{
			name:       "disabled-binary-light",
			size:       fyne.NewSize(1180, 800),
			appearance: appearanceLight,
			setup: func(dg *DeenGUI) {
				binary := make([]byte, 128)
				binary[4] = 0xff
				copy(binary[64:], "deen-binary")
				dg.pipe.SetSource(binary)
				dg.pipe.AddStep("gzip", false)
				dg.pipe.SetStepDisabled(0, true)
				dg.stepOutputView = "Hex"
				dg.selectedStage = 0
			},
			assert: func(t *testing.T, dg *DeenGUI) {
				card := dg.cards[0]
				if card == nil || card.stringsTab == nil || card.viewer.Selected().Text != "Hex" {
					selected := "<none>"
					if card != nil && card.viewer != nil && card.viewer.Selected() != nil {
						selected = card.viewer.Selected().Text
					}
					t.Fatalf("disabled binary scene card=%v strings=%v selected=%q binary=%v output=%d error=%v", card != nil, card != nil && card.stringsTab != nil, selected, pipeline.IsBinaryData(dg.pipe.Output(0)), len(dg.pipe.Output(0)), dg.pipe.Err(0))
				}
				if got := card.headerActions[1].AccessibilityLabel(); got != "Enable step" {
					t.Fatalf("disabled action label = %q, want Enable step", got)
				}
			},
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			dg := newVisualScenarioGUI(t, scenario.appearance)
			scenario.setup(dg)
			dg.rebuild()
			dg.window.Resize(scenario.size)
			dg.window.Show()
			if scenario.afterBuild != nil {
				scenario.afterBuild(dg)
			}
			dg.window.Canvas().Refresh(dg.window.Content())
			fyne.DoAndWait(func() {})
			assertNonNegativeLayout(t, dg.window.Content())
			captured := dg.window.Canvas().Capture()
			writeVisualCapture(t, scenario.name, captured)
			assertVisualCapture(t, captured, scenario.size)
			scenario.assert(t, dg)
		})
	}
}

func assertNonNegativeLayout(t *testing.T, content fyne.CanvasObject) {
	t.Helper()
	for _, object := range fynetest.LaidOutObjects(content) {
		size := object.Size()
		if size.Width < 0 || size.Height < 0 {
			t.Fatalf("negative laid-out size %v for %T at %v", size, object, object.Position())
		}
	}
}

func newVisualScenarioGUI(t *testing.T, appearance appearanceMode) *DeenGUI {
	t.Helper()
	a := fynetest.NewTempApp(t)
	a.Settings().SetTheme(themeForAppearance(appearance))
	dg := &DeenGUI{
		app:             a,
		window:          a.NewWindow("deen"),
		pipe:            pipeline.New(),
		stepsBox:        container.NewVBox(),
		pipelineOutline: container.NewVBox(),
		selectedStage:   pipelineStageInput,
		pipelineNavOpen: true,
		stepEditorSplit: defaultStepEditorSplit,
		sourcePane:      defaultSourcePane,
		sourceView:      defaultSourceView,
		activeTab:       -1,
		tabContent:      container.NewMax(),
		sidebarOpen:     true,
		appearance:      appearance,
	}
	dg.navigationPanel = dg.navigationSidebar()
	dg.homeCommands = dg.homeMenuBar()
	dg.workActivity = widget.NewActivity()
	dg.workActivity.Hide()
	dg.workStatus = widget.NewLabel("Ready")
	dg.workIndicator = container.NewHBox(dg.workActivity, dg.workStatus)
	dg.resultStatus = widget.NewLabel("")
	workspace := container.NewBorder(dg.workspaceToolbar(), dg.statusBar(), nil, nil, dg.tabContent)
	dg.window.SetContent(container.NewPadded(dg.newApplicationShell(workspace)))
	dg.selectTab(0)
	return dg
}

func assertVisualCapture(t *testing.T, captured image.Image, expected fyne.Size) {
	t.Helper()
	bounds := captured.Bounds()
	if bounds.Dx() != int(expected.Width) || bounds.Dy() != int(expected.Height) {
		t.Fatalf("capture size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), int(expected.Width), int(expected.Height))
	}
	colors := make(map[color.NRGBA]struct{})
	frequencies := make(map[color.NRGBA]int)
	darkest, lightest := ^uint32(0), uint32(0)
	samples := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 4 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 4 {
			pixel := color.NRGBAModel.Convert(captured.At(x, y)).(color.NRGBA)
			colors[pixel] = struct{}{}
			frequencies[pixel]++
			samples++
			luma := uint32(pixel.R)*299 + uint32(pixel.G)*587 + uint32(pixel.B)*114
			if luma < darkest {
				darkest = luma
			}
			if luma > lightest {
				lightest = luma
			}
		}
	}
	if len(colors) < 16 {
		t.Fatalf("capture contains only %d sampled colors; scene may have failed to render", len(colors))
	}
	if lightest-darkest < 40_000 {
		t.Fatalf("capture luminance range = %d; controls may be indistinguishable", lightest-darkest)
	}
	mostCommon := 0
	for _, count := range frequencies {
		if count > mostCommon {
			mostCommon = count
		}
	}
	if float64(mostCommon)/float64(samples) > 0.90 {
		t.Fatalf("one color covers %.1f%% of sampled pixels; frame may be incomplete", 100*float64(mostCommon)/float64(samples))
	}
}

func writeVisualCapture(t *testing.T, name string, captured image.Image) {
	t.Helper()
	directory := os.Getenv("DEEN_VISUAL_OUTPUT")
	if directory == "" {
		return
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, fmt.Sprintf("%s.png", name))
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, captured); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
