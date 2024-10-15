package common

import "encoding/json"

type Level = int

// Type indicates the type of log line, e.g. an "error" line.
type Type = int

const (
	// TypeMessage lines do not have extra attributes. They consist solely of
	// the baseline attributes with a `msg` attribute conveying the presentable
	// information.
	TypeMessage Type = iota

	// TypeDataIncluded represents lines that have a `data` field which contains
	// serialized JSON. Such lines may have other attributes, but the `data`
	// attribute is the primary focus.
	TypeDataIncluded

	// TypeEmbeddedDataIncluded represents lines that have inlined serialized
	// JSON. For example, a log was likely generated like:
	//
	//   log.info(`some ${JSON.stringify(data)} data`)
	//
	TypeEmbeddedDataIncluded

	// TypeError represents lines who's extra attributes represent error metadata.
	TypeError

	// TypeExtraAttributes log lines that have added attributes, but are
	// otherwise a regular [TypeMessage] log line.
	TypeExtraAttributes
)

type LogLevel interface {
	IsDebug() bool
	IsError() bool
	IsFatal() bool
	IsInfo() bool
	IsTrace() bool
	IsWarn() bool

	String() string
}

type Envelope interface {
	Kind() Type

	Component() string
	Level() LogLevel
	Message() string

	// TimeStampString returns the current envelope's timestamp as a string.
	TimeStampString() string

	GetEmbeddedData() (json.RawMessage, error)
}
