package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jsumners-nr/nr-node-logviewer/internal/common"
	log "github.com/jsumners-nr/nr-node-logviewer/internal/log"
)

// LOG_LINE_MIN_WIDTH represents the minimum number of characters for each line
// in the viewport, aka how many characters are used by the required fields
// without considering the width of the message string.
const LOG_LINE_MIN_WIDTH = 56

// LinesViewer is a BubbleTea compatible model that represents a scrollable view
// of parsed and colored log lines. Each line in the view has the following
// fields:
//
//   - Cursor (2 characters wide): e.g. `> `or `  `
//   - Timestamp (23 characters wide): e.g `2024-07-03 08:10:41.199`
//   - LogLevel name (6 characters wide): e.g. `Trace `
//   - SourceComponent name (14 characters wide): e.g. `error_tracer  `
//   - Expand indicator (3 characters wide): e.g. ` • `
//   - Log message (remainder of available screen width)
type LinesViewer struct {
	Style lipgloss.Style

	logger *log.Logger

	cursor      int
	currentView string
	needsUpdate bool

	sourceLines []common.Envelope
	totalLines  int
	pageStart   int
	pageEnd     int
}

func NewLinesViewer(sourceLines []common.Envelope, logger *log.Logger) *LinesViewer {
	viewer := &LinesViewer{
		Style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cccccc")).
			Background(lipgloss.NoColor{}).
			Width(1).
			Height(1),

		logger: logger,

		cursor:      0,
		needsUpdate: true,

		sourceLines: sourceLines,
		totalLines:  len(sourceLines),
		pageStart:   0,
		pageEnd:     10,
	}
	viewer.logger.Debug("created new lines viewer")

	return viewer
}

func (v *LinesViewer) Init() tea.Cmd {
	return nil
}

func (v *LinesViewer) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return v, nil
}

func (v *LinesViewer) View() string {
	if v.needsUpdate == false {
		return v.currentView
	}

	if v.pageEnd < 1 {
		// Sometimes we get a screen size of 0x0, which results in pageEnd=-1.
		// We can't do anything with that. So, bail out.
		return ""
	}

	linesInView := v.sourceLines[v.pageStart:v.pageEnd]
	viewContent := ""
	for i, line := range linesInView {
		if v.cursor == i {
			// TODO: fix indexing bug that results in the indicator scrolling off screen
			// It seems like there is an off by at least 1 calculation resulting in the `>`
			// indicator scrolling at least one or two lines off the bottom of the viewport
			// before it will scroll further lines. We need to figure out how to fix
			// the scrolling algorithm. (It seems to be about 3 lines off if we scroll
			// off the screen, back up to a visible cursor, and open the selected
			// log line.)
			viewContent += "> " + v.stylizeLine(line) + "\n"
		} else {
			viewContent += "  " + v.stylizeLine(line) + "\n"
		}
	}

	v.currentView = viewContent
	v.needsUpdate = false
	return v.Style.Render(viewContent)
}

func (v *LinesViewer) DoSelect() ViewModel {
	return NewLineDetailViewer(v.sourceLines[v.cursor], v.Style, v.logger)
}

func (v *LinesViewer) GetStyle() lipgloss.Style {
	return v.Style
}

func (v *LinesViewer) SetStyle(style lipgloss.Style) {
	v.logger.Trace("set style", "width", style.GetWidth(), "height", style.GetHeight())
	v.Style = style
	v.pageEnd = v.pageStart + (style.GetHeight() - 1)
}

