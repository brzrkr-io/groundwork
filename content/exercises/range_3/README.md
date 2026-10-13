# Looping with range: rollout report

After every rollout, your team posts a short report in chat: how many health
checks failed, how each host's deploy went, and how deep the service's DNS name
is. Each number comes from a `range` loop.

**New here:** nothing new; practice (plus integer percentages).

## Learn

### Recap

- `for i, v := range s` walks a slice; put `_` for the part you don't need.
- `for _, r := range str` walks a string one character (*rune*) at a time.
- Compare a rune with a rune literal in single quotes: `r == '-'`.
- `fmt.Sprintf("%d of %d", a, b)` builds a string; each `%d` is filled by the
  next number, in order. It needs `import "fmt"`.
- `n++` adds one to a counter.

> **Why is it written like this?** Single quotes are required for one rune;
> double quotes make a string, and Go won't compare a rune with a string.

### Percentages with integers

Dividing two `int`s throws away the remainder: `200 / 3` is `66`. So multiply
before you divide, or the result collapses to 0:

```go
used, disk := 3, 4
used / disk * 100 // 0, because 3 / 4 is 0
used * 100 / disk // 75
```

Dividing by 0 crashes the program, so check for an empty input first.

## Your task

Write the three numbers for the rollout report.

### 1. Failed health checks

**Goal:** make `FailurePercent(checks)` return the percentage of checks that
failed.

- `true` means passed, `false` means failed.
- Round down (integer division already does this).
- If there are no checks, return 0.

```go
FailurePercent([]bool{true, false, true, true})
// => 25

FailurePercent([]bool{false, false, true})
// => 66

FailurePercent(nil)
// => 0
```

**Think it through:**
- What must you check before dividing?
- Which values do you count, and what do you divide that count by?

### 2. Deploy summary

**Goal:** make `Summary(results)` count the results and return one line in
exactly this format:

```text
<ok> ok, <failed> failed, <skipped> skipped
```

- Every element is exactly `"ok"`, `"failed"` or `"skipped"`.
- Always show all three counts, even when one is 0.
- The three numbers come from your counters; the words, commas and spaces are
  fixed text.

```go
Summary([]string{"ok", "ok", "failed", "skipped", "ok"})
// => "3 ok, 1 failed, 1 skipped"

Summary(nil)
// => "0 ok, 0 failed, 0 skipped"
```

**Think it through:**
- How many counters do you need, and what does each start at?
- Write out the format string with a `%d` where each number goes. Are the
  spaces after the commas there?

### 3. DNS labels

**Goal:** make `Labels(name)` return how many dot-separated labels `name` has.

A DNS name is made of *labels* separated by dots. Kubernetes gives each service
a name like `api.default.svc.cluster.local`, which has 5 labels.

- An empty name has 0 labels.
- Names never start or end with a dot, and never have two dots in a row.
- Range over the string and count the dots.

```go
Labels("api.default.svc.cluster.local")
// => 5

Labels("localhost")
// => 1

Labels("")
// => 0
```

**Think it through:**
- `localhost` has no dots but 1 label. How do dots relate to labels?
- Which input is the special case?

## Learn more

- [Go by Example: Range over Built-in Types](https://gobyexample.com/range-over-built-in-types)
- [Go spec: Arithmetic operators](https://go.dev/ref/spec#Arithmetic_operators)
- [Kubernetes: DNS for Services and Pods](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/)
