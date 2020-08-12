// package ledwidget defines a GIU widget that mimicss the appearance of the
// DE2-115 LED groups.
package ledwidget

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

var ledRadius = 5
var ledBoxSize = 15 // pading "box" around the LED

type ledRenderer struct {
	led        *LedWidget
	ledObjects []fyne.CanvasObject
}

func (l *ledRenderer) MinSize() fyne.Size {
	return fyne.NewSize(l.led.count*ledBoxSize+theme.Padding()*2, int(ledBoxSize+theme.Padding()*2))
}

func (l *ledRenderer) Layout(size fyne.Size) {
}

func (l *ledRenderer) ApplyTheme() {
}

func (l *ledRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (l *ledRenderer) Refresh() {
	for i, v := range l.ledObjects {
		v.(*canvas.Circle).StrokeColor = l.led.getLedColor(i)
		v.(*canvas.Circle).FillColor = l.led.getLedColor(i)
		canvas.Refresh(v)
	}
}

func (l *ledRenderer) Destroy() {
}

func (l *ledRenderer) Objects() []fyne.CanvasObject {
	return l.ledObjects
}

type LedWidget struct {
	widget.BaseWidget
	state    uint32
	count    int
	onColor  color.RGBA
	offColor color.RGBA
}

func (l *LedWidget) Mask() uint32 {
	mask := uint32(0)
	for i := 0; i < l.count; i++ {
		mask = (mask << 1) | 1
	}
	return mask
}

func (l *LedWidget) State() uint32 {
	return l.state
}

func (l *LedWidget) Update(newstate uint32) {
	l.state = newstate & l.Mask()
	widget.Refresh(l)
}

func (l *LedWidget) getLedColor(i int) color.RGBA {
	i = l.count - i - 1
	if ((1 << i) & l.state) == 0 {
		return l.offColor
	} else {
		return l.onColor
	}
}

func (l *LedWidget) Tapped(ev *fyne.PointEvent) {
}

func (l *LedWidget) TappedSecondary(ev *fyne.PointEvent) {
}

func (l *LedWidget) CreateRenderer() fyne.WidgetRenderer {
	r := &ledRenderer{
		led:        l,
		ledObjects: make([]fyne.CanvasObject, 0),
	}

	for i := 0; i < l.count; i++ {
		led := canvas.NewCircle(l.offColor)

		// top-left corner of circle's bounding box
		led.Move(fyne.Position{
			theme.Padding() + i*ledBoxSize + ledRadius,
			theme.Padding() + ledRadius,
		})

		led.Resize(fyne.Size{ledRadius * 2, ledRadius * 2})

		r.ledObjects = append(r.ledObjects, led)
	}

	return r
}

func NewLedWidget(count int, onColor, offColor color.RGBA) *LedWidget {
	l := &LedWidget{
		state:    0,
		count:    count,
		onColor:  onColor,
		offColor: offColor,
	}
	l.ExtendBaseWidget(l)
	return l
}
