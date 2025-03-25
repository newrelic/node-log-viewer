# New Relic Node.js Agent Logs Viewer

> [!IMPORTANT]
> This is an ongoing work in progress. Expect to encounter bugs.
> We expect this tool to improve over time as we use it and learn what we
> want out of it.

This tool reads, parses, and presents [New Relic Node.js Agent][troubleshooting]
troubleshooting logs.

```sh
$ nrlv -f newrelic_agent.log 2>nrlv.log
```

Lines with `»` preceding the message string have more information available by
selecting the line and pressing the `enter` key.

Note: the log view writes its own internal logs to `stderr`.

## Installation

### Go Install

This tool supports being installed directly with the
[`go install`](https://go.dev/ref/mod#go-install) command:

```sh
go install github.com/newrelic/node-log-viewer@latest
```

### Homebrew

This tool is installable via [Homebrew](https://brew.sh/):

```sh
brew install newrelic/agents/node-log-viewer
```

Note: this is the preferred installation method. This is the easiest way to
install the tool on macOS without needing to utilize the
["open anyway"](https://support.apple.com/en-us/102445#openanyway) work around.

## Application Usage

### Navigation

The application has two distinct views: "lines view" and "line detail view."
The lines view is a scrollable list of all log lines showing the timestamp,
log level, log component, and log message for each log line. The line detail
view is what is shown when choosing a log line from the lines view to inspect.
It shows all pertinent information from the log along with any embedded data
in an easy to review format.

+ Lines view:
    * `up arrow`, `j`: move line selection down
    * `down arrow`, `k`: move line selection up
    * `enter`: view detail of selected line
    * `ctrl+s`: open search box
    * `ctrl+g`: open go to line box
    * `ctrl+q`, `ctrl+s`: quit the application
+ Line detail view:
    * up/down navigation is same as lines view
    * `esc`: return to lines view

[troubleshooting]: https://docs.newrelic.com/docs/apm/agents/nodejs-agent/troubleshooting/generate-trace-log-troubleshooting-nodejs/
