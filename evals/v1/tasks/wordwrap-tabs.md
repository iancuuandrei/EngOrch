# wordwrap-tabs

At `mitchellh/go-wordwrap` commit `ecf0936a077a4bd73a1cc2ac5c370f2b55618d62`, add an additive `WrapStringWithTabWidth(s string, limit, tabWidth uint) string` API. It expands each tab to ASCII spaces up to the next tab stop, measured in Unicode code points from the beginning of the current LF-delimited line, then delegates wrapping to the existing `WrapString`. A zero tab width must produce exactly the legacy `WrapString` result. Preserve the existing API and its behavior for all inputs.

Use a positive tab width to verify ordinary and consecutive tabs, tabs after Unicode code points, LF reset, and interaction with the existing wrap limit. The limit and tab stops count Unicode code points, not terminal display columns or grapheme clusters. Add implementation tests and concise package usage documentation. Do not add a display-width dependency or change `WrapString` semantics.

The pinned upstream is MIT licensed (`LICENSE.md`). Its existing implementation already iterates over Unicode runes; this task does not claim to fix rune counting. The pinned source's existing test documents that tabs count as one character, which is the specific behavior the new opt-in API addresses. The pin is from 2020, so the task is a bounded API extension on a small public pure-Go package rather than evidence of active upstream maintenance.
