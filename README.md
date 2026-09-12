# Orchid Logger [![Go Reference](https://pkg.go.dev/badge/github.com/epiphyte/orchid.svg)](https://pkg.go.dev/github.com/epiphyte/orchid) [![CI](https://github.com/epiphyte/orchid/actions/workflows/ci.yml/badge.svg)](https://github.com/epiphyte/orchid/actions/workflows/ci.yml)

Orchid is a small, dependency-free Go library for colorized, structured logs.
It prints color-coded lines to the console and can also write every line to a
file as plain text or JSON.

Severity levels: `INFO`, `OK`, `WARN`, `ERROR`, `FATAL`, `DEBUG`.
`FATAL` exits the program after logging.

## Installation

```bash
go get github.com/epiphyte/orchid
```

```go
import log "github.com/epiphyte/orchid"
```

## Usage

### Default logger

```go
package main

import log "github.com/epiphyte/orchid"

func main() {
	if err := log.Init("main"); err != nil {
		panic(err)
	}
	log.Info("Logger initialized")
	log.Warn("Something ", "to watch")   // arguments are joined with fmt.Sprint
}
```

### Per-module loggers

Each `Logger` carries its own module name. All loggers are safe for
concurrent use.

```go
var db log.Logger
if err := db.Init("database"); err != nil {
	panic(err)
}
db.OK("connected")
```

### File logging

File output is global: one file, shared by the default logger and every
`Logger` instance. Set it once, and close it before exit.

```go
if err := log.SetLogFile("app.log", log.FormatTXT); err != nil {
	panic(err)
}
defer log.Close()
```

Text format:

```
2026-09-12 12:33:38 [INFO] main: Logger initialized
```

JSON format (`log.FormatJSON`) writes one object per line:

```json
{"severity":"INFO","text":"Logger initialized","module":"main","time":"2026-09-12T12:33:38.316-04:00"}
```

If the file cannot be opened, `SetLogFile` returns an error and the previous
file, if any, stays active. If a write fails later, the line still reaches
the console and a single `ORCHID FILE ERROR` is reported on stderr until
writes succeed again or the file is reconfigured.

### Configuration

```go
cfg := log.GetConfiguration()
cfg.SetEnableColors(false)
```

Colors are on by default only when stderr is a terminal and the `NO_COLOR`
environment variable is unset. `SetEnableColors` overrides that in either
direction.

Console output goes through the standard `log` package, so `log.SetOutput`
and `log.SetFlags` from the standard library apply to it.

## Examples

`examples/basic.go` exercises every log level, per-module loggers, and file
logging in both formats:

```bash
cd examples
go run basic.go
make clean      # remove app.log and app.json
```

## Development

```bash
make check      # gofmt, go vet, go test -race, staticcheck, build examples (what CI runs)
make test       # just the tests
make race       # tests with the race detector
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the release process.
