//go:build gui

package gui

import (
	"bytes"
	"fmt"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/takeshixx/deen/internal/pipeline"
	"github.com/takeshixx/deen/pkg/formatters"
)

type urlPartsEditor struct {
	card *stepCard

	view       fyne.CanvasObject
	scroll     *container.Scroll
	content    *fyne.Container
	message    *widget.Label
	doc        *formatters.URLPartsDocument
	lastJSON   string
	showRaw    bool
	refreshing bool

	schemeEntry    *widget.Entry
	hostnameEntry  *widget.Entry
	portEntry      *widget.Entry
	opaqueEntry    *widget.Entry
	fragmentEntry  *widget.Entry
	usernameEntry  *widget.Entry
	passwordEntry  *widget.Entry
	userinfoCheck  *widget.Check
	forceQuery     *widget.Check
	omitHost       *widget.Check
	rawCheck       *widget.Check
	rebuiltEntry   *widget.Entry
	copyURLButton  *widget.Button
	analysisBox    *fyne.Container
	defangedEntry  *widget.Entry
	copyDefanged   *widget.Button
	analysisLabels []*widget.Label

	pathRows          *fyne.Container
	queryRows         *fyne.Container
	addPathButton     *widget.Button
	addQueryButton    *widget.Button
	pathEntries       []*widget.Entry
	queryKeyEntries   []*widget.Entry
	queryValueEntries []*widget.Entry
	staticControls    []fyne.Disableable
	pathControls      []fyne.Disableable
	queryControls     []fyne.Disableable
	analysisControls  []fyne.Disableable
}

func newURLPartsEditor(card *stepCard, data []byte) *urlPartsEditor {
	e := &urlPartsEditor{
		card:    card,
		content: container.NewVBox(),
		message: widget.NewLabel(""),
	}
	e.message.Wrapping = fyne.TextWrapWord
	e.message.Importance = widget.DangerImportance
	e.scroll = container.NewVScroll(e.content)
	e.view = container.New(fixedHeightLayout{height: 240}, e.scroll)
	e.refresh(data)
	return e
}

func (e *urlPartsEditor) refresh(data []byte) {
	if e == nil || e.refreshing || string(data) == e.lastJSON {
		return
	}
	e.refreshing = true
	defer func() { e.refreshing = false }()
	doc, err := formatters.DecodeURLPartsJSON(bytes.NewReader(data))
	e.clearControls()
	e.content.RemoveAll()
	if err != nil {
		e.doc = nil
		e.lastJSON = string(data)
		e.message.SetText("Structured editor unavailable: " + err.Error())
		e.content.Add(container.NewPadded(e.message))
		e.content.Refresh()
		return
	}
	e.doc = doc
	e.lastJSON = string(data)
	e.build()
}

