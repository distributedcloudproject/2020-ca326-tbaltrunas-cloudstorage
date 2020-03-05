package screens

import (
	"cloud/network"
	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func WhitelistScreen(w fyne.Window, c network.Cloud) fyne.CanvasObject {
	addToWhitelist := widget.NewEntry()
	addToWhitelist.SetPlaceHolder("Enter the whitelist ID")

	list := widget.NewVBox()
	wl := c.Whitelist()
	wlEntry := make(map[string]fyne.CanvasObject)

	for i := range wl {
		e := widget.NewEntry()
		e.SetText(wl[i])
		e.Disable()

		wlEntry[wl[i]] = e

		list.Append(e)
	}

	c.Events().WhitelistAdded = func(ID string) {
		e := widget.NewEntry()
		e.SetText(ID)
		e.Disable()

		wlEntry[ID] = e

		list.Append(e)
	}
	c.Events().WhitelistRemoved = func(ID string) {
		e, ok := wlEntry[ID]
		if ok {
			for i := 0; i < len(list.Children); i++ {
				if list.Children[i] == e {
					list.Children = append(list.Children[:i], list.Children[i+1:]...)
					list.Refresh()
				}
			}
		}
	}

	return widget.NewVBox(
		list,
		addToWhitelist,
		fyne.NewContainerWithLayout(layout.NewCenterLayout(), widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), func() {
			err := c.AddToWhitelist(addToWhitelist.Text)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			addToWhitelist.SetText("")
		})),
	)
}
