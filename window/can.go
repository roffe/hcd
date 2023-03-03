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
	if m.portName == "" {
		return nil, fmt.Errorf("No adapter selected") //lint:ignore ST1005 ignore this error
	}

	adapterName := m.adapterSelector.Selected
	if adapterName != "OBDLink MX" {
		adapterName = "J2534"
	}

	dev, err := adapter.New(
		adapterName,
		&gocan.AdapterConfig{
			Port:         m.portName,
			PortBaudrate: 2000000,
			CANRate:      33.3,
			CANFilter:    []uint32{0x64f, 0x649},
			OutputFunc: func(s string) {
				m.writeOutput(s)
			},
			ErrorFunc: func(err error) {
				m.writeOutput("error: " + err.Error())
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return gocan.New(context.Background(), dev)
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
	uec := gmlan.New(c, 0x24F, 0x64f)

	if err := m.getVIN(uec, true); err != nil {
		m.writeOutput(err.Error())
		return
	}

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

	if err := m.getVIN(uec, true); err != nil {
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
