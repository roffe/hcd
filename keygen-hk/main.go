package main

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
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:embed merkel.jpg
var merkel []byte

func main() {
	a := app.NewWithID("com.hcd.keygen")
	w := a.NewWindow("HCD kEyGen v0.1337")
	w.Resize(fyne.NewSize(500, 50))
	a.Settings().SetTheme(theme.DefaultTheme())
	vinInput := widget.NewEntry()
	result := widget.NewEntry()
	img := canvas.NewImageFromResource(fyne.NewStaticResource("merkel.jpg", merkel))
	img.FillMode = canvas.ImageFillContain
	img.ScaleMode = canvas.ImageScaleFastest
	img.SetMinSize(fyne.NewSize(50, 150))
	w.SetContent(container.NewVBox(
		container.NewBorder(
			img,
			nil,
			widget.NewLabel("VIN/UUID:"), widget.NewButton("Generate", func() {
				if len(vinInput.Text) == 17 {
					vinInput.SetText(strings.ToUpper(vinInput.Text))
				}
				result.SetText(CCalc([]byte(vinInput.Text)))
			}),
			vinInput,
		),
		container.NewBorder(
			nil,
			nil,
			widget.NewLabel("License:"),
			nil,
			result,
		),
	))
	w.ShowAndRun()
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
