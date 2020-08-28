package gui

import (
	"log"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/theme"
)

// RunGUI creates a GUI instance
func RunGUI() {
	app := app.NewWithID("io.deen.app")

	w := app.NewWindow("deen")
	dg, err := NewDeenGUI(app, w)
	if err != nil {
		log.Fatal(err)
	}
	w.SetMainMenu(
		fyne.NewMainMenu(
			fyne.NewMenu("File",
				fyne.NewMenuItem("Open", func() {
					fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
						if err == nil && reader == nil {
							return
						}
						if err != nil {
							dialog.ShowError(err, w)
							return
						}

						dg.fileOpened(reader)
					}, w)
					fd.Show()
				}),
				// A quit item will be appended to our first menu
			),
			fyne.NewMenu("Theme",
				fyne.NewMenuItem("Light", func() {
					app.Settings().SetTheme(theme.LightTheme())
					app.Preferences().SetString("theme", "light")
				}),
				fyne.NewMenuItem("Dark", func() {
					app.Settings().SetTheme(theme.DarkTheme())
					app.Preferences().SetString("theme", "dark")
				}),
			),
			fyne.NewMenu("Help",
				fyne.NewMenuItem("About", func() {
					dialog.ShowInformation("About", "deen is a DEcoding/ENcoding application that processes arbitrary input data with a wide range of plugins.", w)
				}),
			)))
	if app.Preferences().String("theme") == "light" {
		app.Settings().SetTheme(theme.LightTheme())
	} else {
		app.Settings().SetTheme(theme.DarkTheme())
	}
	w.SetMaster()
	w.SetContent(dg.Layout)
	w.Resize(fyne.NewSize(640, 480))
	w.ShowAndRun()
}
