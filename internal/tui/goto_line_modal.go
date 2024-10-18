package tui

import (
	"github.com/rivo/tview"
	"strconv"
)

func (t *TUI) initGotoLineModal() {
	// TODO: we'll need other modals. So we should genericize this.
	// See https://github.com/rivo/tview/wiki/Modal
	form := tview.NewForm()
	form.SetBorder(true)
	form.AddInputField(
		"Go to line:",
		"",
		0,
		func(textToCheck string, _ rune) bool {
			_, err := strconv.ParseInt(textToCheck, 10, 0)
			return err == nil
		},
		nil,
	)
	form.AddButton("Go", func() {
		lineNum, _ := strconv.ParseInt(
			form.GetFormItem(0).(*tview.InputField).GetText(),
			10,
			0,
		)
		t.linesTable.Select(int(lineNum)-1, 0)
		t.pages.HidePage(PAGE_GOTO_LINE)
	})
	form.AddButton("Cancel", func() {
		t.pages.HidePage(PAGE_GOTO_LINE)
	})

	containerGrid := tview.NewGrid().
		SetColumns(0, 30, 0).
		SetRows(0, 8, 0).
		AddItem(form, 1, 1, 1, 1, 0, 0, true)

	t.pages.AddPage(PAGE_GOTO_LINE, containerGrid, true, false)
}
