//go:build gui

// Package gui implements the deen desktop interface: a Burp Decoder-style
// chain of plugin transforms backed by the pure internal/pipeline model.
package gui

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/takeshixx/deen/internal/core"
	"github.com/takeshixx/deen/internal/pipeline"
	"github.com/takeshixx/deen/internal/plugins"
)

// DeenGUI is the top-level GUI state.
type DeenGUI struct {
	app    fyne.App
	window fyne.Window

	pipe *pipeline.Pipeline

	sourceEntry           *widget.Entry
	sourceRaw             *widget.Entry
	sourceHex             *widget.Entry
	sourceStrings         *widget.Entry
	sourcePreview         *widget.TextGrid
	sourcePreviewTab      *container.TabItem
	sourceViewer          *container.AppTabs
	sourceWorkspace       *container.AppTabs
	sourceMeta            *widget.Label
	sourceFullControls    *fyne.Container
	sourceName            string
	sourceFullRaw         bool
	sourceFullHex         bool
	sourceFullStrings     bool
	stepsBox              *fyne.Container // focused pipeline detail content
	cards                 []*stepCard     // sparse, indexed by pipe.Steps()
	pipelineOutline       *fyne.Container
	pipelineOutlineCount  *widget.Label
	pipelineNavigator     fyne.CanvasObject
	pipelineShell         *fyne.Container
	pipelineNavButton     *widget.Button
	pipelineStageButtons  []*navTab
	selectedStage         int
	pipelineNavOpen       bool
	stepEditorSplit       float64
	sourcePane            string
	sourceView            string
	stepOutputView        string
	addCatalog            *transformerCatalog
	browserCatalog        *transformerCatalog
	workflowLibrary       *workflowLibrary
	compareWorkspace      *compareWorkspace
	addSuggestions        *fyne.Container
	favoriteTransformers  []string
	recentTransformers    []string
	navigationPanel       fyne.CanvasObject
	sidebarLayout         *fixedWidthLayout
	sidebarBrandCopy      fyne.CanvasObject
	sidebarSection        *widget.Label
	sidebarCollapsed      bool
	applicationShell      *fyne.Container
	tabButtons            []*navTab
	tabContent            *fyne.Container
	tabViews              [5]fyne.CanvasObject
	homeCommands          fyne.CanvasObject
	homeCommandScroll     *container.Scroll
	homeCommandBar        *fyne.Container
	commandLayout         *adaptiveCommandLayout
	workspaceTitle        *widget.Label
	workspaceTitleDivider fyne.CanvasObject
	sidebarButton         *widget.Button
	commandButtons        []*widget.Button
	compactCommandButtons []*widget.Button
	overflowUndoMenuItem  *fyne.MenuItem
	overflowRedoMenuItem  *fyne.MenuItem
	overflowCopyMenuItem  *fyne.MenuItem
	workControls          []fyne.Disableable
	workControlStates     map[fyne.Disableable]bool
	undoCommand           *widget.Button
	redoCommand           *widget.Button
	undoMenuItem          *fyne.MenuItem
	redoMenuItem          *fyne.MenuItem
	addMenuItem           *fyne.MenuItem
	clearMenuItem         *fyne.MenuItem
	previousStageMenuItem *fyne.MenuItem
	nextStageMenuItem     *fyne.MenuItem
	toggleStepMenuItem    *fyne.MenuItem
	moveStepUpMenuItem    *fyne.MenuItem
	moveStepDownMenuItem  *fyne.MenuItem
	duplicateStepMenuItem *fyne.MenuItem
	removeStepMenuItem    *fyne.MenuItem
	workActivity          *widget.Activity
	workStatus            *widget.Label
	workIndicator         *fyne.Container
	resultStatus          *widget.Label
	activeTab             int
	appearance            appearanceMode
	sidebarOpen           bool
	compactSidebar        bool
	compactStages         bool
	sidebarDrawerOpen     bool
	stagesDrawerOpen      bool
	updatingAppearance    bool
	themeRefreshPending   bool
	commandBarCompact     bool
	working               bool
	actionFeedbackActive  bool

	// updating guards programmatic SetText so it does not re-enter OnChanged.
	updating bool
}

type appearanceMode string

const (
	appearanceSystem appearanceMode = "system"
	appearanceDark   appearanceMode = "dark"
	appearanceLight  appearanceMode = "light"

	appearancePreferenceKey           = "gui.appearance"
	sidebarPreferenceKey              = "gui.sidebar"
	pipelineNavPreferenceKey          = "gui.pipeline_navigator"
	stepEditorSplitPreferenceKey      = "gui.step_editor_split"
	stepEditorCompactPreferenceKey    = "gui.step_editor_split_compact"
	sourcePanePreferenceKey           = "gui.source_pane"
	sourceViewPreferenceKey           = "gui.source_view"
	stepOutputPreferenceKey           = "gui.step_output_view"
	addSplitPreferenceKey             = "gui.add_catalog_split"
	addCompactSplitPreferenceKey      = "gui.add_catalog_split_compact"
	addCategoryPreferenceKey          = "gui.add_category"
	browserSplitPreferenceKey         = "gui.transformer_catalog_split"
	browserCompactSplitPreferenceKey  = "gui.transformer_split_compact"
	browserCategoryPreferenceKey      = "gui.transformer_category"
	workflowSplitPreferenceKey        = "gui.workflow_split"
	workflowCompactPreferenceKey      = "gui.workflow_split_compact"
	workflowFilterPreferenceKey       = "gui.workflow_filter"
	compareSplitPreferenceKey         = "gui.compare_split"
	compareCompactPreferenceKey       = "gui.compare_split_compact"
	compareModePreferenceKey          = "gui.compare_mode"
	compareLeftPreferenceKey          = "gui.compare_left"
	compareRightPreferenceKey         = "gui.compare_right"
	favoriteTransformersPreferenceKey = "gui.favorite_transformers"
	recentTransformersPreferenceKey   = "gui.recent_transformers"
	windowWidthPreferenceKey          = "gui.window_width"
	windowHeightPreferenceKey         = "gui.window_height"
	workspacePreferenceKey            = "gui.workspace"
	selectedStagePreferenceKey        = "gui.selected_stage"

	pipelineStageAdd   = -2
	pipelineStageInput = -1
)

func normalizeAppearance(value string) appearanceMode {
	switch appearanceMode(value) {
	case appearanceDark:
		return appearanceDark
	case appearanceLight:
		return appearanceLight
	default:
		return appearanceSystem
	}
}

func themeForAppearance(mode appearanceMode) fyne.Theme {
	switch mode {
	case appearanceDark:
		return newAdversecTheme(theme.VariantDark)
	case appearanceLight:
		return newAdversecTheme(theme.VariantLight)
	default:
		return newSystemAdversecTheme()
	}
}

// NewDeenGUI builds the GUI.
func NewDeenGUI() (*DeenGUI, error) {
	dg := &DeenGUI{
		app:  app.NewWithID("io.deen.app"),
		pipe: pipeline.New(),
	}
	dg.appearance = normalizeAppearance(dg.app.Preferences().StringWithFallback(appearancePreferenceKey, string(appearanceSystem)))
	dg.sidebarOpen = dg.app.Preferences().BoolWithFallback(sidebarPreferenceKey, true)
	dg.pipelineNavOpen = dg.app.Preferences().BoolWithFallback(pipelineNavPreferenceKey, true)
	dg.stepEditorSplit = normalizeStepEditorSplit(dg.app.Preferences().FloatWithFallback(stepEditorSplitPreferenceKey, defaultStepEditorSplit))
	dg.sourcePane = normalizeSourcePane(dg.app.Preferences().StringWithFallback(sourcePanePreferenceKey, defaultSourcePane))
	dg.sourceView = dg.app.Preferences().StringWithFallback(sourceViewPreferenceKey, defaultSourceView)
	dg.stepOutputView = dg.app.Preferences().StringWithFallback(stepOutputPreferenceKey, "")
	dg.favoriteTransformers = sanitizeTransformerHistory(dg.app.Preferences().StringListWithFallback(favoriteTransformersPreferenceKey, nil), maxFavoriteTransformers)
	dg.recentTransformers = sanitizeTransformerHistory(dg.app.Preferences().StringListWithFallback(recentTransformersPreferenceKey, nil), maxRecentTransformers)
	dg.selectedStage = dg.app.Preferences().IntWithFallback(selectedStagePreferenceKey, pipelineStageInput)
	dg.app.Settings().SetTheme(themeForAppearance(dg.appearance))
	dg.window = dg.app.NewWindow("deen")
	dg.window.SetMaster()
	dg.window.SetIcon(deenLogoResource)

	dg.stepsBox = container.NewVBox()
	dg.pipelineOutline = container.NewVBox()
	dg.activeTab = -1
	dg.tabContent = container.NewMax()
	dg.navigationPanel = dg.navigationSidebar()
	dg.homeCommands = dg.homeMenuBar()
	dg.workActivity = widget.NewActivity()
	dg.workActivity.Hide()
	dg.workStatus = widget.NewLabel("Ready")
	dg.workStatus.Importance = widget.LowImportance
	dg.workIndicator = container.NewHBox(dg.workActivity, dg.workStatus)
	dg.resultStatus = widget.NewLabel("")
	dg.resultStatus.Importance = widget.LowImportance
	workspace := container.NewBorder(dg.workspaceToolbar(), dg.statusBar(), nil, nil, dg.tabContent)
	dg.window.SetContent(container.NewPadded(dg.newApplicationShell(workspace)))
	dg.window.SetMainMenu(dg.mainMenu())
	dg.window.Resize(dg.restoredWindowSize())
	dg.window.SetOnDropped(dg.handleDroppedFiles)

	dg.selectTab(normalizeWorkspace(dg.app.Preferences().IntWithFallback(workspacePreferenceKey, 0)))
	dg.rebuild()
	dg.app.Settings().AddListener(func(fyne.Settings) {
		dg.handleSettingsChange()
	})
	return dg, nil
}

