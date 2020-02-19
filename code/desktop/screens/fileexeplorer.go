package screens

import (
	"cloud/datastore"
	"cloud/network"
	"fyne.io/fyne"
	fdialog "fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/sqweek/dialog"
	"os"
	"strconv"
	"strings"
)

func FileExplorerScreen(w fyne.Window, c network.Cloud) fyne.CanvasObject {
	list := widget.NewVBox()
	list2 := widget.NewVBox()
	list2.Append(widget.NewButton("Testing 1", func() {}))
	list2.Append(widget.NewButton("Testing 2", func() {}))
	list2.Append(widget.NewButton("Testing 3", func() {}))
	list2.Append(widget.NewButton("Testing 4", func() {}))
	list2.Append(widget.NewButton("Testing 5", func() {}))
	scroll := widget.NewScrollContainer(list)
	//scroll2 := widget.NewScrollContainer(list2)
	updateList := func() {
		list.Children = []fyne.CanvasObject{}
		network := c.Network()
		for _, f := range network.DataStore.Files {
			list.Append(widget.NewLabel("File: " + string(f.ID)))
			list.Append(widget.NewLabel("Path: " + f.Path))
			list.Append(widget.NewLabel("Size: " + strconv.Itoa(int(f.Size))))

			for i, chunk := range f.Chunks.Chunks {
				nodes := network.ChunkNodes[chunk.ID]
				var nodeOwners []string

				for j := range nodes {
					n, found := c.NodeByID(nodes[j])
					if found {
						nodeOwners = append(nodeOwners, n.Name)
					}
				}

				list.Append(widget.NewLabel("Chunk " + strconv.Itoa(i) + ": " + strings.Join(nodeOwners, ", ")))
			}
			list.Append(widget.NewLabel(" "))
		}
		list.Refresh()
		scroll.Refresh()
	}
	addFile := func() {
		filename, err := dialog.File().Load()
		if err != nil {
			fdialog.ShowError(err, w)
			return
		}
		reader, err := os.Open(filename)
		if err != nil {
			fdialog.ShowError(err, w)
			return
		}
		f, err := datastore.NewFile(reader, filename, 1024*32)
		if err != nil {
			fdialog.ShowError(err, w)
			return
		}

		err = c.AddFile(f)
		if err != nil {
			fdialog.ShowError(err, w)
		}
		c.DistributeFile(f)
		updateList()
	}
	updateList()
	hbox := widget.NewHBox(
		widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
			updateList()
		}),
		widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), addFile))

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, hbox, nil, nil), hbox, scroll)

	//return fyne.NewContainerWithLayout(layout.NewGridLayoutWithRows(3), scroll, scroll2,
	//	widget.NewHBox(
	//		widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
	//			updateList()
	//		}),
	//		widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), addFile),
	//	))

	//return widget.NewVBox(
	//	fyne.NewContainerWithLayout(layout.NewBorderLayout(scroll, nil, nil, nil), scroll, scroll2),
	//	widget.NewHBox(
	//		widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
	//			updateList()
	//		}),
	//		widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), addFile),
	//	),
	//)
}
