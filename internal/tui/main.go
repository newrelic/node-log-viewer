package tui

import (
	"github.com/newrelic/node-log-viewer/internal/common"
	"github.com/newrelic/node-log-viewer/internal/database"
	"github.com/newrelic/node-log-viewer/internal/log"
	"github.com/rivo/tview"
)

type TUI struct {
	App    *tview.Application
	db     *database.LogsDatabase
	logger *log.Logger

	// captureGlobalInput will be true when we are on a "main" view, e.g. the
	// "lines table" view. It will be false when there is some view showing that
	// needs to capture all key presses, e.g. a search modal. The idea being, if
	// this is false, then we can ignore any global keys, e.g. the quit key, until
	// that view is closed.
	captureGlobalInput bool

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

func NewTUI(logLines []common.Envelope, db *database.LogsDatabase, logger *log.Logger) TUI {
	tui := TUI{
		App:                tview.NewApplication(),
		db:                 db,
		logger:             logger,
		lines:              logLines,
		pages:              tview.NewPages(),
		captureGlobalInput: true,
	}

	tui.initLineDetailView()
	tui.initLinesTableView()
	tui.initGotoLineModal()
	tui.initSearchModal()
	tui.initStatusBarView()
	tui.initRootView()

	tui.linesScrollStatus(0, 0) // Initialize the status bar.

	tui.App.SetRoot(tui.root, true)

	return tui
}

func (t *TUI) hidePage(name string) {
	t.pages.HidePage(name)
	t.captureGlobalInput = !t.captureGlobalInput
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

	t.captureGlobalInput = t.pageShouldCaptureGlobalInput(name)
	t.pages.HidePage(currentPageName)
	t.pages.ShowPage(name)
	t.leftStatus.SetText(status)
}

func (t *TUI) hideModal(name string) {
	t.pages.HidePage(name)
	t.captureGlobalInput = !t.captureGlobalInput
}

func (t *TUI) showModal(name string) {
	t.captureGlobalInput = t.pageShouldCaptureGlobalInput(name)
	t.pages.ShowPage(name).SendToFront(name)
}
