package window

import (
	"fmt"
	"os"
	"strings"
)

func GetMachineUUID() (string, error) {
	data, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		return "", err
	}
	uuid := strings.TrimSpace(string(data))
	if len(uuid) != 32 {
		return "", fmt.Errorf("machine-id not valid")
	}
	formattedUUID := fmt.Sprintf("%s-%s-%s-%s-%s", uuid[0:8], uuid[8:12], uuid[12:16], uuid[16:20], uuid[20:32])
	return strings.ToUpper(formattedUUID), nil
}
