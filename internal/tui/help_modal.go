package tui

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var helpText = heredoc.Doc(`
<up arrow>, <j>: Move selection up
<down arrow>, <k>: Move selection down
<enter>: View detail of selection
<s>: Open search box
<e>: Export current result set
<g>: Open go to line box
<esc>, <backspace>: Return to previous view
<q>, <ctrl+c>: Quit the application
`)

func (t *TUI) initHelpModal() {
	view := tview.NewTextView()
	view.SetBorder(true)
	view.SetText(helpText)
	view.SetInputCapture(t.helpModalInputHandler)

	t.pages.AddPage(PAGE_HELP_FORM, modal(view, 75, 16), true, false)
}

func (t *TUI) helpModalInputHandler(event *tcell.EventKey) *tcell.EventKey {
	t.logger.Trace("received key event in line detail view", "key", event.Name(), "rune", event.Rune())
	// The help modal can be closed by pressing any key. If this becomes an
	// issue, we can add a filter at a later date.
	t.hideModal(PAGE_HELP_FORM)
	return event
}
