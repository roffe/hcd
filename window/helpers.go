package window

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"log"
)

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

func freqToSeconds(f float64) float64 {
	return timePerTick * f / 1000
}
