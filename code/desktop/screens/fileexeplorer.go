package screens

import (
	"cloud/datastore"
	"cloud/desktop/resources"
	"cloud/network"
	"errors"
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
			list.Append(NewFileWidget(theme.FolderIcon(), "..", func() {
				folderPaths := strings.Split(folderPath, "/")
				folderPath = "/" + path.Join(folderPaths[:len(folderPaths)-1]...)
				redraw()
			}))
		}

		folder, _ := c.GetFolder(folderPath)
		for i := range folder.SubFolders {
			p := folder.SubFolders[i].Name
			list.Append(NewFileWidget(theme.FolderIcon(), p, func() {
				folderPath = path.Join(folderPath, p)
				redraw()
			}, widget.NewToolbarAction(theme.DeleteIcon(), func() {
				err := c.DeleteDirectory(folderPath + "/" + p)
				if err != nil {
					fdialog.ShowError(err, w)
				}
				redraw()
			})))
		}

		for i := range folder.Files.Files {
			file := folder.Files.Files[i]

			list.Append(NewFileWidget(fileIcon(file.Name), file.Name, nil,
				&toolbarWidget{w: widget.NewLabel("24 KB")},
				widget.NewToolbarSpacer(),
				widget.NewToolbarAction(theme.DeleteIcon(), func() {
					fullpath := folderPath + "/" + file.Name
					if !c.LockFile(fullpath) {
						fdialog.ShowError(errors.New("Could not acquire lock on the file"), w)
						return
					}
					defer c.UnlockFile(fullpath)

					err := c.DeleteFile(folderPath + "/" + file.Name)
					if err != nil {
						fdialog.ShowError(err, w)
					}
					redraw()
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
		fdialog.ShowCustomConfirm("Enter folder name", "Create", "Cancel", content, func(s bool) {
			//c.GetFolder(folderPath + "/" + content.Text)
			if s {
				c.CreateDirectory(folderPath + "/" + content.Text)
				updateList()
			}
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

func fileIcon(filename string) *theme.ThemedResource {
	s := strings.Split(filename, ".")
	if len(s) <= 1 {
		return resources.FileIcons["file"]
	}
	ext := s[len(s)-1]
	t, ok := resources.FileIcons[ext]
	if ok {
		return t
	}

	return resources.FileIcons["file"]
}
