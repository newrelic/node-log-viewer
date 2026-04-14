package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// errorTextView holds a reference to the text view used when displaying
// errors to the user. There isn't a convenient API for getting the control
// out of the constructed modal. So we cheat by keeping the reference here.
var errorTextView *tview.TextView

func (t *TUI) initErrorModal() {
	errorTextView = tview.NewTextView()
	errorTextView.SetBorder(true)

	btn := tview.NewButton("Close")
	btn.SetLabelColor(tcell.ColorWhite)
	btn.SetBackgroundColor(tcell.ColorDarkRed)
	btn.SetBackgroundColorActivated(tcell.ColorRed)
	btn.SetLabelColorActivated(tcell.ColorBlack)
	btn.SetSelectedFunc(func() {
		t.hideModal(PAGE_ERROR_MODAL)
	})

	btnBox := tview.NewGrid()
	btnBox.SetRows(0, 1)
	btnBox.SetColumns(0, 7)
	btnBox.AddItem(errorTextView, 0, 0, 1, 2, 0, 0, true)
	btnBox.AddItem(btn, 1, 1, 1, 1, 0, 0, false)
	btnBox.SetInputCapture(t.errorModalInputHandler)

	t.pages.AddPage(PAGE_ERROR_MODAL, modal(btnBox, 75, 16), true, false)
}

func (t *TUI) setErrorText(text string) {
	errorTextView.SetText(text)
}

func (t *TUI) errorModalInputHandler(event *tcell.EventKey) *tcell.EventKey {
	t.logger.Trace("received key event in line detail view", "key", event.Name(), "rune", event.Rune())

	switch event.Key() {
	case tcell.KeyESC, tcell.KeyEnter, tcell.KeyBackspace, tcell.KeyBackspace2:
		t.hideModal(PAGE_ERROR_MODAL)
	}
	return event
}
