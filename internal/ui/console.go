package ui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type consoleWidget struct {
	widget.BaseWidget
	lines    []string
	maxLines int
}

func newConsole() *consoleWidget {
	c := &consoleWidget{maxLines: 200}
	c.ExtendBaseWidget(c)
	return c
}

func (c *consoleWidget) Append(line string) {
	c.lines = append(c.lines, line)
	if len(c.lines) > c.maxLines {
		c.lines = c.lines[len(c.lines)-c.maxLines:]
	}
	c.Refresh()
}

func (c *consoleWidget) CreateRenderer() fyne.WidgetRenderer {
	c.ExtendBaseWidget(c)
	bg := canvas.NewRectangle(color.NRGBA{0x0a, 0x0a, 0x0a, 0xff})

	rich := widget.NewRichText(&widget.TextSegment{
		Text: "",
		Style: widget.RichTextStyle{
			ColorName: theme.ColorNameSuccess,
			TextStyle: fyne.TextStyle{Monospace: true},
		},
	})
	rich.Wrapping = fyne.TextWrapWord

	scroll := container.NewScroll(rich)

	return &consoleRenderer{
		w:      c,
		bg:     bg,
		rich:   rich,
		scroll: scroll,
	}
}

type consoleRenderer struct {
	w      *consoleWidget
	bg     *canvas.Rectangle
	rich   *widget.RichText
	scroll *container.Scroll
}

func (r *consoleRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	pad := float32(4)
	r.scroll.Move(fyne.NewPos(pad, pad))
	r.scroll.Resize(fyne.NewSize(size.Width-2*pad, size.Height-2*pad))
}

func (r *consoleRenderer) MinSize() fyne.Size {
	return fyne.NewSize(300, 120)
}

func (r *consoleRenderer) Refresh() {
	seg := r.rich.Segments[0].(*widget.TextSegment)
	seg.Text = strings.Join(r.w.lines, "\n")
	r.rich.Refresh()
	r.scroll.Refresh()
}

func (r *consoleRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.scroll}
}

func (r *consoleRenderer) Destroy() {}
