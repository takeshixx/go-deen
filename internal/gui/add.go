//go:build gui

package gui

import (
	"fmt"
	"net/url"
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
	allTransformerCategories    = "All categories"
	favoriteTransformerCategory = "Favorites"
	recentTransformerCategory   = "Recent"
	maxFavoriteTransformers     = 64
	maxRecentTransformers       = 8
)

// selectionSearchEntry keeps focus in the query while Up/Down browse results.
// A regular single-line Entry consumes these keys as cursor movement, which
// makes keyboard-driven result navigation impossible while typing.
type selectionSearchEntry struct {
	widget.Entry
	onMoveSelection func(delta int)
}

func newSelectionSearchEntry(onMoveSelection func(delta int)) *selectionSearchEntry {
	entry := &selectionSearchEntry{onMoveSelection: onMoveSelection}
	entry.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
	entry.ExtendBaseWidget(entry)
	return entry
}

func (entry *selectionSearchEntry) TypedKey(event *fyne.KeyEvent) {
	if !entry.Disabled() && entry.onMoveSelection != nil {
		switch event.Name {
		case fyne.KeyUp:
			entry.onMoveSelection(-1)
			return
		case fyne.KeyDown:
			entry.onMoveSelection(1)
			return
		}
	}
	entry.Entry.TypedKey(event)
}

type transformerCatalog struct {
	gui                   *DeenGUI
	categoryPreferenceKey string
	splitPreferenceKey    string
	compactPreferenceKey  string
	heading               string
	search                *selectionSearchEntry
	category              *widget.Select
	resultSummary         *widget.Label
	results               *widget.List
	matches               []plugins.UIPluginInfo
	detail                *fyne.Container
	split                 *container.Split
	splitController       *responsiveSplitController
	splitContent          fyne.CanvasObject
	controls              []fyne.Disableable
	detailControls        []fyne.Disableable
	selected              int
}