// Run shows the window and blocks until it closes.
func (dg *DeenGUI) Run() {
	defer dg.rememberWorkspaceState()
	dg.window.ShowAndRun()
}

func compactMinWidth(obj fyne.CanvasObject) fyne.CanvasObject {
	return container.New(cappedMinWidthLayout{width: compactControlMinWidth}, obj)
}

type fixedWidthLayout struct {
	width float32
}

func (l fixedWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, object := range objects {
		object.Resize(size)
	}
}

func (l fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	height := float32(0)
	for _, object := range objects {
		if min := object.MinSize(); min.Height > height {
			height = min.Height
		}
	}
	return fyne.NewSize(l.width, height)
}

func (dg *DeenGUI) navigationSidebar() fyne.CanvasObject {
	logo := canvas.NewImageFromResource(deenLogoResource)
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(34, 34))
	title := widget.NewLabelWithStyle("deen", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabel("Local data workbench")
	subtitle.Importance = widget.LowImportance
	dg.sidebarBrandCopy = container.NewVBox(title, subtitle)
	brand := container.NewHBox(logo, dg.sidebarBrandCopy)

	dg.tabButtons = []*navTab{
		newSidebarNavItem("Pipeline", theme.HomeIcon(), func() { dg.selectTab(0) }),
		newSidebarNavItem("Workflows", theme.HistoryIcon(), func() { dg.selectTab(1) }),
		newSidebarNavItem("Transformers", theme.SearchIcon(), func() { dg.selectTab(2) }),
		newSidebarNavItem("About", theme.InfoIcon(), func() { dg.selectTab(3) }),
		newSidebarNavItem("Compare", theme.ViewFullScreenIcon(), func() { dg.selectTab(4) }),
	}
	dg.sidebarSection = widget.NewLabelWithStyle("WORKSPACE", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	dg.sidebarSection.Importance = widget.LowImportance
	primary := container.NewVBox(dg.sidebarSection, dg.tabButtons[0], dg.tabButtons[1], dg.tabButtons[2], dg.tabButtons[4])
	bottom := container.NewVBox(widget.NewSeparator(), dg.tabButtons[3])
	content := container.NewBorder(container.NewVBox(brand, widget.NewSeparator()), bottom, nil, nil, primary)
	panel := widget.NewCard("", "", container.NewPadded(content))
	dg.sidebarLayout = &fixedWidthLayout{width: expandedSidebarWidth}
	return container.New(dg.sidebarLayout, panel)
}

func (dg *DeenGUI) newApplicationShell(workspace fyne.CanvasObject) fyne.CanvasObject {
	layout := &sidebarShellLayout{gui: dg, leading: dg.navigationPanel, content: workspace}
	dg.applicationShell = container.NewWithoutLayout(workspace, dg.navigationPanel)
	dg.applicationShell.Layout = layout
	dg.applicationShell.Resize(fyne.NewSize(compactSidebarBreakpoint, max(workspace.MinSize().Height, dg.navigationPanel.MinSize().Height)))
	return dg.applicationShell
}

func (dg *DeenGUI) setSidebarCollapsed(collapsed bool) {
	if dg.sidebarCollapsed == collapsed {
		return
	}
	dg.sidebarCollapsed = collapsed
	if dg.sidebarLayout != nil {
		if collapsed {
			dg.sidebarLayout.width = collapsedSidebarWidth
		} else {
			dg.sidebarLayout.width = expandedSidebarWidth
		}
	}
	for _, tab := range dg.tabButtons {
		tab.setIconOnly(collapsed)
	}
	setAdaptiveObjectVisible(dg.sidebarBrandCopy, !collapsed)
	setAdaptiveObjectVisible(dg.sidebarSection, !collapsed)
	if dg.navigationPanel != nil {
		dg.navigationPanel.Refresh()
	}
}

func (dg *DeenGUI) setCompactSidebar(compact bool) {
	dg.compactSidebar = compact
	dg.sidebarDrawerOpen = false
	if compact {
		dg.unfocusNavigation(dg.tabButtons)
	}
}

func (dg *DeenGUI) setCompactStages(compact bool) {
	dg.compactStages = compact
	dg.stagesDrawerOpen = false
	if compact {
		dg.unfocusNavigation(dg.pipelineStageButtons)
	}
}

func (dg *DeenGUI) unfocusNavigation(items []*navTab) {
	if dg.window == nil {
		return
	}
	focused := dg.window.Canvas().Focused()
	for _, item := range items {
		if focused == item {
			dg.window.Canvas().Unfocus()
			return
		}
	}
}

func (dg *DeenGUI) refreshAdaptiveNavigation() {
	if dg.applicationShell != nil {
		dg.applicationShell.Refresh()
	}
	if dg.pipelineShell != nil {
		dg.pipelineShell.Refresh()
	}
}

func (dg *DeenGUI) closeCompactNavigation() {
	changed := dg.sidebarDrawerOpen || dg.stagesDrawerOpen
	if dg.sidebarDrawerOpen {
		dg.unfocusNavigation(dg.tabButtons)
	}
	if dg.stagesDrawerOpen {
		dg.unfocusNavigation(dg.pipelineStageButtons)
	}
	dg.sidebarDrawerOpen = false
	dg.stagesDrawerOpen = false
	if changed {
		dg.refreshAdaptiveNavigation()
	}
}

func (dg *DeenGUI) workspaceToolbar() fyne.CanvasObject {
	dg.sidebarButton = widget.NewButtonWithIcon("Sidebar", theme.MenuIcon(), dg.toggleSidebar)
	dg.sidebarButton.Importance = widget.LowImportance
	dg.pipelineNavButton = widget.NewButtonWithIcon("Stages", theme.ListIcon(), dg.togglePipelineNavigator)
	dg.pipelineNavButton.Importance = widget.LowImportance
	dg.workspaceTitle = widget.NewLabelWithStyle("Pipeline", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	dg.workspaceTitleDivider = widget.NewSeparator()
	leading := container.NewHBox(dg.sidebarButton, dg.pipelineNavButton, dg.workspaceTitleDivider, dg.workspaceTitle)
	toolbar := container.NewBorder(nil, nil, leading, nil, dg.homeCommands)
	return widget.NewCard("", "", container.NewPadded(toolbar))
}

func (dg *DeenGUI) statusBar() fyne.CanvasObject {
	content := container.NewBorder(nil, nil, dg.workIndicator, dg.resultStatus)
	return widget.NewCard("", "", container.NewPadded(content))
}

func (dg *DeenGUI) toggleSidebar() {
	if dg.compactSidebar {
		if dg.sidebarDrawerOpen {
			dg.unfocusNavigation(dg.tabButtons)
		}
		dg.sidebarDrawerOpen = !dg.sidebarDrawerOpen
		dg.refreshAdaptiveNavigation()
		if dg.sidebarDrawerOpen && dg.window != nil && dg.activeTab >= 0 && dg.activeTab < len(dg.tabButtons) {
			dg.window.Canvas().Focus(dg.tabButtons[dg.activeTab])
		}
		return
	}
	dg.sidebarOpen = !dg.sidebarOpen
	if dg.app != nil {
		dg.app.Preferences().SetBool(sidebarPreferenceKey, dg.sidebarOpen)
	}
	if dg.applicationShell == nil {
		dg.setSidebarCollapsed(!dg.sidebarOpen)
	}
	dg.refreshAdaptiveNavigation()
}

func (dg *DeenGUI) togglePipelineNavigator() {
	if dg.compactStages {
		if dg.stagesDrawerOpen {
			dg.unfocusNavigation(dg.pipelineStageButtons)
		}
		dg.stagesDrawerOpen = !dg.stagesDrawerOpen
		dg.refreshAdaptiveNavigation()
		position := dg.pipelineStagePosition()
		if dg.stagesDrawerOpen && dg.window != nil && position >= 0 && position < len(dg.pipelineStageButtons) {
			dg.window.Canvas().Focus(dg.pipelineStageButtons[position])
		}
		return
	}
	dg.pipelineNavOpen = !dg.pipelineNavOpen
	if dg.app != nil {
		dg.app.Preferences().SetBool(pipelineNavPreferenceKey, dg.pipelineNavOpen)
	}
	if dg.pipelineShell == nil {
		setAdaptiveObjectVisible(dg.pipelineNavigator, dg.pipelineNavOpen)
	}
	dg.refreshAdaptiveNavigation()
}

func (dg *DeenGUI) homeMenuBar() fyne.CanvasObject {
	open := widget.NewButtonWithIcon("Open file", theme.FolderOpenIcon(), dg.openFile)
	open.Importance = widget.HighImportance
	save := widget.NewButtonWithIcon("Save result", theme.DocumentSaveIcon(), dg.saveResult)
	copyResult := widget.NewButtonWithIcon("Copy result", theme.ContentCopyIcon(), dg.copyResult)
	dg.undoCommand = widget.NewButtonWithIcon("Undo", theme.NavigateBackIcon(), dg.undo)
	dg.redoCommand = widget.NewButtonWithIcon("Redo", theme.NavigateNextIcon(), dg.redo)
	add := widget.NewButtonWithIcon("Add step", theme.ContentAddIcon(), dg.showPluginSearch)
	dg.overflowUndoMenuItem = fyne.NewMenuItemWithIcon("Undo", theme.ContentUndoIcon(), dg.undo)
	dg.overflowRedoMenuItem = fyne.NewMenuItemWithIcon("Redo", theme.ContentRedoIcon(), dg.redo)
	dg.overflowCopyMenuItem = fyne.NewMenuItemWithIcon("Copy result", theme.ContentCopyIcon(), dg.copyResult)
	more := dg.menuButton("More", theme.MoreHorizontalIcon(), fyne.NewMenu("More",
		dg.overflowUndoMenuItem,
		dg.overflowRedoMenuItem,
		dg.overflowCopyMenuItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItemWithIcon("Open chain", theme.FileTextIcon(), dg.openChain),
		fyne.NewMenuItemWithIcon("Save chain", theme.DocumentCreateIcon(), dg.saveChain),
		fyne.NewMenuItemWithIcon("Copy command", theme.MailForwardIcon(), dg.copyCommand),
		fyne.NewMenuItemWithIcon("Presets", theme.HistoryIcon(), dg.showPresets),
		fyne.NewMenuItemWithIcon("Compare", theme.ViewFullScreenIcon(), dg.showCompare),
		fyne.NewMenuItemWithIcon("Clear", theme.ContentClearIcon(), dg.clear),
	))

	dg.commandButtons = []*widget.Button{open, save, copyResult, add, more, dg.undoCommand, dg.redoCommand}
	dg.compactCommandButtons = []*widget.Button{copyResult, dg.undoCommand, dg.redoCommand}
	dg.refreshCommandButtons()
	primarySeparator := widget.NewSeparator()
	historySeparator := widget.NewSeparator()
	objects := []fyne.CanvasObject{open, save, copyResult, primarySeparator, add, more, historySeparator, dg.undoCommand, dg.redoCommand}
	dg.commandLayout = &adaptiveCommandLayout{
		objects: objects,
		compactHidden: map[fyne.CanvasObject]bool{
			copyResult:       true,
			historySeparator: true,
			dg.undoCommand:   true,
			dg.redoCommand:   true,
		},
		onCompact: dg.setCommandBarCompact,
	}
	dg.homeCommandBar = container.New(dg.commandLayout, objects...)
	dg.homeCommandScroll = container.NewHScroll(dg.homeCommandBar)
	return dg.homeCommandScroll
}

func (dg *DeenGUI) setCommandBarCompact(compact bool) {
	dg.commandBarCompact = compact
	if !compact || dg.window == nil {
		return
	}
	focused := dg.window.Canvas().Focused()
	for _, button := range dg.compactCommandButtons {
		if focused == button {
			dg.window.Canvas().Unfocus()
			return
		}
	}
}

func (dg *DeenGUI) menuButton(label string, icon fyne.Resource, menu *fyne.Menu) *widget.Button {
	btn := widget.NewButtonWithIcon(label, icon, nil)
	btn.Importance = widget.LowImportance
	btn.OnTapped = func() {
		widget.ShowPopUpMenuAtRelativePosition(menu, dg.window.Canvas(), fyne.NewPos(0, btn.Size().Height), btn)
	}
	return btn
}

func (dg *DeenGUI) refreshCommandButtons() {
	for _, button := range dg.commandButtons {
		if dg.working {
			button.Disable()
		} else {
			button.Enable()
		}
	}
	if dg.working {
		dg.refreshOverflowMenuState()
		dg.refreshNativeMenuState()
		return
	}
	if dg.undoCommand != nil && !dg.pipe.CanUndo() {
		dg.undoCommand.Disable()
	}
	if dg.redoCommand != nil && !dg.pipe.CanRedo() {
		dg.redoCommand.Disable()
	}
	dg.refreshOverflowMenuState()
	dg.refreshNativeMenuState()
}

func (dg *DeenGUI) refreshOverflowMenuState() {
	if dg.pipe == nil {
		return
	}
	if dg.overflowUndoMenuItem != nil {
		dg.overflowUndoMenuItem.Disabled = dg.working || !dg.pipe.CanUndo()
	}
	if dg.overflowRedoMenuItem != nil {
		dg.overflowRedoMenuItem.Disabled = dg.working || !dg.pipe.CanRedo()
	}
	if dg.overflowCopyMenuItem != nil {
		dg.overflowCopyMenuItem.Disabled = dg.working
	}
}

func (dg *DeenGUI) refreshNativeMenuState() {
	if dg.pipe == nil {
		return
	}
	if dg.undoMenuItem != nil {
		dg.undoMenuItem.Disabled = dg.working || !dg.pipe.CanUndo()
	}
	if dg.redoMenuItem != nil {
		dg.redoMenuItem.Disabled = dg.working || !dg.pipe.CanRedo()
	}
	if dg.addMenuItem != nil {
		dg.addMenuItem.Disabled = dg.working
	}
	if dg.clearMenuItem != nil {
		dg.clearMenuItem.Disabled = dg.working || (dg.pipe.Len() == 0 && len(dg.pipe.Source()) == 0)
	}
	position := dg.pipelineStagePosition()
	if dg.previousStageMenuItem != nil {
		dg.previousStageMenuItem.Disabled = dg.working || dg.activeTab != 0 || position <= 0
	}
	if dg.nextStageMenuItem != nil {
		dg.nextStageMenuItem.Disabled = dg.working || dg.activeTab != 0 || position < 0 || position >= dg.pipe.Len()+1
	}
	selectedStep := dg.activeTab == 0 && dg.selectedStage >= 0 && dg.selectedStage < dg.pipe.Len()
	if dg.toggleStepMenuItem != nil {
		dg.toggleStepMenuItem.Disabled = dg.working || !selectedStep
	}
	if dg.moveStepUpMenuItem != nil {
		dg.moveStepUpMenuItem.Disabled = dg.working || !selectedStep || dg.selectedStage == 0
	}
	if dg.moveStepDownMenuItem != nil {
		dg.moveStepDownMenuItem.Disabled = dg.working || !selectedStep || dg.selectedStage == dg.pipe.Len()-1
	}
	if dg.duplicateStepMenuItem != nil {
		dg.duplicateStepMenuItem.Disabled = dg.working || !selectedStep
	}
	if dg.removeStepMenuItem != nil {
		dg.removeStepMenuItem.Disabled = dg.working || !selectedStep
	}
}

func (dg *DeenGUI) selectTab(index int) {
	if index < 0 || index > 4 || dg.tabContent == nil {
		return
	}
	dg.closeCompactNavigation()
	if dg.activeTab == index {
		return
	}
	dg.activeTab = index
	if dg.app != nil {
		dg.app.Preferences().SetInt(workspacePreferenceKey, index)
	}
	if dg.homeCommands != nil {
		if index == 0 {
			dg.homeCommands.Show()
			setAdaptiveObjectVisible(dg.workspaceTitleDivider, false)
			setAdaptiveObjectVisible(dg.workspaceTitle, false)
		} else {
			dg.homeCommands.Hide()
			setAdaptiveObjectVisible(dg.workspaceTitleDivider, true)
			setAdaptiveObjectVisible(dg.workspaceTitle, true)
		}
	}
	if dg.pipelineNavButton != nil {
		if index == 0 {
			dg.pipelineNavButton.Show()
		} else {
			dg.pipelineNavButton.Hide()
		}
	}
	for i, tab := range dg.tabButtons {
		tab.setActive(i == index)
	}
	dg.refreshWorkspaceTitle()

	dg.tabContent.Objects = []fyne.CanvasObject{dg.cachedTab(index)}
	dg.tabContent.Refresh()
	if index == 1 && dg.workflowLibrary != nil {
		dg.workflowLibrary.refreshSelectedDetail()
	}
	if index == 4 && dg.compareWorkspace != nil {
		dg.compareWorkspace.refreshPoints()
	}
	dg.refreshNativeMenuState()
}

func (dg *DeenGUI) refreshWorkspaceTitle() {
	if dg.workspaceTitle == nil {
		return
	}
	titles := []string{"Pipeline", "Workflows", "Transformers", "About deen", "Compare"}
	if dg.activeTab < 0 || dg.activeTab >= len(titles) {
		return
	}
	dg.workspaceTitle.SetText(titles[dg.activeTab])
}

func (dg *DeenGUI) cachedTab(index int) fyne.CanvasObject {
	if dg.tabViews[index] != nil {
		return dg.tabViews[index]
	}
	var content fyne.CanvasObject
	switch index {
	case 1:
		content = dg.examplesTab()
	case 2:
		content = dg.pluginsTab()
	case 3:
		content = dg.aboutTab()
	case 4:
		content = dg.compareTab()
	default:
		content = dg.homeTab()
	}
	dg.tabViews[index] = compactMinWidth(content)
	return dg.tabViews[index]
}

func (dg *DeenGUI) setWorking(label string, working bool) {
	dg.setTransformerCatalogsDisabled(working)
	if dg.workflowLibrary != nil {
		dg.workflowLibrary.setDisabled(working)
	}
	if dg.compareWorkspace != nil {
		dg.compareWorkspace.setDisabled(working)
	}
	if dg.workStatus == nil || dg.workIndicator == nil {
		return
	}
	if label == "" {
		label = "Processing"
	}
	if working {
		dg.actionFeedbackActive = false
		if dg.window != nil {
			dg.window.Canvas().Unfocus()
		}
		dg.setWorkControlsDisabled(true)
		dg.workStatus.SetText(label + "…")
		dg.workStatus.Importance = widget.LowImportance
		dg.workStatus.Refresh()
		if dg.workActivity != nil {
			dg.workActivity.Show()
			dg.workActivity.Start()
		}
	} else {
		dg.setWorkControlsDisabled(false)
		if dg.workActivity != nil {
			dg.workActivity.Stop()
			dg.workActivity.Hide()
		}
		if !dg.actionFeedbackActive {
			dg.workStatus.SetText("Ready")
			dg.workStatus.Importance = widget.LowImportance
			dg.workStatus.Refresh()
		}
	}
	dg.refreshCommandButtons()
	dg.workIndicator.Refresh()
}

// showActionFeedback reports a completed user action in the persistent status
// area. The next processing cycle clears it; avoiding a timer keeps the message
// available to assistive technology and makes completion deterministic.
func (dg *DeenGUI) showActionFeedback(message string) {
	message = strings.TrimSpace(message)
	if message == "" || dg.workStatus == nil || dg.workIndicator == nil {
		return
	}
	dg.actionFeedbackActive = true
	dg.workStatus.SetText(message)
	dg.workStatus.Importance = widget.SuccessImportance
	dg.workStatus.Refresh()
	dg.workIndicator.Show()
	dg.workIndicator.Refresh()
}

func fileActionFeedback(fallback, verb string, uri fyne.URI) string {
	if uri == nil || strings.TrimSpace(uri.Name()) == "" {
		return fallback
	}
	return fmt.Sprintf("%s “%s”", verb, feedbackSubject(uri.Name()))
}

func feedbackSubject(subject string) string {
	const maxRunes = 40
	subject = strings.Join(strings.Fields(subject), " ")
	runes := []rune(subject)
	if len(runes) <= maxRunes {
		return subject
	}
	return string(runes[:maxRunes-1]) + "…"
}

func writeAndClose(wc fyne.URIWriteCloser, data []byte) error {
	written, writeErr := wc.Write(data)
	if writeErr == nil && written != len(data) {
		writeErr = io.ErrShortWrite
	}
	closeErr := wc.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func (dg *DeenGUI) refreshResultStatus() {
	if dg.resultStatus == nil || dg.pipe == nil {
		return
	}
	steps := dg.pipe.Len()
	stepLabel := "steps"
	if steps == 1 {
		stepLabel = "step"
	}
	dg.resultStatus.SetText(fmt.Sprintf("%d %s  •  %d B result", steps, stepLabel, len(dg.pipe.Result())))
}

func (dg *DeenGUI) registerWorkControl(control fyne.Disableable) {
	if control == nil {
		return
	}
	dg.workControls = append(dg.workControls, control)
	if dg.working {
		if dg.workControlStates == nil {
			dg.workControlStates = make(map[fyne.Disableable]bool)
		}
		dg.workControlStates[control] = control.Disabled()
		control.Disable()
	}
}

func (dg *DeenGUI) setWorkControlsDisabled(disabled bool) {
	if disabled {
		if dg.workControlStates == nil {
			dg.workControlStates = make(map[fyne.Disableable]bool, len(dg.workControls))
		}
		for _, control := range dg.workControls {
			if _, recorded := dg.workControlStates[control]; recorded {
				continue
			}
			dg.workControlStates[control] = control.Disabled()
			control.Disable()
		}
		return
	}
	for control, wasDisabled := range dg.workControlStates {
		if !wasDisabled {
			control.Enable()
		}
	}
	dg.workControlStates = nil
}

func (dg *DeenGUI) runPipelineWork(label string, work func() error, done func()) <-chan struct{} {
	completed := make(chan struct{})
	if dg.working {
		close(completed)
		return completed
	}
	dg.working = true
	dg.setWorking(label, true)
	go func() {
		err := work()
		fyne.Do(func() {
			defer func() {
				dg.working = false
				dg.setWorking("", false)
				if dg.themeRefreshPending {
					dg.themeRefreshPending = false
					dg.refreshAppearanceSurfaces()
				}
				close(completed)
			}()
			if err != nil {
				dialog.ShowError(err, dg.window)
				return
			}
			if done != nil {
				done()
			}
		})
	}()
	return completed
}

// mainMenu builds the window menu (theme switching).
func (dg *DeenGUI) mainMenu() *fyne.MainMenu {
	withShortcut := func(item *fyne.MenuItem, key fyne.KeyName, modifier fyne.KeyModifier) *fyne.MenuItem {
		item.Shortcut = &desktop.CustomShortcut{KeyName: key, Modifier: modifier}
		return item
	}
	standard := fyne.KeyModifierShortcutDefault
	appearanceItem := func(label string, mode appearanceMode, icon fyne.Resource) *fyne.MenuItem {
		item := fyne.NewMenuItemWithIcon(label, icon, func() { dg.applyAppearance(mode) })
		item.Checked = dg.appearance == mode
		return item
	}
	fileMenu := fyne.NewMenu("File",
		withShortcut(fyne.NewMenuItemWithIcon("Open file…", theme.FolderOpenIcon(), dg.openFile), fyne.KeyO, standard),
		withShortcut(fyne.NewMenuItemWithIcon("Save result…", theme.DocumentSaveIcon(), dg.saveResult), fyne.KeyS, standard),
	)
	dg.undoMenuItem = withShortcut(fyne.NewMenuItemWithIcon("Undo", theme.ContentUndoIcon(), dg.undo), fyne.KeyZ, standard)
	dg.redoMenuItem = withShortcut(fyne.NewMenuItemWithIcon("Redo", theme.ContentRedoIcon(), dg.redo), fyne.KeyZ, standard|fyne.KeyModifierShift)
	editMenu := fyne.NewMenu("Edit",
		dg.undoMenuItem,
		dg.redoMenuItem,
		fyne.NewMenuItemSeparator(),
		withShortcut(fyne.NewMenuItemWithIcon("Copy result", theme.ContentCopyIcon(), dg.copyResult), fyne.KeyC, standard|fyne.KeyModifierShift),
	)
	viewMenu := fyne.NewMenu("View",
		withShortcut(fyne.NewMenuItemWithIcon("Toggle sidebar", theme.MenuIcon(), dg.toggleSidebar), fyne.KeyS, standard|fyne.KeyModifierControl),
		withShortcut(fyne.NewMenuItemWithIcon("Toggle pipeline navigator", theme.ListIcon(), dg.togglePipelineNavigator), fyne.KeyP, standard|fyne.KeyModifierControl),
	)
	dg.previousStageMenuItem = withShortcut(fyne.NewMenuItemWithIcon("Previous pipeline stage", theme.MoveUpIcon(), func() {
		dg.navigatePipelineStage(-1)
	}), fyne.KeyUp, standard|fyne.KeyModifierAlt)
	dg.nextStageMenuItem = withShortcut(fyne.NewMenuItemWithIcon("Next pipeline stage", theme.MoveDownIcon(), func() {
		dg.navigatePipelineStage(1)
	}), fyne.KeyDown, standard|fyne.KeyModifierAlt)
	navigateMenu := fyne.NewMenu("Navigate",
		withShortcut(fyne.NewMenuItemWithIcon("Pipeline", theme.HomeIcon(), func() { dg.selectTab(0) }), fyne.Key1, standard),
		withShortcut(fyne.NewMenuItemWithIcon("Workflows", theme.HistoryIcon(), func() { dg.selectTab(1) }), fyne.Key2, standard),
		withShortcut(fyne.NewMenuItemWithIcon("Transformers", theme.SearchIcon(), func() { dg.selectTab(2) }), fyne.Key3, standard),
		withShortcut(fyne.NewMenuItemWithIcon("Compare", theme.ViewFullScreenIcon(), func() { dg.selectTab(4) }), fyne.Key4, standard),
		withShortcut(fyne.NewMenuItemWithIcon("About", theme.InfoIcon(), func() { dg.selectTab(3) }), fyne.Key5, standard),
		fyne.NewMenuItemSeparator(),
		dg.previousStageMenuItem,
		dg.nextStageMenuItem,
	)
	chainMenu := fyne.NewMenu("Chain",
		fyne.NewMenuItemWithIcon("Open chain", theme.FileTextIcon(), dg.openChain),
		fyne.NewMenuItemWithIcon("Save chain", theme.DocumentCreateIcon(), dg.saveChain),
		fyne.NewMenuItemWithIcon("Copy command", theme.MailForwardIcon(), dg.copyCommand),
	)
	dg.addMenuItem = withShortcut(fyne.NewMenuItemWithIcon("Add transformer…", theme.ContentAddIcon(), dg.showPluginSearch), fyne.KeyK, standard)
	dg.clearMenuItem = fyne.NewMenuItemWithIcon("Clear", theme.ContentClearIcon(), dg.clear)
	dg.toggleStepMenuItem = withShortcut(fyne.NewMenuItemWithIcon("Enable or disable selected step", theme.VisibilityIcon(), func() {
		dg.toggleStep(dg.selectedStage)
	}), fyne.KeyE, standard|fyne.KeyModifierShift)
	dg.moveStepUpMenuItem = withShortcut(fyne.NewMenuItemWithIcon("Move selected step up", theme.MoveUpIcon(), func() {
		dg.moveStep(dg.selectedStage, -1)
	}), fyne.KeyUp, standard|fyne.KeyModifierAlt|fyne.KeyModifierShift)
	dg.moveStepDownMenuItem = withShortcut(fyne.NewMenuItemWithIcon("Move selected step down", theme.MoveDownIcon(), func() {
		dg.moveStep(dg.selectedStage, 1)
	}), fyne.KeyDown, standard|fyne.KeyModifierAlt|fyne.KeyModifierShift)
	dg.duplicateStepMenuItem = withShortcut(fyne.NewMenuItemWithIcon("Duplicate selected step", theme.ContentCopyIcon(), func() {
		dg.duplicateStep(dg.selectedStage)
	}), fyne.KeyD, standard|fyne.KeyModifierShift)
	dg.removeStepMenuItem = fyne.NewMenuItemWithIcon("Remove selected step", theme.DeleteIcon(), func() {
		dg.removeStep(dg.selectedStage)
	})
	workflowMenu := fyne.NewMenu("Workflow",
		dg.addMenuItem,
		fyne.NewMenuItemSeparator(),
		dg.toggleStepMenuItem,
		dg.moveStepUpMenuItem,
		dg.moveStepDownMenuItem,
		dg.duplicateStepMenuItem,
		dg.removeStepMenuItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItemWithIcon("Presets", theme.HistoryIcon(), dg.showPresets),
		fyne.NewMenuItemWithIcon("Compare", theme.ViewFullScreenIcon(), dg.showCompare),
		dg.clearMenuItem,
	)
	themeMenu := fyne.NewMenu("Appearance",
		appearanceItem("System", appearanceSystem, theme.SettingsIcon()),
		appearanceItem("Dark", appearanceDark, theme.VisibilityIcon()),
		appearanceItem("Light", appearanceLight, theme.VisibilityOffIcon()),
	)
	help := fyne.NewMenu("Help",
		fyne.NewMenuItemWithIcon("How to use", theme.HelpIcon(), dg.showHelp),
		fyne.NewMenuItemWithIcon("Workflows", theme.HistoryIcon(), func() { dg.selectTab(1) }),
		fyne.NewMenuItemWithIcon("Plugin catalog", theme.SearchIcon(), func() { dg.selectTab(2) }),
		fyne.NewMenuItemWithIcon("Compare pipeline data", theme.ViewFullScreenIcon(), func() { dg.selectTab(4) }),
		fyne.NewMenuItemWithIcon("About", theme.InfoIcon(), func() { dg.selectTab(3) }),
	)
	menu := fyne.NewMainMenu(fileMenu, editMenu, viewMenu, navigateMenu, chainMenu, workflowMenu, themeMenu, help)
	dg.refreshNativeMenuState()
	return menu
}

func (dg *DeenGUI) applyAppearance(mode appearanceMode) {
	if dg.working {
		return
	}
	mode = normalizeAppearance(string(mode))
	dg.appearance = mode
	dg.app.Preferences().SetString(appearancePreferenceKey, string(mode))
	dg.updatingAppearance = true
	dg.app.Settings().SetTheme(themeForAppearance(mode))
	dg.updatingAppearance = false
	dg.themeRefreshPending = false
	dg.refreshAppearanceSurfaces()
	if dg.window != nil {
		dg.window.SetMainMenu(dg.mainMenu())
	}
}

func (dg *DeenGUI) handleSettingsChange() {
	if dg.appearance != appearanceSystem || dg.updatingAppearance {
		return
	}
	if dg.working {
		dg.themeRefreshPending = true
		return
	}
	dg.refreshAppearanceSurfaces()
}

func (dg *DeenGUI) refreshAppearanceSurfaces() {
	for i, tab := range dg.tabButtons {
		tab.setActive(i == dg.activeTab)
	}
	if dg.stepsBox != nil {
		dg.rebuild()
	}
}

func (dg *DeenGUI) homeTab() fyne.CanvasObject {
	if dg.pipelineOutline == nil {
		dg.pipelineOutline = container.NewVBox()
	}
	dg.pipelineOutlineCount = widget.NewLabel("")
	dg.pipelineOutlineCount.Importance = widget.LowImportance
	heading := container.NewVBox(
		widget.NewLabelWithStyle("Pipeline", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		dg.pipelineOutlineCount,
		widget.NewSeparator(),
	)
	outline := container.NewBorder(heading, nil, nil, nil, container.NewVScroll(dg.pipelineOutline))
	dg.pipelineNavigator = container.New(fixedWidthLayout{width: 232}, widget.NewCard("", "", container.NewPadded(outline)))
	detail := container.NewVScroll(dg.stepsBox)
	layout := &adaptiveLeadingLayout{
		leading:    dg.pipelineNavigator,
		content:    detail,
		breakpoint: compactStagesBreakpoint,
		isOpen: func(compact bool) bool {
			if compact {
				return dg.stagesDrawerOpen
			}
			return dg.pipelineNavOpen
		},
		onCompact: dg.setCompactStages,
	}
	dg.pipelineShell = newAdaptiveLeadingContainer(layout)
	return dg.pipelineShell
}

// showHelp displays a usage/info page describing the GUI.
func (dg *DeenGUI) showHelp() {
	md := `## Using deen

deen applies a **chain of transforms** to your input — like Burp Suite's
Decoder. The result of each step feeds into the next.

### Steps
- Use the Input **Editor** to type or paste data, open a file, or drop a file
  anywhere in the window. Switch to **Inspector** for Raw, Hex, Strings, and
  structured Preview representations; only the selected view is shown.
- Open **Add transformer** to search names, aliases, descriptions, and use cases.
  Filter by category, Favorites, or Recent; browse with the keyboard; and press
  Return to add. Favoriting a transformer and adding one updates those reusable
  filters locally.
- Tick **decode** to run a step in reverse (e.g. base64 decode). One-way
  plugins like hashes cannot be decoded.
- Plugin options (e.g. base64 ` + "`-url`" + `, gzip ` + "`-level`" + `) are grouped
  by behavior, values, and sensitive values. Invalid numeric input is highlighted;
  sensitive values are masked but remain plaintext in saved chain files.
- **hex** shows a step's output as a hex dump (read-only).
- Use **Open chain** and **Save chain** to reuse complete transform chains.
- Use **Analyze current result** in Add transformer to show likely next transforms
  inline without changing the pipeline.
- Open **Workflows** to browse runnable Examples and reusable Presets. Examples
  replace input and chain; Presets keep the current input and replace the chain.
- Use **Compare** from the toolbar or Workflow menu to inspect any two pipeline
  points side by side.
- Copy the equivalent shell pipeline from the toolbar.
- Editing any step's output recomputes everything below it.
- Use the disclosure arrow to **collapse/expand** a step, the trash icon
  to remove it.

### Workspace
- Use the sidebar to move between the pipeline, workflows, transformer catalog,
  comparison, and app information. Collapsing it keeps an icon rail visible;
  sidebar destinations and pipeline stages participate in Tab focus and
  activate with Space or Return.
- The **Workflows** workspace searches Examples and Presets together, previews
  bundled example data asynchronously, and labels input-replacement behavior
  before applying anything.
- The **Transformers** workspace shares the searchable catalog, Favorites, and
  Recent history with Add Transformer, while keeping its own browsing layout.
- Within the pipeline, use the stage navigator to focus the input, one
  transformer, or the add-transformer editor. It can be hidden independently
  when more editing space is needed. Command–Option–Up/Down moves to the
  previous or next stage on macOS.
- At compact window widths, the workspace sidebar automatically becomes an
  icon rail and can expand temporarily over the workspace. The stage navigator
  remains a temporary drawer. Selecting a destination closes either expansion,
  and resizing wider restores the saved desktop visibility choices.
- Configuration/output and other master/detail panes stack vertically when a
  compact layout has enough height to benefit. Wide and compact divider
  balances are remembered independently, so resizing never replaces one with
  the other.
- Drag the divider between Configuration and Output to balance the focused
  transformer editor. The chosen balance is remembered.
- Window size, workspace, focused stage, split balances, workflow/catalog
  filters, favorites/recents, and inspection/output tabs are restored on the
  next launch.
- The Pipeline page places file, result, history, and add-step commands in its
  single compact top row. When the complete command row cannot fit, Open, Save,
  Add, and More remain visible while Copy, Undo, and Redo move into More without
  losing their enabled or disabled state.
- The bottom status area reports background activity, step count, and result
  size without moving the editor.

### Menu
- **Appearance**: follow the system theme or force light/dark.
- **Workflow**: enable/disable, move, duplicate, or remove the focused pipeline
  step. Unavailable actions are disabled at the input, Add, and chain boundaries.
- macOS uses standard Command-key shortcuts; other desktops use their standard
  shortcut modifier. Command–Shift–E toggles the focused step,
  Command–Option–Shift–Up/Down reorders it, and Command–Shift–D duplicates it.`

	rich := widget.NewRichTextFromMarkdown(md)
	rich.Wrapping = fyne.TextWrapWord
	scroll := container.NewVScroll(rich)
	scroll.SetMinSize(fyne.NewSize(520, 420))
	dialog.ShowCustom("How to use deen", "Close", scroll, dg.window)
}

// showAbout displays version and project information.
func (dg *DeenGUI) showAbout() {
	dg.selectTab(3)
}

func (dg *DeenGUI) aboutTab() fyne.CanvasObject {
	version := core.Version()
	if b := core.Branch(); b != "" {
		version += " (" + b + ")"
	}
	docsURL, _ := url.Parse("https://deen.adversec.com")
	repoURL, _ := url.Parse("https://github.com/takeshixx/go-deen")

	content := container.NewVBox(
		widget.NewLabel("Version: "+version),
		widget.NewLabel("deen encodes, decodes, hashes, compresses and formats data\nthrough a configurable chain of plugins."),
		widget.NewLabel("Built with Go, Fyne (desktop GUI) and WebAssembly (web)."),
		container.NewHBox(widget.NewLabel("Documentation:"), widget.NewHyperlink("deen.adversec.com", docsURL)),
		container.NewHBox(widget.NewLabel("Source:"), widget.NewHyperlink("github.com/takeshixx/go-deen", repoURL)),
		widget.NewSeparator(),
		widget.NewLabel("All processing happens locally in the desktop GUI. The web UI runs the same pipeline model in WebAssembly."),
	)
	return container.NewPadded(container.NewVScroll(content))
}

func (dg *DeenGUI) examplesTab() fyne.CanvasObject {
	dg.workflowLibrary = newWorkflowLibrary(dg)
	return container.NewPadded(dg.workflowLibrary.splitContent)
}

func exampleSourceSummary(source []byte) string {
	return "Input: " + pipeline.DataMetadata(source, 0).Summary()
}

func exampleDataText(data []byte) string {
	if len(data) == 0 {
		return "(empty)"
	}
	if looksReadable(data) {
		return string(data)
	}
	return base64.StdEncoding.EncodeToString(data)
}

func exampleDataObject(data []byte, fallback fyne.CanvasObject) fyne.CanvasObject {
	img := canvas.NewImageFromReader(bytes.NewReader(data), "example-output")
	if img == nil || img.Image == nil {
		if preview, spans, ok := pipeline.HighlightedPreview(data); ok {
			grid := newPreviewGrid()
			setPreviewText(grid, preview, spans)
			return grid
		}
		return fallback
	}
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(180, 180))
	return container.NewCenter(img)
}

func looksReadable(data []byte) bool {
	if !utf8.Valid(data) {
		return false
	}
	for _, b := range data {
		if b == '\n' || b == '\r' || b == '\t' {
			continue
		}
		if b < 0x20 || b == 0x7f {
			return false
		}
	}
	return true
}

func (dg *DeenGUI) pluginsTab() fyne.CanvasObject {
	dg.browserCatalog = newTransformerCatalog(
		dg,
		"Browse transformers",
		browserCategoryPreferenceKey,
		browserSplitPreferenceKey,
		browserCompactSplitPreferenceKey,
		compactWorkspaceSplit,
	)
	return container.NewPadded(dg.browserCatalog.splitContent)
}

// rebuild recreates the pipeline navigator and the currently focused editor.
// Structural changes use this path; content-only changes use refreshFrom.
func (dg *DeenGUI) rebuild() {
	dg.setWorking("Rendering pipeline", true)
	defer dg.setWorking("", false)
	dg.rememberStepEditorSplit()
	dg.rememberAddSplit()
	dg.workControls = nil
	dg.workControlStates = nil
	dg.stepsBox.RemoveAll()
	dg.clampSelectedStage()
	dg.rebuildPipelineOutline()
	dg.cards = make([]*stepCard, dg.pipe.Len())
	dg.sourceEntry = nil
	dg.sourceRaw = nil
	dg.sourceHex = nil
	dg.sourceStrings = nil
	dg.sourcePreview = nil
	dg.sourcePreviewTab = nil
	dg.sourceViewer = nil
	dg.sourceWorkspace = nil
	dg.sourceMeta = nil
	dg.sourceFullControls = nil
	dg.addCatalog = nil
	dg.addSuggestions = nil

	switch {
	case dg.selectedStage == pipelineStageAdd:
		dg.stepsBox.Add(dg.newAddSlot())
	case dg.selectedStage == pipelineStageInput:
		dg.stepsBox.Add(dg.newSourceCard())
	case dg.selectedStage >= 0 && dg.selectedStage < dg.pipe.Len():
		c := dg.newStepCard(dg.selectedStage)
		dg.cards[dg.selectedStage] = c
		dg.stepsBox.Add(c.container)
	}
	dg.stepsBox.Refresh()
	dg.refreshCommandButtons()
	dg.refreshResultStatus()
	dg.refreshWorkspaceTitle()
	if dg.compareWorkspace != nil {
		dg.compareWorkspace.refreshPoints()
	}
	if dg.app != nil {
		dg.app.Preferences().SetInt(selectedStagePreferenceKey, dg.selectedStage)
	}
}

func (dg *DeenGUI) clampSelectedStage() {
	if dg.selectedStage == pipelineStageInput || dg.selectedStage == pipelineStageAdd {
		return
	}
	if dg.pipe.Len() == 0 {
		dg.selectedStage = pipelineStageInput
		return
	}
	if dg.selectedStage < 0 {
		dg.selectedStage = pipelineStageInput
	} else if dg.selectedStage >= dg.pipe.Len() {
		dg.selectedStage = dg.pipe.Len() - 1
	}
}

func (dg *DeenGUI) selectPipelineStage(stage int) {
	if dg.working || stage < pipelineStageAdd || stage >= dg.pipe.Len() {
		return
	}
	dg.closeCompactNavigation()
	if stage == dg.selectedStage {
		return
	}
	dg.selectedStage = stage
	dg.rebuild()
}

func (dg *DeenGUI) pipelineStagePosition() int {
	if dg.pipe == nil {
		return -1
	}
	switch {
	case dg.selectedStage == pipelineStageInput:
		return 0
	case dg.selectedStage == pipelineStageAdd:
		return dg.pipe.Len() + 1
	case dg.selectedStage >= 0 && dg.selectedStage < dg.pipe.Len():
		return dg.selectedStage + 1
	default:
		return -1
	}
}

func (dg *DeenGUI) navigatePipelineStage(delta int) {
	if dg.working || dg.activeTab != 0 || dg.pipe == nil || (delta != -1 && delta != 1) {
		return
	}
	position := dg.pipelineStagePosition()
	if position < 0 {
		return
	}
	position += delta
	if position < 0 || position > dg.pipe.Len()+1 {
		return
	}
	stage := pipelineStageInput
	switch {
	case position == dg.pipe.Len()+1:
		stage = pipelineStageAdd
	case position > 0:
		stage = position - 1
	}
	dg.selectPipelineStage(stage)
}

func completedGUIAction() <-chan struct{} {
	completed := make(chan struct{})
	close(completed)
	return completed
}

func (dg *DeenGUI) validStepIndex(index int) bool {
	return dg.pipe != nil && index >= 0 && index < dg.pipe.Len()
}

func (dg *DeenGUI) toggleStep(index int) <-chan struct{} {
	if dg.working || !dg.validStepIndex(index) {
		return completedGUIAction()
	}
	return dg.runPipelineWork("Updating step", func() error {
		dg.pipe.SetStepDisabled(index, !dg.pipe.Steps()[index].Disabled)
		return nil
	}, dg.rebuild)
}

func (dg *DeenGUI) moveStep(index, delta int) <-chan struct{} {
	target := index + delta
	if dg.working || !dg.validStepIndex(index) || !dg.validStepIndex(target) || (delta != -1 && delta != 1) {
		return completedGUIAction()
	}
	return dg.runPipelineWork("Moving step", func() error {
		dg.pipe.MoveStep(index, target)
		return nil
	}, func() {
		dg.selectedStage = target
		dg.rebuild()
	})
}

func (dg *DeenGUI) duplicateStep(index int) <-chan struct{} {
	if dg.working || !dg.validStepIndex(index) {
		return completedGUIAction()
	}
	return dg.runPipelineWork("Duplicating step", func() error {
		dg.pipe.DuplicateStep(index)
		return nil
	}, func() {
		dg.selectedStage = index + 1
		dg.rebuild()
	})
}

func (dg *DeenGUI) removeStep(index int) <-chan struct{} {
	if dg.working || !dg.validStepIndex(index) {
		return completedGUIAction()
	}
	return dg.runPipelineWork("Removing step", func() error {
		dg.pipe.RemoveStep(index)
		return nil
	}, func() {
		switch {
		case dg.pipe.Len() == 0:
			dg.selectedStage = pipelineStageInput
		case index >= dg.pipe.Len():
			dg.selectedStage = dg.pipe.Len() - 1
		default:
			dg.selectedStage = index
		}
		dg.rebuild()
	})
}

func (dg *DeenGUI) rebuildPipelineOutline() {
	if dg.pipelineOutline == nil {
		dg.pipelineOutline = container.NewVBox()
	}
	dg.pipelineOutline.RemoveAll()
	dg.pipelineStageButtons = dg.pipelineStageButtons[:0]
	if dg.pipelineOutlineCount != nil {
		steps := dg.pipe.Len()
		label := "steps"
		if steps == 1 {
			label = "step"
		}
		dg.pipelineOutlineCount.SetText(fmt.Sprintf("%d %s", steps, label))
	}

	input := newPipelineStageItem("Input", theme.DocumentIcon(), nil, func() {
		dg.selectPipelineStage(pipelineStageInput)
	})
	dg.pipelineStageButtons = append(dg.pipelineStageButtons, input)
	dg.pipelineOutline.Add(input)

	for i, step := range dg.pipe.Steps() {
		i := i
		stepColor := accent(i)
		icon := theme.NavigateNextIcon()
		direction := "encode"
		if step.Unprocess {
			icon = theme.NavigateBackIcon()
			direction = "decode"
		}
		state := ""
		if step.Disabled {
			stepColor = disabledAccent()
			icon = theme.VisibilityOffIcon()
			state = " · off"
		}
		name := plugins.PluginLabel(step.Plugin)
		if name == "" {
			name = "No transformer"
		}
		label := fmt.Sprintf("%d  %s · %s%s", i+1, name, direction, state)
		item := newPipelineStageItem(label, icon, stepColor, func() { dg.selectPipelineStage(i) })
		dg.pipelineStageButtons = append(dg.pipelineStageButtons, item)
		dg.pipelineOutline.Add(item)
	}

	add := newPipelineStageItem("Add transformer", theme.ContentAddIcon(), nil, func() {
		dg.selectPipelineStage(pipelineStageAdd)
	})
	dg.pipelineStageButtons = append(dg.pipelineStageButtons, add)
	dg.pipelineOutline.Add(add)

	for _, item := range dg.pipelineStageButtons {
		item.setActive(false)
	}
	switch dg.selectedStage {
	case pipelineStageInput:
		input.setActive(true)
	case pipelineStageAdd:
		add.setActive(true)
	default:
		if dg.selectedStage >= 0 && dg.selectedStage < dg.pipe.Len() {
			dg.pipelineStageButtons[dg.selectedStage+1].setActive(true)
		}
	}
	dg.pipelineOutline.Refresh()
}

// refreshFrom updates the displayed output of every card from index `from`
// downward without recreating widgets.
func (dg *DeenGUI) refreshFrom(from int) {
	dg.setWorking("Refreshing output", true)
	defer dg.setWorking("", false)
	if dg.sourceMeta != nil {
		dg.sourceMeta.SetText(dg.sourceMetadataSummary())
	}
	rawNeedsFull, hexNeedsFull, stringsNeedsFull := dg.sourceNeedsFull()
	if !rawNeedsFull {
		dg.sourceFullRaw = false
	}
	if !hexNeedsFull {
		dg.sourceFullHex = false
	}
	if !stringsNeedsFull {
		dg.sourceFullStrings = false
	}
	if dg.sourceEntry != nil {
		text, textCapped := guiTextDisplayMode(dg.pipe.Source(), dg.sourceFullRaw)
		dg.setText(dg.sourceEntry, text)
		if textCapped {
			dg.sourceEntry.Disable()
		} else {
			dg.sourceEntry.Enable()
		}
	}
	if dg.sourceRaw != nil {
		text, _ := guiTextDisplayMode(dg.pipe.Source(), dg.sourceFullRaw)
		dg.setText(dg.sourceRaw, text)
	}
	if dg.sourceHex != nil {
		hexText, _ := guiHexDisplayMode(dg.pipe.Source(), dg.sourceFullHex)
		dg.setText(dg.sourceHex, hexText)
	}
	if dg.sourceStrings != nil {
		stringsText, _ := guiStringsDisplayMode(dg.pipe.Source(), dg.sourceFullStrings)
		dg.setText(dg.sourceStrings, stringsText)
	}
	dg.syncSourcePreviewTab()
	if dg.sourcePreview != nil {
		preview, spans, _ := pipeline.HighlightedPreview(dg.pipe.Source())
		setPreviewText(dg.sourcePreview, preview, spans)
	}
	dg.refreshSourceFullControls(rawNeedsFull, hexNeedsFull, stringsNeedsFull)
	for i := from; i < len(dg.cards); i++ {
		if dg.cards[i] == nil || dg.cards[i].collapsed {
			continue
		}
		dg.cards[i].refresh()
	}
	dg.refreshResultStatus()
	if dg.compareWorkspace != nil {
		dg.compareWorkspace.refreshPoints()
	}
}

// setText updates an entry programmatically without triggering its OnChanged.
func (dg *DeenGUI) setText(e *widget.Entry, s string) {
	dg.updating = true
	e.SetText(s)
	dg.updating = false
}

// --- toolbar actions ---

func (dg *DeenGUI) closeIfWorking(closer io.Closer) bool {
	if !dg.working {
		return false
	}
	_ = closer.Close()
	return true
}

func (dg *DeenGUI) openFile() {
	if dg.working {
		return
	}
	dialog.ShowFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, dg.window)
			return
		}
		if rc == nil {
			return
		}
		dg.loadSourceReader(rc, rc.URI().Name())
	}, dg.window)
}

