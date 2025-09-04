# example

This is an example project.

## how to work with this project

```bash
make build
./bin/demo_x86_64
```

## test hotfix

```bash
telnet localhost 3333
# input
hotfix patch/patch_http.go patch.PatchTestHandler()
# recovery patch/patch_recovery.go
hotfix patch/patch_recovery.go patch.PatchRecoveryHandler()
```

