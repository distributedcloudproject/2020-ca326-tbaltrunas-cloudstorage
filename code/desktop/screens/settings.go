package screens

import (
	"cloud/network"
	"encoding/gob"
	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	sdialog "github.com/sqweek/dialog"
	"os"
)

func SettingScreen(w fyne.Window, c network.Cloud) fyne.CanvasObject {
	return widget.NewVBox(
		widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
			filename, err := sdialog.File().Save()
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			f, err := os.Create(filename)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			defer f.Close()
			encoder := gob.NewEncoder(f)
			err = encoder.Encode(c.SavedNetworkState())
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
		}),
	)
}