func (dg *DeenGUI) handleDroppedFiles(_ fyne.Position, uris []fyne.URI) {
	if dg.working || len(uris) == 0 {
		return
	}
	rc, err := storage.Reader(uris[0])
	if err != nil {
		dialog.ShowError(err, dg.window)
		return
	}
	dg.loadSourceReader(rc, uris[0].Name())
}

func (dg *DeenGUI) loadSourceReader(rc fyne.URIReadCloser, name string) <-chan struct{} {
	if rc == nil || dg.closeIfWorking(rc) {
		completed := make(chan struct{})
		close(completed)
		return completed
	}
	var data []byte
	return dg.runPipelineWork("Processing file", func() error {
		defer rc.Close()
		loaded, err := io.ReadAll(rc)
		if err != nil {
			return err
		}
		data = loaded
		dg.pipe.SetSourceOwned(data)
		return nil
	}, func() {
		dg.sourceName = name
		dg.clearSourceFullViews()
		dg.selectedStage = pipelineStageInput
		dg.rebuild()
		dg.selectTab(0)
		if strings.TrimSpace(name) == "" {
			dg.showActionFeedback("Input loaded")
		} else {
			dg.showActionFeedback(fmt.Sprintf("Loaded “%s”", feedbackSubject(name)))
		}
	})
}

