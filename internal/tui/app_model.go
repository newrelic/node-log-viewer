package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jsumners-nr/nr-node-logviewer/internal/common"
	log "github.com/jsumners-nr/nr-node-logviewer/internal/log"
	"github.com/spf13/cast"
	"math"
)

// ViewModel represents any model being used as the "top" portion of the main
// app view.
type ViewModel interface {
	tea.Model
	GetStyle() lipgloss.Style
	SetStyle(lipgloss.Style)
	ScrollDown(int)
	ScrollUp(int)
	Status() string

	// DoSelect will be invoked when a user "selects" an item from the current
	// view. For example, when a user presses "enter" on a selected log line,
	// the method should return a new view that shows the details of that log
	// line.
	DoSelect() ViewModel
}

// AppModel represents the main TUI "container." It governs the overall window
// size, contains the main view port and status bar, and all metadata needed
// to render the application.
type AppModel struct {
	logger            *log.Logger
	receivedSizeEvent bool
	WindowSize        tea.WindowSizeMsg

	MainView   ViewModel
	StatusLine StatusBar
	InfoLine   InfoBar

	// views represents navigation through the application. For example, when a
	// log line is selected, a new view will be returned that shows the details
	// of that log line. The list of log lines view will be stored as an element
	// in this slice. When the user "navigates back", we will pop the view of this
	// slice and restore it.
	views []ViewModel
}

func NewAppModel(sourceLines []common.Envelope, logger *log.Logger) AppModel {
	return AppModel{
		logger:     logger,
		MainView:   NewLinesViewer(sourceLines, logger),
		StatusLine: NewStatusBar(1, 1, "initialized"),
		InfoLine:   NewInfoBar(1, 1),
	}
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msgType := msg.(type) {

	// tea.WindowSizeMsg is fired when the view is first rendered and on every
	// resize of the window. We use this event to establish the overall window
	// dimensions.
	case tea.WindowSizeMsg:
		m.receivedSizeEvent = true
		m.logger.Trace("got tea.WindowSizeMsg", "width", msgType.Width, "height", msgType.Height)
		m.WindowSize = msgType
		setStyle(&m, msgType.Width, msgType.Height)

	case tea.KeyMsg:
		// We _could_ add a default case that passes the message to child views,
		// but then it gets difficult to synchronize view updates. For example,
		// if we pass a message to a view, and it updates the background color of
		// the view, then we never see that change until we synchronize changes.
		// At least for now, we will simply keep adding new view interface methods
		// and stubbing them on out views that don't need to support whatever
		// functionality the method provides.
		switch msgType.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "g":
		// TODO: implement "go to line"?

		case "h", "?":
		// TODO: implement help modal
		// I like the way `ncdu` presents its help, but `top` and `tmux` just
		// overwrite the current view, and that'll probably be easier.
		//
		// Looks like we can't do modals with BubbleTea:
		// https://github.com/charmbracelet/bubbletea/issues/79

		case "s", "/":
		// TODO: implement a search/filter system (this is the biggest feature)
		// I'm thinking that the lines viewer could have a new "originalLines"
		// property where we can store the originally processed lines from the file.
		// Then we can replace the sourceLines with a filtered set of lines based
		// upon the outcome of the search and all of the implemented navigation
		// and inspection operations will continue to "just work."
		//
		// What I'm not sure about is how we implement a search system. I suppose
		// we can add another interface method like `DoSearch` that accepts some
		// sort of search definition struct. Non-searchable viewers would need to
		// at least stub out the method.
		//
		// I'm also concerned about the memory usage. The app design as it is, is
		// already subject to consuming large amounts of memory based upon how
		// many log lines are present in the parsed file. We're just reading the
		// whole thing into memory and then working with it. If we are going to
		// duplicate a subset of the log lines, that will definitely increase our
		// memory usage. The only solution that I have been able to think of so far
		// is to read everything into a temporary sqlite database and work against
		// that. It'll require significant refactoring, but it might be the way we
		// need to go. I'm open to ideas here. Another might be to memory map the
		// original log file somehow, but I wouldn't know where to begin with that.
		//
		// 2024-10-16: The more I think about it, the more I think using the sqlite
		// database is going to be right direction. If we do that, then we can
		// support caching of a file via automatic and manual naming of the cache
		// file, e.g.:
		//
		// nrlv -f newrelic_agent.log --keep-cache [--cache-file=something.sqlite]
		//
		// If `--cache-file` is not provided, we'd checksum the `newrelic_agent.log`
		// file in order to generate the name for the cache file. Then, when we
		// boot up, we can look in the temporary directory for an existing cache
		// file and skip parsing of the log file if the cache exists.

		case "pgdown":
		// TODO: implement page down

		case "pgup":
		// TODO: implement page up

		case "down", "j":
			m.MainView.ScrollDown(1)
			m.StatusLine.Text = m.MainView.Status()
		case "up", "k":
			m.MainView.ScrollUp(1)
			m.StatusLine.Text = m.MainView.Status()

		case "esc", "backspace":
			if len(m.views) == 0 {
				// Maybe do nothing?
				return m, tea.Quit
			}
			newView := m.views[len(m.views)-1]
			m.views = m.views[:len(m.views)-1]
			m.MainView = newView
			setStyle(&m, m.WindowSize.Width, m.WindowSize.Height)

		case "enter", " ":
			newView := m.MainView.DoSelect()
			m.views = append(m.views, m.MainView)
			m.MainView = newView
			setStyle(&m, m.WindowSize.Width, m.WindowSize.Height)
		}
	}

	return m, nil
}

func (m AppModel) View() string {
	if m.receivedSizeEvent == false {
		// We don't want to render anything until we have some dimensions to
		// work with.
		return ""
	}

	bottomBar := lipgloss.JoinHorizontal(
		lipgloss.Bottom,
		m.StatusLine.View(),
		m.InfoLine.View(),
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		m.MainView.View(),
		bottomBar,
		//m.StatusLine.View(),
	)
}

func setStyle(model *AppModel, width int, height int) {
	model.MainView.SetStyle(
		model.MainView.GetStyle().Width(width).Height(height - 1),
	)

	statusWidth := cast.ToInt(math.Floor(float64(width) * 0.8))
	infoWidth := width - statusWidth

	model.StatusLine.SetStyle(
		model.StatusLine.GetStyle().Width(statusWidth),
	)
	model.StatusLine.Text = model.MainView.Status()

	model.InfoLine.SetStyle(
		model.InfoLine.GetStyle().Width(infoWidth),
	)
}
