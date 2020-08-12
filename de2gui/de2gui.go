// Package de2gui contains code for providing a graphical facsimile of the
// Terasic DE2-115 development board.
package de2gui

import (
	"fmt"
	"image/color"
	"math/rand"
	"os"
	"strconv"

	"fyne.io/fyne"
	"fyne.io/fyne/widget"

	"github.com/herclab/de2gui/de2gui/widgets/hexwidget"
	"github.com/herclab/de2gui/de2gui/widgets/ledwidget"
)

// UIState contains all of the GUI widgets, and the data needed to interact
// with them.
//
// The UI revolves around the assumption that the underlying simulation runs
// in discrete simulation "ticks". The OnTick callback is called whenever
// the user does something that triggers one or more ticks to occur. The
// downstream GUI is expected to update the Tick field appropriately when
// they handle simulation ticks.
//
// Certain aspects of the UI also deal with events happening in the "future".
// Namely, when a KEY is pressed, it stays "pressed" for a variable number of
// ticks (this is to simulate real hardware, where a button will be pressed by
// a human for many thousands of clock cycles). This is handled by scheduling
// a key release for the future. Futures run when the Tick field equals their
// scheduled time to run and a Tick occurs. Futures run before OnTick is
// called.
type UIState struct {
	// state storage
	key     uint32
	futures map[uint64][]func(*UIState)

	// widgets
	ledrWidget   *ledwidget.LedWidget
	ledrLabel    *widget.Label
	ledgWidget   *ledwidget.LedWidget
	ledgLabel    *widget.Label
	hexWidgets   []*hexwidget.HexWidget
	regLabels    []*widget.Label
	cycleLabel   *widget.Label
	switchChecks []*widget.Check
	switchLabel  *widget.Label
	tickEntry    *widget.Entry
	tickEntryVal int

	widgetTree fyne.CanvasObject

	// The Tick value is displayed to the user as the current tick #, and
	// is also used to determine when to run futures
	Tick uint64

	// OnKEY is run when any key is changed (pressed or released)
	OnKEY func(*UIState)

	// OnSW is run with any SW is changed
	OnSW func(*UIState)

	// OnTick is run when any of the tick-related controls is used.
	//
	// The boolean parameter is used as a performance optimization, it will
	// be true if and only if this it the final tick in a range of many
	// ticks which occur at once. For example, it might be best to call
	// functions like SetHex() only when this parameter is true, to avoid
	// spurious UI updates.
	OnTick func(*UIState, bool)

	// OnReset is run when the reset button is used
	OnReset func(*UIState)
}

const numHex int = 8
const numRedLeds int = 18
const numGreenLeds int = 9
const numSwitches int = 18

// ColorRedActive is the color used for red-colored illuminated parts when they
// are active.
var ColorRedActive color.RGBA = color.RGBA{200, 25, 25, 255}

// ColorRedInactive is the color used for red-colored illuminated parts when
// they are not active.
var ColorRedInactive color.RGBA = color.RGBA{25, 15, 15, 64}

// ColorGreenActive is the color used for green-colored illuminated parts when
// they are active.
var ColorGreenActive color.RGBA = color.RGBA{25, 200, 25, 255}

// ColorGreenInactive is the color used for green-colored illuminated parts
// when they are not active.
var ColorGreenInactive color.RGBA = color.RGBA{15, 25, 15, 64}

// KeyPushMinTime is the minimum number of ticks until a key is released after
// it is pushed
var KeyPushMinTime uint64 = 10

// KeyPushMaxTime is the maximum number of ticks until a key is released after
// it is pushed
var KeyPushMaxTime uint64 = 250

