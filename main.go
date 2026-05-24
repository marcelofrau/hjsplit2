package main

import (
	_ "embed"

	"hjsplit2/internal/ui"
)

//go:embed assets/icons/app_icon.ico
var iconData []byte

//go:embed assets/app_icon.png
var aboutIconData []byte

func main() {
	ui.Run(iconData, aboutIconData)
}
