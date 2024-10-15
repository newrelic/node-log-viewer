package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jsumners-nr/nr-node-logviewer/internal/common"
	log "github.com/jsumners-nr/nr-node-logviewer/internal/log"
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
		switch msgType.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		// TODO: implement pgup/pgdwn
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

	return lipgloss.JoinVertical(
		lipgloss.Top,
		m.MainView.View(),
		m.StatusLine.View(),
	)
}

func setStyle(model *AppModel, width int, height int) {
	model.MainView.SetStyle(
		model.MainView.GetStyle().Width(width).Height(height - 1),
	)
	model.StatusLine.SetStyle(
		model.StatusLine.GetStyle().Width(width),
	)
	model.StatusLine.Text = model.MainView.Status()
}
