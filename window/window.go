package window

import (
	_ "embed"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/roffe/gocan"
	"github.com/roffe/gocan/pkg/gmlan"
	"github.com/roffe/hcd/store"
)

//go:embed logo.png
var icon []byte
var iconRes = fyne.NewStaticResource("logo.png", icon)

type Main struct {
	App    fyne.App
	Window fyne.Window
	DB     *store.Store

	output     *widget.List
	outputData binding.StringList

	header      *fyne.Container
	historyList *fyne.Container

	adapterSelector   *widget.Select
	portSelector      *widget.Select
	portOverride      *widget.Entry
	tech2PortSelector *widget.Select

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

	portName string

	uecParkingLightsCheckbox *widget.Check
	uecFoglightsCheckbox     *widget.Check
	uecLowBeamsCheckbox      *widget.Check
	uecHighBeamsCheckbox     *widget.Check
	uecTurnSignalsCheckbox   *widget.Check

	recFoglightsCheckbox     *widget.Check
	recTailLightsCheckbox    *widget.Check
	recReverseLightsCheckbox *widget.Check
	recNumberPlateCheckbox   *widget.Check
	recTurnSignalsCheckbox   *widget.Check
	recStopLightsCheckbox    *widget.Check

	uecPWM *widget.Entry

	customValues *fyne.Container

	uecPWMBytes []byte
	uecBytes    []byte
	recBytes    []byte

	vin []byte
}

const (
	prefsSelectedAdapterKey = "selectedAdapter"
	prefsSelectedPortKey    = "selectedPort"
	timePerTick             = 14.5
)

