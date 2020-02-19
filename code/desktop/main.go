package main

import (
	"cloud/desktop/screens"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/theme"
)

func main() {
	app := app.NewWithID("cloud.desktop")
	app.Settings().SetTheme(theme.DarkTheme())

	w := app.NewWindow("The Cloud")
	app.Settings().SetTheme(theme.DarkTheme())
	w.SetMaster()
	//w.SetContent(widget.NewVBox(
	//	widget.NewLabel("Welcome to the Cloud!"),
	//	widget.NewButton("Quit", func() {
	//		app.Quit()
	//	}),
	//))

	w.SetContent(screens.NewCloudScreen(w))
	w.Resize(fyne.NewSize(640, 280))
	app.Settings().SetTheme(theme.DarkTheme())
	w.ShowAndRun()
}