func (dg *DeenGUI) sourceMetadataSummary() string {
	return metadataSummary(dg.sourceName, pipeline.DataMetadata(dg.pipe.Source(), 0))
}

func metadataSummary(source string, meta pipeline.Metadata) string {
	var lines []string
	if strings.TrimSpace(source) != "" {
		lines = append(lines, "source: "+source)
	}
	for _, field := range meta.Fields() {
		lines = append(lines, field.Label+": "+field.Value)
	}
	return strings.Join(lines, "\n")
}

func (dg *DeenGUI) saveResult() {
	if dg.working {
		return
	}
	dialog.ShowFileSave(func(wc fyne.URIWriteCloser, err error) {
		if err != nil || wc == nil {
			return
		}
		if dg.closeIfWorking(wc) {
			return
		}
		uri := wc.URI()
		if err := writeAndClose(wc, dg.pipe.Result()); err != nil {
			dialog.ShowError(err, dg.window)
			return
		}
		dg.showActionFeedback(fileActionFeedback("Result saved", "Saved", uri))
	}, dg.window)
}

func (dg *DeenGUI) openChain() {
	if dg.working {
		return
	}
	dialog.ShowFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil || rc == nil {
			return
		}
		if dg.closeIfWorking(rc) {
			return
		}
		uri := rc.URI()
		dg.runPipelineWork("Importing chain", func() error {
			defer rc.Close()
			data, err := io.ReadAll(rc)
			if err != nil {
				return err
			}
			dg.clearSourceFullViews()
			if err := dg.pipe.ImportJSON(data); err != nil {
				return err
			}
			return nil
		}, func() {
			dg.selectedStage = pipelineStageInput
			dg.rebuild()
			dg.selectTab(0)
			dg.showActionFeedback(fileActionFeedback("Chain loaded", "Loaded", uri))
		})
	}, dg.window)
}

