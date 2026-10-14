# CI workflows: fix the broken pipeline

A red CI run is a message, and the run log says what went wrong. Here four
workflows broke this week; each file has a `TODO` that quotes the log or the
review. Read the symptom, find the cause, fix only that.

**New here:** reading a run log, YAML numbers vs strings, a first look at
`strategy.matrix`, `needs` with a list, reviewing action pins,
`if: failure()`.

## Learn

### Read the log at the first red step

Open a run in the repository's Actions tab, then a job. Every step has its
own log. The first step with a red ✗ is where things went wrong; the steps
after it are skipped, so their logs are empty.

The error tells you *what* failed, not *why*. Ask: what was on the machine
when that step ran? A job starts on an **empty** runner, and steps run top to
bottom, so a step only sees what the steps **above** it produced.

```yaml
steps:
  - run: shellcheck scripts/*.sh       # fails: shellcheck: command not found
  - run: sudo apt-get install -y shellcheck
```

The tool is installed, just one step too late. The fix is the order, not the
command.

### YAML guesses types

YAML decides what a value is from how it looks. Anything that looks like a
number becomes a number, and a number drops trailing zeros:

| You write | YAML reads | As a version |
|---|---|---|
| `3.9` | the number 3.9 | fine, by luck |
| `3.10` | the number 3.1 | wrong version |
| `'3.10'` | the string "3.10" | right |
| `v3.10` | the string "v3.10" | right (can't be a number) |
| `[3.9, 3.10]` | the list 3.9, 3.1 | one wrong |

GitHub only sees what YAML read, so a step asked for `3.10` gets 3.1. The
tests in this folder read YAML the same way, so they see 3.1 too.

> **Why is it written like this?**
> - **YAML requires:** quotes (single `'...'` or double `"..."`) to make a
>   number-looking value a string.
> - **Convention:** quote *every* version number, even ones that survive
>   unquoted, like `1.26`. Then nobody has to remember which ones are safe.

### A first look at a matrix

A *matrix* runs one job several times, once per value in a list. Inside the
job, `${{ matrix.<name> }}` stands for this copy's value:

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        postgres: ['15', '16']       # two copies of the job
    steps:
      - run: echo "testing against Postgres ${{ matrix.postgres }}"
```

The Actions tab shows each copy with its value in brackets: `test (15)`,
`test (16)`. You'll build matrices yourself in
*CI workflows: a matrix, a cache and artifacts*.

### needs: one job or several

As in *CI workflows*, jobs start **at the same time** unless one `needs`
another. To wait for several jobs, give `needs` a list. The job then starts
only after **all** of them succeed, and it is skipped if any one fails:

```yaml
  publish:
    needs: [build, scan]     # waits for both; skipped if either fails
```

With no `needs`, a job doesn't care how the others end: it may finish before
they do.

### Review the pins

In *CI workflows* you pinned every action to a full commit SHA with its
release in a comment. Here's why that rule is worth the long lines.

A tag is only a name pointing at a commit, and it can be moved. In March
2025 attackers took over the popular `tj-actions/changed-files` action and
moved its version tags to code that printed CI secrets into the build logs.
Workflows that used any of its version tags ran the attacker's code.
Workflows pinned to a commit SHA kept running the old code and were safe.

When you review a `uses` line, look for these:

| You see | Problem | Fix |
|---|---|---|
| `owner/repo` with no `@` | GitHub rejects the workflow file | add the SHA and comment |
| `@main`, `@master` | a branch: new code with every push | the release's SHA |
| `@v4`, `@v4.6.0` | a tag: its owner (or a thief) can move it | the release's SHA |
| `@dbcb813` | a short SHA: GitHub won't resolve it | all 40 characters |
| `@<SHA> # v4.6.0` | the comment must name the release the SHA *is* | check both with `git ls-remote` |

A correct pin, for an action this exercise doesn't use:

```yaml
- uses: docker/login-action@dbcb813823bdd20940b903addbd779551569679f # v4.6.0
```

To check one, list the tag as in *CI workflows* and compare:

```sh
git ls-remote --tags https://github.com/docker/login-action 'v4.6.0*'
# dbcb813823bdd20940b903addbd779551569679f	refs/tags/v4.6.0
```

> **Why is it written like this?**
> - **GitHub requires:** an `@ref` on every `uses` of another repository,
>   and a commit SHA must be the full 40 characters.
> - **Convention:** the `# vX.Y.Z` comment after a SHA. YAML ignores
>   everything after ` #`, so it changes nothing for GitHub.
> - **Watch out:** for an *annotated* tag `git ls-remote` prints two lines,
>   and the one ending in `^{}` holds the commit SHA you want.

### Steps that run on failure

Normally a failed step skips every step after it. A step with
`if: failure()` is the opposite: it runs **only** when a step above it
failed. That's how a job posts "lint is red" to the team chat.

## Your task

Your team's CI broke in four ways this week. Each file is the workflow as it
is now, with a `TODO` above the broken part. Fix the cause, and keep
everything else as it is.

### 1. Check out before testing

**Goal:** in `1_test.yml`, make the `go test` step find the code.

The run log today:

```text
Run go test ./...
go: go.mod file not found in current directory or any parent directory; see 'go help modules'
Error: Process completed with exit code 1.
```

- The job `test` keeps its three steps: `actions/checkout`,
  `actions/setup-go` and the step that runs `go test ./...`.
- They run in this order: checkout, setup-go, `go test ./...`.

**Think it through:**
- What is on the runner when `go test` starts? Which step puts the code
  there?

### 2. Quote the Go versions

**Goal:** in `2_versions.yml`, make the matrix test Go 1.20 and Go 1.26,
not Go 1.2.

- The matrix variable `go` lists exactly `1.20` and `1.26`, in that order.
- Both are strings: each one is in quotes.
- setup-go keeps `go-version: ${{ matrix.go }}`.

Once it's fixed, the Actions tab shows `test (1.20)` and `test (1.26)`.

**Think it through:**
- What does YAML make of `1.20`, and what does setup-go then look for?
- `1.26` works unquoted. Should it stay that way?

### 3. Make deploy wait for the checks

**Goal:** in `3_deploy.yml`, make `deploy` start only after both `lint` and
`test` have passed.

- The job `deploy` needs both `lint` and `test`.
- Keep the three jobs, and keep deploy's steps
  (`./scripts/deploy.sh production`).

```text
today:   lint ──┐                 fixed:  lint ──┐
         test ──┼─ all at once            test ──┴─► deploy
         deploy ┘
```

**Think it through:**
- One job to wait for, or two? Which spelling of `needs` fits?
- What happens to `deploy` when `test` fails?

### 4. Pin every action

**Goal:** in `4_pinned.yml`, pin every `uses` to a commit that can't move
without you.

- Every `uses` line is `owner/repo@<full 40-character commit SHA> # vX.Y.Z`,
  as in *CI workflows*: the SHA and release from the table below.
- The comment names the same release as the SHA.
- Keep all four action steps (checkout, setup-go, golangci-lint,
  slack-notify), with their `if:` and `with:` as they are.

The releases to pin to:

| Action | Release | Commit |
|---|---|---|
| `actions/checkout` | v7.0.1 | `3d3c42e5aac5ba805825da76410c181273ba90b1` |
| `actions/setup-go` | v7.0.0 | `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` |
| `golangci/golangci-lint-action` | v9.3.0 | `ba0d7d2ec06a0ea1cb5fa41b2e4a3ab91d21278a` |
| `acme/slack-notify` (your company's own) | v1.4.2 | `521900d1e65917abfbbe5ca9f186b195f932c7bd` |

**Think it through:**
- Which `uses` line follows a branch, which a tag, which a short SHA, and
  which has no ref at all? The table in Learn names each problem.
- Who could change the code behind each one tomorrow?

## Learn more

- [Using workflow run logs](https://docs.github.com/en/actions/how-tos/monitor-workflows/use-workflow-run-logs)
- [Using jobs in a workflow](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-jobs): `needs`
- [Secure use reference](https://docs.github.com/en/actions/reference/security/secure-use): why to pin actions to a full commit SHA
- [actions/setup-go](https://github.com/actions/setup-go): see the note on quoting `go-version`
