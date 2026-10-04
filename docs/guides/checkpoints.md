# Read-only run checkpoints

`fabric checkpoint RUN` reports one validated controller-journal prefix. It
shows the source commit/tree and their repository identity, the current
candidate and graph bindings, graph task IDs with confirmed completion
evidence, configured native-check statuses, the structured reviewer decision,
and counts of unresolved and uncertain intents. It does not include prompts,
source text, model output, reviewer findings, command lines, or process logs.

`accepted_checkpoint` is true only when the run is `READY`, the workspace and
candidate are confirmed, every configured native check passed against that
same candidate, and the recorded review approved that candidate and exact
verification plan. Any intermediate or uncertain state is reported with
`accepted_checkpoint: false`. A later journal append produces a new head and a
new checkpoint; an earlier checkpoint remains bound to its original prefix.

Resume follows the existing durable workflow. A confirmed completed graph task
can be skipped while a later role task starts with its own fresh thread. An
incomplete runtime invocation is read from its recorded thread; the result is
not regenerated automatically. UNKNOWN outcomes remain unresolved, and repair
attempts already recorded remain consumed. `READY` is terminal; a checkpoint
does not reopen it as another implementation round.

This command is observational. It adds no event, runtime call, workspace write,
or resume authority.