func (e *urlPartsEditor) build() {
	e.clearControls()
	e.content.RemoveAll()
	e.message.SetText("Edits update this step's JSON output and all downstream steps.")
	e.message.Importance = widget.LowImportance
	e.content.Add(e.message)
	e.rebuiltEntry = multilineEntry(2)
	e.rebuiltEntry.Disable()
	e.copyURLButton = widget.NewButtonWithIcon("Copy URL", theme.ContentCopyIcon(), func() {
		if e.rebuiltEntry.Text == "" || e.card.gui.window == nil {
			return
		}
		e.card.gui.window.Clipboard().SetContent(e.rebuiltEntry.Text)
		e.card.gui.showActionFeedback("Rebuilt URL copied")
	})
	e.copyURLButton.Importance = widget.LowImportance
	e.register(e.copyURLButton)
	e.content.Add(widget.NewCard("Rebuilt URL", "Preview generated locally from the fields below.", container.NewBorder(nil, nil, nil, e.copyURLButton, e.rebuiltEntry)))
	e.refreshRebuiltURL()
	e.content.Add(urlPartsSectionTitle("Local analysis"))
	e.analysisBox = container.NewVBox()
	e.content.Add(e.analysisBox)
	e.rebuildAnalysis()

	e.rawCheck = widget.NewCheck("Show original encoded values", nil)
	e.rawCheck.SetChecked(e.showRaw)
	e.rawCheck.OnChanged = func(show bool) {
		e.showRaw = show
		e.rebuildPathRows()
		e.rebuildQueryRows()
	}
	e.register(e.rawCheck)
	e.content.Add(e.rawCheck)

	e.content.Add(urlPartsSectionTitle("Authority"))
	e.schemeEntry = e.field("Scheme", e.doc.Scheme, func(value string) { e.doc.Scheme = value })
	e.hostnameEntry = e.field("Hostname", e.doc.Hostname, func(value string) { e.doc.Hostname = value })
	e.portEntry = e.field("Port", e.doc.Port, func(value string) { e.doc.Port = value })
	e.content.Add(urlPartsFieldGrid(
		"Scheme", e.schemeEntry,
		"Hostname", e.hostnameEntry,
		"Port", e.portEntry,
	))

	e.userinfoCheck = widget.NewCheck("URL contains user information", nil)
	e.userinfoCheck.SetChecked(e.doc.Userinfo != nil)
	e.userinfoCheck.OnChanged = func(enabled bool) {
		if e.refreshing {
			return
		}
		if enabled {
			if e.doc.Userinfo == nil {
				e.doc.Userinfo = &formatters.URLPartsUserinfo{}
			}
		} else {
			e.doc.Userinfo = nil
		}
		e.commit()
		e.build()
	}
	e.register(e.userinfoCheck)
	e.content.Add(e.userinfoCheck)
	if e.doc.Userinfo != nil {
		e.usernameEntry = e.field("Username", e.doc.Userinfo.Username, func(value string) {
			e.doc.Userinfo.Username = value
		})
		e.passwordEntry = widget.NewPasswordEntry()
		e.passwordEntry.SetText(e.doc.Userinfo.Password)
		e.passwordEntry.OnChanged = func(value string) {
			if e.ignoreChange() || e.doc.Userinfo == nil {
				return
			}
			e.doc.Userinfo.Password = value
			e.doc.Userinfo.PasswordSet = true
			e.commit()
		}
		e.register(e.passwordEntry)
		e.content.Add(urlPartsFieldGrid("Username", e.usernameEntry, "Password", e.passwordEntry))
	}

	e.content.Add(urlPartsSectionTitle("Path"))
	e.pathRows = container.NewVBox()
	e.content.Add(e.pathRows)
	e.rebuildPathRows()
	e.addPathButton = widget.NewButtonWithIcon("Add path segment", theme.ContentAddIcon(), func() {
		e.doc.PathSegments = append(e.doc.PathSegments, "")
		e.commit()
		e.rebuildPathRows()
	})
	e.addPathButton.Importance = widget.LowImportance
	e.register(e.addPathButton)
	e.content.Add(e.addPathButton)

	e.content.Add(urlPartsSectionTitle("Query parameters"))
	e.forceQuery = widget.NewCheck("Keep ? when the query is empty", nil)
	e.forceQuery.SetChecked(e.doc.ForceQuery)
	e.forceQuery.OnChanged = func(value bool) {
		if e.ignoreChange() {
			return
		}
		e.doc.ForceQuery = value
		e.commit()
	}
	e.register(e.forceQuery)
	e.content.Add(e.forceQuery)
	e.queryRows = container.NewVBox()
	e.content.Add(e.queryRows)
	e.rebuildQueryRows()
	e.addQueryButton = widget.NewButtonWithIcon("Add query parameter", theme.ContentAddIcon(), func() {
		e.doc.Query = append(e.doc.Query, formatters.URLQueryParameter{HasValue: true})
		e.commit()
		e.rebuildQueryRows()
	})
	e.addQueryButton.Importance = widget.LowImportance
	e.register(e.addQueryButton)
	e.content.Add(e.addQueryButton)

	e.content.Add(urlPartsSectionTitle("Fragment and advanced fields"))
	e.fragmentEntry = e.field("Fragment", e.doc.Fragment, func(value string) { e.doc.Fragment = value })
	e.opaqueEntry = e.field("Opaque data", e.doc.Opaque, func(value string) { e.doc.Opaque = value })
	e.content.Add(urlPartsFieldGrid("Fragment", e.fragmentEntry, "Opaque data", e.opaqueEntry))
	e.omitHost = widget.NewCheck("Host is omitted", nil)
	e.omitHost.SetChecked(e.doc.OmitHost)
	e.omitHost.OnChanged = func(value bool) {
		if e.ignoreChange() {
			return
		}
		e.doc.OmitHost = value
		e.commit()
	}
	e.register(e.omitHost)
	e.content.Add(e.omitHost)
	e.content.Refresh()
	e.scroll.Refresh()
}

