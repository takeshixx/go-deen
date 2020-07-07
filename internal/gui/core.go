package gui

import (
	"fmt"
	"log"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/theme"
)

// RunGUI creates a GUI instance
func RunGUI() {
	app := app.New()

	w := app.NewWindow("deen")
	w.SetMainMenu(
		fyne.NewMainMenu(
			fyne.NewMenu("File",
				fyne.NewMenuItem("Open", func() { fmt.Println("Menu New") }),
				// A quit item will be appended to our first menu
			),
			fyne.NewMenu("Theme",
				fyne.NewMenuItem("Light", func() {
					app.Settings().SetTheme(theme.LightTheme())
				}),
				fyne.NewMenuItem("Dark", func() {
					app.Settings().SetTheme(theme.DarkTheme())
				}),
			),
			fyne.NewMenu("Help",
				fyne.NewMenuItem("About", func() {
					dialog.ShowInformation("About", "deen is a DEcoding/ENcoding application that processes arbitrary input data with a wide range of plugins.", w)
				}),
			)))
	w.SetMaster()

	dg, err := NewDeenGUI(app, w)
	if err != nil {
		log.Fatal(err)
	}

	w.SetContent(dg.Layout)
	w.Resize(fyne.NewSize(640, 480))
	w.ShowAndRun()
}
