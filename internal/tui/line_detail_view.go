package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (t *TUI) initLineDetailView() {
	view := tview.NewTextView()
	view.SetInputCapture(t.lineDetailInputHandler)
	t.lineDetailView = view
	t.pages.AddPage(PAGE_LINE_DETAIL, view, true, false)
}

func (t *TUI) lineDetailInputHandler(event *tcell.EventKey) *tcell.EventKey {
	t.logger.Trace("received key event in line detail view", "key", event.Name(), "rune", event.Rune())

	switch event.Key() {
	case tcell.KeyEsc, tcell.KeyBackspace, tcell.KeyBackspace2:
		t.showPage(PAGE_LINES_TABLE, t.prevPageStatus)
		t.prevPageStatus = ""
	}
	return event
}
