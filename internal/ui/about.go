package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"hjsplit2/internal/version"
)

func showAbout(parent fyne.Window) {
	borderBg := canvas.NewRectangle(color.NRGBA{0x40, 0x10, 0x60, 0xff})
	innerBg := canvas.NewRectangle(color.NRGBA{0x12, 0x00, 0x28, 0xff})

	title := widget.NewLabelWithStyle("hjsplit2", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.TextStyle.Monospace = true

	info := widget.NewLabel(fmt.Sprintf(
		"Version %s\n\n"+
			"A spiritual successor to HJSplit.\n"+
			"Split and join files with ease.\n\n"+
			"Built with Go and Fyne.\n"+
			"Licensed under GPLv3.\n\n"+
			"App icon by Icons8",
		version.Version,
	))
	info.Alignment = fyne.TextAlignCenter
	info.Wrapping = fyne.TextWrapWord

	iconBox := container.NewVBox()
	if appIconAbout != nil {
		iconImg := canvas.NewImageFromResource(appIconAbout)
		iconImg.FillMode = canvas.ImageFillOriginal
		iconBox.Add(iconImg)
	}

	body := container.NewStack(
		borderBg,
		container.NewPadded(
			container.NewStack(
				innerBg,
				container.NewPadded(
					container.NewVBox(
						layout.NewSpacer(),
						iconBox,
						title,
						layout.NewSpacer(),
						info,
						layout.NewSpacer(),
					),
				),
			),
		),
	)

	d := dialog.NewCustom("About hjsplit2", "Close", body, parent)
	d.Resize(fyne.NewSize(420, 380))
	d.Show()
}