func newTransformerCatalog(
	dg *DeenGUI,
	heading,
	categoryPreferenceKey,
	splitPreferenceKey,
	compactPreferenceKey string,
	compactWhen func(fyne.Size) bool,
) *transformerCatalog {
	catalog := &transformerCatalog{
		gui:                   dg,
		categoryPreferenceKey: categoryPreferenceKey,
		splitPreferenceKey:    splitPreferenceKey,
		compactPreferenceKey:  compactPreferenceKey,
		heading:               heading,
		selected:              -1,
	}
	catalog.search = newSelectionSearchEntry(catalog.moveSelection)
	catalog.search.SetPlaceHolder("Search transformers")
	catalog.search.Icon = theme.SearchIcon()
	catalog.category = widget.NewSelect(transformerCategoryLabels(), nil)
	catalog.resultSummary = lowImportanceLabel("")
	catalog.detail = container.NewVBox()

	category := allTransformerCategories
	splitOffset := defaultAddSplit
	if dg.app != nil {
		category = normalizeAddCategory(dg.app.Preferences().StringWithFallback(categoryPreferenceKey, allTransformerCategories))
		splitOffset = normalizeAddSplit(dg.app.Preferences().FloatWithFallback(splitPreferenceKey, defaultAddSplit))
	}
	catalog.category.SetSelected(category)

	catalog.results = widget.NewList(
		func() int { return len(catalog.matches) },
		func() fyne.CanvasObject {
			title := widget.NewLabelWithStyle("Transformer", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			meta := widget.NewLabel("Category")
			meta.Importance = widget.LowImportance
			return container.NewVBox(title, meta)
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {
			if id < 0 || id >= len(catalog.matches) {
				return
			}
			row := object.(*fyne.Container)
			info := catalog.matches[id]
			row.Objects[0].(*widget.Label).SetText(info.Label)
			direction := "encode only"
			if info.CanDecode {
				direction = "encode and decode"
			}
			row.Objects[1].(*widget.Label).SetText(plugins.CategoryLabel(info.Category) + "  •  " + direction)
		},
	)
	catalog.results.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(catalog.matches) {
			return
		}
		catalog.selected = int(id)
		catalog.renderDetail(catalog.matches[id])
	}
	catalog.search.OnChanged = func(string) { catalog.refresh() }
	catalog.search.OnSubmitted = func(string) {
		if catalog.selected >= 0 && catalog.selected < len(catalog.matches) {
			dg.appendTransformer(catalog.matches[catalog.selected], false)
		}
	}
	catalog.category.OnChanged = func(category string) {
		if dg.app != nil {
			dg.app.Preferences().SetString(categoryPreferenceKey, normalizeAddCategory(category))
		}
		catalog.refresh()
	}
	catalog.registerControl(catalog.search)
	catalog.registerControl(catalog.category)

	filters := container.NewVBox(
		widget.NewLabelWithStyle(heading, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		catalog.search,
		catalog.category,
		catalog.resultSummary,
		lowImportanceLabel("Search names, aliases, descriptions, and use cases. Use arrow keys to browse; press Return to add."),
	)
	resultsPane := widget.NewCard("Transformers", "", container.NewBorder(filters, nil, nil, nil, catalog.results))
	detailPane := widget.NewCard("Details", "", container.NewVScroll(catalog.detail))
	catalog.split = container.NewHSplit(resultsPane, detailPane)
	catalog.splitController = newResponsiveSplit(
		dg,
		catalog.split,
		splitPreferenceKey,
		compactPreferenceKey,
		splitOffset,
		defaultCompactMasterDetailSplit,
		normalizeAddSplit,
		compactWhen,
	)
	catalog.splitContent = catalog.splitController.host
	catalog.refresh()
	return catalog
}

// newAddSlot builds the focused, searchable transformer catalog.
func (dg *DeenGUI) newAddSlot() fyne.CanvasObject {
	dg.addCatalog = newTransformerCatalog(
		dg,
		"Find a transformer",
		addCategoryPreferenceKey,
		addSplitPreferenceKey,
		addCompactSplitPreferenceKey,
		func(fyne.Size) bool { return dg.compactStages },
	)
	dg.addSuggestions = container.NewVBox()
	catalog := container.New(fixedHeightLayout{height: focusedStepEditorHeight}, dg.addCatalog.splitContent)

	detect := widget.NewButtonWithIcon("Analyze current result", theme.SearchIcon(), dg.loadInlineSuggestions)
	detect.Importance = widget.HighImportance
	dg.registerWorkControl(detect)
	dg.addSuggestions.Add(container.NewBorder(nil, nil,
		widget.NewLabelWithStyle("Suggested next transforms", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), detect))
	dg.addSuggestions.Add(lowImportanceLabel("Detection stays local and does not change the pipeline until you choose a suggestion."))
	dg.addSuggestions.Add(lowImportanceLabel("Analyze the current result to rank likely decoders and formatters."))

	return container.NewVBox(dg.addSuggestions, catalog)
}

func transformerCategoryLabels() []string {
	labels := []string{allTransformerCategories, favoriteTransformerCategory, recentTransformerCategory}
	for _, category := range plugins.PluginCategories {
		labels = append(labels, plugins.CategoryLabel(category))
	}
	return labels
}

func normalizeAddCategory(label string) string {
	for _, candidate := range transformerCategoryLabels() {
		if label == candidate {
			return label
		}
	}
	return allTransformerCategories
}

func transformerCategoryID(label string) string {
	if label == "" || label == allTransformerCategories {
		return ""
	}
	for _, category := range plugins.PluginCategories {
		if plugins.CategoryLabel(category) == label {
			return category
		}
	}
	return ""
}

func filterTransformerCatalog(query, categoryLabel string, favorites, recents []string) []plugins.UIPluginInfo {
	category := transformerCategoryID(categoryLabel)
	matches := plugins.SearchUICatalog(query)
	if categoryLabel == favoriteTransformerCategory || categoryLabel == recentTransformerCategory {
		orderedNames := favorites
		if categoryLabel == recentTransformerCategory {
			orderedNames = recents
		}
		byName := make(map[string]plugins.UIPluginInfo, len(matches))
		for _, info := range matches {
			byName[info.Name] = info
		}
		filtered := make([]plugins.UIPluginInfo, 0, len(orderedNames))
		for _, name := range orderedNames {
			if info, ok := byName[name]; ok {
				filtered = append(filtered, info)
			}
		}
		return filtered
	}
	if category == "" {
		return matches
	}
	filtered := make([]plugins.UIPluginInfo, 0, len(matches))
	for _, info := range matches {
		if info.Category == category {
			filtered = append(filtered, info)
		}
	}
	return filtered
}

func (catalog *transformerCatalog) refresh() {
	if catalog.results == nil || catalog.detail == nil {
		return
	}
	query := catalog.search.Text
	categoryLabel := catalog.category.Selected
	catalog.matches = filterTransformerCatalog(query, categoryLabel, catalog.gui.favoriteTransformers, catalog.gui.recentTransformers)
	if catalog.resultSummary != nil {
		catalog.resultSummary.SetText(transformerResultSummary(len(catalog.matches)))
	}
	catalog.results.UnselectAll()
	for id := range catalog.matches {
		catalog.results.SetItemHeight(widget.ListItemID(id), 72)
	}
	catalog.results.Refresh()
	if len(catalog.matches) == 0 {
		catalog.selected = -1
		catalog.forgetDetailControls()
		catalog.detail.RemoveAll()
		catalog.detail.Add(wrappingLabel(emptyTransformerCatalogMessage(query, categoryLabel)))
		catalog.detail.Refresh()
		return
	}
	catalog.results.Select(0)
	catalog.results.ScrollToTop()
}

func transformerResultSummary(count int) string {
	if count == 1 {
		return "1 transformer"
	}
	return fmt.Sprintf("%d transformers", count)
}

func emptyTransformerCatalogMessage(query, categoryLabel string) string {
	switch {
	case categoryLabel == favoriteTransformerCategory && strings.TrimSpace(query) == "":
		return "No favorites yet. Mark useful transformers with ☆ Favorite to keep them close."
	case categoryLabel == recentTransformerCategory && strings.TrimSpace(query) == "":
		return "No recent transformers yet. Transformers you add will appear here automatically."
	default:
		return "No transformers match this search and filter."
	}
}

func (catalog *transformerCatalog) moveSelection(delta int) {
	if delta == 0 || catalog.results == nil || len(catalog.matches) == 0 {
		return
	}
	next := catalog.selected + delta
	if catalog.selected < 0 {
		if delta < 0 {
			next = len(catalog.matches) - 1
		} else {
			next = 0
		}
	}
	if next < 0 {
		next = 0
	}
	if next >= len(catalog.matches) {
		next = len(catalog.matches) - 1
	}
	if next == catalog.selected {
		return
	}
	catalog.results.Select(widget.ListItemID(next))
	catalog.results.ScrollTo(widget.ListItemID(next))
}

func (catalog *transformerCatalog) renderDetail(info plugins.UIPluginInfo) {
	dg := catalog.gui
	catalog.forgetDetailControls()
	catalog.detail.RemoveAll()
	title := widget.NewLabelWithStyle(info.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	category := lowImportanceLabel(plugins.CategoryLabel(info.Category) + "  •  command: " + info.Name)
	catalog.detail.Add(title)
	catalog.detail.Add(category)
	if info.Description != "" {
		catalog.detail.Add(wrappingLabel(info.Description))
	}
	if info.UseFor != "" {
		catalog.detail.Add(widget.NewSeparator())
		catalog.detail.Add(widget.NewLabelWithStyle("Best used for", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		catalog.detail.Add(wrappingLabel(info.UseFor))
	}
	var metadata []string
	if len(info.Aliases) > 0 {
		metadata = append(metadata, "aliases: "+strings.Join(info.Aliases, ", "))
	}
	if count := len(pipeline.PluginOptions(info.Name)); count > 0 {
		metadata = append(metadata, fmt.Sprintf("%d configurable options", count))
	}
	if len(metadata) > 0 {
		catalog.detail.Add(lowImportanceLabel(strings.Join(metadata, "  •  ")))
	}

	addEncode := widget.NewButtonWithIcon("Add encode", theme.ContentAddIcon(), func() { dg.appendTransformer(info, false) })
	addEncode.Importance = widget.HighImportance
	favoriteLabel := "☆ Favorite"
	if containsTransformer(dg.favoriteTransformers, info.Name) {
		favoriteLabel = "★ Favorited"
	}
	favorite := widget.NewButton(favoriteLabel, func() { dg.toggleFavoriteTransformer(info) })
	favorite.Importance = widget.LowImportance
	actions := container.NewHBox(addEncode)
	catalog.registerDetailControl(addEncode)
	if info.CanDecode {
		addDecode := widget.NewButtonWithIcon("Add decode", theme.NavigateBackIcon(), func() { dg.appendTransformer(info, true) })
		actions.Add(addDecode)
		catalog.registerDetailControl(addDecode)
	}
	actions.Add(favorite)
	catalog.registerDetailControl(favorite)
	catalog.detail.Add(actions)

	if len(info.Examples) > 0 {
		catalog.detail.Add(widget.NewSeparator())
		catalog.detail.Add(widget.NewLabelWithStyle("Example", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		example := info.Examples[0]
		catalog.detail.Add(wrappingLabel(example.Label + "\nInput: " + example.Input + "\nOutput: " + example.Output))
	}
	if len(info.References) > 0 {
		catalog.detail.Add(widget.NewSeparator())
		catalog.detail.Add(widget.NewLabelWithStyle("References", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		for _, reference := range info.References {
			if link := catalogReferenceLink(reference); link != nil {
				catalog.detail.Add(link)
			}
		}
	}
	catalog.detail.Refresh()
}

func (catalog *transformerCatalog) registerControl(control fyne.Disableable) {
	catalog.controls = append(catalog.controls, control)
	if catalog.gui.working {
		control.Disable()
	}
}

func (catalog *transformerCatalog) registerDetailControl(control fyne.Disableable) {
	catalog.detailControls = append(catalog.detailControls, control)
	catalog.registerControl(control)
}

func (catalog *transformerCatalog) forgetDetailControls() {
	if len(catalog.detailControls) == 0 || len(catalog.controls) == 0 {
		catalog.detailControls = nil
		return
	}
	stale := make(map[fyne.Disableable]bool, len(catalog.detailControls))
	for _, control := range catalog.detailControls {
		stale[control] = true
	}
	retained := catalog.controls[:0]
	for _, control := range catalog.controls {
		if !stale[control] {
			retained = append(retained, control)
		}
	}
	catalog.controls = retained
	catalog.detailControls = nil
}

func (catalog *transformerCatalog) setDisabled(disabled bool) {
	for _, control := range catalog.controls {
		if disabled {
			control.Disable()
		} else {
			control.Enable()
		}
	}
}

func (dg *DeenGUI) appendTransformer(info plugins.UIPluginInfo, decode bool) {
	if dg.working {
		return
	}
	dg.runPipelineWork("Adding transformer", func() error {
		dg.pipe.AddStep(info.Name, decode)
		return nil
	}, func() {
		dg.recordRecentTransformer(info.Name)
		dg.selectedStage = dg.pipe.Len() - 1
		dg.rebuild()
		dg.selectTab(0)
	})
}

func (dg *DeenGUI) loadInlineSuggestions() {
	if dg.working || dg.addSuggestions == nil {
		return
	}
	var suggestions []pipeline.Suggestion
	dg.runPipelineWork("Detecting transforms", func() error {
		suggestions = pipeline.Suggestions(dg.pipe.Result())
		return nil
	}, func() {
		if dg.selectedStage != pipelineStageAdd || dg.addSuggestions == nil {
			return
		}
		dg.renderInlineSuggestions(suggestions)
	})
}

func (dg *DeenGUI) renderInlineSuggestions(suggestions []pipeline.Suggestion) {
	dg.addSuggestions.RemoveAll()
	heading := widget.NewLabelWithStyle("Suggested next transforms", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	dg.addSuggestions.Add(heading)
	if len(suggestions) == 0 {
		dg.addSuggestions.Add(lowImportanceLabel("No likely transforms were detected for the current result."))
		dg.addSuggestions.Refresh()
		return
	}
	sort.SliceStable(suggestions, func(i, j int) bool { return suggestions[i].Confidence > suggestions[j].Confidence })
	limit := len(suggestions)
	if limit > 3 {
		limit = 3
	}
	for _, suggestion := range suggestions[:limit] {
		suggestion := suggestion
		detail := suggestion.Reason
		if suggestion.Confidence > 0 {
			detail += fmt.Sprintf("  •  %d%% confidence", suggestion.Confidence)
		}
		add := widget.NewButton("Add", func() { dg.appendSuggestion(suggestion) })
		dg.registerWorkControl(add)
		body := container.NewVBox(
			widget.NewLabelWithStyle(suggestion.Label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			wrappingLabel(detail),
		)
		if suggestion.Preview != "" {
			body.Add(lowImportanceLabel("Preview: " + guiSafeSuggestionPreview(suggestion.Preview)))
		}
		dg.addSuggestions.Add(widget.NewCard("", "", container.NewBorder(nil, nil, nil, add, body)))
	}
	dg.addSuggestions.Refresh()
}

func (dg *DeenGUI) appendSuggestion(suggestion pipeline.Suggestion) {
	if dg.working {
		return
	}
	dg.runPipelineWork("Adding suggestion", func() error {
		dg.pipe.AddSuggestion(suggestion)
		return nil
	}, func() {
		for _, step := range suggestion.Steps {
			dg.recordRecentTransformer(step.Plugin)
		}
		if len(suggestion.Steps) == 0 {
			dg.recordRecentTransformer(suggestion.Plugin)
		}
		dg.selectedStage = dg.pipe.Len() - 1
		dg.rebuild()
		dg.selectTab(0)
	})
}

func sanitizeTransformerHistory(names []string, limit int) []string {
	if limit <= 0 {
		return nil
	}
	result := make([]string, 0, min(len(names), limit))
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if seen[name] || plugins.CategoryOf(name) == "" {
			continue
		}
		seen[name] = true
		result = append(result, name)
		if len(result) == limit {
			break
		}
	}
	return result
}

func containsTransformer(names []string, name string) bool {
	for _, candidate := range names {
		if candidate == name {
			return true
		}
	}
	return false
}

func prependTransformer(names []string, name string, limit int) []string {
	return sanitizeTransformerHistory(append([]string{name}, names...), limit)
}

func (dg *DeenGUI) toggleFavoriteTransformer(info plugins.UIPluginInfo) {
	favorites := make([]string, 0, len(dg.favoriteTransformers)+1)
	if containsTransformer(dg.favoriteTransformers, info.Name) {
		for _, name := range dg.favoriteTransformers {
			if name != info.Name {
				favorites = append(favorites, name)
			}
		}
	} else {
		favorites = prependTransformer(dg.favoriteTransformers, info.Name, maxFavoriteTransformers)
	}
	dg.favoriteTransformers = favorites
	if dg.app != nil {
		dg.app.Preferences().SetStringList(favoriteTransformersPreferenceKey, favorites)
	}
	for _, catalog := range dg.transformerCatalogs() {
		if catalog.category.Selected == favoriteTransformerCategory {
			catalog.refresh()
			continue
		}
		if catalog.selected >= 0 && catalog.selected < len(catalog.matches) && catalog.matches[catalog.selected].Name == info.Name {
			catalog.renderDetail(info)
		}
	}
}

func (dg *DeenGUI) recordRecentTransformer(name string) {
	if plugins.CategoryOf(name) == "" {
		return
	}
	dg.recentTransformers = prependTransformer(dg.recentTransformers, name, maxRecentTransformers)
	if dg.app != nil {
		dg.app.Preferences().SetStringList(recentTransformersPreferenceKey, dg.recentTransformers)
	}
	for _, catalog := range dg.transformerCatalogs() {
		if catalog.category.Selected == recentTransformerCategory {
			catalog.refresh()
		}
	}
}

func (dg *DeenGUI) transformerCatalogs() []*transformerCatalog {
	catalogs := make([]*transformerCatalog, 0, 2)
	if dg.addCatalog != nil {
		catalogs = append(catalogs, dg.addCatalog)
	}
	if dg.browserCatalog != nil && dg.browserCatalog != dg.addCatalog {
		catalogs = append(catalogs, dg.browserCatalog)
	}
	return catalogs
}

func (dg *DeenGUI) setTransformerCatalogsDisabled(disabled bool) {
	for _, catalog := range dg.transformerCatalogs() {
		catalog.setDisabled(disabled)
	}
}

func lowImportanceLabel(text string) *widget.Label {
	label := wrappingLabel(text)
	label.Importance = widget.LowImportance
	return label
}

func wrappingLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	return label
}

func catalogReferenceLink(reference plugins.Reference) *widget.Hyperlink {
	u, err := url.Parse(reference.URL)
	if err != nil {
		return nil
	}
	return widget.NewHyperlink(reference.Label, u)
}
