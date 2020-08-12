// This example application creates a window with a DE2GUI instance int it.
// The red LEDs are used to show the tick number.
package main

import (
	"fmt"

	"fyne.io/fyne/app"

	"github.com/herclab/de2gui/de2gui"
)

func main() {
	app := app.New()
	w := app.NewWindow("de2gui demo")

	w.SetMaster()

	s := de2gui.NewUIState()

	s.OnKEY = func(s *de2gui.UIState) {
		fmt.Printf("KEY pressed, key state is: 0x%x\n", s.KEY())
	}

	s.OnSW = func(s *de2gui.UIState) {
		fmt.Printf("SW changed, switch state is 0x%x\n", s.SW())
	}

	// It is important to define an OnTick function, as de2gui does not
	// increment it's internal tick count on it's own.
	s.OnTick = func(s *de2gui.UIState, final bool) {
		// the caller has to maintain the simulation tick #
		s.Tick++

		// show the tick number on the red LEDs
		if final {
			s.SetLEDR(uint32(s.Tick))
		}
	}

	// You probably want to define an OnRest() as well, to allow your
	// user to re-start the application in place.
	s.OnReset = func(s *de2gui.UIState) {
		// the caller handles any reset operations that need to happen
		// too
		s.Tick = 0
		s.SetLEDR(0)
		s.SetLEDG(0)
		s.ClearFutures()
		s.ClearSW()
		for i := 0; i < 8; i++ {
			s.SetHEX(i, 0xff)
		}
	}

	w.SetContent(s.FyneObject())

	w.ShowAndRun()
}
