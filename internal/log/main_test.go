package log

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"regexp"
	"testing"
)

func Test_BasicFunctions(t *testing.T) {
	dest := &bytes.Buffer{}
	logger, err := New(WithLevel(slog.LevelDebug), WithDestination(dest))
	assert.Nil(t, err)

	logger.Info("a test")
	assert.Regexp(t, regexp.MustCompile(`^INFO\s+a test {\n\t`), dest.String())
	dest.Reset()

	logger.Debug("with data", "foo", 42)
	assert.Regexp(t, regexp.MustCompile(`^DEBUG\s+with data {\n\t`), dest.String())
	dest.Reset()
}
