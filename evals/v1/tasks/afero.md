# afero task

At `spf13/afero` commit `eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e`, correct `CopyOnWriteFs.RemoveAll`: deleting a path that exists only in the base filesystem must return an error matching `syscall.EPERM` and leave base contents untouched. Preserve deletion behavior for overlay-only and shadowed paths. Add focused regression tests and run the relevant package tests. Do not commit or push.
