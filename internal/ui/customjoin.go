package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"hjsplit2/internal/core"
)

type customJoinUI struct {
	parent    fyne.Window
	files     []string
	selected  int
	list      *widget.List
	output    *widget.Entry
	progress  *widget.ProgressBar
	startBtn  *widget.Button
	status    *widget.Label
}

func ShowCustomJoin(parent fyne.Window) {
	ui := &customJoinUI{parent: parent, selected: -1}
	ui.build()
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

	listCard := widget.NewCard("Files to join (order matters)", "", container.NewPadded(ui.list))

	addBtn := widget.NewButton("Add files...", ui.addFiles)
	removeBtn := widget.NewButton("Remove", ui.remove)
	moveUp := widget.NewButton("↑", ui.moveUp)
	moveDown := widget.NewButton("↓", ui.moveDown)
	sortBtn := widget.NewButton("Sort", ui.autoSort)

	btnRow := container.NewHBox(addBtn, removeBtn, layout.NewSpacer(), moveUp, moveDown, sortBtn)

	ui.output = widget.NewEntry()
	browseBtn := widget.NewButton("Browse...", func() {
		d := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
			if err == nil && w != nil {
				ui.output.SetText(w.URI().Path())
				w.Close()
			}
		}, ui.parent)
		d.Resize(fyne.NewSize(700, 500))
		d.Show()
	})
	outputRow := container.NewHBox(widget.NewLabel("Output:"), ui.output, browseBtn)

	ui.status = widget.NewLabel("")
	ui.status.Alignment = fyne.TextAlignCenter

	ui.progress = widget.NewProgressBar()
	ui.progress.Hidden = true

	ui.startBtn = widget.NewButton("START JOIN", ui.startJoin)
	ui.startBtn.Importance = widget.HighImportance

	content := container.NewBorder(
		nil, nil,
		nil, nil,
		container.NewVBox(
			listCard,
			btnRow,
			outputRow,
			ui.progress,
			ui.status,
			container.NewHBox(layout.NewSpacer(), ui.startBtn, layout.NewSpacer()),
		),
	)

	d := dialog.NewCustom("Custom Join", "Close", content, ui.parent)
	d.Resize(fyne.NewSize(560, 480))
	d.Show()
}

func (ui *customJoinUI) addFiles() {
	d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil || r == nil {
			return
		}
		path := r.URI().Path()
		r.Close()
		ui.addSingleFile(path)
		// Keep adding more
		ui.addFiles()
	}, ui.parent)
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
		dialog.ShowInformation("No files", "Add at least one file to join.", ui.parent)
		return
	}
	dest := ui.output.Text
	if dest == "" {
		dialog.ShowInformation("No output", "Specify an output file path.", ui.parent)
		return
	}

	for _, f := range ui.files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			dialog.ShowError(fmt.Errorf("file not found: %s", filepath.Base(f)), ui.parent)
			return
		}
	}

	ui.startBtn.Disable()
	ui.progress.Hidden = false
	ui.progress.SetValue(0)
	ui.status.SetText(fmt.Sprintf("Joining %d files...", len(ui.files)))

	go func() {
		err := core.JoinMulti(ui.files, dest, func(cur, total int64) {
			if total > 0 {
				ui.progress.SetValue(float64(cur) / float64(total))
			}
		})
		if err != nil {
			ui.status.SetText(fmt.Sprintf("Error: %v", err))
		} else {
			ui.progress.SetValue(1)
			ui.status.SetText(fmt.Sprintf("Done! → %s", filepath.Base(dest)))
		}
		ui.startBtn.Enable()
	}()
}
