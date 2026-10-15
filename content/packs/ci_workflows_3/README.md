# CI workflows: a matrix, a cache and artifacts

Real CI does three more things than the workflows so far: it tests every
combination you ship, it skips work it already did last time, and it passes
files from one job to the next. A *matrix*, a *cache* and *artifacts* do
exactly that.

**New here:** `${{ }}` expressions, the `matrix` and `runner` contexts,
`hashFiles`, `strategy.matrix`, `fail-fast`, `actions/cache` (`path`, `key`,
`restore-keys`), `actions/upload-artifact`, `actions/download-artifact`.

## Learn

### Expressions: `${{ ... }}`

GitHub replaces every `${{ expression }}` in a workflow with its value just
before the step runs. An expression reads a *context* (facts about the run)
or calls a function:

| Expression | Value |
|---|---|
| `${{ matrix.region }}` | this job copy's value of the matrix variable `region` |
| `${{ runner.os }}` | the runner's system: `Linux`, `macOS` or `Windows` |
| `${{ github.sha }}` | the commit being tested |
| `${{ hashFiles('**/package-lock.json') }}` | a hash of the matching files; it changes when they change |

Text inside an expression takes **single** quotes, as in
`hashFiles('**/go.sum')`. Double quotes there are an error.

### A matrix runs one job many times

Under `strategy`, a `matrix` lists variables, each with a list of values.
GitHub runs the job once for **every combination**, all at the same time:

```yaml
jobs:
  test:
    strategy:
      matrix:
        python: ['3.12', '3.13']
        os: [ubuntu-latest, windows-latest]
    runs-on: ${{ matrix.os }}             # matrix values work anywhere in the job
    steps:
      - run: echo "Python ${{ matrix.python }} on ${{ runner.os }}"
```

That is 2 × 2 = 4 jobs, shown as `test (3.12, ubuntu-latest)`,
`test (3.12, windows-latest)`, `test (3.13, ubuntu-latest)` and
`test (3.13, windows-latest)`.

`actions/setup-go` takes the version either from a file
(`go-version-file: go.mod`) or directly (`go-version: '1.26'`). If you give
both, `go-version` wins, so a matrix job keeps just `go-version`.

### fail-fast

By default, as soon as one copy fails, GitHub **cancels** the copies that are
still running. That saves minutes, but you only learn about the first
failure. With `fail-fast: false` every copy runs to the end, and you see
exactly which combinations are broken.

> **Why is it written like this?**
> - **GitHub requires:** `fail-fast` sits under `strategy`, next to
>   `matrix`. Every key **inside** `matrix` is a variable, so
>   `fail-fast` there would become a third variable, not a setting.
> - **GitHub requires:** `runs-on` and `steps` stay at the job level; only
>   the values come from `${{ matrix... }}`.
> - **Convention:** quote versions in a matrix, as in
>   *CI workflows: fix the broken pipeline*.

### Caches: skip work you did last time

A fresh runner has none of the Go modules your code needs and none of the
compiled packages. Go keeps them in two folders (on Linux):

| Folder | What's in it | `go env` name |
|---|---|---|
| `~/go/pkg/mod` | downloaded modules | `GOMODCACHE` |
| `~/.cache/go-build` | compiled packages and test results | `GOCACHE` |

`actions/cache` saves folders at the end of a successful job and puts them
back in a later run. It takes three inputs:

| Input | Meaning |
|---|---|
| `path` | the folders to save and restore, one per line |
| `key` | the cache's name. Same key next run: *cache hit*, restore it |
| `restore-keys` | on a *miss*, restore the newest cache whose key **starts with** one of these lines |

A good key changes exactly when the folders' contents should: build it from
`${{ runner.os }}` (a Linux cache is useless on macOS) and a hash of the
lock file. For a Node.js project it looks like this:

```yaml
- uses: actions/cache@55cc8345863c7cc4c66a329aec7e433d2d1c52a9 # v6.1.0
  with:
    path: |
      ~/.npm
    key: ${{ runner.os }}-npm-${{ hashFiles('**/package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-npm-
```

`path: |` starts a *block*: each indented line below it is one line of the
value. When the lock file changes, the key misses, `restore-keys` brings back
the last cache for this OS, and the job only downloads what's new. At the
end, it saves a cache under the new key. After an exact hit it saves
nothing: a cache, once saved, is never overwritten.

`hashFiles` hashes files of the checked-out code. Before
`actions/checkout`, there is no lock file and the hash is empty.

> **Why is it written like this?**
> - **Author's choice:** `actions/setup-go` already caches both folders by
>   default, keyed on `go.mod` (or `cache-dependency-path`). Write your own
>   `actions/cache` step when you want to choose the key yourself. Then turn
>   the built-in cache off with `cache: false`, or both would restore and
>   save the same folders.
> - **Convention:** the cache step goes right after setup-go, before
>   anything that downloads or builds.

### Artifacts: hand files to the next job

Every job gets its own fresh machine, so a file built in one job is gone in
the next. `actions/upload-artifact` stores files under a name, and
`actions/download-artifact` fetches them in a later job:

