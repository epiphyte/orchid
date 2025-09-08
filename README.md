# Orchid Logger

Orchid is a go library for printing pretty logs. It is currently under heavy development.


# Installation

To install the orchid logger:
```bash
$ go get github.com/epiphyte/orchid
```

To use it in your code:
```go
import log "github.com/epiphyte/orchid"
```

# Usage

```go
package main

import log "github.com/epiphyte/orchid"

func main() {
	log.Init("Main")
	
	log.Info("Logger Initialized")
}

```

# Examples

The `examples/` directory contains sample programs demonstrating various features:

- `basic.go` - Comprehensive example showing all log levels, custom loggers, and file logging in both text and JSON formats

## Running Examples

To run the basic example:

```bash
cd examples/
go run basic.go
```

To build the example:

```bash
cd examples/
make
./basic
```

To clean up generated files:

```bash
cd examples/
make clean
```