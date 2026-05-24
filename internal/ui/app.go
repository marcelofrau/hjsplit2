package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"hjsplit2/internal/core"
	"hjsplit2/internal/version"
)

var appIcon fyne.Resource
var appIconAbout fyne.Resource

func Run(iconData, aboutIconData []byte) {
	a := app.New()
	a.Settings().SetTheme(&synthwaveTheme{})

	if iconData != nil {
		appIcon = fyne.NewStaticResource("app_icon.ico", iconData)
		a.SetIcon(appIcon)
	}
	if aboutIconData != nil {
		appIconAbout = fyne.NewStaticResource("app_icon.png", aboutIconData)
	}

	w := a.NewWindow("hjsplit2")
	w.Resize(fyne.NewSize(640, 520))
	w.SetFixedSize(true)
	w.CenterOnScreen()

	ui := &appUI{app: a, window: w}
	ui.buildGUI()

	w.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		if len(uris) > 0 {
			ui.handleFile(uris[0].Path())
		}
	})

	w.ShowAndRun()
}

type appUI struct {
	app    fyne.App
	window fyne.Window
	tabs   *container.AppTabs

	filePath  string
	splitLabel *widget.Label
	joinLabel  *widget.Label
	sizeEntry *widget.Entry
	unitSel   *widget.Select
	splitBtn  *widget.Button
	joinBtn   *widget.Button
	progress  *widget.ProgressBar
	console   *consoleWidget
}

func (ui *appUI) log(format string, args ...any) {
	ui.console.Append(fmt.Sprintf(format, args...))
}

func (ui *appUI) buildGUI() {
	// Menu
	ui.window.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Open...", ui.browseFile),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() { ui.window.Close() }),
		),
		fyne.NewMenu("Help",
			fyne.NewMenuItem("About hjsplit2", func() { showAbout(ui.window) }),
		),
	))

	// Shared labels
	ui.splitLabel = widget.NewLabel("No file selected")
	ui.splitLabel.Wrapping = fyne.TextTruncate
	ui.joinLabel = widget.NewLabel("No file selected")
	ui.joinLabel.Wrapping = fyne.TextTruncate

	// Drop buttons
	splitDrop := widget.NewButton("Drag & Drop a file here\nor click to browse", ui.browseFile)
	joinDrop := widget.NewButton("Drag & Drop a file here\nor click to browse", ui.browseFile)

	// Custom Join button (Join tab only)
	customJoinBtn := widget.NewButton("Custom Join...", func() {
		ShowCustomJoin(ui.window)
	})

	// ── Split tab ──────────────────────────────────────────────
	ui.sizeEntry = widget.NewEntry()
	ui.sizeEntry.SetText("5")
	ui.unitSel = widget.NewSelect([]string{"KB", "MB", "GB"}, nil)
	ui.unitSel.SetSelected("MB")

	sizeRow := container.NewHBox(
		widget.NewLabel("Chunk size:"),
		ui.sizeEntry,
		ui.unitSel,
	)

	ui.splitBtn = widget.NewButton("Split", ui.startOp)
	ui.splitBtn.Importance = widget.HighImportance

	splitRight := container.NewVBox(
		layout.NewSpacer(),
		sizeRow,
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), ui.splitBtn),
	)

	splitCard := widget.NewCard("File Explorer", "",
		container.NewPadded(container.NewVBox(splitDrop, ui.splitLabel)))

	splitContent := container.NewBorder(nil, nil, nil, splitRight, splitCard)

	// ── Join tab ───────────────────────────────────────────────
	ui.joinBtn = widget.NewButton("Join", ui.startOp)
	ui.joinBtn.Importance = widget.HighImportance

	joinRight := container.NewVBox(
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), ui.joinBtn),
	)

	joinCard := widget.NewCard("File Explorer", "",
		container.NewPadded(container.NewVBox(joinDrop, ui.joinLabel, customJoinBtn)))

	joinContent := container.NewBorder(nil, nil, nil, joinRight, joinCard)

	// ── Tabs ───────────────────────────────────────────────────
	ui.tabs = container.NewAppTabs(
		container.NewTabItem("Split", splitContent),
		container.NewTabItem("Join", joinContent),
	)

	// ── Progress ───────────────────────────────────────────────
	ui.progress = widget.NewProgressBar()
	ui.progress.Hidden = true

	// ── Console ────────────────────────────────────────────────
	ui.console = newConsole()
	consolePadded := container.NewPadded(ui.console)

	// ── Layout ─────────────────────────────────────────────────
	topSection := container.NewVBox(ui.tabs, ui.progress)

	content := container.NewBorder(
		topSection,
		nil, nil, nil,
		consolePadded,
	)

	ui.window.SetContent(content)
	ui.log("hjsplit2 v%s", version.Version)
}

func (ui *appUI) browseFile() {
	fd := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		if r == nil {
			return
		}
		ui.handleFile(r.URI().Path())
		r.Close()
	}, ui.window)
	fd.Resize(fyne.NewSize(700, 500))
	fd.Show()
}

func (ui *appUI) handleFile(path string) {
	ui.filePath = path
	ui.splitLabel.SetText(path)
	ui.joinLabel.SetText(path)
	ui.log("Selected: %s", filepath.Base(path))
}

func (ui *appUI) startOp() {
	if ui.filePath == "" {
		dialog.ShowInformation("No file", "Select or drop a file first.", ui.window)
		return
	}

	if _, err := os.Stat(ui.filePath); os.IsNotExist(err) {
		dialog.ShowError(fmt.Errorf("file not found: %s", ui.filePath), ui.window)
		return
	}

	ui.progress.SetValue(0)
	ui.progress.Hidden = false

	if ui.tabs.SelectedIndex() == 0 {
		// Split
		ui.splitBtn.Disable()

		sizeVal, err := strconv.ParseInt(ui.sizeEntry.Text, 10, 64)
		if err != nil || sizeVal <= 0 {
			dialog.ShowError(fmt.Errorf("enter a positive number for chunk size"), ui.window)
			ui.splitBtn.Enable()
			ui.progress.Hidden = true
			return
		}
		chunkSize := sizeVal * unitMul(ui.unitSel.Selected)

		ui.log("Splitting: %s (%d %s chunks)", filepath.Base(ui.filePath), sizeVal, ui.unitSel.Selected)
		go func() {
			parts, err := core.Split(ui.filePath, chunkSize, func(cur, total int64) {
				if total > 0 {
					ui.progress.SetValue(float64(cur) / float64(total))
				}
			})
			ui.finishOp(ui.splitBtn, fmt.Sprintf("Split into %d parts", len(parts)), err)
		}()
	} else {
		// Join
		ui.joinBtn.Disable()

		ui.log("Joining: %s", filepath.Base(ui.filePath))
		go func() {
			joined, err := core.Join(ui.filePath, func(cur, total int64) {
				if total > 0 {
					ui.progress.SetValue(float64(cur) / float64(total))
				}
			})
			ui.finishOp(ui.joinBtn, fmt.Sprintf("Joined: %s", filepath.Base(joined)), err)
		}()
	}
}

func (ui *appUI) finishOp(btn *widget.Button, msg string, err error) {
	if err != nil {
		ui.log("Error: %v", err)
	} else {
		ui.progress.SetValue(1)
		ui.log(msg)
	}
	btn.Enable()
}

func unitMul(unit string) int64 {
	switch unit {
	case "KB":
		return 1024
	case "MB":
		return 1024 * 1024
	case "GB":
		return 1024 * 1024 * 1024
	}
	return 1024 * 1024
}
