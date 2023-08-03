package main

import (
	_ "embed"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hirschmann-koxha-gbr/hcd/window"
)

//go:embed merkel.jpg
var merkel []byte

func main() {
	a := app.NewWithID("com.hcd.keygen")
	w := a.NewWindow("HCD kEyGen v0.1337")
	w.Resize(fyne.NewSize(500, 50))
	a.Settings().SetTheme(theme.DefaultTheme())
	vinInput := widget.NewEntry()
	result := widget.NewEntry()
	img := canvas.NewImageFromResource(fyne.NewStaticResource("merkel.jpg", merkel))
	img.FillMode = canvas.ImageFillContain
	img.ScaleMode = canvas.ImageScaleFastest
	img.SetMinSize(fyne.NewSize(50, 150))
	w.SetContent(container.NewVBox(
		container.NewBorder(
			img,
			nil,
			widget.NewLabel("VIN/UUID:"), widget.NewButton("Generate", func() {
				if len(vinInput.Text) == 17 {
					vinInput.SetText(strings.ToUpper(vinInput.Text))
				}
				result.SetText(window.CCalc([]byte(vinInput.Text)))
			}),
			vinInput,
		),
		container.NewBorder(
			nil,
			nil,
			widget.NewLabel("License:"),
			nil,
			result,
		),
	))
	w.ShowAndRun()
}