// NewUIState initializes a new instance of the DE2GUI's state object along
// with all of the needed widgets. After calling this, FyneObject() can
// safely be called.
//
// EtraWidgets, if non-nil, will be inserted into the left panel of the
// created GUI elements.
func NewUIState() *UIState {
	s := &UIState{
		futures:      make(map[uint64][]func(*UIState)),
		ledrWidget:   ledwidget.NewLedWidget(numRedLeds, ColorRedActive, ColorRedInactive),
		ledrLabel:    widget.NewLabelWithStyle("(0x00000)", fyne.TextAlignLeading, fyne.TextStyle{false, false, true}),
		ledgWidget:   ledwidget.NewLedWidget(numGreenLeds, ColorGreenActive, ColorGreenInactive),
		ledgLabel:    widget.NewLabelWithStyle("(0x000)", fyne.TextAlignLeading, fyne.TextStyle{false, false, true}),
		hexWidgets:   make([]*hexwidget.HexWidget, numHex),
		cycleLabel:   widget.NewLabel("cycle# --"),
		switchChecks: make([]*widget.Check, numSwitches),
		switchLabel:  widget.NewLabelWithStyle("(0x00000)", fyne.TextAlignLeading, fyne.TextStyle{false, false, true}),
		tickEntry:    widget.NewEntry(),
	}

	// Create the HEX widgets and initialize them to completely off.
	for i := 0; i < numHex; i++ {
		s.hexWidgets[i] = hexwidget.NewHexWidget()
		s.hexWidgets[i].Update(0xff) // remember they are active low
	}

	// now we will set up a container to store the checkboxes used
	// as switches, and initialize the checks themselves
	checkcontainer := widget.NewHBox(widget.NewLabel("SW:"))
	for i := 0; i < numSwitches; i++ {
		s.switchChecks[i] = widget.NewCheck("", func(dummy bool) { s.switchUpdate() })
		checkcontainer.Children = append(checkcontainer.Children, s.switchChecks[i])
	}

	// setup s.tickEntryVal to update when the entry is changed
	s.tickEntry.OnChanged = func(str string) {
		n, err := strconv.Atoi(str)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid tick entry value '%s': %v\n", str, err)
			s.tickEntryVal = 0
		} else {
			s.tickEntryVal = n
		}
	}

	// now we create the structure of the window in proper
	s.widgetTree = widget.NewVBox(
		widget.NewHBox(
			s.hexWidgets[0],
			s.hexWidgets[1],
			s.hexWidgets[2],
			s.hexWidgets[3],
			s.hexWidgets[4],
			s.hexWidgets[5],
			s.hexWidgets[6],
			s.hexWidgets[7],
		),
		widget.NewHBox(
			widget.NewLabel("LEDR:"),
			s.ledrWidget,
			s.ledrLabel,
		),
		widget.NewHBox(
			widget.NewLabel("LEDG:"),
			s.ledgWidget,
			s.ledgLabel,
		),
		checkcontainer,
		widget.NewHBox(
			widget.NewButton("KEY3", func() { s.pushKey(3) }),
			widget.NewButton("KEY2", func() { s.pushKey(2) }),
			widget.NewButton("KEY1", func() { s.pushKey(1) }),
			widget.NewButton("KEY0", func() { s.pushKey(0) }),
		),
		widget.NewHBox(
			s.cycleLabel,
			widget.NewButton("Tick 1", func() { s.tick(1) }),
			widget.NewButton("Tick 10", func() { s.tick(10) }),
			widget.NewButton("Tick 100", func() { s.tick(100) }),
			widget.NewLabel("n="),
			s.tickEntry,
			widget.NewButton("Tick N", func() { s.tick(s.tickEntryVal) }),
			widget.NewButton("Reset", func() {
				if s.OnReset != nil {
					s.OnReset(s)
				}
			}),
		),
	)

	return s
}

// Internal function wired into key presses
func (s *UIState) pushKey(i int) {
	r := uint64(rand.Float64()*float64(KeyPushMaxTime) + float64(KeyPushMinTime))
	release := s.Tick + r
	s.ScheduleFuture(release, func(uistate *UIState) {
		s.releaseKey(i)
	})

	s.key |= (1 << i)

	if s.OnKEY != nil {
		s.OnKEY(s)
	}
}

