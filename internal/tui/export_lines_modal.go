package tui

import (
	"fmt"
	"os"

	"github.com/rivo/tview"
)

func (t *TUI) initExportLinesModal() {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetButtonsAlign(tview.AlignRight)

	form.AddInputField("Export file:", "", 0, nil, nil)

	form.AddButton("Export", func() { t.handleExport(form) })
	form.AddButton("Cancel", func() { t.hideModal(PAGE_EXPORT_LINES) })

	t.pages.AddPage(PAGE_EXPORT_LINES, modal(form, 50, 8), true, false)
}

func (t *TUI) handleExport(form *tview.Form) {

	// TODO: we can't update the buttons because the button handler blocks the
	// main thread. We'd have to wrap that handler in a goroutine in order to
	// perform the UI updates (along with wrapping the meat of this function).
	// But we can't get the error pop-up to show when we do that. So, for now,
	// we are leaving the unresponsive UI in place until we can devise a
	// solution.
	//
	// exportBtn := form.GetButton(0)
	// cancelBtn := form.GetButton(1)
	// t.App.QueueUpdateDraw(func() {
	// 	exportBtn.SetDisabled(true).SetLabel("Exporting ...")
	// 	cancelBtn.SetDisabled(true)
	// })
	//
	// resetButtons := func() {
	// 	t.hideModal(PAGE_EXPORT_LINES)
	// 	exportBtn.SetDisabled(false).SetTitle("Export")
	// 	cancelBtn.SetDisabled(false)
	// }

	fileName := form.GetFormItem(0).(*tview.InputField).GetText()
	t.logger.Trace("attempting to export filtered lines", "fileName", fileName)

	rows, err := t.query.AllResults()
	if err != nil {
		t.hideModal(PAGE_EXPORT_LINES)
		t.logExportError(err, "Could not fetch filtered lines: %s", err.Error())
		return
	}

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.hideModal(PAGE_EXPORT_LINES)
		t.logExportError(err, "Could not open file for writing: %s", err.Error())
		return
	}
	defer file.Close()

	for _, row := range rows {
		_, err = file.WriteString(fmt.Sprintf("%s\n", row.Original))
		if err != nil {
			t.hideModal(PAGE_EXPORT_LINES)
			t.logExportError(err, "Could not write to file (%s): %s", fileName, err.Error())
			return
		}
	}

	t.logger.Trace("successfully exported filtered lines", "fileName", fileName)
	t.hideModal(PAGE_EXPORT_LINES)
}

func (t *TUI) logExportError(err error, msg string, a ...any) {
	formattedMsg := fmt.Sprintf(msg, a...)
	t.logger.Error("error exporting data", "errorMsg", formattedMsg, "error", err)
	t.setErrorText(formattedMsg)
	t.showModal(PAGE_ERROR_MODAL)
}
