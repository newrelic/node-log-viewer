package v0

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jsumners-nr/nr-node-logviewer/internal/common"
	"github.com/jsumners/go-rfc3339"
	"github.com/perimeterx/marshmallow"
	"github.com/spf13/cast"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type LineEnvelope struct {
	Version         int              `json:"v"`
	LogLevel        *Level           `json:"level"`
	Name            string           `json:"name"`
	Hostname        string           `json:"hostname"`
	Pid             int              `json:"pid"`
	Time            rfc3339.DateTime `json:"time"`
	LogMessage      string           `json:"msg"`
	SourceComponent string           `json:"component,omitempty"`
	OtherFields     map[string]any   `json:"-"`
}

type ErrorLine struct {
	Version         int
	LogLevel        *Level
	Name            string
	Hostname        string
	Pid             int
	Time            rfc3339.DateTime
	LogMessage      string
	SourceComponent string
	Error           Error
}

type Error struct {
	ErrNo   int
	Code    string
	Syscall string
	Path    string
	Stack   []string
	Message string
}

type Level struct {
	number int
}

const (
	TRACE = 10
	DEBUG = 20
	INFO  = 30
	WARN  = 40
	ERROR = 50
	FATAL = 60
)

func (l *Level) IsDebug() bool {
	return l.number == DEBUG
}

func (l *Level) IsError() bool {
	return l.number == ERROR
}

func (l *Level) IsFatal() bool {
	return l.number == FATAL
}

func (l *Level) IsInfo() bool {
	return l.number == INFO
}

func (l *Level) IsTrace() bool {
	return l.number == TRACE
}

func (l *Level) IsWarn() bool {
	return l.number == WARN
}

func (l *Level) String() string {
	var level string
	switch l.number {
	case TRACE:
		level = "Trace"
	case DEBUG:
		level = "Debug"
	case INFO:
		level = "Info"
	case WARN:
		level = "Warn"
	case ERROR:
		level = "Error"
	case FATAL:
		level = "Fatal"
	}
	return level
}

func (l *Level) UnmarshalJSON(data []byte) error {
	if data == nil {
		return nil
	}
	numStr := string(data)
	parsed, err := strconv.ParseInt(numStr, 10, 0)
	if err != nil {
		return err
	}
	l.number = int(parsed)
	return nil
}

var ErrFieldsMissing = errors.New("could not convert line type due to missing fields")

func (e *LineEnvelope) UnmarshalJSON(data []byte) error {
	var envelope LineEnvelope

	extra, err := marshmallow.Unmarshal(data, &envelope, marshmallow.WithExcludeKnownFieldsFromMap(true))
	if err != nil {
		return err
	}

	// We need to perform a "deep copy" of the map, because there's some bug
	// wherein line1 gets a correct map, but then once another line is parsed
	// that has extra fields, line1's map gets corrupted.
	// https://goplay.tools/snippet/16Z61ZrXIB8 should be a reproduction of the
	// issue, but it is not showing the problem 😕
	// TODO: figure out what needs to be done to remove this deep copy hack
	envelope.OtherFields = make(map[string]any)
	copyMap(extra, &envelope.OtherFields)
	// envelope.OtherFields = extra
	*e = envelope

	return nil
}

func copyMap(in, out any) {
	// https://stackoverflow.com/a/67139854
	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(in)
	gob.NewDecoder(buf).Decode(out)
}

func (e *LineEnvelope) Kind() common.Type {
	var result common.Type

	switch {
	case LooksLikeDataIncludedLine(e) == true:
		result = common.TypeDataIncluded

	case LooksLikeEmbeddedDataIncludedLine(e) == true:
		result = common.TypeEmbeddedDataIncluded

	case LooksLikeErrorLine(e) == true:
		result = common.TypeError

	case LooksLikeExtraAttributesLine(e) == true:
		result = common.TypeExtraAttributes

	default:
		result = common.TypeMessage
	}

	return result
}

func (e *LineEnvelope) Component() string {
	if e.SourceComponent == "" {
		return "<none>"
	}
	return e.SourceComponent
}

func (e *LineEnvelope) Level() common.LogLevel {
	return e.LogLevel
}

func (e *LineEnvelope) Message() string {
	return e.LogMessage
}

func (e *LineEnvelope) TimeStampString() string {
	return e.Time.In(time.Now().Location()).Format("2006-01-02 15:04:05.000")
}

func (e *LineEnvelope) GetEmbeddedData() (json.RawMessage, error) {
	var result json.RawMessage
	var err error

	msg := e.LogMessage
	i := strings.Index(msg, `"`)
	j := strings.LastIndex(msg, `"`)
	if i < 0 || j < 0 {
		return nil, fmt.Errorf("embedded JSON is missing start or end")
	}

	msg = msg[i+1 : j]
	msg = strings.ReplaceAll(msg, "\\\"", "\"")
	err = json.Unmarshal([]byte(msg), &result)

	return result, err
}

func EnvelopeToError(envelope LineEnvelope) ErrorLine {
	result := ErrorLine{
		Version:         envelope.Version,
		LogLevel:        envelope.LogLevel,
		Name:            envelope.Name,
		Hostname:        envelope.Hostname,
		Pid:             envelope.Pid,
		Time:            envelope.Time,
		LogMessage:      envelope.LogMessage,
		SourceComponent: envelope.SourceComponent,
		Error: Error{
			ErrNo:   cast.ToInt(envelope.OtherFields["errno"]),
			Code:    envelope.OtherFields["code"].(string),
			Syscall: envelope.OtherFields["syscall"].(string),
			Path:    envelope.OtherFields["path"].(string),
			Message: envelope.OtherFields["message"].(string),
		},
	}

	stack := envelope.OtherFields["stack"].(string)
	result.Error.Stack = strings.Split(stack, "\n")

	return result
}

func LooksLikeErrorLine(envelope any) bool {
	l := envelopeToLineEnvelope(envelope)
	_, errNoFound := l.OtherFields["errno"]
	_, codeFound := l.OtherFields["code"]
	return errNoFound == true && codeFound == true
}

func LooksLikeDataIncludedLine(envelope any) bool {
	l := envelopeToLineEnvelope(envelope)
	if len(l.OtherFields) < 1 {
		return false
	}
	return l.OtherFields["data"] != nil
}

func LooksLikeEmbeddedDataIncludedLine(envelope any) bool {
	l := envelopeToLineEnvelope(envelope)
	hasStart, _ := regexp.MatchString(`"[{\[]`, l.LogMessage)
	hasEnd, _ := regexp.MatchString(`[}\]]"`, l.LogMessage)
	return hasStart == true && hasEnd == true
}

func LooksLikeExtraAttributesLine(envelope any) bool {
	l := envelopeToLineEnvelope(envelope)
	return len(l.OtherFields) > 0
}

// envelopeToLineEnvelope performs a type assertion conversion of the provided
// envelope to a [LineEnvelope]. If neither a value reference nor a pointer
// reference to a [v0.LineEnvelope] is provided, this function will panic.
func envelopeToLineEnvelope(envelope any) LineEnvelope {
	// We _could_ `reflect.ValueOf(envelope).Type().String()` and inspect if
	// the string contains the correct target name ("v0.LineEnvelope"), but then
	// we'd need to return an error. We really just want a helper function so
	// that we can unbox the value regardless if it is a pointer or value
	// reference.
	kind := reflect.ValueOf(envelope).Kind()
	if kind == reflect.Pointer {
		return *envelope.(*LineEnvelope)
	}
	return envelope.(LineEnvelope)
}
