package screens

import (
	"cloud/datastore"
	"cloud/desktop/resources"
	"cloud/desktop/screens/widgets"
	"cloud/network"
	"errors"
	"fmt"
	"fyne.io/fyne"
	fdialog "fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func fancySizePrint(num float64, suffix string) string {
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(num) < 1024.0 {
			return fmt.Sprintf("%3.1f %s%s", num, unit, suffix)
		}
		num /= 1024.0
	}
	return fmt.Sprintf("%.1f %s%s", num, "Yi", suffix)
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
			},
				widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
					filename, err := BrowseDirDialog()
					if err != nil {
						fdialog.ShowError(err, w)
						return
					}

					fullpath := folderPath + "/" + p
					if err := c.SyncFolder(fullpath, filename); err != nil {
						fdialog.ShowError(err, w)
					}
				}),
				widget.NewToolbarAction(theme.DeleteIcon(), func() {
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
				&toolbarWidget{w: widget.NewLabel(fancySizePrint(float64(file.Size), "B"))},
				widget.NewToolbarSpacer(),
				widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
					filename, err := SaveFileDialog()
					if err != nil {
						fdialog.ShowError(err, w)
						return
					}

					fullpath := folderPath + "/" + file.Name
					if err := c.SyncFile(fullpath, filename); err != nil {
						fdialog.ShowError(err, w)
					}
				}),
				widget.NewToolbarAction(theme.MoveDownIcon(), func() {
					filename, err := SaveFileDialog()
					if err != nil {
						fdialog.ShowError(err, w)
						return
					}

					fullpath := folderPath + "/" + file.Name
					var q *network.DownloadQueue
					q = c.DownloadManager().QueueDownload(fullpath, filename, func(event network.DownloadEvent) {
						completed := false
						if event == network.DownloadCompleted {
							completed = true
						}
						if event == network.InfoRetrieved {
							go func() {
								content := widget.NewHBox(
									widgets.NewChunkProgressBar(q.ChunkDownloaded),
								)
								fdialog.ShowCustom("Downloading", "Dismiss", content, w)
								for !completed {
									content.Refresh()
									time.Sleep(time.Millisecond * 100)
								}
							}()
						}
					})
				}),
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
		filename, err := LoadFileDialog()
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
		f, err := datastore.NewFile(reader, fname, 1024*1024*4)
		reader.Close()
		if err != nil {
			fdialog.ShowError(err, w)
			return
		}

		err = c.AddFile(f, path.Join(folderPath, fname), filename)
		if err != nil {
			fdialog.ShowError(err, w)
		}
		// c.DistributeFile(f)
		updateList()
	}
	syncFile := func() {
		filename, err := LoadFileDialog()
		if err != nil {
			fdialog.ShowError(err, w)
			return
		}

		content := widget.NewEntry()
		content.SetPlaceHolder("file")
		fdialog.ShowCustomConfirm("Enter file name", "Sync", "Cancel", content, func(s bool) {
			if s {
				if err := c.SyncFile(folderPath+"/"+content.Text, filename); err != nil {
					fdialog.ShowError(err, w)
				}
				updateList()
			}
		}, w)
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
		widget.NewButtonWithIcon("Sync", theme.ContentAddIcon(), syncFile),
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
