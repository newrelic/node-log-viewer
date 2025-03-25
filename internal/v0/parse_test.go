package v0

import (
	"encoding/json"
	"github.com/jsumners/go-rfc3339"
	"github.com/newrelic/node-log-viewer/internal/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_BasicLine(t *testing.T) {
	line := `{
		"v":0,
		"level":20,
		"name":"newrelic",
		"hostname":"foobar",
		"pid":5362,
		"time":"2024-07-03T12:10:41.199Z",
		"msg":"Using configuration file /foo/newrelic.js."
	}`
	expected := LineEnvelope{
		Version:         0,
		LogLevel:        &Level{DEBUG},
		Name:            "newrelic",
		Hostname:        "foobar",
		Pid:             5362,
		Time:            rfc3339.MustParseDateTimeString("2024-07-03T12:10:41.199Z"),
		LogMessage:      "Using configuration file /foo/newrelic.js.",
		SourceComponent: "",
		OtherFields:     make(map[string]any),
	}

	var found LineEnvelope
	err := json.Unmarshal([]byte(line), &found)
	assert.Nil(t, err)
	assert.Equal(t, expected, found)
}

func Test_LineWithComponent(t *testing.T) {
	line := `{
		"v":0,
		"level":20,
		"name":"newrelic",
		"hostname":"foobar",
		"pid":5362,
		"time":"2024-07-03T12:10:42.171Z",
		"msg":"Aborting request for metadata at \"{\\\"headers\\\":{\\\"X-aws-ec2-metadata-token-ttl-seconds\\\":\\\"21600\\\"},\\\"host\\\":\\\"127.0.0.1\\\",\\\"method\\\":\\\"PUT\\\",\\\"path\\\":\\\"/latest/api/token\\\",\\\"timeout\\\":500}\"",
		"component":"utilization-request"
	}`
	expected := LineEnvelope{
		Version:         0,
		LogLevel:        &Level{DEBUG},
		Name:            "newrelic",
		Hostname:        "foobar",
		Pid:             5362,
		Time:            rfc3339.MustParseDateTimeString("2024-07-03T12:10:42.171Z"),
		LogMessage:      "Aborting request for metadata at \"{\\\"headers\\\":{\\\"X-aws-ec2-metadata-token-ttl-seconds\\\":\\\"21600\\\"},\\\"host\\\":\\\"127.0.0.1\\\",\\\"method\\\":\\\"PUT\\\",\\\"path\\\":\\\"/latest/api/token\\\",\\\"timeout\\\":500}\"",
		SourceComponent: "utilization-request",
		OtherFields:     make(map[string]any),
	}

	var found LineEnvelope
	err := json.Unmarshal([]byte(line), &found)
	assert.Nil(t, err)
	assert.Equal(t, expected, found)
}

func Test_LooksLikeEmbeddedDataIncludedLine(t *testing.T) {
	line := `{
		"v":0,
		"level":20,
		"name":"newrelic",
		"hostname":"foobar",
		"pid":5362,
		"time":"2024-07-03T12:10:42.171Z",
		"msg":"Aborting request for metadata at \"{\\\"headers\\\":{\\\"X-aws-ec2-metadata-token-ttl-seconds\\\":\\\"21600\\\"},\\\"host\\\":\\\"127.0.0.1\\\",\\\"method\\\":\\\"PUT\\\",\\\"path\\\":\\\"/latest/api/token\\\",\\\"timeout\\\":500}\"",
		"component":"utilization-request"
	}`

	var envelope LineEnvelope
	err := json.Unmarshal([]byte(line), &envelope)
	assert.Nil(t, err)
	assert.Equal(t, true, LooksLikeEmbeddedDataIncludedLine(envelope))

	line = `{
		"v":0,
		"level":20,
		"name":"newrelic",
		"hostname":"foobar",
		"pid":5362,
		"time":"2024-07-03T12:10:42.171Z",
		"msg":"Aborting request for metadata at \"[{\\\"foo\\\":42}]\"",
		"component":"utilization-request"
	}`
	err = json.Unmarshal([]byte(line), &envelope)
	assert.Nil(t, err)
	assert.Equal(t, true, LooksLikeEmbeddedDataIncludedLine(envelope))
	data, err := envelope.GetEmbeddedData()
	assert.Nil(t, err)
	var d []map[string]any
	err = json.Unmarshal(data, &d)
	assert.Nil(t, err)
	assert.Equal(t, float64(42), d[0]["foo"])
}

