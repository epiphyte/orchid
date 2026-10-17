.PHONY: test race vet fmt fmt-check lint examples check index clean

MODULE  := github.com/epiphyte/orchid
# Version to publish; defaults to the most recent tag. Override with
# `make index VERSION=v1.2.3`.
VERSION ?= $(shell git describe --tags --abbrev=0)

test:
	go test ./...

race:
	go test -race -count=1 ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

fmt-check:
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "files need gofmt:"; echo "$$out"; exit 1; fi

lint:
	@command -v staticcheck >/dev/null || { echo "install with: go install honnef.co/go/tools/cmd/staticcheck@latest"; exit 1; }
	staticcheck ./...

examples:
	cd examples && go build -o /dev/null .

# Everything CI runs.
check: fmt-check vet race lint examples

# Ask the Go module proxy to fetch the tagged version so it shows up on
# pkg.go.dev and for `go get` without waiting for the next crawl. The tag
# must already be pushed.
index:
	@test -n "$(VERSION)" || { echo "no tag found; pass VERSION=vX.Y.Z"; exit 1; }
	GOPROXY=proxy.golang.org go list -m $(MODULE)@$(VERSION)

clean:
	rm -f *.log *.test coverage.out
	$(MAKE) -C examples clean
