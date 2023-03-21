package window

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (m *Main) Layout() fyne.CanvasObject {
	m.customValues = container.New(layout.NewGridLayout(2),
		container.NewMax(
			container.NewBorder(nil, nil, widget.NewLabel("UEC value:"), nil, m.uecInput),
		),
		container.NewMax(
			container.NewBorder(nil, nil, widget.NewLabel("REC value:"), nil, m.recInput),
		),
	)
	m.customValues.Hidden = !m.App.Preferences().Bool("customValues")

	m.header = container.NewBorder(nil, nil, widget.NewLabel("License:"), nil, m.licenseInput)

	//helpBtn := widget.NewButtonWithIcon("Help", theme.HelpIcon(), func() {
	//	help.NewWindow(m.App)
	//})

	footer := container.NewVBox(
		container.New(layout.NewGridLayout(2),
			container.NewMax(),
			container.NewBorder(nil, nil, widget.NewLabel("UEC PWM:"), nil, m.uecPWM),
		),
		container.New(layout.NewGridLayout(2),
			container.NewBorder(nil, nil, widget.NewLabel("UEC Check Freq:"), m.calculatedUECFreq, widget.NewSliderWithData(0, 255, m.uecCheckFreq)),
			container.NewBorder(nil, nil, widget.NewLabel("REC Check Freq:"), m.calculatedRECFreq, widget.NewSliderWithData(0, 255, m.recCheckFreq)),
		),
		m.customValues,
		container.New(layout.NewGridLayout(3), m.getBtn, m.vinBtn, m.setBtn),
	)

	m.portOverride = &widget.Entry{
		PlaceHolder: "Enter port name",
		OnChanged: func(s string) {
			m.portName = s
			m.App.Preferences().SetString(prefsSelectedPortKey, s)
		},
	}
	m.portOverride.Hide()

	return container.NewBorder(
		nil,
		footer,
		nil,
		nil,
		&container.Split{
			Horizontal: true,
			Offset:     0.8,
			Leading: container.NewBorder(
				m.header,
				nil,
				nil,
				nil,
				m.output,
			),
			Trailing: container.NewVBox(
				m.adapterSelector,
				m.portSelector,
				m.portOverride,
				m.tech2PortSelector,
				widget.NewForm(
					widget.NewFormItem("Front Parking Lights", m.uecParkingLightsCheckbox),
					widget.NewFormItem("Front Fog Lights", m.uecFoglightsCheckbox),
					widget.NewFormItem("Rear Fog Lights", m.recFoglightsCheckbox),
					widget.NewFormItem("Low Beams", m.uecLowBeamsCheckbox),
					widget.NewFormItem("High Beams", m.uecHighBeamsCheckbox),
					widget.NewFormItem("Front Turn Signals", m.uecTurnSignalsCheckbox),
					widget.NewFormItem("Rear Turn Signals", m.recTurnSignalsCheckbox),
					widget.NewFormItem("Tail Lights", m.recTailLightsCheckbox),
					widget.NewFormItem("Stop Lights", m.recStopLightsCheckbox),
					widget.NewFormItem("Reverse Light", m.recReverseLightsCheckbox),
					widget.NewFormItem("License Plate", m.recNumberPlateCheckbox),
				),
			),
		},
	)
}
