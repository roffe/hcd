package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hirschmann-koxha-gbr/hcd/window"
)

func main() {
	a := app.NewWithID("com.hcd.keygen")
	w := a.NewWindow("HCD kEyGen v0.1337")
	w.Resize(fyne.NewSize(500, 50))
	a.Settings().SetTheme(theme.DefaultTheme())
	vinInput := widget.NewEntry()
	result := widget.NewEntry()
	w.SetContent(container.NewVBox(
		container.NewBorder(
			nil,
			nil,
			widget.NewLabel("VIN/UUID:"), widget.NewButton("Generate", func() {
				if strings.HasPrefix(vinInput.Text, "ys") {
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
