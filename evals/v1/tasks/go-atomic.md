# go-atomic task

At `uber-go/atomic` commit `2d2bdbd262f95d0a890c696fb6f91d4776cb142c`, add `encoding.TextMarshaler` and `encoding.TextUnmarshaler` support to `*Bool`. Encode with canonical `strconv` boolean text, accept the same input forms as `strconv.ParseBool`, and leave the stored value unchanged on invalid text. Keep JSON behavior unchanged. Add focused tests and run the relevant package tests. Do not commit or push.
