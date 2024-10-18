package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (t *TUI) initStatusBarView() {
	statusBar := tview.NewGrid().SetRows(1).SetColumns(0, 0)

	leftStatus := tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetText("").
		SetTextColor(tcell.ColorBlack)
	leftStatus.SetBackgroundColor(tcell.GetColor("#73d4e9"))

	rightStatus := tview.NewTextView().
		SetTextAlign(tview.AlignRight).
		SetText("(s)earch | (h)elp").
		SetTextColor(tcell.ColorBlack)
	rightStatus.SetBackgroundColor(tcell.GetColor("#73d4e9"))

	statusBar.AddItem(leftStatus, 0, 0, 1, 1, 0, 0, false)
	statusBar.AddItem(rightStatus, 0, 1, 1, 1, 0, 0, false)

	t.statusBar = statusBar
	t.leftStatus = leftStatus
	t.rightStatus = rightStatus
}
