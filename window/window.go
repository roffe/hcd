package window

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/roffe/gocan"
	"github.com/roffe/gocan/adapter"
	"github.com/roffe/gocan/adapter/j2534"
	"github.com/roffe/gocan/pkg/gmlan"
)

var enableLicenseCheck = true

//go:embed hk_white.png
var icon []byte
var iconRes = fyne.NewStaticResource("hk_white.png", icon)

type Main struct {
	App    fyne.App
	Window fyne.Window

	output     *widget.List
	outputData binding.StringList

	dllSelector *widget.Select

	getBtn *widget.Button
	setBtn *widget.Button

	licenseInput *widget.Entry

	uecInput *widget.Entry
	recInput *widget.Entry

	uecCheckFreq binding.Float
	recCheckFreq binding.Float

	calculatedUECFreq *widget.Label
	calculatedRECFreq *widget.Label

	portsList   []j2534.J2534DLL
	selectedDLL string

	uecFoglightsCheckbox   *widget.Check
	uecLowBeamsCheckbox    *widget.Check
	uecHighBeamsCheckbox   *widget.Check
	uecTurnSignalsCheckbox *widget.Check

	recFoglightsCheckbox     *widget.Check
	recTailLightsCheckbox    *widget.Check
	recReverseLightsCheckbox *widget.Check
	recNumberPlateCheckbox   *widget.Check
	recTurnSignalsCheckbox   *widget.Check
	recStopLightsCheckbox    *widget.Check

	customValues *fyne.Container

	ub []byte
	rb []byte

	licenseKey string

	vin []byte
}

const (
	prefsSelectedDLLKey = "selectedDLL"
	timePerTick         = 14.5
)

