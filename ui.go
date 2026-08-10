// SPDX-License-Identifier: MIT

package main

import (
	"image/color"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// statusIdle is shown when the editor has nothing more interesting to say.
const statusIdle = "Ctrl/Cmd+S to save · macOS: Fn Fn to dictate"

// UI is the whole editor screen: a text area plus a status line.
type UI struct {
	theme  *material.Theme
	editor widget.Editor
	doc    Document
	status string
}

// NewUI builds the editor UI, optionally opening path on startup.
func NewUI(path string) *UI {
	ui := &UI{theme: material.NewTheme(), status: statusIdle}
	// A code-ish editor wants a monospace face and real newlines, not submits.
	ui.editor.SingleLine = false
	ui.editor.Submit = false
	ui.editor.WrapPolicy = text.WrapHeuristically

	if path != "" {
		content, err := ui.doc.Load(path)
		if err != nil {
			ui.status = err.Error()
		} else {
			ui.editor.SetText(content)
		}
	}
	return ui
}

// Title is the caption for the OS window.
func (ui *UI) Title() string { return ui.doc.Title() + " — HelloWorldEditor" }

// save persists the buffer and reports the outcome on the status line.
func (ui *UI) save() {
	if err := ui.doc.Save(ui.editor.Text()); err != nil {
		ui.status = err.Error()
		return
	}
	ui.status = "Saved " + ui.doc.Name()
}

// handleShortcuts drains global key events. Registering the filter every frame
// is the Gio idiom: filters live for exactly one frame.
func (ui *UI) handleShortcuts(gtx layout.Context) {
	saveKey := key.Filter{Name: "S", Required: key.ModShortcut}
	for {
		ev, ok := gtx.Event(saveKey)
		if !ok {
			return
		}
		if e, ok := ev.(key.Event); ok && e.State == key.Press {
			ui.save()
		}
	}
}

// Layout draws one frame.
func (ui *UI) Layout(gtx layout.Context) layout.Dimensions {
	ui.handleShortcuts(gtx)

	// Any edit (including text injected by macOS Dictation) may change the
	// dirty state, so re-sync against the last saved content.
	for {
		ev, ok := ui.editor.Update(gtx)
		if !ok {
			break
		}
		if _, isChange := ev.(widget.ChangeEvent); isChange {
			ui.doc.Sync(ui.editor.Text())
			ui.status = statusIdle
		}
	}

	// The editor should own the keyboard as soon as the window opens, so
	// dictation and typing land in the buffer without a click first.
	if !gtx.Focused(&ui.editor) {
		gtx.Execute(key.FocusCmd{Tag: &ui.editor})
	}

	paint.Fill(gtx.Ops, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF})

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				ed := material.Editor(ui.theme, &ui.editor, "Start typing — or dictate…")
				ed.TextSize = unit.Sp(15)
				ed.Font.Typeface = "monospace"
				return ed.Layout(gtx)
			})
		}),
		layout.Rigid(ui.layoutStatus),
	)
}

// layoutStatus draws the bottom bar: file name on the left, message on the right.
// The message is elided rather than clipped when the window gets narrow.
func (ui *UI) layoutStatus(gtx layout.Context) layout.Dimensions {
	bg := color.NRGBA{R: 0xF0, G: 0xF0, B: 0xF4, A: 0xFF}
	macro := op.Record(gtx.Ops)
	dims := layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		name := material.Body2(ui.theme, ui.doc.Title())
		name.MaxLines = 1
		msg := material.Body2(ui.theme, ui.status)
		msg.MaxLines = 1
		msg.Alignment = text.End
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
			layout.Rigid(name.Layout),
			layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
			layout.Flexed(1, msg.Layout),
		)
	})
	call := macro.Stop()

	// Paint the bar background behind the text we just recorded.
	defer clip.Rect{Max: dims.Size}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, bg)
	call.Add(gtx.Ops)
	return dims
}

// compile-time nudge: Editor must remain a valid event tag.
var _ event.Tag = (*widget.Editor)(nil)
