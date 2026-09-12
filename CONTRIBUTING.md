# Contributing Guide

## Before opening a pull request

Run the same checks as CI:

```
make check
```

This runs gofmt, `go vet`, the test suite under the race detector,
[staticcheck](https://staticcheck.dev), and builds the example program.
Install staticcheck with `go install honnef.co/go/tools/cmd/staticcheck@latest`.

## Publish Module

0. Add a section for the new version to `CHANGELOG.md`.
1. Tidy up the module dependencies
   ```
   go mod tidy
   ```
2. Run the checks
   ```
   make check
   ```
3. Tag the change
   ```
   git tag v1.0.0
   git push origin v1.0.0
   ```
4. Publish the new version
   ```
   GOPROXY=proxy.golang.org go list -m github.com/epiphyte/orchid@v1.0.0
   ```
