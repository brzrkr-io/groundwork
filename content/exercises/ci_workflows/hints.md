## Task 1: Say hello on every push

A workflow has three top-level keys: `name`, `on` and `jobs`. The job
needs a machine (`runs-on`) and a list of `steps`.

### Closer

`on: push` fits on one line. Under `jobs:`, the job id `hello:` is indented
two spaces, and its keys two more. A shell step is `- run: <command>`.

### Answer

```yaml
name: hello

on: push

jobs:
  hello:
    runs-on: ubuntu-latest
    steps:
      - run: go version
```

## Task 2: Test every push and pull request

Start from the shape of task 1. The triggers now have settings, and the job
gets four steps: two actions, then two commands. Checkout must be first,
because the runner starts empty.

### Closer

`push:` gets `branches: [main]` indented under it; `pull_request:` stays
bare. The setup-go step needs a `with:` block indented under it, holding
`go-version-file: go.mod`.

### Answer

```yaml
name: test

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
      - run: go vet ./...
      - run: go test ./...
```

## Task 3: Build after the tests pass

The new job sits beside `test:` under `jobs:`. It starts on its own empty
machine, so it needs its own checkout and Go before it can build.

### Closer

Add `needs: test` next to `runs-on`. The steps are the same two setup steps
as in the `test` job, then `- run: go build -o bin/statusd .`.

### Answer

```yaml
  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
      - run: go build -o bin/statusd .
```
