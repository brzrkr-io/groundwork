# Everyday commands for golinks. Run `make` alone to see them.

IMAGE   ?= golinks:dev
CLUSTER ?= golinks
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: help run test vet build image kind-up kind-load deploy logs kind-down

help: ## list the targets
	@grep -E '^[a-z-]+:.*## ' Makefile | awk -F ':.*## ' '{printf "  %-10s %s\n", $$1, $$2}'

run: ## run golinks on this machine (http://localhost:8080)
	go run ./cmd/golinks

test: ## run every test with the race detector
	go test -race ./...

vet: ## run go vet
	go vet ./...

build: ## build bin/golinks with the version stamped in
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)" -o bin/golinks ./cmd/golinks

image: ## build the container image (IMAGE=golinks:dev)
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE) .

kind-up: ## create a local kind cluster
	kind create cluster --name $(CLUSTER)

kind-load: image ## build the image and copy it into the kind cluster
	kind load docker-image $(IMAGE) --name $(CLUSTER)

deploy: ## apply the dev overlay to the current kubectl context
	kubectl apply -k deploy/overlays/dev
	kubectl rollout status deployment/golinks --timeout=120s

logs: ## follow the logs of the running pod
	kubectl logs -f deployment/golinks

kind-down: ## delete the kind cluster
	kind delete cluster --name $(CLUSTER)
