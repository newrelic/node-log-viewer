package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/newrelic/node-log-viewer/internal/common"
	"github.com/newrelic/node-log-viewer/internal/database"
	"github.com/rivo/tview"
)

const LINES_TABLE_COLUMN_COUNT = 5

type LinesTableContent struct {
	// Embedding the [tview.TableContentReadOnly] type allows us to implement
	// only the read methods in order to satisfy the type implementation
	// requirements.
	tview.TableContentReadOnly
	query *database.Query
}

func NewLinesTableContent(query *database.Query) *LinesTableContent {
	return &LinesTableContent{
		query: query,
	}
}

func (t *LinesTableContent) GetCell(rowNumber int, columnNumber int) *tview.TableCell {
	// The sqlite windowing function `row_number()` starts numbering at 1.
	// The tview widget starts numbering at 0.
	// So we always need to increment the row number by 1.
	envelope := t.query.GetRow(rowNumber + 1)
	if envelope == nil {
		return nil
	}

	cell := tview.NewTableCell("")
	switch columnNumber {
	case 0: // Timestamp
		cell.SetMaxWidth(23).
			SetText(envelope.TimeStampString()).
			SetTextColor(tcell.ColorYellow)
	case 1: // LogLevel
		cell.SetMaxWidth(6).
			SetText(envelope.Level().String()).
			SetTextColor(t.levelColor(envelope.Level())).
			SetAlign(tview.AlignLeft)
	case 2: // SourceComponent
		cell.SetMaxWidth(0).SetText(envelope.Component())
	case 3: // Expand indicator
		cell.SetMaxWidth(3).
			SetText(t.expandIndicator(envelope)).
			SetTextColor(tcell.GetColor("#BB5FB9"))
	case 4: // Log message
		cell.SetExpansion(1).SetText(envelope.Message())
	}

	return cell
}

func (t *LinesTableContent) GetRowCount() int {
	return t.query.NumRows()
}

func (t *LinesTableContent) GetColumnCount() int {
	// We have columns:
	// 1. Timestamp (23 characters wide): e.g `2024-07-03 08:10:41.199`
	// 2. LogLevel name (6 characters wide): e.g. `Trace `
	// 3. SourceComponent name (variable width): e.g. `error_tracer  `
	// 4. Expand indicator (3 characters wide): e.g. ` » `
	// 5. Log message (remainder of available screen width)
	return LINES_TABLE_COLUMN_COUNT
}

func (t *LinesTableContent) levelColor(level common.LogLevel) tcell.Color {
	var color tcell.Color
	switch {
	case level.IsTrace():
		color = tcell.GetColor("#039BE5")
	case level.IsDebug():
		color = tcell.GetColor("#43A047")
	case level.IsInfo():
		color = tcell.GetColor("#9E9E9E")
	case level.IsWarn():
		color = tcell.GetColor("#F57F17")
	case level.IsError():
		color = tcell.GetColor("#FF0000")
	case level.IsFatal():
		color = tcell.GetColor("#6A1B9A")
	}
	return color
}

func (t *LinesTableContent) expandIndicator(line common.Envelope) string {
	indicator := " » "
	result := "   "
	lineKind := line.Kind()
	switch {
	case lineKind == common.TypeDataIncluded:
		result = indicator
	case lineKind == common.TypeEmbeddedDataIncluded:
		result = indicator
	case lineKind == common.TypeError:
		result = indicator
	case lineKind == common.TypeExtraAttributes:
		result = indicator
	}
	return result
}
