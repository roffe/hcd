package help

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func NewWindow(a fyne.App) {
	w := a.NewWindow("Help")
	w.Resize(fyne.NewSize(500, 800))
	w.SetContent(container.NewMax())
	w.Show()
}
