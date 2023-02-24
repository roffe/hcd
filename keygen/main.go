package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hirschmann-koxha-gbr/hcd/window"
)

func main() {
	a := app.NewWithID("com.hcd.keygen")
	w := a.NewWindow("HCD kEyGen v0.1")
	w.Resize(fyne.NewSize(400, 50))
	a.Settings().SetTheme(theme.DefaultTheme())

	vinInput := widget.NewEntry()
	result := widget.NewLabel("")

	generateButton := widget.NewButton("Generate", func() {
		res := window.CCalc([]byte(vinInput.Text))
		result.SetText(res)
		w.Clipboard().SetContent(res)
	})

	window := container.NewVBox(
		container.New(
			layout.NewGridLayout(1),
			container.NewBorder(nil, nil, widget.NewLabel("VIN:"), generateButton, vinInput),
		),
		container.New(
			layout.NewGridLayout(1),
			container.NewBorder(nil, nil, widget.NewLabel("License:"), nil, result),
		),
	)

	w.SetContent(window)
	w.ShowAndRun()
}
