package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StatusBar struct {
	Style lipgloss.Style
	Text  string
}

func NewStatusBar(width int, height int, status string) StatusBar {
	return StatusBar{
		Text: status,
		Style: lipgloss.
			NewStyle().
			Height(height).
			Width(width).
			Background(lipgloss.Color("#cccccc")).
			Foreground(lipgloss.Color("#111111")).
			PaddingLeft(2),
	}
}

func (s *StatusBar) Init() tea.Cmd {
	return nil
}

func (s *StatusBar) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return s, nil
}

func (s *StatusBar) View() string {
	return s.Style.Render(s.Text)
}

func (s *StatusBar) GetStyle() lipgloss.Style {
	return s.Style
}

func (s *StatusBar) SetStyle(style lipgloss.Style) {
	s.Style = style
}
