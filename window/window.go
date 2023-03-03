package window

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
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
	"github.com/roffe/gocan/adapter/j2534"
	"github.com/roffe/gocan/pkg/gmlan"
	"golang.org/x/sys/windows/registry"
)

//go:embed hk_white.png
var icon []byte
var iconRes = fyne.NewStaticResource("hk_white.png", icon)

type Main struct {
	App    fyne.App
	Window fyne.Window

	output     *widget.List
	outputData binding.StringList

	header *fyne.Container

	adapterSelector *widget.Select
	portSelector    *widget.Select

	getBtn *widget.Button
	vinBtn *widget.Button
	setBtn *widget.Button

	licenseInput *widget.Entry

	uecInput *widget.Entry
	recInput *widget.Entry

	uecCheckFreq binding.Float
	recCheckFreq binding.Float

	calculatedUECFreq *widget.Label
	calculatedRECFreq *widget.Label

	j2534List []j2534.J2534DLL
	portName  string

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
	prefsSelectedAdapterKey = "selectedAdapter"
	prefsSelectedPortKey    = "selectedPort"
	timePerTick             = 14.5
)

func NewMainWindow() *Main {
	a := app.NewWithID("com.hirschmannkoxha.hcd")
	a.Settings().SetTheme(theme.DefaultTheme())
	w := a.NewWindow(fmt.Sprintf("Saab NG9-3 Halogen Check Disabler %s Build %d", a.Metadata().Version, a.Metadata().Build))

	w.SetIcon(iconRes)
	w.Resize(fyne.NewSize(850, 500))

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

	w.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyG, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			if val, err := getMachineUUID(); err == nil {
				m.writeOutput("copied machine UUID to clipboard")
				w.Clipboard().SetContent(val)
			}
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
			trimmed := strings.TrimSpace(s)
			m.licenseKey = trimmed
			m.App.Preferences().SetString("licenseKey", trimmed)
			m.licenseInput.SetText(trimmed)
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

	m.j2534List = j2534.FindDLLs()
	adapters := []string{"OBDLink MX"}

	for _, p := range m.j2534List {
		adapters = append(adapters, p.Name)
	}

	m.portSelector = &widget.Select{
		Alignment: fyne.TextAlignCenter,
		Options:   m.listPorts(),
		OnChanged: func(s string) {
			m.portName = s
			m.App.Preferences().SetString(prefsSelectedPortKey, s)
		},
	}

	m.adapterSelector = &widget.Select{
		Alignment: fyne.TextAlignCenter,
		Options:   adapters,
		OnChanged: func(s string) {
			if s == "OBDLink MX" {
				m.portSelector.Show()
				m.App.Preferences().SetString(prefsSelectedAdapterKey, s)
			} else {
				m.portSelector.Hide()
				for _, p := range m.j2534List {
					if p.Name == s {
						m.portName = p.FunctionLibrary
						m.App.Preferences().SetString(prefsSelectedAdapterKey, p.Name)
						break
					}
				}
			}
		},
	}

	if selected := m.App.Preferences().String(prefsSelectedAdapterKey); selected != "" {
		if selected == "OBDLink MX" {
			if selectedPort := m.App.Preferences().String(prefsSelectedPortKey); selectedPort != "" {
				m.portSelector.SetSelected(selectedPort)
			}
			m.portSelector.Show()
		} else {
			m.portSelector.Hide()
		}
		m.adapterSelector.SetSelected(selected)
	}

	m.setBtn = &widget.Button{
		Text:     "Save configuration",
		OnTapped: m.setStates,
		Icon:     theme.DocumentSaveIcon(),
	}

	m.vinBtn = &widget.Button{
		Text:     "Copy VIN",
		OnTapped: m.displayVin,
		Icon:     theme.SearchIcon(),
	}

	m.getBtn = &widget.Button{
		Text:     "Load configuration",
		OnTapped: m.getStates,
		Icon:     theme.DownloadIcon(),
	}

	w.SetContent(m.Layout())
	return m
}

func (m *Main) displayVin() {
	c, err := m.initCAN()
	if err != nil {
		m.writeOutput(err.Error())
		return
	}
	defer c.Close()

	if err := m.getVIN(gmlan.New(c, 0x24F, 0x64F), false); err != nil {
		m.writeOutput(err.Error())
		return
	}
	m.Window.Clipboard().SetContent(string(m.vin))
	dialog.ShowInformation("VIN", fmt.Sprintf("VIN: %s copied to Clipboard", m.vin), m.Window)
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

	m.header = container.NewBorder(nil, nil, widget.NewLabel("License:"), nil, m.licenseInput)

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
		container.New(layout.NewGridLayout(3), m.getBtn, m.vinBtn, m.setBtn),
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
				m.header,
				nil,
				nil,
				nil,
				m.output,
			),
			Trailing: container.NewVBox(
				m.adapterSelector,
				m.portSelector,
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

func (m *Main) disableButtons() {
	m.getBtn.Disable()
	m.vinBtn.Disable()
	m.setBtn.Disable()

}

func (m *Main) enableButtons() {
	m.getBtn.Enable()
	m.vinBtn.Enable()
	m.setBtn.Enable()
}

func (m *Main) writeOutput(s string) {
	m.outputData.Append(s)
	m.output.ScrollToBottom()
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
	if len(m.vin) == 0 {
		return fmt.Errorf("Load configuration first") //lint:ignore ST1005 ignore this error
	}

	if uuid, err := getMachineUUID(); err == nil {
		if CCalc([]byte(uuid)) == m.licenseKey {
			return nil
		}
	} else {
		m.writeOutput("Could not read machine UUID: " + err.Error())
	}

	if CCalc(m.vin) != m.licenseKey {
		return fmt.Errorf("Invalid license for VIN: %s", m.vin) //lint:ignore ST1005 ignore this error
	}
	return nil

}

func freqToSeconds(f float64) float64 {
	return timePerTick * f / 1000
}

func getMachineUUID() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return "", err
	}
	defer k.Close()

	uuid, _, err := k.GetStringValue("MachineGuid")
	if err != nil {
		return "", err
	}
	return uuid, nil
}
