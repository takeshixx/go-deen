//go:build gui

package gui

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/takeshixx/deen/internal/pipeline"
	"github.com/takeshixx/deen/internal/plugins"
)

const (
	allWorkflowsFilter     = "All workflows"
	exampleWorkflowsFilter = "Examples"
	presetWorkflowsFilter  = "Presets"
	workflowPreviewRows    = 14
)

type workflowKind string

const (
	workflowExample workflowKind = "example"
	workflowPreset  workflowKind = "preset"
)

type workflowItem struct {
	kind    workflowKind
	name    string
	detail  string
	example pipeline.Example
	preset  pipeline.Preset
}

func (item workflowItem) key() string {
	return string(item.kind) + "\x00" + item.name
}

func (item workflowItem) steps() []pipeline.PresetStep {
	if item.kind == workflowExample {
		return item.example.Steps
	}
	return item.preset.Steps
}

func builtinWorkflowItems() []workflowItem {
	examples := pipeline.BuiltinExamples()
	presets := pipeline.BuiltinPresets()
	items := make([]workflowItem, 0, len(examples)+len(presets))
	for _, example := range examples {
		items = append(items, workflowItem{
			kind:    workflowExample,
			name:    example.Name,
			detail:  example.Description,
			example: example,
		})
	}
	for _, preset := range presets {
		items = append(items, workflowItem{
			kind:   workflowPreset,
			name:   preset.Name,
			detail: preset.Description,
			preset: preset,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		return strings.ToLower(items[i].name) < strings.ToLower(items[j].name)
	})
	return items
}

func normalizeWorkflowFilter(value string) string {
	switch value {
	case exampleWorkflowsFilter, presetWorkflowsFilter:
		return value
	default:
		return allWorkflowsFilter
	}
}

func workflowMatches(item workflowItem, query, filter string) bool {
	filter = normalizeWorkflowFilter(filter)
	if filter == exampleWorkflowsFilter && item.kind != workflowExample {
		return false
	}
	if filter == presetWorkflowsFilter && item.kind != workflowPreset {
		return false
	}
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	parts := []string{item.name, item.detail, string(item.kind)}
	if item.kind == workflowExample {
		parts = append(parts, "example sample replaces input", string(item.example.Source), item.example.WantContains)
	} else {
		parts = append(parts, "preset recipe keeps input")
	}
	for _, step := range item.steps() {
		parts = append(parts, step.Plugin, plugins.PluginLabel(step.Plugin))
		if step.Unprocess {
			parts = append(parts, "decode", "."+step.Plugin)
		} else {
			parts = append(parts, "encode")
		}
		for key, value := range step.Options {
			parts = append(parts, key, value)
		}
	}
	return strings.Contains(strings.ToLower(strings.Join(parts, " ")), query)
}

type workflowLibrary struct {
	gui                 *DeenGUI
	items               []workflowItem
	matches             []workflowItem
	search              *selectionSearchEntry
	filter              *widget.Select
	resultSummary       *widget.Label
	results             *widget.List
	detail              *fyne.Container
	detailScroll        *container.Scroll
	split               *container.Split
	splitController     *responsiveSplitController
	splitContent        fyne.CanvasObject
	controls            []fyne.Disableable
	detailControls      []fyne.Disableable
	selected            int
	detailVersion       int
	currentInputSummary *widget.Label
	previewSlot         *fyne.Container
	previewInput        *widget.Entry
	previewOutput       *widget.Entry
}

func newWorkflowLibrary(dg *DeenGUI) *workflowLibrary {
	library := &workflowLibrary{
		gui:      dg,
		items:    builtinWorkflowItems(),
		selected: -1,
		detail:   container.NewVBox(),
	}
	library.search = newSelectionSearchEntry(library.moveSelection)
	library.search.SetPlaceHolder("Search workflows")
	library.search.Icon = theme.SearchIcon()
	library.filter = widget.NewSelect([]string{allWorkflowsFilter, exampleWorkflowsFilter, presetWorkflowsFilter}, nil)
	library.resultSummary = lowImportanceLabel("")

	filter := allWorkflowsFilter
	splitOffset := defaultAddSplit
	if dg.app != nil {
		filter = normalizeWorkflowFilter(dg.app.Preferences().StringWithFallback(workflowFilterPreferenceKey, allWorkflowsFilter))
		splitOffset = normalizeAddSplit(dg.app.Preferences().FloatWithFallback(workflowSplitPreferenceKey, defaultAddSplit))
	}
	library.filter.SetSelected(filter)

	library.results = widget.NewList(
		func() int { return len(library.matches) },
		func() fyne.CanvasObject {
			title := widget.NewLabelWithStyle("Workflow", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			meta := widget.NewLabel("Type • steps")
			meta.Importance = widget.LowImportance
			return container.NewVBox(title, meta)
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {
			if id < 0 || id >= len(library.matches) {
				return
			}
			item := library.matches[id]
			row := object.(*fyne.Container)
			row.Objects[0].(*widget.Label).SetText(item.name)
			kind := "Preset • keeps current input"
			if item.kind == workflowExample {
				kind = "Example • includes sample input"
			}
			row.Objects[1].(*widget.Label).SetText(kind + " • " + workflowStepCountLabel(len(item.steps())))
		},
	)
	library.results.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(library.matches) {
			return
		}
		library.selected = int(id)
		library.renderDetail(library.matches[id])
	}
	library.search.OnChanged = func(string) { library.refresh() }
	library.search.OnSubmitted = func(string) {
		if item, ok := library.selectedItem(); ok {
			dg.applyWorkflow(item)
		}
	}
	library.filter.OnChanged = func(value string) {
		if dg.app != nil {
			dg.app.Preferences().SetString(workflowFilterPreferenceKey, normalizeWorkflowFilter(value))
		}
		library.refresh()
	}
	library.registerControl(library.search)
	library.registerControl(library.filter)

	filters := container.NewVBox(
		widget.NewLabelWithStyle("Find a workflow", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		library.search,
		library.filter,
		library.resultSummary,
		lowImportanceLabel("Examples replace input and chain. Presets keep input and replace only the chain. Use arrow keys to browse."),
	)
	resultsPane := widget.NewCard("Workflows", "", container.NewBorder(filters, nil, nil, nil, library.results))
	library.detailScroll = container.NewVScroll(library.detail)
	detailPane := widget.NewCard("Details", "", library.detailScroll)
	library.split = container.NewHSplit(resultsPane, detailPane)
	library.splitController = newResponsiveSplit(
		dg,
		library.split,
		workflowSplitPreferenceKey,
		workflowCompactPreferenceKey,
		splitOffset,
		defaultCompactMasterDetailSplit,
		normalizeAddSplit,
		compactWorkspaceSplit,
	)
	library.splitContent = library.splitController.host
	library.refresh()
	return library
}

func (library *workflowLibrary) refresh() {
	if library.results == nil || library.detail == nil {
		return
	}
	library.matches = library.matchesFor(library.search.Text, library.filter.Selected)
	library.resultSummary.SetText(workflowResultSummary(len(library.matches)))
	library.results.UnselectAll()
	for id := range library.matches {
		library.results.SetItemHeight(widget.ListItemID(id), 72)
	}
	library.results.Refresh()
	if len(library.matches) == 0 {
		library.selected = -1
		library.forgetDetailControls()
		library.detail.RemoveAll()
		library.detail.Add(wrappingLabel("No workflows match this search and filter."))
		library.detail.Refresh()
		return
	}
	library.results.Select(0)
	library.results.ScrollToTop()
}

func (library *workflowLibrary) matchesFor(query, filter string) []workflowItem {
	matches := make([]workflowItem, 0, len(library.items))
	for _, item := range library.items {
		if workflowMatches(item, query, filter) {
			matches = append(matches, item)
		}
	}
	return matches
}

func workflowResultSummary(count int) string {
	if count == 1 {
		return "1 workflow"
	}
	return fmt.Sprintf("%d workflows", count)
}

func workflowStepCountLabel(count int) string {
	if count == 1 {
		return "1 step"
	}
	return fmt.Sprintf("%d steps", count)
}

func (library *workflowLibrary) selectedItem() (workflowItem, bool) {
	if library.selected < 0 || library.selected >= len(library.matches) {
		return workflowItem{}, false
	}
	return library.matches[library.selected], true
}

func (library *workflowLibrary) moveSelection(delta int) {
	if delta == 0 || library.results == nil || len(library.matches) == 0 {
		return
	}
	next := library.selected + delta
	if library.selected < 0 {
		if delta < 0 {
			next = len(library.matches) - 1
		} else {
			next = 0
		}
	}
	if next < 0 {
		next = 0
	}
	if next >= len(library.matches) {
		next = len(library.matches) - 1
	}
	if next == library.selected {
		return
	}
	library.results.Select(widget.ListItemID(next))
	library.results.ScrollTo(widget.ListItemID(next))
}

func (library *workflowLibrary) renderDetail(item workflowItem) {
	library.detailVersion++
	library.currentInputSummary = nil
	library.previewSlot = nil
	library.previewInput = nil
	library.previewOutput = nil
	library.forgetDetailControls()
	library.detail.RemoveAll()
	library.detail.Add(widget.NewLabelWithStyle(item.name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	kind := "Chain preset • keeps the current input"
	if item.kind == workflowExample {
		kind = "Runnable example • replaces input and chain"
	}
	library.detail.Add(lowImportanceLabel(kind))
	library.detail.Add(wrappingLabel(item.detail))
	library.detail.Add(widget.NewSeparator())
	library.detail.Add(widget.NewLabelWithStyle("Pipeline", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	library.detail.Add(workflowStepList(item.steps()))

	if item.kind == workflowExample {
		library.renderExampleActions(item)
	} else {
		library.renderPresetActions(item)
	}
	library.detail.Refresh()
	if library.detailScroll != nil {
		library.detailScroll.Refresh()
		library.detailScroll.ScrollToTop()
	}
}

func (library *workflowLibrary) renderExampleActions(item workflowItem) {
	library.detail.Add(widget.NewSeparator())
	library.detail.Add(widget.NewLabelWithStyle("Bundled data", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	library.detail.Add(lowImportanceLabel(exampleSourceSummary(item.example.Source)))
	if item.example.WantContains != "" {
		library.detail.Add(lowImportanceLabel("Expected result contains: " + item.example.WantContains))
	}
	load := widget.NewButtonWithIcon("Load example", theme.MediaPlayIcon(), func() { library.gui.applyWorkflow(item) })
	load.Importance = widget.HighImportance
	library.previewSlot = container.NewVBox()
	version := library.detailVersion
	preview := widget.NewButtonWithIcon("Preview bundled data", theme.VisibilityIcon(), func() {
		library.previewExample(item, version, library.previewSlot)
	})
	library.registerDetailControl(load)
	library.registerDetailControl(preview)
	library.detail.Add(container.NewHBox(load, preview))
	library.detail.Add(lowImportanceLabel("Loading is undoable, but it replaces the current input and chain."))
	library.detail.Add(library.previewSlot)
}

func (library *workflowLibrary) renderPresetActions(item workflowItem) {
	library.detail.Add(widget.NewSeparator())
	library.detail.Add(widget.NewLabelWithStyle("Current input", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	library.currentInputSummary = lowImportanceLabel(pipeline.DataMetadata(library.gui.pipe.Source(), 0).Summary())
	library.detail.Add(library.currentInputSummary)
	apply := widget.NewButtonWithIcon("Apply to current input", theme.ConfirmIcon(), func() { library.gui.applyWorkflow(item) })
	apply.Importance = widget.HighImportance
	library.registerDetailControl(apply)
	library.detail.Add(apply)
	library.detail.Add(lowImportanceLabel("Applying is undoable and replaces only the current transform chain."))
}

func (library *workflowLibrary) refreshSelectedDetail() {
	if item, ok := library.selectedItem(); ok {
		library.renderDetail(item)
	}
}

func workflowStepList(steps []pipeline.PresetStep) fyne.CanvasObject {
	rows := container.NewVBox()
	for index, step := range steps {
		direction := "encode"
		if step.Unprocess {
			direction = "decode"
		}
		title := widget.NewLabelWithStyle(fmt.Sprintf("%d  %s • %s", index+1, plugins.PluginLabel(step.Plugin), direction), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		row := container.NewVBox(title)
		if len(step.Options) > 0 {
			keys := make([]string, 0, len(step.Options))
			for key := range step.Options {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			options := make([]string, 0, len(keys))
			for _, key := range keys {
				options = append(options, key+"="+step.Options[key])
			}
			row.Add(lowImportanceLabel(strings.Join(options, "  •  ")))
		}
		rows.Add(widget.NewCard("", "", row))
	}
	return rows
}

func (library *workflowLibrary) previewExample(item workflowItem, version int, previewSlot *fyne.Container) <-chan struct{} {
	if library.gui.working {
		return completedGUIAction()
	}
	var result []byte
	return library.gui.runPipelineWork("Preparing example preview", func() error {
		var err error
		result, err = pipeline.ExampleResult(item.example)
		return err
	}, func() {
		selected, ok := library.selectedItem()
		if !ok || selected.key() != item.key() || library.detailVersion != version {
			return
		}
		previewSlot.RemoveAll()
		previewSlot.Add(lowImportanceLabel("Output: " + pipeline.DataMetadata(result, len(item.example.Source)).Summary()))
		inputEntry := multilineEntry(workflowPreviewRows)
		inputEntry.SetText(exampleDataText(item.example.Source))
		inputEntry.Disable()
		outputEntry := multilineEntry(workflowPreviewRows)
		outputEntry.SetText(exampleDataText(result))
		outputEntry.Disable()
		library.previewInput = inputEntry
		library.previewOutput = outputEntry
		previewSlot.Add(container.NewVBox(
			widget.NewCard("Input data", "", exampleDataObject(item.example.Source, inputEntry)),
			widget.NewCard("Output result", "", exampleDataObject(result, outputEntry)),
		))
		previewSlot.Refresh()
		library.detail.Refresh()
		if library.detailScroll != nil {
			library.detailScroll.Refresh()
			library.detailScroll.ScrollToOffset(fyne.NewPos(0, previewSlot.Position().Y))
		}
	})
}

func (dg *DeenGUI) applyWorkflow(item workflowItem) <-chan struct{} {
	if dg.working || (item.kind != workflowExample && item.kind != workflowPreset) {
		return completedGUIAction()
	}
	return dg.runPipelineWork("Loading workflow", func() error {
		if item.kind == workflowExample {
			dg.pipe.ApplyExample(item.example)
		} else {
			dg.pipe.ApplyPreset(item.preset)
		}
		return nil
	}, func() {
		if item.kind == workflowExample {
			dg.sourceName = ""
		}
		dg.selectedStage = pipelineStageInput
		dg.rebuild()
		dg.selectTab(0)
		if item.kind == workflowExample {
			dg.showActionFeedback(fmt.Sprintf("Loaded example “%s”", feedbackSubject(item.name)))
		} else {
			dg.showActionFeedback(fmt.Sprintf("Applied preset “%s”", feedbackSubject(item.name)))
		}
	})
}

func (library *workflowLibrary) registerControl(control fyne.Disableable) {
	library.controls = append(library.controls, control)
	if library.gui.working {
		control.Disable()
	}
}

func (library *workflowLibrary) registerDetailControl(control fyne.Disableable) {
	library.detailControls = append(library.detailControls, control)
	library.registerControl(control)
}

func (library *workflowLibrary) forgetDetailControls() {
	if len(library.detailControls) == 0 || len(library.controls) == 0 {
		library.detailControls = nil
		return
	}
	stale := make(map[fyne.Disableable]bool, len(library.detailControls))
	for _, control := range library.detailControls {
		stale[control] = true
	}
	retained := library.controls[:0]
	for _, control := range library.controls {
		if !stale[control] {
			retained = append(retained, control)
		}
	}
	library.controls = retained
	library.detailControls = nil
}

func (library *workflowLibrary) setDisabled(disabled bool) {
	for _, control := range library.controls {
		if disabled {
			control.Disable()
		} else {
			control.Enable()
		}
	}
}
