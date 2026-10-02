# logr task

At `go-logr/logr` commit `e99cde667024c0da5e20141f88fb7952e5f4e582`, make `Logger.WithValues()` with zero arguments return the equivalent logger without invoking `LogSink.WithValues`. Preserve one delegation and existing behavior when one or more values are supplied. Add a focused regression test and run the package tests. Do not commit or push.
