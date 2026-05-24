package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
var globalApp fyne.App

func Run(iconData, aboutIconData []byte) {
	a := app.New()
	globalApp = a
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

	filePath    string
	splitLabel  *widget.Label
	joinLabel   *widget.Label
	sizeEntry   *widget.Entry
	unitSel     *widget.Select
	splitBtn    *widget.Button
	joinBtn     *widget.Button
	abortBtn    *widget.Button
	progress    *widget.ProgressBar
	console     *consoleWidget
	cancel      context.CancelFunc
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
	ui.sizeEntry.SetText("200")
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

	// ── Abort button ───────────────────────────────────────────
	ui.abortBtn = widget.NewButton("Abort", ui.abortOp)
	ui.abortBtn.Importance = widget.DangerImportance
	ui.abortBtn.Hide()

	abortRow := container.NewHBox(layout.NewSpacer(), ui.abortBtn)

	// ── Progress ───────────────────────────────────────────────
	ui.progress = widget.NewProgressBar()
	ui.progress.Hidden = true

	// ── Console ────────────────────────────────────────────────
	ui.console = newConsole()
	consolePadded := container.NewPadded(ui.console)

	// ── Layout ─────────────────────────────────────────────────
	topSection := container.NewVBox(ui.tabs, ui.progress, abortRow)

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

	info, err := os.Stat(ui.filePath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("file not found: %s", ui.filePath), ui.window)
		return
	}

	fileSize := info.Size()

	if ui.tabs.SelectedIndex() == 0 {
		// ── Split mode ────────────────────────────────────────
		sizeVal, err := strconv.ParseInt(ui.sizeEntry.Text, 10, 64)
		if err != nil || sizeVal <= 0 {
			dialog.ShowError(fmt.Errorf("enter a positive number for chunk size"), ui.window)
			return
		}
		chunkSize := sizeVal * unitMul(ui.unitSel.Selected)

		if chunkSize <= 0 {
			dialog.ShowError(fmt.Errorf("invalid chunk size"), ui.window)
			return
		}

		if fileSize <= chunkSize {
			dialog.ShowInformation("File too small",
				fmt.Sprintf("The file (%s) is smaller than the chunk size (%d %s). Nothing to split.",
					fmtBytes(fileSize), sizeVal, ui.unitSel.Selected), ui.window)
			return
		}

		estimatedParts := int((fileSize + chunkSize - 1) / chunkSize)
		if estimatedParts > 20 {
			suggestedMB := int(fileSize / (20 * 1024 * 1024))
			if suggestedMB < 1 {
				suggestedMB = 1
			}
			confirm := dialog.NewConfirm("Many parts",
				fmt.Sprintf("This will create %d parts.\nRecommended: increase chunk size to at least %d MB.\n\nContinue anyway?", estimatedParts, suggestedMB),
				func(ok bool) {
					if ok {
						ui.runSplit(chunkSize, fileSize)
					}
				}, ui.window)
			confirm.Show()
			return
		}

		ui.runSplit(chunkSize, fileSize)

	} else {
		// ── Join mode ─────────────────────────────────────────
		ui.runJoin()
	}
}

func (ui *appUI) runSplit(chunkSize int64, fileSize int64) {
	ui.splitBtn.Hide()
	ui.abortBtn.Show()
	ui.progress.SetValue(0)
	ui.progress.Hidden = false

	ctx, cancel := context.WithCancel(context.Background())
	ui.cancel = cancel

	ui.log("Splitting: %s (%s chunks of %s)",
		filepath.Base(ui.filePath), ui.sizeEntry.Text, ui.unitSel.Selected)

	go func() {
		parts, err := core.Split(ctx, ui.filePath, chunkSize, func(cur, total int64) {
			if total > 0 {
				ui.progress.SetValue(float64(cur) / float64(total))
			}
		})
		if err != nil && ctx.Err() != nil {
			ui.log("Aborted. Partial files removed.")
		}
		ui.finishOp(ui.splitBtn, fmt.Sprintf("Split into %d parts", len(parts)), err)
	}()
}

func (ui *appUI) runJoin() {
	ui.joinBtn.Hide()
	ui.abortBtn.Show()
	ui.progress.SetValue(0)
	ui.progress.Hidden = false

	ctx, cancel := context.WithCancel(context.Background())
	ui.cancel = cancel

	ui.log("Joining: %s", filepath.Base(ui.filePath))

	go func() {
		joined, err := core.Join(ctx, ui.filePath, func(cur, total int64) {
			if total > 0 {
				ui.progress.SetValue(float64(cur) / float64(total))
			}
		})
		if err != nil && ctx.Err() != nil {
			ui.log("Aborted. Output file removed.")
		}
		ui.finishOp(ui.joinBtn, fmt.Sprintf("Joined: %s", filepath.Base(joined)), err)
		if err == nil {
			ui.askDeleteParts(ui.filePath)
		}
	}()
}

func (ui *appUI) abortOp() {
	if ui.cancel == nil {
		return
	}
	confirm := dialog.NewConfirm("Abort operation",
		"Cancel the current operation? Partial files will be removed.",
		func(ok bool) {
			if ok {
				ui.cancel()
				ui.cancel = nil
			}
		}, ui.window)
	confirm.Show()
}

func (ui *appUI) finishOp(btn *widget.Button, msg string, err error) {
	ui.abortBtn.Hide()
	btn.Show()

	if err != nil {
		ui.log("Error: %v", err)
	} else {
		ui.progress.SetValue(1)
		ui.log(msg)
	}
	btn.Enable()
	ui.cancel = nil
}

func (ui *appUI) askDeleteParts(firstPart string) {
	basePath := strings.TrimSuffix(firstPart, ".001")
	if basePath == firstPart {
		return
	}
	var parts []string
	for i := 1; ; i++ {
		p := fmt.Sprintf("%s.%03d", basePath, i)
		if _, err := os.Stat(p); err != nil {
			break
		}
		parts = append(parts, p)
	}
	if len(parts) == 0 {
		return
	}

	dialog.NewConfirm("Delete source files",
		fmt.Sprintf("Delete the %d part file(s) after joining?", len(parts)),
		func(ok bool) {
			if !ok {
				return
			}
			for _, p := range parts {
				os.Remove(p)
				ui.log("Deleted: %s", filepath.Base(p))
			}
			ui.log("Source files removed.")
		}, ui.window).Show()
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

func fmtBytes(n int64) string {
	for _, unit := range []string{"B", "KB", "MB", "GB"} {
		if n < 1024 {
			return fmt.Sprintf("%d %s", n, unit)
		}
		n /= 1024
	}
	return fmt.Sprintf("%d TB", n)
}
