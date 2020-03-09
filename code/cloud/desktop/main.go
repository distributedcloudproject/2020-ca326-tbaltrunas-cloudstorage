// +build desktop

package main

import (
	"cloud/desktop/screens"
	"cloud/utils"
	"flag"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/theme"
	"github.com/sqweek/dialog"
)

func main() {
	utils.NewLoggerFromLevel("INFO")
	dialogFlag := flag.String("dialog", "", "")
	flag.Parse()
	if *dialogFlag == "file-load" {
		f, err := dialog.File().Load()
		if err != nil {
			fmt.Printf("!%v", err)
			return
		}
		fmt.Printf(f)
		return
	}
	if *dialogFlag == "file-save" {
		f, err := dialog.File().Save()
		if err != nil {
			fmt.Printf("!%v", err)
			return
		}
		fmt.Printf(f)
		return
	}
	if *dialogFlag == "dir" {
		f, err := dialog.Directory().Browse()
		if err != nil {
			fmt.Printf("!%v", err)
			return
		}
		fmt.Printf(f)
		return
	}

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
