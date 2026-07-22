package window

import (
	"fyne.io/fyne/v2/widget"
)

func (m *Main) createTech2PortSelector() {
	m.tech2PortSelector = widget.NewSelect([]string{"Not supported"}, func(s string) {
		m.portName = s
	})
	m.tech2PortSelector.PlaceHolder = "Select Tech2 Port"
}

func checkAdmin() (bool, error) {
	return false, nil
}
