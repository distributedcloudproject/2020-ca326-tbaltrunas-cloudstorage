package screens

import (
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/widget"
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
)

type toolbarWidget struct {
	w fyne.CanvasObject
}

func (t *toolbarWidget) ToolbarObject() fyne.CanvasObject {
	return t.w
}

// Toolbar widget creates a horizontal list of tool buttons
type FileWidget struct {
	widget.BaseWidget

	Items []widget.ToolbarItem

	objs []fyne.CanvasObject

	hovered  bool
	selected bool
	onclick  func()
}

func (f *FileWidget) append(item widget.ToolbarItem) {
	f.objs = append(f.objs, item.ToolbarObject())
}

func (f *FileWidget) prepend(item widget.ToolbarItem) {
	f.objs = append([]fyne.CanvasObject{item.ToolbarObject()}, f.objs...)
}

func (f *FileWidget) CreateRenderer() fyne.WidgetRenderer {
	f.ExtendBaseWidget(f)
	for _, item := range f.Items {
		f.append(item)
	}

	return &fileWidgetRenderer{toolbar: f, layout: layout.NewHBoxLayout()}
}

func (f *FileWidget) Append(item widget.ToolbarItem) {
	f.Items = append(f.Items, item)

	f.append(item)
}

func (f *FileWidget) Prepend(item widget.ToolbarItem) {
	f.Items = append([]widget.ToolbarItem{item}, f.Items...)

	f.prepend(item)
}

func (f *FileWidget) MinSize() fyne.Size {
	f.ExtendBaseWidget(f)
	return f.BaseWidget.MinSize()
}

func (f *FileWidget) Tapped(*fyne.PointEvent) {
}

func (f *FileWidget) TappedSecondary(*fyne.PointEvent) {
}
func (f *FileWidget) DoubleTapped(*fyne.PointEvent) {
	if f.onclick != nil {
		f.onclick()
	}
}

func (f *FileWidget) MouseIn(*desktop.MouseEvent) {
	f.hovered = true
	f.Refresh()
}

func (f *FileWidget) MouseOut() {
	f.hovered = false
	f.Refresh()
}

func (f *FileWidget) MouseMoved(*desktop.MouseEvent) {
}

func (f *FileWidget) FocusGained() {
	f.selected = true
	f.Refresh()
}

func (f *FileWidget) FocusLost() {
	f.selected = false
	f.Refresh()
}

func (f *FileWidget) Focused() bool {
	return f.selected
}

func (f *FileWidget) TypedRune(_ rune) {
}

func (f *FileWidget) TypedKey(_ *fyne.KeyEvent) {
}

// NewFileWidget creates a new file widget.
func NewFileWidget(icon fyne.Resource, label string, onclick func(), actions ...widget.ToolbarItem) *FileWidget {
	t := &FileWidget{
		onclick: onclick,
		Items: append([]widget.ToolbarItem{
			&toolbarWidget{widget.NewIcon(icon)},
			&toolbarWidget{widget.NewLabel(label)},
			widget.NewToolbarSpacer(),
		}, actions...),
	}
	t.ExtendBaseWidget(t)

	t.Refresh()
	return t
}

type fileWidgetRenderer struct {
	layout fyne.Layout

	objects []fyne.CanvasObject
	toolbar *FileWidget
}

func (r *fileWidgetRenderer) MinSize() fyne.Size {
	return r.layout.MinSize(r.objects)
}

func (r *fileWidgetRenderer) Layout(size fyne.Size) {
	r.layout.Layout(r.objects, size)
}

func (r *fileWidgetRenderer) BackgroundColor() color.Color {
	switch {
	case r.toolbar.selected:
		return theme.ButtonColor()
	case r.toolbar.hovered:
		return theme.HoverColor()
	default:
		return theme.ButtonColor()
	}
}

func (r *fileWidgetRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *fileWidgetRenderer) Refresh() {
	r.objects = r.toolbar.objs

	for i, item := range r.toolbar.Items {
		if _, ok := item.(*widget.ToolbarSeparator); ok {
			rect := r.objects[i].(*canvas.Rectangle)
			rect.FillColor = theme.TextColor()
		}
	}

	canvas.Refresh(r.toolbar)
}

func (r *fileWidgetRenderer) Destroy() {
}
