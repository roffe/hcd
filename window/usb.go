package window

import (
	"fmt"

	"go.bug.st/serial/enumerator"
)

func (m *Main) listPorts() []string {
	var portsList []string
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		m.writeOutput(err.Error())
		return []string{}
	}
	if len(ports) == 0 {
		m.writeOutput("No serial ports found!")
		//return []string{}
	}
	for _, port := range ports {
		m.writeOutput(fmt.Sprintf("Found port: %s", port.Name))
		if port.IsUSB {
			m.writeOutput(fmt.Sprintf("  USB ID     %s:%s", port.VID, port.PID))
			m.writeOutput(fmt.Sprintf("  USB serial %s", port.SerialNumber))
			portsList = append(portsList, port.Name)
		}
	}

	return portsList
}
