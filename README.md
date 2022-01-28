# Orchid Logger

Orchid is a go library for printing pretty logs. It is currently under heavy development.


# Installation

To install the orchid logger:
```bash
$ go get github.com/epiphyte/orchid
```

To use it in your code:
```go
import "github.com/epiphyte/orchid"
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