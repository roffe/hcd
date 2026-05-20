package window

import (
	"context"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/roffe/hcd/ent"
)

func (m *Main) refreshHistory() {
	m.historyList.RemoveAll()

	cars, err := m.DB.GetCars(context.TODO())
	if err != nil {
		m.writeOutput(err.Error())
		return
	}

	for _, car := range cars {
		title := car.VIN
		if car.Name != "" {
			title = car.Name + " (" + car.VIN + ")"
		}
		m.historyList.Add(widget.NewButtonWithIcon(title, theme.InfoIcon(), func() {
			var popup *widget.PopUp

			hist, err := m.DB.GetValueHistory(context.TODO(), car.VIN)
			if err != nil {
				m.writeOutput(err.Error())
			}

			entries := container.NewVBox()
			for _, h := range hist {
				//entry := m.createHistoryEntry(h, popup)
				createdAt := widget.NewLabel(h.CreatedAt.Format("02/01-06 15:04:05"))

				uecBytes := widget.NewLabel(fmt.Sprintf("%02X", h.UEC))
				uecBytes.Selectable = true

				uecPWMBytes := widget.NewLabel(fmt.Sprintf("%02X", h.UECPWM))
				uecPWMBytes.Selectable = true

				recBytes := widget.NewLabel(fmt.Sprintf("%02X", h.REC))
				recBytes.Selectable = true

				typeLabel := widget.NewLabel(h.Type.String())

				loadBtn := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
					m.setInternalUECState(h.UEC)
					m.setInternalRECState(h.REC)
					m.uecPWMBytes = h.UECPWM
					m.uecPWM.SetText(fmt.Sprintf("%02X", h.UECPWM))
					popup.Hide()
				})
				loadBtn.Importance = widget.LowImportance

				entry := container.NewGridWithColumns(6, typeLabel, createdAt, uecBytes, uecPWMBytes, recBytes, loadBtn)

				entries.Add(entry)
			}

			customName := widget.NewEntry()
			customName.SetText(car.Name)
			customName.OnChanged = func(s string) {
				if err := car.Update().SetName(s).Exec(context.TODO()); err != nil {
					m.writeOutput(fmt.Sprintf("failed updating car name: %v", err))
				}
			}

			content := container.NewBorder(
				container.NewBorder(
					nil,
					nil,
					nil,
					widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
						popup.Hide()
						m.refreshHistory()
					}),
					container.NewGridWithColumns(2,
						container.NewBorder(
							nil,
							nil,
							widget.NewLabel("Name:"),
							nil,
							customName,
						),
						widget.NewLabelWithStyle(car.VIN, fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
					),
				),
				nil,
				nil,
				nil,
				container.NewBorder(
					nil,
					nil,
					nil,
					nil,
					container.NewBorder(
						container.NewGridWithColumns(6,
							widget.NewLabelWithStyle("Type", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
							widget.NewLabelWithStyle("Date", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
							widget.NewLabelWithStyle("UEC", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
							widget.NewLabelWithStyle("UEC PWM", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
							widget.NewLabelWithStyle("REC", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
							widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
						),
						nil,
						nil,
						nil,
						container.NewVScroll(entries),
					),
				),
			)

			sh := &desktop.CustomShortcut{
				KeyName:  fyne.KeyW,
				Modifier: fyne.KeyModifierShortcutDefault,
			}

			m.Window.Canvas().AddShortcut(
				sh,
				func(shortcut fyne.Shortcut) {
					m.Window.Canvas().RemoveShortcut(sh)
					popup.Hide()
					m.refreshHistory()
				},
			)

			popup = widget.NewModalPopUp(content, m.Window.Canvas())
			popup.Resize(m.Window.Canvas().Size().Subtract(fyne.NewSize(100, 100)))
			popup.Show()
		}))
	}
}

func (m *Main) createHistoryEntry(entry *ent.ValueHistory, popup *widget.PopUp) fyne.CanvasObject {
	createdAt := widget.NewLabel(entry.CreatedAt.Format("02/01-06 15:04:05"))

	uecBytes := widget.NewLabel(fmt.Sprintf("%02X", entry.UEC))
	uecBytes.Selectable = true

	uecPWMBytes := widget.NewLabel(fmt.Sprintf("%02X", entry.UECPWM))
	uecPWMBytes.Selectable = true

	recBytes := widget.NewLabel(fmt.Sprintf("%02X", entry.REC))
	recBytes.Selectable = true

	typeLabel := widget.NewLabel(entry.Type.String())

	loadBtn := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		m.setInternalUECState(entry.UEC)
		m.setInternalRECState(entry.REC)
		m.uecPWMBytes = entry.UECPWM
		m.uecPWM.SetText(fmt.Sprintf("%02X", entry.UECPWM))
		log.Println(popup)
		popup.Hide()
	})
	loadBtn.Importance = widget.LowImportance

	return container.NewGridWithColumns(6, typeLabel, createdAt, uecBytes, uecPWMBytes, recBytes, loadBtn)

}
