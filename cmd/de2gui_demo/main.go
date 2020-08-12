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

	s.OnTick = func(s *de2gui.UIState, final bool) {
		// the caller has to maintain the simulation tick #
		s.Tick++

		// show the tick number on the red LEDs
		if final {
			s.SetLEDR(uint32(s.Tick))
		}
	}

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
