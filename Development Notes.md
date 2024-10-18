# Development Notes

This document provides some helpful tips for folks developing this log viewer
application.

## Monitor The App Log

The application log will provide a lot of helpful information that cannot
be displayed on screen due to the TUI taking over the whole terminal. So it's
best to start a test of the application by launching like so:

```sh
$ go run ./... -f newrelic_agent.log -l trace -k -c cache.sqlite 2>viewer.log
```

If your IDE/editor of choice will monitor for external changes to a file, you
can open `viewer.log` and see a running log of messages. Otherwise, launch
another terminal and run:

```sh
$ tail -f viewer.log 
```
