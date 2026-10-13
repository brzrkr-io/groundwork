# Looping with range: rollout report

After every rollout, your team posts a short report in chat: how many health
checks failed, how each host's deploy went, and how deep the service's DNS name
is. Each number comes from a `range` loop.

## Learn

### Recap

`for i, v := range s` walks a slice with index and element; use `_` for the one
you don't need. `for i, r := range str` walks a string one character (*rune*)
at a time, and you compare a rune with a rune literal in single quotes:
`r == '.'`.

### Percentages with integers

Dividing two `int`s throws away the remainder: `200 / 3` is `66`. Multiply
before you divide, or the answer collapses to 0:

```go
failed, total := 2, 3
failed / total * 100 // 0, because 2 / 3 is 0
failed * 100 / total // 66
```

## Your task

Write the three numbers for the rollout report.

### 1. Failed health checks

`FailurePercent(checks)` returns the percentage of checks that failed, where
`true` means the check passed and `false` means it failed. Round down, which is
what integer division does. If there are no checks, return 0.

```go
FailurePercent([]bool{true, false, true, true})
// => 25

FailurePercent([]bool{false, false, true})
// => 66

FailurePercent(nil)
// => 0
```

### 2. Deploy summary

`Summary(results)` counts the deploy results and returns them as one line in
exactly this format:

```text
<ok> ok, <failed> failed, <skipped> skipped
```

Every element of `results` is exactly `"ok"`, `"failed"` or `"skipped"`.
Always show all three counts, even when one is 0.

```go
Summary([]string{"ok", "ok", "failed", "skipped", "ok"})
// => "3 ok, 1 failed, 1 skipped"

Summary(nil)
// => "0 ok, 0 failed, 0 skipped"
```

### 3. DNS labels

A DNS name is made of *labels* separated by dots. Kubernetes gives every
service a name like `api.default.svc.cluster.local`, which has 5 labels.
`Labels(name)` returns how many labels `name` has. An empty name has 0. Names
never start or end with a dot and never contain two dots in a row. Range over
the string and count the dots.

```go
Labels("api.default.svc.cluster.local")
// => 5

Labels("localhost")
// => 1

Labels("")
// => 0
```

## Learn more

- [Go by Example: Range over Built-in Types](https://gobyexample.com/range-over-built-in-types)
- [Go spec: Arithmetic operators](https://go.dev/ref/spec#Arithmetic_operators)
- [Kubernetes: DNS for Services and Pods](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/)