func (v *LinesViewer) ScrollDown(numLines int) {
	if v.cursor+1 == v.totalLines {
		v.logger.Trace("end of list, scrolling down not possible")
		// TODO: check what happens when trying to cursor to the last line in the log file
		return
	}
	if v.cursor+1 > v.pageEnd-1 {
		v.logger.Trace(
			fmt.Sprintf("moving cursor down and adjusting lines in view by %d lines", 1),
			"cursor_pos", v.cursor,
			"new_cursor_pos", v.cursor+1,
		)
		v.pageStart += 1
		v.pageEnd += 1
		v.needsUpdate = true
		return
	}
	// Cursor is not at the last line. Increment it so the view will move it
	// one line down.
	v.logger.Trace(
		fmt.Sprintf("moving cursor down %d lines", 1),
		"cursor_pos", v.cursor,
		"new_cursor_pos", v.cursor+1,
	)
	v.cursor += 1
	v.needsUpdate = true
}

func (v *LinesViewer) ScrollUp(numLines int) {
	v.needsUpdate = true
	if v.cursor-1 < 0 {
		if v.pageStart-1 < 0 {
			// We are at the first log line. So no other lines have been scrolled
			// out of view that would need to be shifted into view.
			v.logger.Trace("top of list, scrolling up not possible")
			v.cursor = 0
			return
		}

		v.logger.Trace(
			fmt.Sprintf("moving cursor up and adjusting lines in view by %d lines", 1),
			"cursor_pos", v.cursor,
			"new_cursor_pos", v.cursor-1,
		)
		v.pageStart -= 1
		v.pageEnd -= 1
		return
	}
	v.logger.Trace(
		fmt.Sprintf("moving cursor up %d lines", 1),
		"cursor_pos", v.cursor,
		"new_cursor_pos", v.cursor-1,
	)
	v.cursor -= 1
}

func (v *LinesViewer) Status() string {
	return fmt.Sprintf("%d / %d ~ (start: %d, end: %d)", v.cursor+1, v.totalLines, v.pageStart, v.pageEnd)
}

var tstampStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FDD835"))
var traceLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#039BE5"))
var debugLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#43A047"))
var infoLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#9E9E9E"))
var warnLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F57F17"))
var errorLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
var fatalLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6A1B9A"))

var moreDataIndicatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#BB5FB9"))
var moreDataIndicator = moreDataIndicatorStyle.Render(" » ")

func (v *LinesViewer) stylizeLine(line common.Envelope) string {
	// v.logger.Trace("stylizing line", "width", v.Style.GetWidth())
	tstamp := tstampStyle.Render(line.TimeStampString())
	level := renderLevel(line.Level())

	component := line.Component()
	if len(component) > 14 {
		component = component[:11] + "..."
	}

	messageTruncated := false
	availableWidth := abs(v.Style.GetWidth() - LOG_LINE_MIN_WIDTH)
	message := line.Message()
	if len(message) > availableWidth {
		message = message[:(availableWidth-3)] + "..."
		messageTruncated = true
	}

	expansion := "   "
	lineKind := line.Kind()
	switch {
	case messageTruncated == true:
		expansion = moreDataIndicator

	case lineKind == common.TypeDataIncluded:
		expansion = moreDataIndicator
	case lineKind == common.TypeEmbeddedDataIncluded:
		expansion = moreDataIndicator
	case lineKind == common.TypeError:
		expansion = moreDataIndicator
	case lineKind == common.TypeExtraAttributes:
		expansion = moreDataIndicator
	}

	return fmt.Sprintf("%s %-6s [%-14s] %s %s", tstamp, level, component, expansion, message)
}

func abs(input int) int {
	if input < 0 {
		return input * -1
	}
	return input
}

func renderLevel(level common.LogLevel) string {
	rendered := fmt.Sprintf("%-6s", level.String())

	switch {
	case level.IsTrace():
		rendered = traceLevelStyle.Render(rendered)
	case level.IsDebug():
		rendered = debugLevelStyle.Render(rendered)
	case level.IsInfo():
		rendered = infoLevelStyle.Render(rendered)
	case level.IsWarn():
		rendered = warnLevelStyle.Render(rendered)
	case level.IsError():
		rendered = errorLevelStyle.Render(rendered)
	case level.IsFatal():
		rendered = fatalLevelStyle.Render(rendered)
	}

	return rendered
}