```yaml
jobs:
  report:
    runs-on: ubuntu-latest
    steps:
      - run: ./make-report.sh            # writes out/coverage.html
      - uses: actions/upload-artifact@cf430e030ddbb5b0abf93d22962f4752f3646cd9 # v7.0.2
        with:
          name: coverage
          path: out/coverage.html

  publish:
    needs: report                        # the artifact must exist first
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@9000827ccba6bdab643e8b6fd33ac0654aef8333 # v8.0.2
        with:
          name: coverage                 # lands as ./coverage.html
      - run: ls -l coverage.html
```

Three details that trip people up:

- An uploaded file keeps its name but not its folders: `out/coverage.html`
  downloads as `coverage.html`, into the working folder when you leave out
  `path`.
- By default an artifact is a zip file, and uploading doesn't keep the
  "executable" permission. A downloaded program needs `chmod +x` before it
  runs.
- The downloading job must `needs` the uploading one; otherwise it may start
  before the artifact exists.

| | Cache | Artifact |
|---|---|---|
| purpose | speed | keep or pass on an output |
| if it's missing | the job is slower, still correct | the next job fails |

## Your task

Three services need better CI: `inventory` must work on every Go version and
system it ships for, `billing`'s tests are slow, and `statusd`'s smoke test
should run the very binary that `build` produced.

### 1. Test on two Go versions and two systems

**Goal:** in `1_matrix.yml`, run the `test` job on Go 1.26 and Go 1.27,
each on Linux and macOS: four copies.

- The job `test` gets a `strategy` with a `matrix` of exactly two
  variables (the values in any order):
  - `go`: `'1.26'` and `'1.27'`, both quoted,
  - `os`: `ubuntu-latest` and `macos-latest`.
- `fail-fast: false`.
- `runs-on: ${{ matrix.os }}`.
- setup-go has `go-version: ${{ matrix.go }}`, and no `go-version-file`.
- The steps stay checkout, setup-go, `go test ./...`, in that order.

The Actions tab then shows `test (1.26, ubuntu-latest)`,
`test (1.26, macos-latest)`, `test (1.27, ubuntu-latest)` and
`test (1.27, macos-latest)`.

**Think it through:**
- Which two lines of the job read a matrix value?
- `fail-fast` and `matrix` are neighbours. Under which key do both go?

### 2. Cache modules and build output

**Goal:** in `2_cache.yml`, cache Go's module and build folders with
`actions/cache`, keyed on the runner's OS and `go.sum`.

- setup-go gets `cache: false`.
- A new step uses `actions/cache`, pinned as in *CI workflows*:
  `actions/cache@55cc8345863c7cc4c66a329aec7e433d2d1c52a9 # v6.1.0`.
- The steps run in this order: checkout, setup-go, cache, `go test ./...`.
- `path` lists two lines: `~/go/pkg/mod` and `~/.cache/go-build`.
- `key` contains `${{ runner.os }}` and `${{ hashFiles('**/go.sum') }}`,
  with any fixed text you like around them (such as `-go-`).
- `restore-keys` has one line: the start of your key, up to the
  `hashFiles` part. It keeps `${{ runner.os }}` and drops the hash.

The cache step's log, first run and second run:

```text
Cache not found for input keys: Linux-go-5d1b0a9c..., Linux-go-
Cache restored from key: Linux-go-5d1b0a9c...
```

**Think it through:**
- What must be on the runner before `hashFiles` can hash `go.sum`?
- When `go.sum` changes, why is last week's cache still worth restoring?

### 3. Hand the binary to the next job

**Goal:** in `3_artifacts.yml`, upload the binary from `build` and run it in
a new job `smoke`.

- In the job `build`, after `go build -o bin/statusd .`, add a step with
  `actions/upload-artifact`, `name: statusd` and `path: bin/statusd`.
- Add the job `smoke`:
  - it `needs` `build` and runs on `ubuntu-latest`,
  - step 1: `actions/download-artifact` with `name: statusd` and no
    `path`,
  - then make the file executable with `chmod +x statusd` and run
    `./statusd -version`: one step with `&&` between the two commands, or
    two steps.
- Pin both actions as in *CI workflows*:
  `actions/upload-artifact@cf430e030ddbb5b0abf93d22962f4752f3646cd9 # v7.0.2`
  and
  `actions/download-artifact@9000827ccba6bdab643e8b6fd33ac0654aef8333 # v8.0.2`.

`statusd -version` prints the version and exits. If the binary can't start,
the step fails, and the run goes red before anything ships.

**Think it through:**
- Does `smoke` need the source code or Go? Which steps does it really need?
- Where does the downloaded file land, and why not in `bin/`?

## Learn more

- [Running variations of jobs in a workflow](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/run-job-variations)
- [Evaluate expressions in workflows and actions](https://docs.github.com/en/actions/reference/workflows-and-actions/expressions): contexts and `hashFiles`
- [Dependency caching reference](https://docs.github.com/en/actions/reference/workflows-and-actions/dependency-caching)
- [Store and share data with workflow artifacts](https://docs.github.com/en/actions/tutorials/store-and-share-data)
