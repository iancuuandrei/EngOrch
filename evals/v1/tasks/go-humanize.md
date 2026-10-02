# go-humanize task

At `dustin/go-humanize` commit `a1b4e66b9a6d890e9e15e7091cf16c8032367d6e`, extend `ParseBytes` to accept correctly grouped underscore separators in integer quantities. Reject leading, trailing, or repeated separators. Preserve the parser's exact uint64 overflow rejection and existing decimal/unit behavior. Add focused public tests and documentation if appropriate; run the package's tests. Do not commit or push.
