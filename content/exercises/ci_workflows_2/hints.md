## Task 1: Check out before testing

The runner starts empty, and steps run top to bottom. When `go test` runs,
which step has already happened, and which hasn't yet?

### Closer

The commands are all right; only their order is wrong. The step that clones
the code has to come before every step that needs the code. Move one whole
`- uses: ...` line; setup-go keeps its `with:` block.

### Answer

```yaml
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
        with:
          go-version: '1.26'
      - run: go test ./...
```

## Task 2: Quote the Go versions

YAML turned `1.20` into a number and dropped the zero, so setup-go was asked
for Go 1.2. What makes YAML keep a value exactly as written?

### Closer

Put quotes around each item of the list: `['1.20', ...]`. Do the same for
`1.26`, even though it survives unquoted; the test checks both.

### Answer

```yaml
      matrix:
        go: ['1.20', '1.26']
```

## Task 3: Make deploy wait for the checks

Without `needs`, deploy starts at the same moment as the checks and doesn't
care how they end. It has to wait for two jobs, not one.

### Closer

`needs` takes a list for several jobs: `needs: [first, second]`. Put it in
the `deploy` job, at the same indentation as `runs-on`.

### Answer

```yaml
  deploy:
    needs: [lint, test]
    runs-on: ubuntu-latest
```

## Task 4: Pin every action

Look at each `uses` line. One has a short SHA, one a tag, one a branch, and
one no `@` at all. None of them is safe yet. The README's table has the
releases to pin to.

### Closer

Every line becomes `owner/repo@<all 40 hex characters> # vX.Y.Z`: the SHA
and the release from the same row of the table. Keep `if: failure()` and
`with:` as they are.

### Answer

```yaml
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
        with:
          go-version-file: go.mod
      - uses: golangci/golangci-lint-action@ba0d7d2ec06a0ea1cb5fa41b2e4a3ab91d21278a # v9.3.0
      - if: failure()
        uses: acme/slack-notify@521900d1e65917abfbbe5ca9f186b195f932c7bd # v1.4.2
        with:
          channel: statusd-alerts
```
