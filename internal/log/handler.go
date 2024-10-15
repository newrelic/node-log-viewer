package log

import (
	"context"
	"encoding"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var levelStrings = map[slog.Level]string{
	slog.LevelDebug: "DEBUG",
	slog.LevelInfo:  "INFO",
	slog.LevelError: "ERROR",
	slog.LevelWarn:  "WARN",
	LevelTrace:      "TRACE",
}

// Handler is a custom [slog.Handler] that writes out application logs
// in a human-readable format. We don't really need structured logs for this
// application; it's just that slog is very easy to work with.
//
// Loosely based on https://github.com/telemachus/humane/blob/0bf66cc42a0f7164f38b525497a74dabd4510ed8/humane.go
type Handler struct {
	dest      io.Writer
	level     slog.Leveler
	addSource bool

	attrs []slog.Attr
}

type HandlerOptions struct {
	Level     slog.Level
	AddSource bool
}

type metadata struct {
	source *sourceMeta
	attrs  sourceAttrs
}

type sourceMeta struct {
	file     string
	function string
	line     int
}

type sourceAttrs map[string]string

func NewHandler(dest io.Writer, opts *HandlerOptions) *Handler {
	if opts == nil {
		opts = &HandlerOptions{
			Level: slog.LevelInfo,
		}
	}

	return &Handler{
		dest:      dest,
		level:     opts.Level,
		addSource: opts.AddSource,
	}
}

func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *Handler) Handle(_ context.Context, record slog.Record) error {
	log := &strings.Builder{}

	h.appendLevel(log, record.Level)
	log.WriteString(record.Message)

	haveMeta := false
	meta := metadata{}
	if h.addSource && record.PC != 0 {
		meta.source = h.newSourceAttr(record.PC)
		haveMeta = true
	}

	if record.NumAttrs() > 0 {
		metaAttrs := make(sourceAttrs)
		record.Attrs(func(attr slog.Attr) bool {
			metaAttrs[attr.Key] = valueToString(attr.Value)
			return true
		})
		meta.attrs = metaAttrs
		haveMeta = true
	}

	if haveMeta == true {
		log.WriteString(" {\n")

		if meta.source != nil {
			log.WriteString("\tfile: " + meta.source.file + "\n")
			log.WriteString("\tfunction: " + meta.source.function + "\n")
			log.WriteString("\tline: " + strconv.Itoa(meta.source.line) + "\n")
		}

		if meta.attrs != nil {
			for key, val := range meta.attrs {
				log.WriteString(fmt.Sprintf("\t%s: %s\n", key, val))
			}
		}

		log.WriteString("}")
	}

	log.WriteString("\n")

	// Writes to the output stream should be done synchronously. So we create
	// a lock to enqueue our write position, and free it once we are complete.
	mutex := sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()

	_, err := io.WriteString(h.dest, log.String())
	return err
}

func (h *Handler) WithAttrs([]slog.Attr) slog.Handler {
	// Not implemented.
	return h
}

func (h *Handler) WithGroup(string) slog.Handler {
	// Not implemented.
	return h
}

func (h *Handler) clone() *Handler {
	return &Handler{
		dest:      h.dest,
		level:     h.level,
		attrs:     h.attrs,
		addSource: h.addSource,
	}
}

func (h *Handler) appendLevel(builder *strings.Builder, level slog.Level) {
	val, ok := levelStrings[level]
	if ok == false {
		val = "UNKNOWN"
	}
	builder.WriteString(fmt.Sprintf("%-10s", val))
}

func (h *Handler) newSourceAttr(pc uintptr) *sourceMeta {
	source := frame(pc)
	return &sourceMeta{
		file:     source.File,
		function: source.Function,
		line:     source.Line,
	}
}

func frame(pc uintptr) runtime.Frame {
	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()
	return f
}

func valueToString(val slog.Value) string {
	var result string

	switch val.Kind() {
	case slog.KindString:
		result = val.String()
	case slog.KindInt64:
		result = strconv.FormatInt(val.Int64(), 10)
	case slog.KindUint64:
		result = strconv.FormatUint(val.Uint64(), 10)
	case slog.KindFloat64:
		result = strconv.FormatFloat(val.Float64(), 'g', -1, 64)
	case slog.KindBool:
		result = strconv.FormatBool(val.Bool())
	case slog.KindDuration:
		result = val.Duration().String()
	case slog.KindTime:
		result = val.Time().String()
	case slog.KindAny, slog.KindGroup, slog.KindLogValuer:
		if tm, ok := val.Any().(encoding.TextMarshaler); ok {
			data, err := tm.MarshalText()
			if err != nil {
				// TODO: should this append an error?
				return ""
			}
			result = string(data)
		} else {
			result = fmt.Sprint(val.Any())
		}
	}

	return result
}