func (dg *DeenGUI) saveChain() {
	if dg.working {
		return
	}
	dialog.ShowFileSave(func(wc fyne.URIWriteCloser, err error) {
		if err != nil || wc == nil {
			return
		}
		if dg.closeIfWorking(wc) {
			return
		}
		data, err := dg.pipe.ExportJSON()
		if err != nil {
			_ = wc.Close()
			dialog.ShowError(err, dg.window)
			return
		}
		uri := wc.URI()
		if err := writeAndClose(wc, data); err != nil {
			dialog.ShowError(err, dg.window)
			return
		}
		dg.showActionFeedback(fileActionFeedback("Chain saved", "Saved", uri))
	}, dg.window)
}

func (dg *DeenGUI) copyResult() {
	if dg.working {
		return
	}
	if pipeline.IsLargeData(dg.pipe.Result()) {
		dialog.ShowInformation("Copy result", "Result is too large to copy safely from the GUI. Use Save result instead.", dg.window)
		return
	}
	dg.window.Clipboard().SetContent(string(dg.pipe.Result()))
	dg.showActionFeedback("Result copied")
}

func (dg *DeenGUI) copyCommand() {
	if dg.working {
		return
	}
	command := dg.pipe.CommandLine()
	if command == "" {
		dialog.ShowInformation("Command line", "No enabled transforms to export.", dg.window)
		return
	}
	dg.window.Clipboard().SetContent(command)
	dg.showActionFeedback("Command copied")
}

