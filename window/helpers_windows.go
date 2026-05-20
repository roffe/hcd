package window

import "golang.org/x/sys/windows/registry"

func GetMachineUUID() (string, error) {
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
