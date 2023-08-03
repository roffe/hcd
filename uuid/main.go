package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/roffe/hcd/window"
)

func main() {
	a := app.NewWithID("com.example.app")
	w := a.NewWindow("UUID")
	w.Resize(fyne.NewSize(300, 50))
	e := widget.NewEntry()

	win, err := window.GetMachineUUID()
	if err != nil {
		dialog.ShowError(err, w)
	}
	e.SetText(win)

	w.SetContent(container.NewMax(e))
	w.ShowAndRun()
}
