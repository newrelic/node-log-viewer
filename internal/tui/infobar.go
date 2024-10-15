package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type InfoBar struct {
	Style lipgloss.Style
	Text  string
}

var fields = []string{
	"(s)earch",
	"(h)elp",
}

func NewInfoBar(width int, height int) InfoBar {
	return InfoBar{
		Text: strings.Join(fields, " | "),
		Style: lipgloss.
			NewStyle().
			Height(height).
			Width(width).
			Background(lipgloss.Color("#cccccc")).
			Foreground(lipgloss.Color("#111111")).
			AlignHorizontal(lipgloss.Right).
			PaddingRight(2),
	}
}

func (i *InfoBar) Init() tea.Cmd {
	return nil
}

func (i *InfoBar) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return i, nil
}

func (i *InfoBar) View() string {
	return i.Style.Render(i.Text)
}

func (i *InfoBar) GetStyle() lipgloss.Style {
	return i.Style
}

func (i *InfoBar) SetStyle(style lipgloss.Style) {
	i.Style = style
}
