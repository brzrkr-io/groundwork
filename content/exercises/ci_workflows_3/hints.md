## Task 1: Test on two Go versions and two systems

The job needs a `strategy` block with two pieces: the `matrix` with its two
variables, and the setting that stops one failure from cancelling the rest.
Then two lines of the job read their values from the matrix.

### Closer

`strategy:` goes inside the job, at the same depth as `runs-on:`. Under it,
`fail-fast: false` and `matrix:` are neighbours; `go:` and `os:` go under
`matrix:`. The job reads them as `${{ matrix.os }}` and `${{ matrix.go }}`:
one in `runs-on`, one in setup-go's `with`, replacing `go-version-file`.

### Answer

```yaml
  test:
    strategy:
      fail-fast: false
      matrix:
        go: ['1.26', '1.27']
        os: [ubuntu-latest, macos-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version: ${{ matrix.go }}
      - run: go test ./...
```

## Task 2: Cache modules and build output

Two changes: switch off setup-go's own cache, then add one `actions/cache`
step between setup-go and the tests. The key is the runner's OS and a hash of
`go.sum`; the restore key is the same text without the hash.

### Closer

`cache: false` goes under setup-go's `with:`, next to `go-version-file`. The
new step's `with:` has three inputs; `path` and `restore-keys` are blocks:

```yaml
          path: |
            first/folder
            second/folder
          key: ${{ runner.os }}-go-...
```

### Answer

```yaml
      - uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
          cache: false
      - uses: actions/cache@v6
        with:
          path: |
            ~/go/pkg/mod
            ~/.cache/go-build
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-
      - run: go test ./...
```

## Task 3: Hand the binary to the next job

The `build` job ends with one more step that uploads the file. The new job
`smoke` waits for `build`, and needs no source code and no Go: only the
artifact and a shell.

### Closer

Upload with `actions/upload-artifact@v7` and `with:` `name` and `path`.
`smoke:` goes under `jobs:`, beside `build:`, with `needs: build`. Its first
step is `actions/download-artifact@v8` with the same `name`; the second runs
the exact command from the README.

### Answer

```yaml
      - run: go build -o bin/statusd .
      - uses: actions/upload-artifact@v7
        with:
          name: statusd
          path: bin/statusd

  smoke:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v8
        with:
          name: statusd
      - run: chmod +x statusd && ./statusd -version
```
