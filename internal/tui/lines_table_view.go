package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/newrelic/node-log-viewer/internal/common"
	"github.com/rivo/tview"
)

func (t *TUI) initLinesTableView() {
	table := tview.NewTable()

	// The `.SetEvaluateAllRows(true)` tells tview, at a minimum, to calculate
	// the column widths in all possible rows. This makes column widths consistent
	// while scrolling the table. However, this is an expensive operation. It is
	// prohibitively expensive for tables with a significant number of rows, and
	// for "virtual tables." Unfortunately, we hit both of those. Fortunately,
	// by defining our virtual table correctly, we don't really need this setting
	// to get consistent rows. But we are leaving this here just in case we
	// need to stop using a virtual table, and need the hint in the future.
	// table.SetEvaluateAllRows(true)

	table.SetContent(NewLinesTableContent(t.query))
	table.SetSelectable(true, false) // Select by rows only.
	table.SetSelectedStyle(
		tcell.Style{}.
			Background(tcell.GetColor("#40ea37")).
			Foreground(tcell.ColorBlack),
	)
	table.SetInputCapture(t.linesTableInputHandler)
	table.SetSelectionChangedFunc(t.linesScrollStatus)
	table.SetSelectedFunc(t.lineSelected)
	t.linesTable = table
	t.pages.AddPage(PAGE_LINES_TABLE, table, true, true)
}

func (t *TUI) linesTableInputHandler(event *tcell.EventKey) *tcell.EventKey {
	t.logger.Trace("received key event in lines table view", "key", event.Name(), "rune", event.Rune())

	// TODO: modals are retaining state between invocations, they shouldn't
	switch event.Rune() {
	case 'g':
		t.logger.Trace("showing go to line modal")
		t.showModal(PAGE_GOTO_LINE)
		return nil

	case 's':
		t.logger.Trace("showing search modal")
		t.showModal(PAGE_SEARCH_FORM)
		return nil
	}
	return event
}

// linesScrollStatus is a callback invoked by the log lines table to indicate
// which line has been highlighted. We use this to update the status bar to
// show which line, out of the total, is currently highlighted.
func (t *TUI) linesScrollStatus(row int, _ int) {
	totalRows := t.linesTable.GetRowCount()
	t.leftStatus.SetText(fmt.Sprintf("%d / %d", row+1, totalRows))
}

// lineSelected is a callback invoked by the [tview.Table] when a row has been
// selected, i.e. the user pressed the "enter" key while the row was
// highlighted. This handler will determine the kind of the log line, prepare
// the line for detailed view, and switch to the detail view.
func (t *TUI) lineSelected(row int, _ int) {
	// The UI references rows starting from 0.
	// The database references rows starting from 1.
	line := t.query.GetRow(row + 1)
	lines := strings.Split(line.Message(), "\n")

	switch line.Kind() {
	case common.TypeDataIncluded:
		preparedLines, err := prepareDataIncludedLines(line)
		if err != nil {
			lines = []string{"Error preparing selected line:", "\t" + err.Error()}
		} else {
			lines = append(lines, preparedLines...)
		}

	case common.TypeEmbeddedDataIncluded:
		preparedLines, err := prepareEmbeddedDataLines(line)
		if err != nil {
			t.logger.Error("failed to prepare embedded data line", "error", err)
			lines = []string{"Error preparing selected line:", "\t" + err.Error()}
		} else {
			lines = append(lines, preparedLines...)
		}

	case common.TypeError:
		lines = append(lines, prepareErrorLines(line)...)

	case common.TypeExtraAttributes:
		preparedLines, err := prepareExtraAttrsLines(line)
		if err != nil {
			t.logger.Error("failed to prepare extra attributes line for viewing", "error", err)
			lines = []string{"Error preparing selected line:", "\t" + err.Error()}
		} else {
			lines = append(lines, preparedLines...)
		}

	case common.TypeMessage:
		// Nothing to do.
	}

	t.lineDetailView.SetText(strings.Join(lines, "\n"))
	t.showPage(
		PAGE_LINE_DETAIL,
		fmt.Sprintf("component: %s -- level: %s", line.Component(), line.Level()),
	)
}
