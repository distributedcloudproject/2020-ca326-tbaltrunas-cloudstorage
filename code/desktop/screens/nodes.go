package screens

import (
	"cloud/network"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

type nodeResource struct {
	icon   *widget.Icon
	entry  *widget.Entry
	online bool
	name   string
	ip     string
}

var nodeResourceCache map[string]nodeResource
var nodesList *widget.Box

func init() {
	nodeResourceCache = make(map[string]nodeResource)
}

func newNodeEntry(n network.Node, online bool) *fyne.Container {
	nodeResourceCache[n.ID] = nodeResource{
		icon:   widget.NewIcon(theme.RadioButtonIcon()),
		entry:  widget.NewEntry(),
		online: online,
		name:   n.Name,
		ip:     n.IP,
	}
	updateNodeEntry(n.ID)
	item := fyne.NewContainerWithLayout(layout.NewFormLayout(), nodeResourceCache[n.ID].icon, nodeResourceCache[n.ID].entry)
	return item
}

func updateNodeEntry(ID string) {
	if _, ok := nodeResourceCache[ID]; !ok {
		return
	}
	if nodeResourceCache[ID].online {
		nodeResourceCache[ID].icon.SetResource(theme.RadioButtonCheckedIcon())
	} else {
		nodeResourceCache[ID].icon.SetResource(theme.RadioButtonIcon())
	}
	nodeResourceCache[ID].entry.SetText(fmt.Sprintf("%s [%s]", nodeResourceCache[ID].name, nodeResourceCache[ID].ip))
	nodeResourceCache[ID].entry.Disable()
}

func updateNodeOnline(ID string, online bool) {
	e, ok := nodeResourceCache[ID]
	if !ok {
		return
	}
	e.online = online
	nodeResourceCache[ID] = e
}

func nodesEventNodeConnected(ID string) {
	updateNodeOnline(ID, true)
	updateNodeEntry(ID)
}

func nodesEventNodeDisconnected(ID string) {
	updateNodeOnline(ID, false)
	updateNodeEntry(ID)
}

func nodesEventNodeUpdated(n network.Node) {
	updateNodeEntry(n.ID)
}

func nodesEventNodeAdded(n network.Node) {
	if nodesList != nil {
		nodesList.Append(newNodeEntry(n, false))
	}
}

func NodesScreen(w fyne.Window, c network.Cloud) fyne.CanvasObject {
	nodesList = widget.NewVBox()

	nw := c.Network()
	for i := range nw.Nodes {
		nodesList.Append(newNodeEntry(nw.Nodes[i], c.IsNodeOnline(nw.Nodes[i].ID)))
	}

	return widget.NewVBox(
		nodesList,
	)
}