func Test_ErrorLine(t *testing.T) {
	line := `{
		"v":0,
		"level":10,
		"name":"newrelic",
		"hostname":"foobar",
		"pid":5362,
		"time":"2024-07-03T12:10:41.634Z",
		"msg":"Could not list packages in /lib/node_modules (probably not an error)",
		"component":"environment",
		"errno":-2,
		"code":"ENOENT",
		"syscall":"scandir",
		"path":"/lib/node_modules",
		"stack":"Error: ENOENT: no such file or directory, scandir '/lib/node_modules'\n    at async Object.readdir (node:internal/fs/promises:948:18)\n    at async listPackages (/foo/node_modules/newrelic/lib/environment.js:76:18)\n    at async getPackages (/foo/node_modules/newrelic/lib/environment.js:205:3)\n    at async time (/foo/node_modules/newrelic/lib/environment.js:417:16)\n    at async Promise.all (index 1)\n    at async findPackages (/foo/node_modules/newrelic/lib/environment.js:388:40)\n    at async refresh (/foo/node_modules/newrelic/lib/environment.js:486:3)\n    at async Object.getJSON (/foo/node_modules/newrelic/lib/environment.js:498:5)\n    at async Promise.all (index 1)\n    at async facts (/foo/node_modules/newrelic/lib/collector/facts.js:25:35)",
		"message":"ENOENT: no such file or directory, scandir '/lib/node_modules'"
	}`
	expected := LineEnvelope{
		Version:         0,
		LogLevel:        &Level{TRACE},
		Name:            "newrelic",
		Hostname:        "foobar",
		Pid:             5362,
		Time:            rfc3339.MustParseDateTimeString("2024-07-03T12:10:41.634Z"),
		LogMessage:      "Could not list packages in /lib/node_modules (probably not an error)",
		SourceComponent: "environment",
		OtherFields: map[string]any{
			"errno":   float64(-2),
			"code":    "ENOENT",
			"syscall": "scandir",
			"path":    "/lib/node_modules",
			"message": "ENOENT: no such file or directory, scandir '/lib/node_modules'",
			"stack":   "Error: ENOENT: no such file or directory, scandir '/lib/node_modules'\n    at async Object.readdir (node:internal/fs/promises:948:18)\n    at async listPackages (/foo/node_modules/newrelic/lib/environment.js:76:18)\n    at async getPackages (/foo/node_modules/newrelic/lib/environment.js:205:3)\n    at async time (/foo/node_modules/newrelic/lib/environment.js:417:16)\n    at async Promise.all (index 1)\n    at async findPackages (/foo/node_modules/newrelic/lib/environment.js:388:40)\n    at async refresh (/foo/node_modules/newrelic/lib/environment.js:486:3)\n    at async Object.getJSON (/foo/node_modules/newrelic/lib/environment.js:498:5)\n    at async Promise.all (index 1)\n    at async facts (/foo/node_modules/newrelic/lib/collector/facts.js:25:35)",
		},
	}

	var found LineEnvelope
	err := json.Unmarshal([]byte(line), &found)
	assert.Nil(t, err)
	assert.Equal(t, expected, found)

	expectedErrorLine := ErrorLine{
		Version:         0,
		LogLevel:        &Level{TRACE},
		Name:            "newrelic",
		Hostname:        "foobar",
		Pid:             5362,
		Time:            rfc3339.MustParseDateTimeString("2024-07-03T12:10:41.634Z"),
		LogMessage:      "Could not list packages in /lib/node_modules (probably not an error)",
		SourceComponent: "environment",
		Error: Error{
			ErrNo:   -2,
			Code:    "ENOENT",
			Syscall: "scandir",
			Path:    "/lib/node_modules",
			Message: "ENOENT: no such file or directory, scandir '/lib/node_modules'",
			Stack: []string{
				"Error: ENOENT: no such file or directory, scandir '/lib/node_modules'",
				"    at async Object.readdir (node:internal/fs/promises:948:18)",
				"    at async listPackages (/foo/node_modules/newrelic/lib/environment.js:76:18)",
				"    at async getPackages (/foo/node_modules/newrelic/lib/environment.js:205:3)",
				"    at async time (/foo/node_modules/newrelic/lib/environment.js:417:16)",
				"    at async Promise.all (index 1)",
				"    at async findPackages (/foo/node_modules/newrelic/lib/environment.js:388:40)",
				"    at async refresh (/foo/node_modules/newrelic/lib/environment.js:486:3)",
				"    at async Object.getJSON (/foo/node_modules/newrelic/lib/environment.js:498:5)",
				"    at async Promise.all (index 1)",
				"    at async facts (/foo/node_modules/newrelic/lib/collector/facts.js:25:35)",
			},
		},
	}

	assert.Equal(t, true, LooksLikeErrorLine(&found))
	foundErrorLine := EnvelopeToError(found)
	assert.Equal(t, expectedErrorLine, foundErrorLine)
}

