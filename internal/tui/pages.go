package tui

const (
	PAGE_GOTO_LINE   string = "goto_line_modal"
	PAGE_LINES_TABLE        = "lines_table"
	PAGE_LINE_DETAIL        = "line_detail"
	PAGE_SEARCH_FORM        = "search_form"
)

func (t *TUI) pageShouldCaptureGlobal(pageName string) bool {
	switch pageName {
	case PAGE_GOTO_LINE:
		return false
	case PAGE_LINE_DETAIL:
		return false
	case PAGE_LINES_TABLE:
		return true
	case PAGE_SEARCH_FORM:
		return false
	}
	return false
}