func (e *urlPartsEditor) field(_ string, value string, apply func(string)) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetText(value)
	entry.OnChanged = func(value string) {
		if e.ignoreChange() {
			return
		}
		apply(value)
		e.commit()
	}
	e.register(entry)
	return entry
}

func (e *urlPartsEditor) ignoreChange() bool {
	return e.refreshing || e.card.gui.updating || e.card.gui.working || e.doc == nil
}

func (e *urlPartsEditor) commit() {
	if e.doc == nil || e.refreshing {
		return
	}
	var output bytes.Buffer
	if err := formatters.EncodeURLPartsJSON(&output, e.doc); err != nil {
		e.message.Importance = widget.DangerImportance
		e.message.SetText("Could not update URL Parts JSON: " + err.Error())
		return
	}
	data := output.Bytes()
	e.lastJSON = string(data)
	e.card.gui.pipe.EditOutput(e.card.index, data)
	e.card.gui.setText(e.card.body, e.lastJSON)
	e.refreshRebuiltURL()
	e.rebuildAnalysis()
	e.card.meta.SetText(metadataSummary("", pipeline.DataMetadata(data, len(e.card.gui.pipe.Input(e.card.index)))))
	e.card.syncPreviewTab(data)
	if e.card.preview != nil {
		preview, spans, _ := pipeline.HighlightedPreview(data)
		setPreviewText(e.card.preview, preview, spans)
	}
	e.card.gui.refreshFrom(e.card.index + 1)
}

func (e *urlPartsEditor) rebuildPathRows() {
	if e.pathRows == nil || e.doc == nil {
		return
	}
	e.unregisterControls(e.pathControls)
	e.pathControls = nil
	e.pathRows.RemoveAll()
	e.pathEntries = nil
	if len(e.doc.PathSegments) == 0 {
		e.pathRows.Add(urlPartsEmptyLabel("No path segments"))
	}
	for i, value := range e.doc.PathSegments {
		i := i
		entry := widget.NewEntry()
		entry.SetText(value)
		entry.OnChanged = func(value string) {
			if e.ignoreChange() || i >= len(e.doc.PathSegments) {
				return
			}
			e.doc.PathSegments[i] = value
			e.commit()
		}
		e.registerPath(entry)
		e.pathEntries = append(e.pathEntries, entry)
		actions := e.rowActions(e.registerPath, i, len(e.doc.PathSegments), func(delta int) {
			e.doc.PathSegments[i], e.doc.PathSegments[i+delta] = e.doc.PathSegments[i+delta], e.doc.PathSegments[i]
			e.commit()
			e.rebuildPathRows()
		}, func() {
			e.doc.PathSegments = slices.Insert(e.doc.PathSegments, i+1, e.doc.PathSegments[i])
			e.commit()
			e.rebuildPathRows()
		}, func() {
			e.doc.PathSegments = slices.Delete(e.doc.PathSegments, i, i+1)
			e.commit()
			e.rebuildPathRows()
		})
		row := container.NewBorder(nil, nil, widget.NewLabel(fmt.Sprintf("%d", i+1)), actions, entry)
		e.pathRows.Add(row)
	}
	if e.showRaw && e.doc.Raw != nil {
		e.pathRows.Add(urlPartsRawLabel("Encoded path: " + e.doc.Raw.Path))
	}
	e.pathRows.Refresh()
	e.scroll.Refresh()
}

