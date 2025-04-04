package tui

import (
	"github.com/rivo/tview"
	"strconv"
)

func (t *TUI) initGotoLineModal() {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetButtonsAlign(tview.AlignCenter)

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
		t.hideModal(PAGE_GOTO_LINE)
	})

	form.AddButton("Cancel", func() {
		t.hideModal(PAGE_GOTO_LINE)
	})

	t.pages.AddPage(PAGE_GOTO_LINE, modal(form, 30, 7), true, false)
}
