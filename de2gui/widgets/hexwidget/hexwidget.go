// Package hexwidget implements a 7-segment style hexadecimal display
package hexwidget

import (
	"image"
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

// size in pixels of the hex widget
var hexHeight float32 = 75.0
var hexWidth float32 = hexHeight * (7.5 / 14.0)

// slant angle
var hexOffset float32 = 0.1 * hexWidth

var hexSegmentWidth float32 = 0.2 * hexWidth
var hexSegmentVLength float32 = (9.14 / (2 * 14)) * hexHeight
var hexSegmentHLength float32 = (4.8 / 7.5) * hexWidth

var hexOnColor color.RGBA = color.RGBA{200, 25, 25, 255}
var hexOffColor color.RGBA = color.RGBA{25, 15, 15, 64}

type hexRenderer struct {
	hex            *HexWidget
	segmentObjects []fyne.CanvasObject
}

func (h *hexRenderer) MinSize() fyne.Size {
	return fyne.NewSize(int(hexWidth)+theme.Padding()*2, int(hexHeight)+theme.Padding()*2)
}

func (h *hexRenderer) Layout(size fyne.Size) {
}

func (h *hexRenderer) ApplyTheme() {
}

func (h *hexRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (h *hexRenderer) Refresh() {
	for i, v := range h.segmentObjects {
		v.(*canvas.Line).StrokeColor = h.hex.getSegmentColor(i)
		canvas.Refresh(v)
	}
}

func (h *hexRenderer) Destroy() {
}

func (h *hexRenderer) Objects() []fyne.CanvasObject {
	return h.segmentObjects
}

// HexWidget represents a 7-segment hexadecimal display. The segments
// of the display mapped active-low onto 7 state bits, with segment 0 in
// the least significant bit.
//
//       0
//     -----
//    |     |
//  5 |     | 1
//    |  6  |
//     -----
//    |     |
//  4 |     | 2
//    |  3  |
//     -----
type HexWidget struct {
	widget.BaseWidget
	segments uint8
}

func ptToPos(pt image.Point) fyne.Position {
	return fyne.NewPos(pt.X, pt.Y)
}

func setLineEndpoints(l *canvas.Line, pt1, pt2 image.Point) {
	l.Move(fyne.NewPos(pt1.X, pt1.Y))
	l.Resize(fyne.NewSize(pt2.X-pt1.X, pt2.Y-pt1.Y))
}

func (h *HexWidget) getSegmentColor(segno int) color.RGBA {
	if (h.segments & (1 << segno)) == 0 {
		return hexOnColor
	}

	return hexOffColor
}

// CreateRenderer implements fyne.Widget
func (h *HexWidget) CreateRenderer() fyne.WidgetRenderer {
	seg0 := canvas.NewLine(hexOffColor)
	seg1 := canvas.NewLine(hexOffColor)
	seg2 := canvas.NewLine(hexOffColor)
	seg3 := canvas.NewLine(hexOffColor)
	seg4 := canvas.NewLine(hexOffColor)
	seg5 := canvas.NewLine(hexOffColor)
	seg6 := canvas.NewLine(hexOffColor)

	seg0.StrokeWidth = float32(hexSegmentWidth / 2)
	seg1.StrokeWidth = float32(hexSegmentWidth / 2)
	seg2.StrokeWidth = float32(hexSegmentWidth / 2)
	seg3.StrokeWidth = float32(hexSegmentWidth / 2)
	seg4.StrokeWidth = float32(hexSegmentWidth / 2)
	seg5.StrokeWidth = float32(hexSegmentWidth / 2)
	seg6.StrokeWidth = float32(hexSegmentWidth / 2)

	pos := image.Pt(0, 0)

	// pt0Center := image.Pt(pos.X+int(hexWidth/2.0+hexOffset), pos.Y+int((hexHeight-2.0*hexSegmentVLength)/2))
	pt0Center := image.Pt(pos.X+int(hexWidth/2.0+hexOffset), pos.Y)
	pt05 := image.Pt(int(float32(pt0Center.X)-(hexSegmentHLength/2)), pt0Center.Y)
	pt01 := image.Pt(int(float32(pt0Center.X)+(hexSegmentHLength/2)), pt0Center.Y)

	pt6Center := image.Pt(pos.X+int(hexWidth/2.0), int(float32(pt0Center.Y)+hexSegmentVLength))
	pt65 := image.Pt(int(float32(pt6Center.X)-(hexSegmentHLength/2)), pt6Center.Y)
	pt61 := image.Pt(int(float32(pt6Center.X)+(hexSegmentHLength/2)), pt6Center.Y)

	pt3Center := image.Pt(pos.X+int(hexWidth/2.0-hexOffset), int(float32(pt0Center.Y)+2*hexSegmentVLength))
	pt34 := image.Pt(int(float32(pt3Center.X)-(hexSegmentHLength/2)), pt3Center.Y)
	pt32 := image.Pt(int(float32(pt3Center.X)+(hexSegmentHLength/2)), pt3Center.Y)

	// segment 0
	setLineEndpoints(seg0, pt05, pt01)

	// segment 1
	setLineEndpoints(seg1, pt01, pt61)

	// segment 2
	setLineEndpoints(seg2, pt61, pt32)

	// segment 3
	setLineEndpoints(seg3, pt32, pt34)

	// segment 4
	setLineEndpoints(seg4, pt34, pt65)

	// segment 5
	setLineEndpoints(seg5, pt65, pt05)

	// segment 6
	setLineEndpoints(seg6, pt65, pt61)

	return &hexRenderer{
		hex:            h,
		segmentObjects: []fyne.CanvasObject{seg0, seg1, seg2, seg3, seg4, seg5, seg6},
	}
}

// NewHexWidget instantiates a new widget instance, with all of the segments
// disabled.
func NewHexWidget() *HexWidget {
	h := &HexWidget{
		segments: 0xff,
	}

	h.ExtendBaseWidget(h)
	return h
}

// Update changes the state of the segments and causes the widget to refresh so
// the changes are visible to the user.
func (h *HexWidget) Update(segments uint8) {
	h.segments = segments
	widget.Refresh(h)
}