func NewMainWindow() *Main {
	a := app.NewWithID("com.trionictuning.hcd")
	a.Settings().SetTheme(theme.DefaultTheme())

	w := a.NewWindow("Saab NG9-3 Halogen Check Disabler")
	w.SetIcon(iconRes)
	w.Resize(fyne.NewSize(850, 450))

	m := &Main{
		App:               a,
		Window:            w,
		outputData:        binding.NewStringList(),
		uecInput:          widget.NewEntry(),
		recInput:          widget.NewEntry(),
		uecCheckFreq:      binding.NewFloat(),
		recCheckFreq:      binding.NewFloat(),
		calculatedUECFreq: widget.NewLabel("N/A"),
		calculatedRECFreq: widget.NewLabel("N/A"),
	}

	w.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyTab, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			if m.customValues.Hidden {
				m.customValues.Show()
				a.Preferences().SetBool("customValues", true)
			} else {
				m.customValues.Hide()
				a.Preferences().SetBool("customValues", false)
			}
			w.Content().Refresh()
		},
	)

	m.output = widget.NewListWithData(
		m.outputData,
		func() fyne.CanvasObject {
			w := widget.NewLabel("")
			w.Alignment = fyne.TextAlignCenter
			w.Wrapping = fyne.TextWrapBreak
			w.TextStyle.Monospace = true
			return w
		},
		func(item binding.DataItem, obj fyne.CanvasObject) {
			i := item.(binding.String)
			txt, err := i.Get()
			if err != nil {
				m.writeOutput(err.Error())
				return
			}
			obj.(*widget.Label).SetText(txt)
		},
	)

	m.licenseInput = &widget.Entry{
		PlaceHolder: "Enter license key",
		OnChanged: func(s string) {
			m.licenseKey = s
			m.App.Preferences().SetString("licenseKey", s)
		},
		Password: true,
	}
	m.licenseInput.SetText(m.App.Preferences().StringWithFallback("licenseKey", ""))

	m.uecFoglightsCheckbox = widget.NewCheck("", m.setUECMask(1, 6))
	m.uecLowBeamsCheckbox = widget.NewCheck("", m.setUECMask(1, 5))
	m.uecHighBeamsCheckbox = widget.NewCheck("", m.setUECMask(1, 4))
	m.uecTurnSignalsCheckbox = widget.NewCheck("", m.setUECMask(1, 0))

	m.recTailLightsCheckbox = widget.NewCheck("", m.setRECMask(1, 7))
	m.recFoglightsCheckbox = widget.NewCheck("", m.setRECMask(1, 6))
	m.recStopLightsCheckbox = widget.NewCheck("", m.setRECMask(1, 5))
	m.recReverseLightsCheckbox = widget.NewCheck("", m.setRECMask(1, 4))
	m.recNumberPlateCheckbox = widget.NewCheck("", m.setRECMask(1, 3))
	m.recTurnSignalsCheckbox = widget.NewCheck("", m.setRECMask(2, 3))

	m.uecCheckFreq.AddListener(binding.NewDataListener(func() {
		f, err := m.uecCheckFreq.Get()
		if err != nil {
			m.writeOutput(err.Error())
			return
		}
		if len(m.ub) > 0 {
			m.ub[0] = byte(f)
		}
		if freq := freqToSeconds(f); freq > 0 {
			m.calculatedUECFreq.SetText(fmt.Sprintf("%.02fs", freq))
		} else {
			m.calculatedUECFreq.SetText("OFF")
		}

		m.uecInput.SetText(strings.ToUpper(hex.EncodeToString(m.ub)))
	}))

	m.recCheckFreq.AddListener(binding.NewDataListener(func() {
		f, err := m.recCheckFreq.Get()
		if err != nil {
			m.writeOutput(err.Error())
			return
		}
		if len(m.rb) > 0 {
			m.rb[0] = byte(f)
		}
		if freq := freqToSeconds(f); freq > 0 {
			m.calculatedRECFreq.SetText(fmt.Sprintf("%.02fs", freq))
		} else {
			m.calculatedRECFreq.SetText("OFF")
		}
		m.recInput.SetText(strings.ToUpper(hex.EncodeToString(m.rb)))
	}))

	m.uecInput.OnChanged = func(s string) {
		if len(s)%2 == 0 {
			uecData, err := hex.DecodeString(s)
			if err != nil {
				m.writeOutput("UEC: " + err.Error())
				return
			}
			m.setInternalUECState(uecData)
		}
	}

	m.recInput.OnChanged = func(s string) {
		if len(s)%2 == 0 {
			recData, err := hex.DecodeString(s)
			if err != nil {
				m.writeOutput("REC: " + err.Error())
				return
			}
			m.setInternalRECState(recData)
		}
	}

	m.portsList = j2534.FindDLLs()
	var adapters []string
	for _, p := range m.portsList {
		adapters = append(adapters, p.Name)
	}

	m.dllSelector = &widget.Select{
		Alignment: fyne.TextAlignCenter,
		Options:   adapters,
		OnChanged: func(s string) {
			for _, p := range m.portsList {
				if p.Name == s {
					m.selectedDLL = p.FunctionLibrary
					m.App.Preferences().SetString(prefsSelectedDLLKey, p.Name)
					break
				}
			}
		},
	}

	if selected := m.App.Preferences().String(prefsSelectedDLLKey); selected != "" {
		m.dllSelector.SetSelected(selected)
	}

	m.setBtn = widget.NewButton("Save configuration", m.setState)
	m.setBtn.Icon = theme.DocumentSaveIcon()
	m.getBtn = widget.NewButton("Load configuration", func() {
		if err := m.getStates(); err != nil {
			m.writeOutput(err.Error())
		}
	})
	m.getBtn.Icon = theme.DownloadIcon()

	w.SetContent(m.Layout())
	return m
}

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

	header := container.New(
		layout.NewGridLayout(1),
		container.NewBorder(nil, nil, widget.NewLabel("License:"), nil, m.licenseInput),
	)

	header.Hidden = !enableLicenseCheck

	footer := container.NewVBox(
		container.New(layout.NewGridLayout(2),
			container.NewMax(
				container.NewBorder(nil, nil, widget.NewLabel("UEC check interval:"), m.calculatedUECFreq, widget.NewSliderWithData(0, 255, m.uecCheckFreq)),
				widget.NewLabel(""),
			),
			container.NewMax(
				container.NewBorder(nil, nil, widget.NewLabel("REC check interval:"), m.calculatedRECFreq, widget.NewSliderWithData(0, 255, m.recCheckFreq)),
			),
		),
		m.customValues,
		container.New(layout.NewGridLayout(2), m.getBtn, m.setBtn),
	)

	return container.NewBorder(
		nil,
		footer,
		nil,
		nil,
		&container.Split{
			Horizontal: true,
			Offset:     0.8,
			Leading: container.NewBorder(
				header,
				nil,
				nil,
				nil,
				m.output,
			),
			Trailing: container.NewVBox(
				m.dllSelector,
				widget.NewForm(
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

func (m *Main) initCAN() (*gocan.Client, error) {
	if m.selectedDLL == "" {
		return nil, fmt.Errorf("No adapter selected") //lint:ignore ST1005 ignore this error
	}
	dev, err := adapter.New(
		"j2534",
		&gocan.AdapterConfig{
			Port:         m.selectedDLL,
			PortBaudrate: 0,
			CANRate:      33.3,
			CANFilter:    []uint32{0x64f, 0x649},
			Output: func(s string) {
				m.writeOutput(s)
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return gocan.New(context.Background(), dev)
}

func (m *Main) disableButtons() {
	m.getBtn.Disable()
	m.setBtn.Disable()

}

func (m *Main) enableButtons() {
	m.getBtn.Enable()
	m.setBtn.Enable()
}

func (m *Main) writeOutput(s string) {
	m.outputData.Append(s)
	m.output.Refresh()
	m.output.ScrollToBottom()
}

func (m *Main) getVIN(gm *gmlan.Client) error {
	var err error
	m.vin, err = gm.ReadDataByIdentifier(context.TODO(), 0x90)
	if err != nil {
		return err
	}
	return m.checkLicense()
}

func (m *Main) getStates() error {
	m.disableButtons()
	defer m.enableButtons()

	c, err := m.initCAN()
	if err != nil {
		return err
	}
	defer c.Close()

	// UEC
	uec := gmlan.New(c, 0x24F, 0x64f)

	if err := m.getVIN(uec); err != nil {
		return err
	}

	//m.writeOutput(fmt.Sprintf("VIN: %s", vin))

	uecBytes, err := uec.ReadDataByIdentifier(context.TODO(), 0x4D)
	if err != nil {
		return err
	}
	m.setInternalUECState(uecBytes)

	// REC
	rec := gmlan.New(c, 0x249, 0x649)
	recBytes, err := rec.ReadDataByIdentifier(context.TODO(), 0x45)
	if err != nil {
		return err
	}
	m.setInternalRECState(recBytes)
	m.writeOutput("Read configuration from car successfully")
	return nil
}

func (m *Main) setInternalUECState(data []byte) {
	if len(data) < 4 {
		return
	}
	m.ub = data
	if data[1]&0x40 == 0x40 {
		m.uecFoglightsCheckbox.SetChecked(true)
	} else {
		m.uecFoglightsCheckbox.SetChecked(false)
	}
	if data[1]&0x20 == 0x20 {
		m.uecLowBeamsCheckbox.SetChecked(true)
	} else {
		m.uecLowBeamsCheckbox.SetChecked(false)
	}
	if data[1]&0x10 == 0x10 {
		m.uecHighBeamsCheckbox.SetChecked(true)
	} else {
		m.uecHighBeamsCheckbox.SetChecked(false)
	}
	if data[1]&0x01 == 0x01 {
		m.uecTurnSignalsCheckbox.SetChecked(true)
	} else {
		m.uecTurnSignalsCheckbox.SetChecked(false)
	}
	if err := m.uecCheckFreq.Set(float64(data[0])); err != nil {
		m.writeOutput(err.Error())
	}
	m.uecInput.SetText(strings.ToUpper(hex.EncodeToString(m.ub)))
}

func (m *Main) setInternalRECState(data []byte) {
	if len(data) != 6 {
		return
	}
	m.rb = data
	if err := m.recCheckFreq.Set(float64(m.rb[0])); err != nil {
		m.writeOutput(err.Error())
	}
	if data[1]&0x80 == 0x80 {
		m.recTailLightsCheckbox.SetChecked(true)
	} else {
		m.recTailLightsCheckbox.SetChecked(false)
	}
	if data[1]&0x40 == 0x40 {
		m.recFoglightsCheckbox.SetChecked(true)
	} else {
		m.recFoglightsCheckbox.SetChecked(false)
	}
	if data[1]&0x20 == 0x20 {
		m.recStopLightsCheckbox.SetChecked(true)
	} else {
		m.recStopLightsCheckbox.SetChecked(false)
	}
	if data[1]&0x10 == 0x10 {
		m.recReverseLightsCheckbox.SetChecked(true)
	} else {
		m.recReverseLightsCheckbox.SetChecked(false)
	}
	if data[1]&0x08 == 0x08 {
		m.recNumberPlateCheckbox.SetChecked(true)
	} else {
		m.recNumberPlateCheckbox.SetChecked(false)
	}

	if data[2]&0x08 == 0x08 {
		m.recTurnSignalsCheckbox.SetChecked(true)
	} else {
		m.recTurnSignalsCheckbox.SetChecked(false)
	}

	m.recInput.SetText(strings.ToUpper(hex.EncodeToString(m.rb)))
}

func (m *Main) setState() {
	m.disableButtons()
	defer m.enableButtons()

	c, err := m.initCAN()
	if err != nil {
		m.writeOutput(err.Error())
		return
	}
	defer c.Close()

	// UEC
	uec := gmlan.New(c, 0x24F, 0x64f)

	if err := m.getVIN(uec); err != nil {
		m.writeOutput(err.Error())
		return
	}

	if len(m.ub) < 4 || len(m.rb) != 6 {
		dialog.ShowError(errors.New("Load car configuration first"), m.Window) //lint:ignore ST1005 ignore this
		return
	}

	if err := uec.WriteDataByIdentifier(context.TODO(), 0x4D, m.ub); err != nil {
		m.outputData.Append("UEC: " + err.Error())
		return
	}

	// REC
	rec := gmlan.New(c, 0x249, 0x649)
	if err := rec.WriteDataByIdentifier(context.TODO(), 0x45, m.rb); err != nil {
		m.outputData.Append("REC: " + err.Error())
		return
	}

	//m.writeOutput("Wrote configuration to car successfully")
	dialog.ShowInformation("Success", "Wrote configuration to car successfully", m.Window)
}

func (m *Main) setUECMask(bytePos, pos uint) func(v bool) {
	return func(v bool) {
		if len(m.ub) < 2 {
			return
		}
		m.ub[bytePos] = setBit(m.ub[bytePos], pos, v)
		m.uecInput.SetText(strings.ToUpper(hex.EncodeToString(m.ub)))
	}
}

func (m *Main) setRECMask(bytePos, pos uint) func(v bool) {
	return func(v bool) {
		if len(m.rb) < 2 {
			return
		}
		m.rb[bytePos] = setBit(m.rb[bytePos], pos, v)
		m.recInput.SetText(strings.ToUpper(hex.EncodeToString(m.rb)))
	}
}

func setBit(value byte, position uint, bitValue bool) byte {
	if position > 7 {
		panic("position must be between 0 and 7")
	}
	// Shift 1 to the left by the position to create a mask
	var mask byte = 1 << position
	// Set the bit at the given position to 1 or 0
	if bitValue {
		value |= mask // Set bit to 1
	} else {
		value &= ^mask // Set bit to 0
	}
	return value
}

func CCalc(b []byte) string {
	crc := crc32.Checksum(b, crc32.MakeTable(0xDEADBEEF)) ^ 0xc0ffee
	var buff bytes.Buffer
	if err := binary.Write(&buff, binary.BigEndian, crc); err != nil {
		log.Fatal(err)
	}
	str := hex.EncodeToString(buff.Bytes())
	var val int
	for i, c := range b {
		val += int(c) ^ (int(c) - i)
	}
	return fmt.Sprintf("hk-%s%s%d", str[4:8], str[0:4], val)
}

func (m *Main) checkLicense() error {
	if enableLicenseCheck {
		if strings.HasPrefix(m.licenseKey, "mk-") {
			return nil
		}

		if len(m.vin) == 0 {
			return fmt.Errorf("Load configuration first") //lint:ignore ST1005 ignore this error
		}
		if CCalc(m.vin) != m.licenseKey {
			return fmt.Errorf("Invalid license for VIN: %s", m.vin) //lint:ignore ST1005 ignore this error
		}
		return nil
	}
	return nil
}

func freqToSeconds(f float64) float64 {
	return timePerTick * f / 1000
}
