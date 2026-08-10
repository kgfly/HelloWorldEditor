// SPDX-License-Identifier: MIT

// Command helloworldeditor is a minimal cross-platform text editor built with Gio.
//
// It runs on macOS, Linux and Windows from a single source tree. On macOS the
// window is a real NSTextInputClient, so system Dictation (voice typing) works
// out of the box — no extra code, no extra permissions.
//
// Usage:
//
//	helloworldeditor [file]
package main

import (
	"fmt"
	"os"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/unit"
)

func main() {
	var path string
	if args := os.Args[1:]; len(args) > 0 {
		path = args[0]
	}

	go func() {
		if err := run(path); err != nil {
			fmt.Fprintln(os.Stderr, "helloworldeditor:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

// run owns the window and its event loop.
func run(path string) error {
	ui := NewUI(path)

	w := new(app.Window)
	w.Option(
		app.Title(ui.Title()),
		app.Size(unit.Dp(900), unit.Dp(620)),
	)

	var ops op.Ops
	title := ui.Title()
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			ui.Layout(gtx)
			e.Frame(gtx.Ops)

			// Keep the OS title in sync with the dirty marker.
			if t := ui.Title(); t != title {
				title = t
				w.Option(app.Title(t))
			}
		}
	}
}
