package widgets

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"image/color"
)

type progressRenderer struct {
	objects []fyne.CanvasObject

	bars []*canvas.Rectangle

	progress *ChunkProgress
}

func (p *progressRenderer) MinSize() fyne.Size {
	return fyne.NewSize(100+theme.Padding()*4+(2*p.progress.NumChunks), theme.Padding()*6)
}

func (p *progressRenderer) updateBar() {
	//size := p.progress.Size()
	for i := range p.bars {
		barcolor := theme.ButtonColor()
		if p.progress.Chunks[i] {
			barcolor = theme.PrimaryColor()
		}
		p.bars[i].FillColor = barcolor
		//p.bars[i].SetMinSize(fyne.NewSize(size.Width/p.progress.NumChunks, size.Height))
		// p.bars[i].Refresh()
	}
}

// Layout the components of the check widget
func (p *progressRenderer) Layout(size fyne.Size) {
	p.updateBar()
	p.layout(p.objects, size)
}

func (g *progressRenderer) layout(objects []fyne.CanvasObject, size fyne.Size) {
	x, y := 0, 0
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		child.Move(fyne.NewPos(x, y))
		x += child.Size().Width
	}
}

func (p *progressRenderer) BackgroundColor() color.Color {
	return theme.ButtonColor()
}

func (p *progressRenderer) Refresh() {
	p.updateBar()
	canvas.Refresh(p.progress)
}

func (p *progressRenderer) Objects() []fyne.CanvasObject {
	return p.objects
}

func (p *progressRenderer) Destroy() {
}

// ProgressBar widget creates a horizontal panel that indicates progress
type ChunkProgress struct {
	widget.BaseWidget

	NumChunks int
	Chunks    []bool
}

// MinSize returns the size that this widget should not shrink below
func (p *ChunkProgress) MinSize() fyne.Size {
	p.ExtendBaseWidget(p)
	return p.BaseWidget.MinSize()
	//return fyne.NewSize((100 + theme.Padding()*4 + (2 * p.NumChunks)), s.Height)
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (p *ChunkProgress) CreateRenderer() fyne.WidgetRenderer {
	p.ExtendBaseWidget(p)

	// size := p.Size()
	chunkWidth := (100 + theme.Padding()*4 + (2 * p.NumChunks)) / p.NumChunks

	pr := &progressRenderer{objects: []fyne.CanvasObject{}, progress: p}
	for i := range p.Chunks {
		barcolor := theme.ButtonColor()
		if p.Chunks[i] {
			barcolor = theme.PrimaryColor()
		}
		bar := canvas.NewRectangle(barcolor)
		bar.SetMinSize(fyne.NewSize(chunkWidth, theme.Padding()*6))
		bar.Resize(fyne.NewSize(chunkWidth, theme.Padding()*6))
		pr.bars = append(pr.bars, bar)
		pr.objects = append(pr.objects, bar)
	}
	return pr
}

func NewChunkProgressBar(chunks []bool) *ChunkProgress {
	p := &ChunkProgress{NumChunks: len(chunks), Chunks: chunks}

	//widget.Renderer(p).Layout(p.MinSize())

	return p
}
