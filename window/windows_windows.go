package window

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func (m *Main) createTech2PortSelector() {
	m.tech2PortSelector = &widget.Select{
		Alignment: fyne.TextAlignCenter,
		Options:   m.listPorts(),
		OnChanged: func(s string) {
			p, err := strconv.Atoi(strings.TrimPrefix(s, "COM"))
			if err != nil {
				m.writeOutput(err.Error())
				return
			}
			isadm, err := checkAdmin()
			if err != nil {
				m.writeOutput(err.Error())
				return
			}
			if !isadm {
				m.writeOutput("Please run as administrator to change Tech2 port")
				return
			}
			k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\PassThruSupport.04.04\GM - Tech2`, registry.ALL_ACCESS|registry.WOW64_32KEY)
			if err != nil {

				m.writeOutput(err.Error())
				return
			}
			m.writeOutput(fmt.Sprintf("Setting Tech2 port to COM%d", p))
			if err := k.SetDWordValue("Port", uint32(p)); err != nil {
				m.writeOutput(err.Error())
				return
			}
		},
	}
	m.tech2PortSelector.Hide()
}

func checkAdmin() (bool, error) {
	var sid *windows.SID

	// Although this looks scary, it is directly copied from the
	// official windows documentation. The Go API for this is a
	// direct wrap around the official C++ API.
	// See https://docs.microsoft.com/en-us/windows/desktop/api/securitybaseapi/nf-securitybaseapi-checktokenmembership
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false, fmt.Errorf("SID Error: %w", err)
	}
	defer windows.FreeSid(sid)

	// This appears to cast a null pointer so I'm not sure why this
	// works, but this guy says it does and it Works for Me™:
	// https://github.com/golang/go/issues/28804#issuecomment-438838144
	token := windows.Token(0)

	member, err := token.IsMember(sid)
	if err != nil {
		return false, fmt.Errorf("Token Membership Error: %w", err) //lint:ignore ST1005 ignore error string
	}

	// Also note that an admin is _not_ necessarily considered
	// elevated.
	// For elevation see https://github.com/mozey/run-as-admin
	if member {
		return true, nil
	}

	if token.IsElevated() {
		return true, nil
	}
	return false, nil
}