func NewMainWindow(db *store.Store) *Main {
	a := app.NewWithID("com.roffe.lightcontrol")
	a.Settings().SetTheme(theme.DefaultTheme())
	w := a.NewWindow(fmt.Sprintf("Saab NG9-3 Light Control %s Build %d", a.Metadata().Version, a.Metadata().Build))

	w.SetIcon(iconRes)

	w.Resize(fyne.NewSize(850, 705))
	//w.SetFixedSize(true)

	w.SetMaster()

	m := &Main{
		App:               a,
		DB:                db,
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
			if val, err := GetMachineUUID(); err == nil {
				m.writeOutput("copied machine UUID to clipboard")
				fyne.CurrentApp().Clipboard().SetContent(val)
			} else {
				m.writeOutput("Error loading uuid: " + err.Error())
			}
		},
	)

	w.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			if m.portOverride.Visible() {
				m.portOverride.Hide()
				m.Window.Content().Refresh()
			} else {
				m.portOverride.Show()
				m.Window.Content().Refresh()
			}
		},
	)

	m.historyList = container.NewVBox()
	m.refreshHistory()

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

	m.uecPWM = &widget.Entry{
		OnChanged: func(s string) {
			if len(s)%2 == 0 {
				b, err := hex.DecodeString(s)
				if err != nil {
					m.writeOutput("invalid PWM hex value")
					return
				}
				m.uecPWMBytes = b
			}
		},
	}

	m.licenseInput = &widget.Entry{
		PlaceHolder: "Enter license key",
		OnChanged: func(s string) {
			trimmed := strings.TrimSpace(s)
			m.App.Preferences().SetString("licenseKey", trimmed)
			m.licenseInput.SetText(trimmed)
		},
		Password: true,
	}
	m.licenseInput.SetText(m.App.Preferences().StringWithFallback("licenseKey", ""))

	m.uecParkingLightsCheckbox = widget.NewCheck("", m.setUECMask(1, 7))
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
		if len(m.uecBytes) > 0 {
			m.uecBytes[0] = byte(f)
		}
		if freq := freqToSeconds(f); freq > 0 {
			m.calculatedUECFreq.SetText(fmt.Sprintf("%.02fs", freq))
		} else {
			m.calculatedUECFreq.SetText("OFF")
		}

		m.uecInput.SetText(strings.ToUpper(hex.EncodeToString(m.uecBytes)))
	}))

	m.recCheckFreq.AddListener(binding.NewDataListener(func() {
		f, err := m.recCheckFreq.Get()
		if err != nil {
			m.writeOutput(err.Error())
			return
		}
		if len(m.recBytes) > 0 {
			m.recBytes[0] = byte(f)
		}
		if freq := freqToSeconds(f); freq > 0 {
			m.calculatedRECFreq.SetText(fmt.Sprintf("%.02fs", freq))
		} else {
			m.calculatedRECFreq.SetText("OFF")
		}
		m.recInput.SetText(strings.ToUpper(hex.EncodeToString(m.recBytes)))
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

	var adapters []string
	for _, p := range gocan.ListAdapters() {
		if p.Capabilities.SWCAN {
			adapters = append(adapters, p.Name)
		}
	}
	sort.Slice(adapters, func(i, j int) bool { return strings.ToLower(adapters[i]) < strings.ToLower(adapters[j]) })

	m.portSelector = &widget.Select{
		Alignment: fyne.TextAlignCenter,
		Options:   m.listPorts(),
		OnChanged: func(s string) {
			m.portName = s
			m.App.Preferences().SetString(prefsSelectedPortKey, s)
		},
	}

	m.tech2PortSelector = &widget.Select{
		Alignment: fyne.TextAlignCenter,
		Options:   m.listPorts(),
		OnChanged: func(s string) {
			p, err := strconv.Atoi(strings.TrimPrefix(s, "COM"))
			if err != nil {
				m.writeOutput(err.Error())
				return
			}
			isadm, err := CheckAdmin()
			if err != nil {
				m.writeOutput(err.Error())
				return
			}
			if !isadm {
				m.writeOutput("Please run as administrator to change Tech2 port")
				return
			}
			if err := SetTech2Port(p); err != nil {
				m.writeOutput(err.Error())
				return
			}

			//k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\PassThruSupport.04.04\GM - Tech2`, registry.ALL_ACCESS|registry.WOW64_32KEY)
			//if err != nil {
			//	m.writeOutput(err.Error())
			//	return
			//}
			//if err := k.SetDWordValue("Port", uint32(p)); err != nil {
			//	m.writeOutput(err.Error())
			//	return
			//}
			m.writeOutput(fmt.Sprintf("Setting Tech2 port to COM%d", p))
		},
	}
	m.tech2PortSelector.Hide()

	m.adapterSelector = &widget.Select{
		Alignment: fyne.TextAlignCenter,
		Options:   adapters,
		OnChanged: func(s string) {
			m.App.Preferences().SetString(prefsSelectedAdapterKey, s)

			for _, adapter := range gocan.ListAdapters() {
				if adapter.Name == s {
					if adapter.RequiresSerialPort {
						m.portSelector.Show()
						m.tech2PortSelector.Hide()
						m.portName = m.portSelector.Selected
						return
					} else if strings.HasPrefix(strings.ToLower(adapter.Name), "tech2") {
						m.tech2PortSelector.Show()
						m.portSelector.Hide()
					} else {
						m.portSelector.Hide()
						m.tech2PortSelector.Hide()
					}
				}
			}
		},
	}

	if selected := m.App.Preferences().String(prefsSelectedAdapterKey); selected != "" {
		for _, adapter := range gocan.ListAdapters() {
			if adapter.Name == selected {

				//if adapter, found := gocan.GetAdapterMap()[selected]; found {
				if adapter.RequiresSerialPort {
					if selectedPort := m.App.Preferences().String(prefsSelectedPortKey); selectedPort != "" {
						m.portSelector.SetSelected(selectedPort)
					}
					m.portSelector.Show()
					m.tech2PortSelector.Hide()
				} else if strings.HasPrefix(strings.ToLower(selected), "tech2") {
					m.tech2PortSelector.Show()
					m.portSelector.Hide()
				} else {
					m.portSelector.Hide()
					m.tech2PortSelector.Hide()
				}
				m.adapterSelector.SetSelected(selected)
			}
		}
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
	fyne.CurrentApp().Clipboard().SetContent(string(m.vin))
	dialog.ShowInformation("VIN", fmt.Sprintf("VIN: %s copied to Clipboard", m.vin), m.Window)
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
		if len(m.uecBytes) < 2 {
			return
		}
		m.uecBytes[bytePos] = setBit(m.uecBytes[bytePos], pos, v)
		m.uecInput.SetText(strings.ToUpper(hex.EncodeToString(m.uecBytes)))
	}
}

func (m *Main) setRECMask(bytePos, pos uint) func(v bool) {
	return func(v bool) {
		if len(m.recBytes) < 2 {
			return
		}
		m.recBytes[bytePos] = setBit(m.recBytes[bytePos], pos, v)
		m.recInput.SetText(strings.ToUpper(hex.EncodeToString(m.recBytes)))
	}
}

func (m *Main) checkLicense() error {
	if len(m.vin) == 0 {
		return fmt.Errorf("Load configuration first") //lint:ignore ST1005 ignore this error
	}

	licenseKey := m.App.Preferences().String("licenseKey")

	if uuid, err := GetMachineUUID(); err == nil {
		if CCalc([]byte(uuid)) == licenseKey {
			return nil
		}
	}

	if CCalc(m.vin) != licenseKey {
		return fmt.Errorf("Invalid license for VIN: %s", m.vin) //lint:ignore ST1005 ignore this error
	}
	return nil

}
