//go:build gui

package gui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2/container"
	fynetest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/takeshixx/deen/internal/pipeline"
)

var guiBenchmarkSink any

func newBenchmarkGUI() *DeenGUI {
	dg := &DeenGUI{
		app:           fynetest.NewApp(),
		pipe:          pipeline.New(),
		stepsBox:      container.NewVBox(),
		selectedStage: pipelineStageInput,
		activeTab:     -1,
		tabContent:    container.NewMax(),
		workStatus:    widget.NewLabel(""),
	}
	dg.app.Settings().SetTheme(newAdversecTheme(theme.VariantDark))
	dg.navigationPanel = dg.navigationSidebar()
	return dg
}

func addAlternatingBase64Steps(p *pipeline.Pipeline, count int) {
	for i := 0; i < count; i++ {
		p.AddStep("base64", i%2 == 1)
	}
}

func BenchmarkGUISelectTabs(b *testing.B) {
	dg := newBenchmarkGUI()
	dg.pipe.SetSource([]byte("deen"))
	addAlternatingBase64Steps(dg.pipe, 4)
	dg.rebuild()
	dg.selectTab(1)
	dg.selectTab(2)
	dg.selectTab(3)
	dg.selectTab(0)
	b.ResetTimer()

	b.ReportAllocs()
	for b.Loop() {
		dg.selectTab(1)
		dg.selectTab(2)
		dg.selectTab(3)
		dg.selectTab(0)
	}
	guiBenchmarkSink = dg.tabContent
}

func BenchmarkGUIRebuildEmptyPipeline(b *testing.B) {
	dg := newBenchmarkGUI()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		dg.rebuild()
	}
	guiBenchmarkSink = dg.stepsBox
}

func BenchmarkGUIRebuildEightSteps(b *testing.B) {
	dg := newBenchmarkGUI()
	dg.pipe.SetSource([]byte(strings.Repeat("deen", 256)))
	addAlternatingBase64Steps(dg.pipe, 8)
	dg.selectedStage = 7

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		dg.rebuild()
	}
	guiBenchmarkSink = dg.stepsBox
}

func BenchmarkGUIRefreshFromLargeSource(b *testing.B) {
	dg := newBenchmarkGUI()
	dg.pipe.SetSource([]byte(strings.Repeat("deen", (pipeline.LargeDataThreshold/4)+1)))
	addAlternatingBase64Steps(dg.pipe, 4)
	dg.selectedStage = 3
	dg.rebuild()
	b.ResetTimer()

	b.ReportAllocs()
	for b.Loop() {
		dg.refreshFrom(0)
	}
	guiBenchmarkSink = dg.cards
}
