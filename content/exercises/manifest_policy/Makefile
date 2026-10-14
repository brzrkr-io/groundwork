# Everyday commands. `make` alone runs the tests and the policy check.

BIN := bin/kubepolicy

.PHONY: all build test fuzz golden policy image clean

all: test policy

build:
	go build -o $(BIN) ./cmd/kubepolicy

test:
	go vet ./...
	go test -race ./...

# Let the fuzzer look for inputs that break Split.
fuzz:
	go test ./internal/manifest -run '^$$' -fuzz FuzzSplit -fuzztime 30s

# Rewrite the golden files after a deliberate change to the output; review the diff!
golden:
	go test ./internal/cli -run TestGolden -update

# Check the deploy/ manifests the way CI does.
policy: build
	$(BIN) deploy/base
	kubectl kustomize deploy/overlays/production | $(BIN) -

image:
	docker build -t kubepolicy:dev .

clean:
	rm -rf bin
