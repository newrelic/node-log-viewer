package tui

const (
	PAGE_GOTO_LINE    string = "goto_line_modal"
	PAGE_LINES_TABLE         = "lines_table"
	PAGE_LINE_DETAIL         = "line_detail"
	PAGE_SEARCH_FORM         = "search_form"
	PAGE_HELP_FORM           = "help_form"
	PAGE_EXPORT_LINES        = "export_lines"
	PAGE_ERROR_MODAL         = "error_modal"
)

func (t *TUI) pageShouldCaptureGlobalInput(pageName string) bool {
	switch pageName {
	case PAGE_GOTO_LINE:
		return false
	case PAGE_LINE_DETAIL:
		return false
	case PAGE_LINES_TABLE:
		return true
	case PAGE_SEARCH_FORM:
		return false
	case PAGE_HELP_FORM:
		return false
	case PAGE_EXPORT_LINES:
		return false
	case PAGE_ERROR_MODAL:
		return false
	}
	return false
}
