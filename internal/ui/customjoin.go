package ui

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"hjsplit2/internal/core"
)

type customJoinUI struct {
	parent    fyne.Window
	win       fyne.Window
	files     []string
	selected  int
	list      *widget.List
	output    *widget.Entry
	progress  *widget.ProgressBar
	startBtn  *widget.Button
	abortBtn  *widget.Button
	status    *widget.Label
	cancel    context.CancelFunc
}

func ShowCustomJoin(parent fyne.Window) {
	win := globalApp.NewWindow("Custom Join")
	win.Resize(fyne.NewSize(640, 500))

	ui := &customJoinUI{parent: parent, win: win, selected: -1}
	ui.build()

	win.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		for _, u := range uris {
			ui.addSingleFile(u.Path())
		}
	})

	// Center before showing
	win.CenterOnScreen()
	win.Show()
}

func (ui *customJoinUI) build() {
	ui.list = widget.NewList(
		func() int { return len(ui.files) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(filepath.Base(ui.files[id]))
		},
	)
	ui.list.OnSelected = func(id widget.ListItemID) {
		ui.selected = id
	}
	ui.list.OnUnselected = func(widget.ListItemID) {
		ui.selected = -1
	}

	// Navigation buttons alongside list
	moveTop := widget.NewButton("⤒", func() { ui.moveTop() })
	moveUp := widget.NewButton("↑", func() { ui.moveUp() })
	moveDown := widget.NewButton("↓", func() { ui.moveDown() })
	moveBottom := widget.NewButton("⤓", func() { ui.moveBottom() })

	navCol := container.NewVBox(
		moveTop,
		moveUp,
		layout.NewSpacer(),
		moveDown,
		moveBottom,
	)

	listRow := container.NewBorder(nil, nil, nil, navCol, ui.list)

	listBg := canvas.NewRectangle(color.NRGBA{0x0b, 0x25, 0x48, 0xff})
	listWithBg := container.NewStack(listBg, container.NewPadded(listRow))

	listCard := widget.NewCard("Files to join (order matters)", "", listWithBg)

	// Bottom buttons
	addBtn := widget.NewButton("Add files...", ui.addFiles)
	removeBtn := widget.NewButton("Remove", ui.remove)
	sortBtn := widget.NewButton("Sort", ui.autoSort)

	btnRow := container.NewHBox(addBtn, removeBtn, layout.NewSpacer(), sortBtn)

	// Output
	ui.output = widget.NewEntry()
	browseBtn := widget.NewButton("Browse...", func() {
		d := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
			if err == nil && w != nil {
				ui.output.SetText(w.URI().Path())
				w.Close()
			}
		}, ui.win)
		d.Resize(fyne.NewSize(700, 500))
		d.Show()
	})
	outputRow := container.NewBorder(nil, nil, widget.NewLabel("Output:"), browseBtn, ui.output)

	// Status
	ui.status = widget.NewLabel("")
	ui.status.Alignment = fyne.TextAlignCenter

	// Progress
	ui.progress = widget.NewProgressBar()
	ui.progress.Hidden = true

	// Start / Abort buttons (left side)
	ui.startBtn = widget.NewButton("START JOIN", ui.startJoin)
	ui.startBtn.Importance = widget.HighImportance

	ui.abortBtn = widget.NewButton("Abort", ui.abortOp)
	ui.abortBtn.Importance = widget.DangerImportance
	ui.abortBtn.Hide()

	leftCol := container.NewVBox(
		ui.startBtn,
		ui.abortBtn,
	)

	// Drop hint (left side)
	dropHint := widget.NewLabelWithStyle("Tip: Drag files anywhere onto this window", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

	// Bottom area: left (drop hint) + right (buttons)
	bottomRow := container.NewGridWithColumns(2,
		container.NewHBox(dropHint, layout.NewSpacer()),
		container.NewHBox(layout.NewSpacer(), leftCol),
		// right-aligned buttons
	)

	// Main content
	content := container.NewBorder(
		nil, nil,
		nil, nil,
		container.NewVBox(
			listCard,
			btnRow,
			outputRow,
			ui.progress,
			ui.status,
			bottomRow,
		),
	)

	ui.win.SetContent(content)
}

func (ui *customJoinUI) addFiles() {
	d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil || r == nil {
			return
		}
		path := r.URI().Path()
		r.Close()
		ui.addSingleFile(path)
		ui.addFiles()
	}, ui.win)
	d.Resize(fyne.NewSize(700, 500))
	d.Show()
}

func (ui *customJoinUI) addSingleFile(path string) {
	for _, f := range ui.files {
		if f == path {
			return
		}
	}
	ui.files = append(ui.files, path)
	ui.list.Refresh()
	if ui.output.Text == "" {
		ui.guessOutput()
	}
}

func (ui *customJoinUI) remove() {
	if ui.selected < 0 || ui.selected >= len(ui.files) {
		return
	}
	ui.files = append(ui.files[:ui.selected], ui.files[ui.selected+1:]...)
	ui.selected = -1
	ui.list.UnselectAll()
	ui.list.Refresh()
}

