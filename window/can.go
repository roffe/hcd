package window

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"fyne.io/fyne/v2/dialog"
	"github.com/roffe/gocan"
	"github.com/roffe/gocan/adapter"
	"github.com/roffe/gocan/pkg/gmlan"
)

func (m *Main) initCAN() (*gocan.Client, error) {
	if m.adapterSelector.Selected == "" {
		return nil, fmt.Errorf("No adapter selected") //lint:ignore ST1005 ignore this error
	}
	device, err := adapter.New(
		m.adapterSelector.Selected,
		&gocan.AdapterConfig{
			Port:         m.portName,
			PortBaudrate: 2000000,
			CANRate:      33.3,
			CANFilter:    []uint32{0x64F, 0x649},
			OnMessage: func(s string) {
				m.writeOutput(s)
			},
			OnError: func(err error) {
				m.writeOutput("err: " + err.Error())
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return gocan.New(context.Background(), device)
}

func (m *Main) getVIN(gm *gmlan.Client, checkLicense bool) error {
	var err error
	m.vin, err = gm.ReadDataByIdentifier(context.TODO(), 0x90)
	if err != nil {
		return err
	}
	if !checkLicense {
		return nil
	}
	return m.checkLicense()
}

func (m *Main) getStates() {
	m.disableButtons()
	defer m.enableButtons()

	c, err := m.initCAN()
	if err != nil {
		m.writeOutput(err.Error())
		return
	}
	defer c.Close()

	// UEC
	uec := gmlan.New(c, 0x24F, 0x64F)

	if err := uec.TesterPresentNoResponseAllowed(); err != nil {
		m.writeOutput("TPNRA: " + err.Error())
	}

	if err := m.getVIN(uec, true); err != nil {
		m.writeOutput(err.Error())
		return
	}
	uecPWMBytes, err := uec.ReadDataByIdentifier(context.TODO(), 0x40)
	if err != nil {
		m.writeOutput(err.Error())
		return
	}
	m.uecPWMBytes = uecPWMBytes
	m.uecPWM.SetText(hex.EncodeToString(uecPWMBytes))

	uecBytes, err := uec.ReadDataByIdentifier(context.TODO(), 0x4D)
	if err != nil {
		m.writeOutput(err.Error())
		return
	}
	m.setInternalUECState(uecBytes)

	// REC
	rec := gmlan.New(c, 0x249, 0x649)
	recBytes, err := rec.ReadDataByIdentifier(context.TODO(), 0x45)
	if err != nil {
		m.writeOutput(err.Error())
		return
	}
	m.setInternalRECState(recBytes)
	m.writeOutput("Read configuration from car successfully")
}

func (m *Main) setInternalUECState(data []byte) {
	if len(data) < 4 {
		return
	}
	m.uecBytes = data
	if data[1]&0x80 == 0x80 {
		m.uecParkingLightsCheckbox.SetChecked(true)
	} else {
		m.uecParkingLightsCheckbox.SetChecked(false)
	}
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
	m.uecInput.SetText(strings.ToUpper(hex.EncodeToString(m.uecBytes)))
}

func (m *Main) setInternalRECState(data []byte) {
	if len(data) != 6 {
		return
	}
	m.recBytes = data
	if err := m.recCheckFreq.Set(float64(m.recBytes[0])); err != nil {
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

	m.recInput.SetText(strings.ToUpper(hex.EncodeToString(m.recBytes)))
}

func (m *Main) setStates() {
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

	if err := uec.TesterPresentNoResponseAllowed(); err != nil {
		m.writeOutput("TPNRA: " + err.Error())
	}

	if err := m.getVIN(uec, true); err != nil {
		m.writeOutput(err.Error())
		return
	}

	if len(m.uecBytes) < 4 || len(m.recBytes) != 6 || len(m.vin) == 0 || len(m.uecPWMBytes) != 2 {
		dialog.ShowError(errors.New("Press \"Load configuration\" first"), m.Window) //lint:ignore ST1005 ignore this
		return
	}

	if err := uec.WriteDataByIdentifier(context.TODO(), 0x4D, m.uecBytes); err != nil {
		m.outputData.Append("UEC: " + err.Error())
		return
	}

	if err := uec.WriteDataByIdentifier(context.TODO(), 0x40, m.uecPWMBytes); err != nil {
		m.outputData.Append("UEC PWM: " + err.Error())
		return
	}

	// REC
	rec := gmlan.New(c, 0x249, 0x649)
	if err := rec.WriteDataByIdentifier(context.TODO(), 0x45, m.recBytes); err != nil {
		m.outputData.Append("REC: " + err.Error())
		return
	}
	m.writeOutput("Wrote configuration to car successfully")
}
