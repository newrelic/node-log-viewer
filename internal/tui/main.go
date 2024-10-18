package tui

import (
	"github.com/jsumners-nr/nr-node-logviewer/internal/common"
	"github.com/jsumners-nr/nr-node-logviewer/internal/log"
	"github.com/rivo/tview"
)

type TUI struct {
	App    *tview.Application
	logger *log.Logger

	// root is the overall application frame.
	root *tview.Grid
	// pages is the primary top view of the application, i.e. everything
	// that is not the bottom status bar. It is a set of named pages.
	pages *tview.Pages

	// linesTable is the primary view table that shows the set of log lines
	// matching whatever search filter has been applied (or all lines if no
	// filter is applied). It is named PAGE_LINES_TABLE in the pages set.
	linesTable *tview.Table

	// lineDetailView is used to display the details of a selected log line.
	// It is named PAGE_LINE_DETAIL in the pages set.
	lineDetailView *tview.TextView

	statusBar   *tview.Grid
	leftStatus  *tview.TextView
	rightStatus *tview.TextView

	// lines is the current set of log lines available for navigation and
	// inspection.
	lines []common.Envelope

	// prevPageStatus is used to cache the left status text when switching
	// between pages. This allows us to restore the status when we return to
	// that page.
	prevPageStatus string
}

func NewTUI(logLines []common.Envelope, logger *log.Logger) TUI {
	tui := TUI{
		App:    tview.NewApplication(),
		logger: logger,
		lines:  logLines,
		pages:  tview.NewPages(),
	}

	tui.initLineDetailView()
	tui.initLinesTableView()
	tui.initGotoLineModal()
	tui.initStatusBarView()
	tui.initRootView()

	tui.linesScrollStatus(0, 0) // Initialize the status bar.

	tui.App.SetRoot(tui.root, true)

	return tui
}

// showPage hides the current page, caches the text of the left status
// indicator, updates the left status, and shows the new page.
func (t *TUI) showPage(name string, status string) {
	currentPageName, currentPage := t.pages.GetFrontPage()
	if currentPage != nil {
		t.prevPageStatus = t.leftStatus.GetText(false)
	} else {
		t.prevPageStatus = ""
	}

	t.pages.HidePage(currentPageName)
	t.pages.ShowPage(name)
	t.leftStatus.SetText(status)
}
