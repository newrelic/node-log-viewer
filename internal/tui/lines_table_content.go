package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/jsumners-nr/nr-node-logviewer/internal/common"
	"github.com/rivo/tview"
)

const LINES_TABLE_COLUMN_COUNT = 5

type LinesTableContent struct {
	lines []common.Envelope

	totalPossibleRows int
}

func NewLinesTableContent(lines []common.Envelope) *LinesTableContent {
	return &LinesTableContent{lines: lines}
}

func (t *LinesTableContent) GetCell(row int, col int) *tview.TableCell {
	if row < 0 || col < 0 {
		return nil
	}
	if row >= len(t.lines) {
		return nil
	}

	r := t.lines[row]
	if col >= LINES_TABLE_COLUMN_COUNT {
		return nil
	}

	cell := tview.NewTableCell("")
	switch col {
	case 0: // Timestamp
		cell.SetMaxWidth(23).
			SetText(r.TimeStampString()).
			SetTextColor(tcell.ColorYellow)
	case 1: // LogLevel
		cell.SetMaxWidth(6).
			SetText(r.Level().String()).
			SetTextColor(t.levelColor(r.Level())).
			SetAlign(tview.AlignLeft)
	case 2: // SourceComponent
		cell.SetMaxWidth(0).SetText(r.Component())
	case 3: // Expand indicator
		cell.SetMaxWidth(3).
			SetText(t.expandIndicator(r)).
			SetTextColor(tcell.GetColor("#BB5FB9"))
	case 4: // Log message
		cell.SetExpansion(1).SetText(r.Message())
	}

	return cell
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

func (t *LinesTableContent) GetRowCount() int {
	return len(t.lines)
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

func (t *LinesTableContent) SetCell(row int, col int, cell *tview.TableCell) {}

func (t *LinesTableContent) RemoveRow(row int) {}

func (t *LinesTableContent) RemoveColumn(col int) {}

func (t *LinesTableContent) InsertRow(row int) {}

func (t *LinesTableContent) InsertColumn(col int) {}

func (t *LinesTableContent) Clear() {}
