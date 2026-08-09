package window

import (
	_ "embed"
	"image/color"
	_ "image/jpeg"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

//go:embed assets/front.jpg
var frontJPG []byte

//go:embed assets/rear.jpg
var rearJPG []byte

var (
	frontRes = fyne.NewStaticResource("front.jpg", frontJPG)
	rearRes  = fyne.NewStaticResource("rear.jpg", rearJPG)
)

const (
	panelW = 420
	panelH = 240
)

// marker overlays a checkbox chip on a photo with a leader line to the bulb.
// chip and bulb are in design coordinates (panelW x panelH) and are scaled
// with the panel.
type marker struct {
	check *widget.Check
	chip  fyne.Position // top-left of the chip
	bulb  fyne.Position // bulb location on the photo
}

type panelMarker struct {
	marker
	dot  *canvas.Circle
	line *canvas.Line
	bg   *canvas.Rectangle
}

// panelLayout scales the photo (keeping aspect, centered) and re-projects the
// marker positions with it. Chips keep their natural size.
type panelLayout struct {
	img     *canvas.Image
	markers []panelMarker
}

func (p *panelLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(panelW, panelH)
}

func (p *panelLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	s := fyne.Min(size.Width/panelW, size.Height/panelH)
	ox := (size.Width - panelW*s) / 2
	oy := (size.Height - panelH*s) / 2

	p.img.Move(fyne.NewPos(ox, oy))
	p.img.Resize(fyne.NewSize(panelW*s, panelH*s))

	for _, m := range p.markers {
		chipSize := m.check.MinSize()
		chip := fyne.NewPos(ox+m.chip.X*s, oy+m.chip.Y*s)
		bulb := fyne.NewPos(ox+m.bulb.X*s, oy+m.bulb.Y*s)

		m.dot.Resize(fyne.NewSize(10, 10))
		m.dot.Move(fyne.NewPos(bulb.X-5, bulb.Y-5))

		m.line.Position1 = chipAnchor(chip, chipSize, bulb)
		m.line.Position2 = bulb
		m.line.Refresh()

		m.bg.Resize(chipSize)
		m.bg.Move(chip)

		m.check.Resize(chipSize)
		m.check.Move(chip)
	}
}

func lightPanel(res fyne.Resource, markers []marker) fyne.CanvasObject {
	img := canvas.NewImageFromResource(res)
	img.FillMode = canvas.ImageFillStretch

	pl := &panelLayout{img: img}
	objs := []fyne.CanvasObject{img}
	for _, m := range markers {
		dot := canvas.NewCircle(color.NRGBA{R: 0xFF, G: 0xD5, B: 0x2E, A: 0xB4})
		dot.StrokeColor = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xE6}
		dot.StrokeWidth = 1.5

		line := canvas.NewLine(color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xB4})
		line.StrokeWidth = 1.5

		bg := canvas.NewRectangle(color.NRGBA{A: 0xA5})
		bg.StrokeColor = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x5A}
		bg.StrokeWidth = 1

		pl.markers = append(pl.markers, panelMarker{marker: m, dot: dot, line: line, bg: bg})
		objs = append(objs, line, dot, bg, m.check)
	}
	return container.New(pl, objs...)
}

// chipAnchor picks the point on the chip edge the leader line starts from.
func chipAnchor(chip fyne.Position, size fyne.Size, bulb fyne.Position) fyne.Position {
	cx := chip.X + size.Width/2
	cy := chip.Y + size.Height/2
	switch {
	case bulb.Y > chip.Y+size.Height: // bulb below chip
		return fyne.NewPos(cx, chip.Y+size.Height)
	case bulb.Y < chip.Y: // bulb above chip
		return fyne.NewPos(cx, chip.Y)
	case bulb.X > cx: // bulb right of chip
		return fyne.NewPos(chip.X+size.Width, cy)
	default:
		return fyne.NewPos(chip.X, cy)
	}
}

func (m *Main) frontPanel() fyne.CanvasObject {
	return lightPanel(frontRes, []marker{
		{m.uecTurnSignalsCheckbox, fyne.NewPos(6, 4), fyne.NewPos(32, 60)},
		{m.uecLowBeamsCheckbox, fyne.NewPos(110, 4), fyne.NewPos(58, 64)},
		{m.uecHighBeamsCheckbox, fyne.NewPos(224, 4), fyne.NewPos(340, 74)},
		{m.uecParkingLightsCheckbox, fyne.NewPos(316, 105), fyne.NewPos(363, 79)},
		{m.uecFoglightsCheckbox, fyne.NewPos(6, 200), fyne.NewPos(48, 162)},
	})
}

func (m *Main) rearPanel() fyne.CanvasObject {
	return lightPanel(rearRes, []marker{
		{m.recTailLightsCheckbox, fyne.NewPos(6, 4), fyne.NewPos(230, 78)},
		{m.recReverseLightsCheckbox, fyne.NewPos(150, 4), fyne.NewPos(250, 96)},
		{m.recNumberPlateCheckbox, fyne.NewPos(310, 4), fyne.NewPos(350, 64)},
		{m.recStopLightsCheckbox, fyne.NewPos(6, 60), fyne.NewPos(220, 78)},
		{m.recTurnSignalsCheckbox, fyne.NewPos(6, 110), fyne.NewPos(210, 96)},
		{m.recFoglightsCheckbox, fyne.NewPos(280, 200), fyne.NewPos(357, 176)},
	})
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

	footer := container.NewVBox(
		container.New(layout.NewGridLayout(2),
			container.NewMax(),
			container.NewBorder(nil, nil, widget.NewLabel("UEC PWM:"), nil, m.uecPWM),
		),
		container.New(layout.NewGridLayout(2),
			container.NewBorder(nil, nil, widget.NewLabel("UEC Check Freq:"), m.calculatedUECFreq, widget.NewSliderWithData(0, 255, m.uecCheckFreq)),
			container.NewBorder(nil, nil, widget.NewLabel("REC Check Freq:"), m.calculatedRECFreq, widget.NewSliderWithData(0, 255, m.recCheckFreq)),
		),
		m.customValues,
		container.New(layout.NewGridLayout(2), m.getBtn, m.setBtn),
	)

	m.portOverride = &widget.Entry{
		PlaceHolder: "Enter port name",
		OnChanged: func(s string) {
			m.portName = s
			m.App.Preferences().SetString(prefsSelectedPortKey, s)
		},
	}
	m.portOverride.Hide()

	top := container.New(layout.NewGridLayout(2),
		m.adapterSelector,
		container.NewVBox(m.portSelector, m.portOverride, m.tech2PortSelector),
	)

	photos := container.New(layout.NewGridLayout(2), m.frontPanel(), m.rearPanel())

	split := container.NewVSplit(photos, m.output)
	split.Offset = 0.62

	return container.NewBorder(
		top,
		footer,
		nil,
		nil,
		split,
	)
}
