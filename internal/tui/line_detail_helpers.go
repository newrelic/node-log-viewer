package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jsumners-nr/nr-node-logviewer/internal/common"
	v0 "github.com/jsumners-nr/nr-node-logviewer/internal/v0"
	"gopkg.in/yaml.v3"
	"strconv"
)

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
