package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jsumners-nr/nr-node-logviewer/internal/common"
	"github.com/jsumners-nr/nr-node-logviewer/internal/log"
	v0 "github.com/jsumners-nr/nr-node-logviewer/internal/v0"
	"github.com/muesli/reflow/wordwrap"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
)

type LineDetailViewer struct {
	Style lipgloss.Style

	logger *log.Logger

	line common.Envelope

	viewLines []string
	start     int
	end       int
}

func NewLineDetailViewer(line common.Envelope, style lipgloss.Style, logger *log.Logger) *LineDetailViewer {
	viewWidth := style.GetWidth()
	viewHeight := style.GetHeight()

	viewer := &LineDetailViewer{
		Style:  style,
		logger: logger,
		line:   line,
		start:  0,
		end:    viewHeight - 1,
	}
	viewer.viewLines = viewer.prepareViewLines(line, viewWidth, viewHeight)

	viewer.logger.Debug("created new detail viewer")
	return viewer
}

func (v *LineDetailViewer) Init() tea.Cmd {
	return nil
}

func (v *LineDetailViewer) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return v, nil
}

func (v *LineDetailViewer) View() string {
	// TODO: implement render caching
	linesInView := v.viewLines[v.start:v.end]
	return v.Style.Render(strings.Join(linesInView, "\n"))
}

func (v *LineDetailViewer) DoSelect() ViewModel {
	return v
}

func (v *LineDetailViewer) GetStyle() lipgloss.Style {
	return v.Style
}

func (v *LineDetailViewer) SetStyle(style lipgloss.Style) {
	v.logger.Trace("set style", "width", style.GetWidth(), "height", style.GetHeight())
	v.Style = style
}

func (v *LineDetailViewer) ScrollDown(int) {
	// TODO: implement ScrollDown
}

func (v *LineDetailViewer) ScrollUp(int) {
	// TODO: implement ScrollUp
}

func (v *LineDetailViewer) Status() string {
	return fmt.Sprintf("component: %s level: %s", v.line.Component(), v.line.Level())
}

func (v *LineDetailViewer) prepareViewLines(line common.Envelope, viewWidth int, viewHeight int) []string {
	wrapped := wordwrap.String(line.Message(), viewWidth)
	lines := strings.Split(wrapped, "\n")

	switch line.Kind() {
	case common.TypeDataIncluded:
		// TODO: for some reason the initial message line is getting cut off during rendering
		preparedLines, err := prepareDataIncludedLines(line)
		if err != nil {
			lines = []string{"Error preparing selected line:", "\t" + err.Error()}
		} else {
			lines = append(lines, preparedLines...)
		}

	case common.TypeEmbeddedDataIncluded:
		preparedLines, err := prepareEmbeddedDataLines(line)
		if err != nil {
			v.logger.Error("failed to prepare embedded data line", "error", err)
			lines = []string{"Error preparing selected line:", "\t" + err.Error()}
		} else {
			lines = append(lines, preparedLines...)
		}

	case common.TypeError:
		lines = append(lines, prepareErrorLines(line)...)

	case common.TypeExtraAttributes:
		preparedLines, err := prepareExtraAttrsLines(line)
		if err != nil {
			v.logger.Error("failed to prepare extra attributes line for viewing", "error", err)
			lines = []string{"Error preparing selected line:", "\t" + err.Error()}
		} else {
			lines = append(lines, preparedLines...)
		}

	case common.TypeMessage:
		// Nothing to do.
	}

	// We need to pad out the view lines or else the status bar will
	// not be at the bottom of the screen (really think this is dumb, but I
	// cannot figure out how to get BubbleTea to do it correctly otherwise).
	for range viewHeight - len(lines) {
		lines = append(lines, "")
	}

	return lines
}

func prepareDataIncludedLines(line common.Envelope) ([]string, error) {
	l := line.(*v0.LineEnvelope)
	result := make([]string, 0)
	dataAttribute := l.OtherFields["data"]
	attrs := make(map[string]any)

	for k, v := range l.OtherFields {
		if k == "data" {
			continue
		}
		attrs[k] = v
	}
	if len(attrs) > 0 {
		result = append(result, "\nAttributes:")
		newLines, err := appendAsYaml(attrs, result)
		if err != nil {
			return nil, err
		}
		result = newLines
	}

	result = append(result, "\nData:")
	var js any
	err := json.Unmarshal([]byte(dataAttribute.(string)), &js)
	if err != nil {
		return nil, err
	}
	result, err = appendAsYaml(js, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func prepareEmbeddedDataLines(line common.Envelope) ([]string, error) {
	l := line.(*v0.LineEnvelope)
	embeddedData, err := l.GetEmbeddedData()
	if err != nil {
		return nil, err
	}

	result := []string{"\nData:"}
	var js any
	err = json.Unmarshal(embeddedData, &js)
	if err != nil {
		return nil, fmt.Errorf("unable to coerce embedded data to an interface: %w", err)
	}
	result, err = appendAsYaml(js, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func prepareErrorLines(line common.Envelope) []string {
	l := v0.EnvelopeToError(*line.(*v0.LineEnvelope))

	errorDetail := []string{
		"{",
		"\tErrorNo: " + strconv.Itoa(l.Error.ErrNo),
		"\tCode: " + l.Error.Code,
		"\tSyscall: " + l.Error.Syscall,
		"\tPath: " + l.Error.Path,
		"\tMessage: " + l.Error.Message,
		"\tStack:",
	}

	stack := make([]string, 0)
	for _, s := range l.Error.Stack {
		stack = append(stack, "\t\t"+s)
	}

	errorDetail = append(errorDetail, stack...)
	errorDetail = append(errorDetail, "}")

	return errorDetail
}

func prepareExtraAttrsLines(line common.Envelope) ([]string, error) {
	l := line.(*v0.LineEnvelope)
	result := []string{
		"\nAttributes:",
	}

	result, err := appendAsYaml(l.OtherFields, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func appendAsYaml(input any, lines []string) ([]string, error) {
	yml, err := yaml.Marshal(input)
	if err != nil {
		return nil, err
	}

	splitYml := bytes.Split(yml, []byte("\n"))
	for _, b := range splitYml {
		if len(b) == 0 {
			continue
		}
		lines = append(lines, "\t"+string(b))
	}

	return lines, nil
}