func (dg *DeenGUI) showCompare() {
	if dg.working {
		return
	}
	dg.selectTab(4)
}

func guiSafeSuggestionPreview(preview string) string {
	var b strings.Builder
	for _, r := range preview {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			b.WriteRune(r)
		case strconv.IsPrint(r):
			b.WriteRune(r)
		case r <= 0xff:
			fmt.Fprintf(&b, `\x%02x`, r)
		case r <= 0xffff:
			fmt.Fprintf(&b, `\u%04x`, r)
		default:
			fmt.Fprintf(&b, `\U%08x`, r)
		}
	}
	return b.String()
}

func (dg *DeenGUI) showPluginSearch() {
	if dg.working {
		return
	}
	dg.selectTab(0)
	if dg.selectedStage != pipelineStageAdd {
		dg.selectedStage = pipelineStageAdd
		dg.rebuild()
	}
	if dg.addCatalog != nil && dg.window != nil {
		dg.window.Canvas().Focus(dg.addCatalog.search)
	}
}

func (dg *DeenGUI) showPresets() {
	if dg.working {
		return
	}
	dg.selectTab(1)
	if dg.workflowLibrary == nil {
		return
	}
	dg.workflowLibrary.search.SetText("")
	dg.workflowLibrary.filter.SetSelected(presetWorkflowsFilter)
	if dg.window != nil {
		dg.window.Canvas().Focus(dg.workflowLibrary.search)
	}
}

func (dg *DeenGUI) undo() {
	if dg.working {
		return
	}
	dg.runPipelineWork("Undoing", func() error {
		dg.pipe.Undo()
		return nil
	}, dg.rebuild)
}

func (dg *DeenGUI) redo() {
	if dg.working {
		return
	}
	dg.runPipelineWork("Redoing", func() error {
		dg.pipe.Redo()
		return nil
	}, dg.rebuild)
}

func (dg *DeenGUI) clear() {
	if dg.working {
		return
	}
	dg.sourceName = ""
	dg.selectedStage = pipelineStageInput
	dg.runPipelineWork("Clearing", func() error {
		dg.pipe.Clear()
		return nil
	}, func() {
		dg.rebuild()
		dg.selectTab(0)
	})
}