// Internal function to handle key releases
func (s *UIState) releaseKey(i int) {
	s.key &= ^(1 << i)
	if s.OnKEY != nil {
		s.OnKEY(s)
	}
}

// Internal function wired into switch change callbacks
func (s *UIState) switchUpdate() {
	if s.OnSW != nil {
		s.OnSW(s)
	}
}

// Internal function which handles tick events
func (s *UIState) tick(count int) {

	// don't trigger updates on 0-tick events
	if count == 0 {
		return
	}

	for i := 0; i < count; i++ {
		// handle future that need to run on this tick
		for k, futurelist := range s.futures {
			if s.Tick >= k {
				for _, future := range futurelist {
					future(s)
				}
				delete(s.futures, k)
			}
		}

		if s.OnTick != nil {
			s.OnTick(s, (i+1) >= (count))
		}
	}

	s.cycleLabel.SetText(fmt.Sprintf("cycle# %d", s.Tick))
}

// ClearFutures removes all functions scheduled to run in the future.  You
// almost certainly want to call this in your OnRest() method.
func (s *UIState) ClearFutures() {
	s.futures = make(map[uint64][]func(*UIState))
}

// ClearSW resets all switches to the "off" state. You might want to call
// this in your OnRest() method.
func (s *UIState) ClearSW() {
	for i := 0; i < numSwitches; i++ {
		s.switchChecks[i].Checked = false
		widget.Refresh(s.switchChecks[i])
	}
}

// FyneObject will return a Fyne canvas object which contains all of the
// widgets and such relating to this instance of the UIState. This should be
// suitable for use with Window.SetContent. However for more advanced use
// cases, it can be embedded in a container as needed.
func (s *UIState) FyneObject() fyne.CanvasObject {

	return s.widgetTree
}

// ScheduleFuture will cause the provided callback to be executed whenever
// a tick occurs and s.Tick is at least equal to `when`.
func (s *UIState) ScheduleFuture(when uint64, f func(*UIState)) {
	_, ok := s.futures[when]
	if !ok {
		s.futures[when] = make([]func(*UIState), 0)
	}

	s.futures[when] = append(s.futures[when], f)
}

// SetHEX updates the state of the i-th HEX display. Hex display 0 is the
// rightmost (least significant)
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
//
// Segments are packed into a uint8 as shown in the above diagram. Segments
// are active-low.
func (s *UIState) SetHEX(i int, state uint8) {
	s.hexWidgets[i%numHex].Update(state)
}

// SetLEDR sets the LEDR display. There are 18 red LEDs. The least significant
// bit codes for the rightmost LED. LEDs are active-high. Unused higher order
// bits are ignored.
func (s *UIState) SetLEDR(state uint32) {
	s.ledrWidget.Update(state)
	s.ledrLabel.SetText(fmt.Sprintf("(0x%05x)", s.ledrWidget.State()))
}

// SetLEDG sets the LEDG display. There are 9 green LEDs. the least significant
// bit codes for the rightmost LED. LEDs are active-high. Unused higher order
// bits are ignored.
func (s *UIState) SetLEDG(state uint32) {
	s.ledgWidget.Update(state)
	s.ledgLabel.SetText(fmt.Sprintf("(0x%03x)", s.ledrWidget.State()))
}

// SW gets the current value of the SW(itch) controls. There are 18
// switches. The rightmost switch is assigned to the least-significant bit.
// Unused higher order bits are left as zero.
func (s *UIState) SW() uint32 {
	val := uint32(0)
	for i := 0; i < numSwitches; i++ {
		if s.switchChecks[i].Checked {
			val |= 1 << (numSwitches - 1 - i)
		}
	}
	return val
}

// KEY returns the current value of the KEY controls. There are 4 keys.
// The rightmost key is the least-significant bit. Unused higher order bits
// are left as zero.
func (s *UIState) KEY() uint32 {
	return s.key
}