func Test_Level(t *testing.T) {
	level := &Level{TRACE}
	assert.Equal(t, true, level.IsTrace())
	assert.Equal(t, "Trace", level.String())

	level = &Level{DEBUG}
	assert.Equal(t, true, level.IsDebug())
	assert.Equal(t, "Debug", level.String())

	level = &Level{INFO}
	assert.Equal(t, true, level.IsInfo())
	assert.Equal(t, "Info", level.String())

	level = &Level{WARN}
	assert.Equal(t, true, level.IsWarn())
	assert.Equal(t, "Warn", level.String())

	level = &Level{ERROR}
	assert.Equal(t, true, level.IsError())
	assert.Equal(t, "Error", level.String())

	level = &Level{FATAL}
	assert.Equal(t, true, level.IsFatal())
	assert.Equal(t, "Fatal", level.String())

	err := json.Unmarshal([]byte("10"), level)
	assert.Nil(t, err)
	assert.Equal(t, true, level.IsTrace())
}

func Test_envelopetToLineEnvelope(t *testing.T) {
	line := `{
		"v":0,
		"level":20,
		"name":"newrelic",
		"hostname":"foobar",
		"pid":5362,
		"time":"2024-07-03T12:10:42.171Z",
		"msg":"Aborting request for metadata at \"{\\\"headers\\\":{\\\"X-aws-ec2-metadata-token-ttl-seconds\\\":\\\"21600\\\"},\\\"host\\\":\\\"127.0.0.1\\\",\\\"method\\\":\\\"PUT\\\",\\\"path\\\":\\\"/latest/api/token\\\",\\\"timeout\\\":500}\"",
		"component":"utilization-request"
	}`

	var pointer *LineEnvelope
	err := json.Unmarshal([]byte(line), &pointer)
	assert.Nil(t, err)
	assert.NotPanics(t, func() {
		l := envelopeToLineEnvelope(pointer)
		assert.Equal(t, l.Kind(), common.TypeEmbeddedDataIncluded)
	})

	var notPointer LineEnvelope
	err = json.Unmarshal([]byte(line), &notPointer)
	assert.Nil(t, err)
	assert.NotPanics(t, func() {
		l := envelopeToLineEnvelope(notPointer)
		assert.Equal(t, l.Kind(), common.TypeEmbeddedDataIncluded)
	})

	var notEnvelope map[string]any
	data := []byte(`{"foo":"foo"}`)
	err = json.Unmarshal(data, &notEnvelope)
	assert.Nil(t, err)
	assert.Panics(t, func() {
		envelopeToLineEnvelope(notEnvelope)
	})
}
