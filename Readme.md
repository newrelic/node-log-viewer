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

[troubleshooting]: https://docs.newrelic.com/docs/apm/agents/nodejs-agent/troubleshooting/generate-trace-log-troubleshooting-nodejs/
