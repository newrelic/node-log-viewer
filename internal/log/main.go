package log

import (
	"io"
	"log/slog"
	"os"
)

const LevelTrace = slog.Level(-8)

// Logger is a wrapper over [slog.Logger] that writes application logs in its
// own format.
type Logger struct {
	*slog.Logger
	level slog.Level
	dest  io.Writer
}

type Option func(*Logger) error

func WithDestination(dest io.Writer) Option {
	return func(logger *Logger) error {
		logger.dest = dest
		return nil
	}
}

func WithLevel(level slog.Level) Option {
	return func(logger *Logger) error {
		logger.level = level
		return nil
	}
}

func New(opts ...Option) (*Logger, error) {
	logger := &Logger{
		level: slog.LevelInfo,
		dest:  os.Stderr,
	}

	for _, opt := range opts {
		err := opt(logger)
		if err != nil {
			return nil, err
		}
	}

	handlerOpts := &HandlerOptions{Level: logger.level, AddSource: true}
	logger.Logger = slog.New(NewHandler(logger.dest, handlerOpts))

	return logger, nil
}

func NewDiscardLogger() *Logger {
	return &Logger{
		Logger: slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
	}
}

func (l *Logger) Trace(msg string, args ...any) {
	l.Logger.Log(nil, LevelTrace, msg, args...)
}

func (l *Logger) WithGroup(name string) *Logger {
	child := l.Logger.WithGroup(name)
	return &Logger{
		Logger: child,
		level:  l.level,
		dest:   l.dest,
	}
}