func (e *urlPartsEditor) rebuildQueryRows() {
	if e.queryRows == nil || e.doc == nil {
		return
	}
	e.unregisterControls(e.queryControls)
	e.queryControls = nil
	e.queryRows.RemoveAll()
	e.queryKeyEntries = nil
	e.queryValueEntries = nil
	if len(e.doc.Query) == 0 {
		e.queryRows.Add(urlPartsEmptyLabel("No query parameters"))
	}
	for i := range e.doc.Query {
		i := i
		parameter := &e.doc.Query[i]
		key := widget.NewEntry()
		key.SetPlaceHolder("Key")
		key.SetText(parameter.Key)
		key.OnChanged = func(value string) {
			if e.ignoreChange() || i >= len(e.doc.Query) {
				return
			}
			e.doc.Query[i].Key = value
			e.commit()
		}
		value := widget.NewEntry()
		value.SetPlaceHolder("Value")
		value.SetText(parameter.Value)
		value.OnChanged = func(value string) {
			if e.ignoreChange() || i >= len(e.doc.Query) {
				return
			}
			e.doc.Query[i].Value = value
			e.commit()
		}
		if !parameter.HasValue {
			value.Disable()
		}
		hasValue := widget.NewCheck("has value", nil)
		hasValue.SetChecked(parameter.HasValue)
		hasValue.OnChanged = func(enabled bool) {
			if e.ignoreChange() || i >= len(e.doc.Query) {
				return
			}
			e.doc.Query[i].HasValue = enabled
			if !enabled {
				e.doc.Query[i].Value = ""
			}
			e.commit()
			e.rebuildQueryRows()
		}
		e.registerQuery(key)
		e.registerQuery(value)
		e.registerQuery(hasValue)
		e.queryKeyEntries = append(e.queryKeyEntries, key)
		e.queryValueEntries = append(e.queryValueEntries, value)
		actions := e.rowActions(e.registerQuery, i, len(e.doc.Query), func(delta int) {
			e.doc.Query[i], e.doc.Query[i+delta] = e.doc.Query[i+delta], e.doc.Query[i]
			e.commit()
			e.rebuildQueryRows()
		}, func() {
			e.doc.Query = slices.Insert(e.doc.Query, i+1, e.doc.Query[i])
			e.commit()
			e.rebuildQueryRows()
		}, func() {
			e.doc.Query = slices.Delete(e.doc.Query, i, i+1)
			e.commit()
			e.rebuildQueryRows()
		})
		fields := container.NewGridWithColumns(2, key, value)
		row := container.NewBorder(nil, container.NewHBox(hasValue), widget.NewLabel(fmt.Sprintf("%d", i+1)), actions, fields)
		e.queryRows.Add(row)
		if e.showRaw {
			e.queryRows.Add(urlPartsRawLabel(fmt.Sprintf("Encoded: %s=%s", parameter.RawKey, parameter.RawValue)))
		}
	}
	e.queryRows.Refresh()
	e.scroll.Refresh()
}

func (e *urlPartsEditor) rowActions(register func(fyne.Disableable), index, count int, move func(int), duplicate, remove func()) fyne.CanvasObject {
	up := stepIconButton("Move up", theme.MoveUpIcon(), func() { move(-1) })
	down := stepIconButton("Move down", theme.MoveDownIcon(), func() { move(1) })
	duplicateButton := stepIconButton("Duplicate", theme.ContentCopyIcon(), duplicate)
	deleteButton := stepIconButton("Remove", theme.DeleteIcon(), remove)
	if index == 0 {
		up.Disable()
	}
	if index == count-1 {
		down.Disable()
	}
	register(up)
	register(down)
	register(duplicateButton)
	register(deleteButton)
	return container.NewHBox(up, down, duplicateButton, deleteButton)
}

func (e *urlPartsEditor) refreshRebuiltURL() {
	if e.doc == nil || e.rebuiltEntry == nil {
		return
	}
	rebuilt, err := formatters.RebuildURLParts(e.doc)
	if err != nil {
		e.rebuiltEntry.SetText("")
		e.message.Importance = widget.DangerImportance
		e.message.SetText("URL fields are not valid yet: " + err.Error())
		if e.copyURLButton != nil {
			e.copyURLButton.Disable()
		}
		return
	}
	e.rebuiltEntry.SetText(rebuilt)
	e.message.Importance = widget.LowImportance
	e.message.SetText("Edits update this step's JSON output and all downstream steps.")
	if e.copyURLButton != nil && !e.card.gui.working {
		e.copyURLButton.Enable()
	}
}

