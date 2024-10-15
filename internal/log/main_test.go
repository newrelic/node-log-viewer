package log

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"strings"
	"testing"
)

func Test_BasicFunctions(t *testing.T) {
	dest := &bytes.Buffer{}
	logger, err := New(WithLevel(slog.LevelDebug), WithDestination(dest))
	assert.Nil(t, err)

	logger.Info("a test")
	expected := strings.Join(
		[]string{
			"INFO      a test {",
			"\tfile: /Users/jsumners/Projects/logviewer/internal/log/main_test.go",
			"\tfunction: github.com/jsumners-nr/nr-node-logviewer/internal/log.Test_BasicFunctions",
			"\tline: 16",
			"}\n",
		},
		"\n",
	)
	assert.Equal(t, expected, dest.String())
	dest.Reset()

	logger.Debug("with data", "foo", 42)
	expected = strings.Join(
		[]string{
			"DEBUG     with data {",
			"\tfile: /Users/jsumners/Projects/logviewer/internal/log/main_test.go",
			"\tfunction: github.com/jsumners-nr/nr-node-logviewer/internal/log.Test_BasicFunctions",
			"\tline: 30",
			"\tfoo: 42",
			"}\n",
		},
		"\n",
	)
	assert.Equal(t, expected, dest.String())
	dest.Reset()
}
