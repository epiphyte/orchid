.PHONY: test race vet fmt fmt-check lint examples check clean

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

clean:
	rm -f *.log *.test coverage.out
	$(MAKE) -C examples clean