func (e *urlPartsEditor) rebuildAnalysis() {
	if e.analysisBox == nil || e.doc == nil {
		return
	}
	e.unregisterControls(e.analysisControls)
	e.analysisControls = nil
	e.analysisLabels = nil
	e.analysisBox.RemoveAll()

	e.defangedEntry = multilineEntry(2)
	e.defangedEntry.Disable()
	e.copyDefanged = widget.NewButtonWithIcon("Copy defanged URL", theme.ContentCopyIcon(), func() {
		if e.doc == nil || e.doc.Analysis == nil || e.doc.Analysis.DefangedURL == "" || e.card.gui.window == nil {
			return
		}
		e.card.gui.window.Clipboard().SetContent(e.doc.Analysis.DefangedURL)
		e.card.gui.showActionFeedback("Defanged URL copied")
	})
	e.copyDefanged.Importance = widget.LowImportance
	e.registerAnalysis(e.copyDefanged)
	if e.doc.Analysis == nil {
		e.copyDefanged.Disable()
		e.analysisBox.Add(widget.NewCard("Defanged URL", "Unavailable while URL fields are invalid.", container.NewBorder(nil, nil, nil, e.copyDefanged, e.defangedEntry)))
		e.analysisBox.Refresh()
		return
	}
	e.defangedEntry.SetText(e.doc.Analysis.DefangedURL)
	e.analysisBox.Add(widget.NewCard("Defanged URL", "Safer to paste into tickets or chat; this does not change the pipeline URL.", container.NewBorder(nil, nil, nil, e.copyDefanged, e.defangedEntry)))

	if len(e.doc.Analysis.Indicators) == 0 {
		e.addAnalysisLabel("No common indicators detected. This is not a safety verdict.", widget.LowImportance)
	} else {
		for _, indicator := range e.doc.Analysis.Indicators {
			importance := widget.LowImportance
			prefix := "Info"
			if indicator.Severity == "warning" {
				importance = widget.WarningImportance
				prefix = "Warning"
			}
			e.addAnalysisLabel(prefix+": "+indicator.Message, importance)
		}
	}
	if len(e.doc.Analysis.NestedURLs) > 0 {
		e.analysisBox.Add(urlPartsSectionTitle("Nested URLs"))
		for _, nested := range e.doc.Analysis.NestedURLs {
			e.addAnalysisLabel(fmt.Sprintf("Query #%d (%s): %s", nested.ParameterIndex, nested.Key, nested.URL), widget.LowImportance)
		}
	}
	if len(e.doc.Analysis.TrackingParameters) > 0 {
		e.analysisBox.Add(urlPartsSectionTitle("Common tracking parameters"))
		for _, tracking := range e.doc.Analysis.TrackingParameters {
			e.addAnalysisLabel(fmt.Sprintf("Query #%d: %s", tracking.ParameterIndex, tracking.Key), widget.LowImportance)
		}
	}
	e.analysisBox.Refresh()
	e.scroll.Refresh()
}

func (e *urlPartsEditor) addAnalysisLabel(text string, importance widget.Importance) {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapBreak
	label.Importance = importance
	e.analysisLabels = append(e.analysisLabels, label)
	e.analysisBox.Add(label)
}

func (e *urlPartsEditor) register(control fyne.Disableable) {
	e.card.gui.registerWorkControl(control)
	e.staticControls = append(e.staticControls, control)
}

func (e *urlPartsEditor) registerPath(control fyne.Disableable) {
	e.card.gui.registerWorkControl(control)
	e.pathControls = append(e.pathControls, control)
}

func (e *urlPartsEditor) registerQuery(control fyne.Disableable) {
	e.card.gui.registerWorkControl(control)
	e.queryControls = append(e.queryControls, control)
}

func (e *urlPartsEditor) registerAnalysis(control fyne.Disableable) {
	e.card.gui.registerWorkControl(control)
	e.analysisControls = append(e.analysisControls, control)
}

func (e *urlPartsEditor) clearControls() {
	e.unregisterControls(e.staticControls)
	e.unregisterControls(e.pathControls)
	e.unregisterControls(e.queryControls)
	e.unregisterControls(e.analysisControls)
	e.staticControls = nil
	e.pathControls = nil
	e.queryControls = nil
	e.analysisControls = nil
}

func (e *urlPartsEditor) unregisterControls(controls []fyne.Disableable) {
	if len(controls) == 0 || e.card == nil || e.card.gui == nil {
		return
	}
	remove := make(map[fyne.Disableable]struct{}, len(controls))
	for _, control := range controls {
		remove[control] = struct{}{}
		delete(e.card.gui.workControlStates, control)
	}
	e.card.gui.workControls = slices.DeleteFunc(e.card.gui.workControls, func(control fyne.Disableable) bool {
		_, found := remove[control]
		return found
	})
}

func urlPartsSectionTitle(text string) *widget.Label {
	label := widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	label.Importance = widget.MediumImportance
	return label
}

func urlPartsFieldGrid(items ...any) fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, 0, len(items))
	for i := 0; i+1 < len(items); i += 2 {
		objects = append(objects, widget.NewLabel(items[i].(string)), items[i+1].(fyne.CanvasObject))
	}
	return container.NewGridWithColumns(2, objects...)
}

func urlPartsEmptyLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Importance = widget.LowImportance
	return label
}

func urlPartsRawLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Importance = widget.LowImportance
	label.TextStyle.Monospace = true
	label.Wrapping = fyne.TextWrapBreak
	return label
}
