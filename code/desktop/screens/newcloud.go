package screens

import (
	"cloud/network"
	"crypto/rsa"
	"crypto/x509"
	"encoding/gob"
	"encoding/pem"
	"errors"
	"fmt"
	"fyne.io/fyne"
	fdialog "fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

var progressBar *widget.ProgressBar

var newCloudForm struct {
	newNetwork bool

	networkName *widget.Entry
	networkIP   *widget.Entry

	nodeName           *widget.Entry
	nodeIP             *widget.Entry
	nodePort           *widget.Entry
	nodePrivateKeyPath *widget.Entry
	nodeFileStorageDir *widget.Entry
}

func setProgressBarAnimated(newValue float64, timeToUpdate time.Duration) {
	val := progressBar.Value
	increments := (newValue - val) / float64(timeToUpdate.Milliseconds()/33)
	for val < newValue {
		val += increments
		if val > newValue {
			val = newValue
		}
		progressBar.SetValue(val)
		time.Sleep(time.Millisecond * 33)
	}
}

func createWidgets() {
	newCloudForm.networkName = &widget.Entry{
		PlaceHolder: "Network Name",
	}
	newCloudForm.networkIP = &widget.Entry{
		PlaceHolder: "Network IP",
	}

	newCloudForm.nodeName = &widget.Entry{
		PlaceHolder: "Node Name",
	}
	newCloudForm.nodeIP = &widget.Entry{
		PlaceHolder: "Your IP address - only needed if connecting to local network node",
	}
	newCloudForm.nodePort = &widget.Entry{
		PlaceHolder: "Port to listen on",
	}
	newCloudForm.nodePrivateKeyPath = &widget.Entry{
		PlaceHolder: "Path to your private key",
	}

	newCloudForm.nodeFileStorageDir = &widget.Entry{
		PlaceHolder: "Path to directory to store files",
	}
}

func navButtons(back func(), next func()) fyne.CanvasObject {
	var backButton, nextButton *widget.Button
	if back != nil {
		backButton = widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), back)
	}
	if next == nil {
		return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, backButton, nil),
			backButton)
	}
	nextButton = widget.NewButtonWithIcon("Continue", theme.NavigateNextIcon(), next)
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, backButton, nextButton),
		backButton, nextButton)
}

func NewCloudScreen(win fyne.Window) fyne.CanvasObject {
	createWidgets()

	progressBar = &widget.ProgressBar{
		Min:   0,
		Max:   100,
		Value: 0,
	}

	w := widget.NewVBox(
		progressBar,
		widget.NewLabelWithStyle("Would you like to create a new Cloud Network or join an existing one?",
			fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		&widget.Box{},
		widget.NewButton("Load from file", func() {
			filename, err := LoadFileDialog()
			if err == nil {
				r, err := os.Open(filename)
				if err == nil {
					defer r.Close()
					decoder := gob.NewDecoder(r)
					var savedNetwork network.SavedNetworkState
					err := decoder.Decode(&savedNetwork)
					if err == nil {
						c := network.LoadNetwork(savedNetwork)
						err = c.Listen()
						if err != nil {
							fdialog.ShowError(err, win)
							return
						}
						fmt.Printf("%v\n", c)
						win.SetContent(connectedToNetwork(win, c))
					}
				}
			}
		}),
		widget.NewButton("Create", func() {
			newCloudForm.newNetwork = true
			win.SetContent(createCloudScreen(win))
		}),
		widget.NewButton("Join", func() {
			newCloudForm.newNetwork = false
			win.SetContent(joinCloudScreen(win))
		}),
	)

	return w
}

func createCloudScreen(win fyne.Window) fyne.CanvasObject {
	go setProgressBarAnimated(25.0, time.Millisecond*300)

	w := widget.NewVBox(
		progressBar,
		widget.NewLabelWithStyle("Name your Cloud!",
			fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		&widget.Box{},
		newCloudForm.networkName,
		&widget.Box{},
		navButtons(func() {
			// Back
			win.SetContent(NewCloudScreen(win))
		}, func() {
			// Continue
			win.SetContent(createNodeScreen(win))
		}),
	)
	return w
}

func joinCloudScreen(win fyne.Window) fyne.CanvasObject {
	go setProgressBarAnimated(25.0, time.Millisecond*300)

	w := widget.NewVBox(
		progressBar,
		widget.NewLabelWithStyle("Enter the IP of an online node in the Cloud!",
			fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		&widget.Box{},
		newCloudForm.networkIP,
		&widget.Box{},
		navButtons(func() {
			// Back
			win.SetContent(NewCloudScreen(win))
		}, func() {
			// Continue
			win.SetContent(createNodeScreen(win))
		}),
	)
	return w
}

func createNodeScreen(win fyne.Window) fyne.CanvasObject {
	go setProgressBarAnimated(50.0, time.Millisecond*300)
	browseButton := widget.NewButton("Browse", func() {
		filename, err := LoadFileDialog()
		if err == nil {
			newCloudForm.nodePrivateKeyPath.SetText(filename)
		}
	})
	browseFileStorageButton := widget.NewButton("Browse", func() {
		filename, err := BrowseDirDialog()
		if err == nil {
			newCloudForm.nodeFileStorageDir.SetText(filename)
		}
	})
	keyID := widget.NewEntry()
	keyID.Disable()

	newCloudForm.nodePrivateKeyPath.OnChanged = func(new string) {
		key, err := readKey(new)
		if err != nil {
			keyID.SetText("")
			return
		}
		id, err := network.PublicKeyToID(&key.PublicKey)
		if err != nil {
			keyID.SetText("")
			return
		}
		keyID.SetText(id)
	}

	w := widget.NewVBox(
		progressBar,
		widget.NewLabelWithStyle("Enter your Node information",
			fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		&widget.Box{},
		newCloudForm.nodeName,
		newCloudForm.nodeIP,
		newCloudForm.nodePort,
		fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, newCloudForm.nodePrivateKeyPath, browseButton),
			newCloudForm.nodePrivateKeyPath, browseButton),
		fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, newCloudForm.nodeFileStorageDir, browseFileStorageButton),
			newCloudForm.nodeFileStorageDir, browseFileStorageButton),
		widget.NewHBox(keyID, widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
			clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
			clipboard.SetContent(keyID.Text)
		})),

		&widget.Box{},
		navButtons(func() {
			// Back
			if newCloudForm.newNetwork {
				win.SetContent(createCloudScreen(win))
			} else {
				win.SetContent(joinCloudScreen(win))
			}
		}, func() {
			// Continue
			win.SetContent(connectingToNetworkScreen(win))
		}),
	)
	return w
}

