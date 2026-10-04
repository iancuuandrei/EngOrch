# Windows Git worktree long-path diagnosis

The v1.0.14 evaluation recorded a blocked worktree-creation attempt with a
branch reference lock path of 261 characters. The worktree Git command factory
intentionally removes inherited `GIT_*` variables and runs with
`GIT_CONFIG_GLOBAL=NUL` and `GIT_CONFIG_NOSYSTEM=1`; the source repository also
had `core.longpaths=false`. A machine-wide `core.longpaths=true` therefore
could not affect that command. The retained failure evidence has no Git stderr,
so this is a strongly supported path-length risk diagnosis, not a proven exact
error cause.

The bounded fix adds `-c core.longpaths=true` to each worktree Git command.
This is command-scoped, works while global/system configuration is masked, and
does not mutate user or repository configuration. A Windows regression creates
a fresh fixture whose harness branch ref-lock path exceeds 260 characters,
keeps repository-local `core.longpaths=false`, confirms the masked command sees
the explicit override, then exercises ordinary worktree creation and
observation. It does not access or reconcile the historical blocked run.
