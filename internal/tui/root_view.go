package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (t *TUI) initRootView() {
	t.root = tview.NewGrid().SetRows(0, 1).SetColumns(0)
	t.root.AddItem(t.pages, 0, 0, 1, 1, 0, 0, true)
	t.root.AddItem(t.statusBar, 1, 0, 1, 1, 0, 0, false)
	t.root.SetInputCapture(t.rootInputHandler)
}

// rootInputHandler captures any key inputs before they are forwarded to the
// currently focused primitive. This is where we control application-wide
// inputs, e.g. "quit" or "help". Return `nil` to prevent an event from being
// forwarded to the current primitive.
//
// The majority of event handlers should be located on primitives.
func (t *TUI) rootInputHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlH:
		t.leftStatus.SetText("help invoked")
		return nil

	case tcell.KeyCtrlQ:
		t.App.Stop()
		return nil
	}

	return event
}
