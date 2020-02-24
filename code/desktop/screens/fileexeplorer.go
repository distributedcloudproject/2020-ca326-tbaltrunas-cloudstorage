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
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func FileListScreen(w fyne.Window, c network.Cloud) fyne.CanvasObject {
	list := widget.NewVBox()
	scroll := widget.NewScrollContainer(list)
	updateList := func() {
		list.Children = []fyne.CanvasObject{}

		nw := c.Network()
		for _, f := range nw.DataStore.Files {
			list.Append(widget.NewLabel("File: " + string(f.ID)))
			list.Append(widget.NewLabel("Path: " + f.Name))
			list.Append(widget.NewLabel("Size: " + strconv.Itoa(int(f.Size))))

			for i, chunk := range f.Chunks.Chunks {
				nodes := nw.ChunkNodes[chunk.ID]
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

		err = c.AddFile(f, path.Base(filename))
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

func FileExplorerScreen(w fyne.Window, c network.Cloud) fyne.CanvasObject {
	list := widget.NewVBox()
	scroll := widget.NewScrollContainer(list)
	folderPath := "/"
	var redraw func()
	updateList := func() {
		list.Children = []fyne.CanvasObject{}
		list.Append(widget.NewLabel(folderPath))
		if folderPath != "/" && folderPath != "" {
			list.Append(widget.NewHBox(widget.NewButtonWithIcon("..", theme.FolderIcon(), func() {
				folderPaths := strings.Split(folderPath, "/")
				folderPath = "/" + path.Join(folderPaths[:len(folderPaths)-1]...)
				redraw()
			})))
		}

		folder, _ := c.GetFolder(folderPath)
		for i := range folder.SubFolders {
			p := folder.SubFolders[i].Name
			list.Append(widget.NewHBox(widget.NewButtonWithIcon(p, theme.FolderIcon(), func() {
				folderPath = path.Join(folderPath, p)
				redraw()
			})))
		}

		for i := range folder.Files.Files {
			file := folder.Files.Files[i]
			list.Append(widget.NewHBox(widget.NewButtonWithIcon(path.Base(file.Name), theme.ContentPasteIcon(), func() {
				// Nothing
			})))
		}

		list.Refresh()
		scroll.Refresh()
	}
	// Temp hack
	redraw = updateList
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

		_, fname := filepath.Split(filename)
		f, err := datastore.NewFile(reader, fname, 1024*32)
		if err != nil {
			fdialog.ShowError(err, w)
			return
		}

		err = c.AddFile(f, path.Join(folderPath, fname))
		if err != nil {
			fdialog.ShowError(err, w)
		}
		// c.DistributeFile(f)
		updateList()
	}
	addFolder := func() {
		content := widget.NewEntry()
		content.SetPlaceHolder("folder")
		fdialog.ShowCustomConfirm("Enter folder name", "Create", "Cancel", content, func(bool) {
			c.GetFolder(folderPath + "/" + content.Text)
			updateList()
		}, w)
	}
	updateList()
	hbox := widget.NewHBox(
		widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
			updateList()
		}),
		widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), addFile),
		widget.NewButtonWithIcon("Add Folder", theme.ContentAddIcon(), addFolder))

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, hbox, nil, nil), hbox, scroll)
}
