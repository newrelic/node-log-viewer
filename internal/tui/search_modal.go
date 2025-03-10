package tui

import "github.com/rivo/tview"

func (t *TUI) initSearchModal() {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetButtonsAlign(tview.AlignRight)

	// TODO: create a robust form that allows for searching by component or message. Maybe
	// tailor it to FTS5 specific queries 🤷‍♂️.
	form.AddInputField(
		"Search term(s):",
		"",
		0,
		nil,
		nil,
	)

	form.AddButton("Search", func() { t.handleSearch(form) })
	form.AddButton("Cancel", func() { t.pages.HidePage(PAGE_SEARCH_FORM) })

	t.pages.AddPage(PAGE_SEARCH_FORM, modal(form, 50, 8), true, false)
}

func (t *TUI) handleSearch(form *tview.Form) {
	searchTerm := form.GetFormItem(0).(*tview.InputField).GetText()
	searchResult, err := t.db.Search(searchTerm)
	if err != nil {
		t.logger.Error("search failed", "error", err)
		t.pages.HidePage(PAGE_SEARCH_FORM)
		// TODO: show error modal
		return
	}

	lines := searchResult.ToLines()
	content := NewLinesTableContent(lines)
	t.lines = lines
	t.linesTable.SetContent(content)
	t.pages.HidePage(PAGE_SEARCH_FORM)
	t.linesScrollStatus(0, 0)
	t.linesTable.Select(0, 0)
}