func connectingToNetworkScreen(win fyne.Window) fyne.CanvasObject {
	go setProgressBarAnimated(75.0, time.Millisecond*300)

	w := widget.NewVBox(
		progressBar,
		widget.NewLabelWithStyle("Connecting", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		widget.NewProgressBarInfinite(),
		layout.NewSpacer(),
	)
	go func() {
		displayError := func(err error) {
			win.SetContent(widget.NewVBox(
				progressBar,
				layout.NewSpacer(),
				widget.NewLabelWithStyle(err.Error(), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				layout.NewSpacer(),
				navButtons(func() {
					win.SetContent(createNodeScreen(win))
				}, nil),
				layout.NewSpacer(),
			))
		}
		key, err := readKey(newCloudForm.nodePrivateKeyPath.Text)
		if err != nil {
			displayError(err)
			return
		}
		id, err := network.PublicKeyToID(&key.PublicKey)
		if err != nil {
			displayError(err)
			return
		}
		port, err := strconv.Atoi(newCloudForm.nodePort.Text)
		if err != nil {
			displayError(errors.New("Invalid port: " + err.Error()))
			return
		}
		me := network.Node{
			ID:        id,
			Name:      newCloudForm.nodeName.Text,
			IP:        newCloudForm.nodeIP.Text + ":" + newCloudForm.nodePort.Text,
			PublicKey: key.PublicKey,
		}
		config := network.CloudConfig{FileStorageDir: newCloudForm.nodeFileStorageDir.Text}
		if newCloudForm.newNetwork {
			c := network.SetupNetwork(network.Network{
				Name:        newCloudForm.networkName.Text,
				RequireAuth: true,
				Whitelist:   true,
			}, me, key)
			c.SetConfig(config)
			err = c.ListenOnPort(port)
			if err != nil {
				fdialog.ShowError(err, win)
			}
			win.SetContent(connectedToNetwork(win, c))
		} else {
			c, err := network.BootstrapToNetwork(newCloudForm.networkIP.Text, me, key, config)
			if err != nil {
				displayError(err)
				return
			}
			err = c.ListenOnPort(port)
			if err != nil {
				displayError(err)
				return
			}
			win.SetContent(connectedToNetwork(win, c))
		}
	}()
	return w
}

func connectedToNetwork(win fyne.Window, c network.Cloud) fyne.CanvasObject {
	go setProgressBarAnimated(100.0, time.Millisecond*300)

	go func() {
		c.Accept()
	}()

	w := widget.NewVBox(
		progressBar,
		layout.NewSpacer(),
		widget.NewLabelWithStyle("Connected to "+c.Network().Name+"! You are all set.", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		fyne.NewContainerWithLayout(layout.NewCenterLayout(), widget.NewButtonWithIcon("Continue", theme.ConfirmIcon(), func() {
			Navigation(win, c)
		})),
		layout.NewSpacer(),
	)
	return w
}

func readKey(file string) (*rsa.PrivateKey, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	bb, _ := pem.Decode(b)
	if bb.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("invalid type " + bb.Type + " want: RSA PRIVATE KEY")
	}

	key, err := x509.ParsePKCS1PrivateKey(bb.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}
