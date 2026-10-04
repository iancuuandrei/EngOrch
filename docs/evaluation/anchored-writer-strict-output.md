# Strict anchored writer output

A bounded read-only diagnosis identified the concrete cause of the preserved
context-v2 godotenv blocker. Run
`ba0a3da6d601f337b7c5c06520be2b44ebbc5a546ceb097ed8166e982d557619`
ended after observing repair writer invocation
`d5604d5b2e542b3a536d2dcaad5a8fc8eee7386b35dfc151e972f8bda5f81ba2`.
The final output included an existing README file with a valid before hash,
empty `edits` and null new content. Four anchor validations passed for the
other files. They did not validate the no-op README entry.

The existing decoder correctly rejected that entry before proposal admission
or a file effect. The original candidate, repair 1/2 and BLOCKED outcome remain
unchanged; the run is not resumed or relabeled. No uncertain effect is resent.
The prior comparison record accurately stated that cause was unknown before
this subsequent output inspection.

Evidence hashes: main event 61
`aeb4b79e30d4443580a880fa5fd18bad8d9d0e7b8e8bbffbe66c1fbd53292b89`;
runtime result `b9105bf8ca9ea815fd360c438af3be6e1bd1a64ee0a8b9f4cab95e797cb885f0`;
runtime event 54 `d0ad9c62d4dc052755ca703fb0ab6cb04a7dc8973e13c3e38781254bc4b868a1`;
output SHA-256 `91e428de78b94803f863477053b958b3bcd13f710c0c4fc8c350754a05ccbb68`.
No raw transcript or source-bearing runtime output is published.

## Product correction

`fabric init --strict-writer-edits` opts new Codex runs into
`anchored-edits-v3`; omission preserves v1 and `--validate-writer-edits`
preserves v2. The strict option is mutually exclusive with the v2 option.
V3 retains the anchored decoder and candidate effect boundaries. Its schema
has two disjoint nested `anyOf` forms: existing files require a string before
hash, one or more anchored edits and null new content; new files require a
null before hash, no edits and explicit string content. The prompt tells the
writer to omit unchanged files. Runtime schema authentication still requires
the exact candidate-bound schema. No controller retry or budget is added.

Nested `anyOf` and array bounds are in the documented provider subset; this
is schema compatibility guidance, not proof of a live V3 task succeeding.
See [OpenAI structured outputs](https://developers.openai.com/api/docs/guides/structured-outputs).

The Native evaluation runner exposes `-StrictWriterEdits` on both Prepare
and Evaluate and verifies the inspected contract. Tests preserve historical
V1/V2 schema/prompt paths and reject no-op existing-file entries. Focused
cross-package regressions and independent core/runner review passed. Clean full
qualification and actual model exercise remain required before adoption.