package screens

import (
	"cloud/network"
	"fyne.io/fyne"
	"fyne.io/fyne/widget"
	"strconv"
)

var nodesOnlineWidget *widget.Label
var nodesOnline func() int

func homeEventNodeConnected(ID string) {
	if nodesOnlineWidget != nil && nodesOnline != nil {
		nodesOnlineWidget.SetText("Nodes Online: " + strconv.Itoa(nodesOnline()))
	}
}

func homeEventNodeDisconnected(ID string) {
	if nodesOnlineWidget != nil && nodesOnline != nil {
		nodesOnlineWidget.SetText("Nodes Online: " + strconv.Itoa(nodesOnline()))
	}
}

func HomeScreen(w fyne.Window, c network.Cloud) fyne.CanvasObject {
	infoWidget := func(label, value string) *widget.Box {
		e := widget.NewEntry()
		e.SetText(value)
		e.Disable()

		return widget.NewHBox(
			widget.NewLabel(label),
			e,
		)
	}
	nodesOnline = c.OnlineNodesNum
	nodesOnlineWidget = widget.NewLabel("Nodes Online: " + strconv.Itoa(nodesOnline()))
	myNode := c.MyNode()
	return widget.NewVBox(
		widget.NewLabel("Network Name: "+c.Network().Name),
		nodesOnlineWidget,
		widget.NewLabel(""),
		widget.NewLabel("Node Name: "+myNode.Name),
		infoWidget("Node IP: ", myNode.IP),
		infoWidget("Node ID: ", myNode.ID),
	)
}