func (ui *customJoinUI) moveTop() {
	if ui.selected <= 0 {
		return
	}
	f := ui.files[ui.selected]
	ui.files = append(ui.files[:ui.selected], ui.files[ui.selected+1:]...)
	ui.files = append([]string{f}, ui.files...)
	ui.selected = 0
	ui.list.Refresh()
	ui.list.Select(ui.selected)
}

func (ui *customJoinUI) moveUp() {
	if ui.selected <= 0 || ui.selected >= len(ui.files) {
		return
	}
	i := ui.selected
	ui.files[i], ui.files[i-1] = ui.files[i-1], ui.files[i]
	ui.selected = i - 1
	ui.list.Refresh()
	ui.list.Select(ui.selected)
}

func (ui *customJoinUI) moveDown() {
	if ui.selected < 0 || ui.selected >= len(ui.files)-1 {
		return
	}
	i := ui.selected
	ui.files[i], ui.files[i+1] = ui.files[i+1], ui.files[i]
	ui.selected = i + 1
	ui.list.Refresh()
	ui.list.Select(ui.selected)
}

func (ui *customJoinUI) moveBottom() {
	if ui.selected < 0 || ui.selected >= len(ui.files)-1 {
		return
	}
	f := ui.files[ui.selected]
	ui.files = append(ui.files[:ui.selected], ui.files[ui.selected+1:]...)
	ui.files = append(ui.files, f)
	ui.selected = len(ui.files) - 1
	ui.list.Refresh()
	ui.list.Select(ui.selected)
}

func (ui *customJoinUI) autoSort() {
	sort.Slice(ui.files, func(i, j int) bool {
		return strings.ToLower(filepath.Base(ui.files[i])) < strings.ToLower(filepath.Base(ui.files[j]))
	})
	ui.selected = -1
	ui.list.UnselectAll()
	ui.list.Refresh()
}

func (ui *customJoinUI) guessOutput() {
	if len(ui.files) == 0 {
		return
	}
	name := filepath.Base(ui.files[0])
	for _, suffix := range []string{".part00", ".part01", ".001", ".002", ".part0", ".part1"} {
		if strings.HasSuffix(name, suffix) {
			name = name[:len(name)-len(suffix)]
			break
		}
	}
	dest := filepath.Join(filepath.Dir(ui.files[0]), name+"_joined")
	ui.output.SetText(dest)
}

func (ui *customJoinUI) startJoin() {
	if len(ui.files) == 0 {
		dialog.ShowInformation("No files", "Add at least one file to join.", ui.win)
		return
	}
	dest := ui.output.Text
	if dest == "" {
		dialog.ShowInformation("No output", "Specify an output file path.", ui.win)
		return
	}

	for _, f := range ui.files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			dialog.ShowError(fmt.Errorf("file not found: %s", filepath.Base(f)), ui.win)
			return
		}
	}

	ui.startBtn.Hide()
	ui.abortBtn.Show()
	ui.progress.Hidden = false
	ui.progress.SetValue(0)
	ui.status.SetText(fmt.Sprintf("Joining %d files...", len(ui.files)))

	ctx, cancel := context.WithCancel(context.Background())
	ui.cancel = cancel

	go func() {
		err := core.JoinMulti(ctx, ui.files, dest, func(cur, total int64) {
			if total > 0 {
				ui.progress.SetValue(float64(cur) / float64(total))
			}
		})
		if err != nil && ctx.Err() != nil {
			ui.status.SetText("Aborted. Output file removed.")
		}
		ui.finishJoin(err, dest)
		if err == nil {
			ui.askDeleteSources()
		}
	}()
}

func (ui *customJoinUI) abortOp() {
	if ui.cancel == nil {
		return
	}
	confirm := dialog.NewConfirm("Abort operation",
		"Cancel the current join? The output file will be removed.",
		func(ok bool) {
			if ok {
				ui.cancel()
				ui.cancel = nil
			}
		}, ui.win)
	confirm.Show()
}

func (ui *customJoinUI) askDeleteSources() {
	if len(ui.files) == 0 {
		return
	}
	dialog.NewConfirm("Delete source files",
		fmt.Sprintf("Delete the %d source file(s) after joining?", len(ui.files)),
		func(ok bool) {
			if !ok {
				return
			}
			for _, p := range ui.files {
				os.Remove(p)
			}
		}, ui.win).Show()
}

func (ui *customJoinUI) finishJoin(err error, dest string) {
	ui.abortBtn.Hide()
	ui.startBtn.Show()
	if err != nil {
		if ui.cancel == nil {
			ui.status.SetText(fmt.Sprintf("Error: %v", err))
		}
	} else {
		ui.progress.SetValue(1)
		ui.status.SetText(fmt.Sprintf("Done! → %s", filepath.Base(dest)))
	}
	ui.startBtn.Enable()
	ui.cancel = nil
}
