package screens

import (
	"cloud/network"
	"fyne.io/fyne"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func EmptyScreen(w fyne.Window, c *network.Cloud) fyne.CanvasObject {
	return widget.NewHBox()
}

func Navigation(w fyne.Window, c network.Cloud) {
	RegisterEvents(c)
	tabs := widget.NewTabContainer(
		widget.NewTabItemWithIcon("Home", theme.HomeIcon(), HomeScreen(w, c)),
		widget.NewTabItemWithIcon("Nodes", theme.ContentCopyIcon(), NodesScreen(w, c)),
		widget.NewTabItemWithIcon("Whitelist", theme.ContentClearIcon(), WhitelistScreen(w, c)),
		widget.NewTabItemWithIcon("File Explorer", theme.FolderIcon(), FileExplorerScreen(w, c)),
		widget.NewTabItemWithIcon("Settings", theme.SettingsIcon(), SettingScreen(w, c)),
	)

	tabs.SetTabLocation(widget.TabLocationLeading)
	w.SetContent(tabs)
}
