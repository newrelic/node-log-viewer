package tui

import (
	"github.com/newrelic/node-log-viewer/internal/database"
	"github.com/rivo/tview"
)

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
	form.AddButton("Cancel", func() { t.hideModal(PAGE_SEARCH_FORM) })

	t.pages.AddPage(PAGE_SEARCH_FORM, modal(form, 50, 8), true, false)
}

func (t *TUI) handleSearch(form *tview.Form) {
	searchTerm := form.GetFormItem(0).(*tview.InputField).GetText()

	var query *database.Query
	if searchTerm == "" {
		query = database.SelectAllQuery(t.db, t.logger)
	} else {
		query = database.SearchQuery(searchTerm, t.db, t.logger)
	}
	content := NewLinesTableContent(query)

	t.query = query
	t.linesTable.SetContent(content)
	t.hideModal(PAGE_SEARCH_FORM)
	t.linesScrollStatus(0, 0)
	t.linesTable.Select(0, 0)
}
