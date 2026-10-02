# go-difflib task

At `pmezard/go-difflib` commit `5d4384ee4fb2527b0a1256a821ebfc92f91efefc`, fix `difflib.SplitLines` so empty input returns no lines and an existing trailing newline does not produce a phantom extra line. Nonempty text without a trailing newline must retain the current convention of returning its last line with a newline; existing newlines must be preserved. Add boundary tests and run the package tests. Do not commit or push.
