# Drift detection: ignore AWS-managed tags

The nightly drift check pages the on-call engineer every night, even though
nobody touched the payments stack. Two bugs make the report useless; fix them.

**New here:** AWS-managed `aws:` tags, `strings.HasPrefix`, `continue`.

## Learn

### Recap

As in *Drift detection: a tiny terraform plan*, you compare desired and actual
state by `kind/name` IDs. Missing in AWS → "create". Different tags → "update"
with the sorted differing keys. Unmanaged → "delete".

### AWS-managed tags

Some services tag your resources for you. The payments stack was created by
CloudFormation, so every instance carries tags like these:

| Key | Value |
|---|---|
| `aws:cloudformation:stack-name` | `payments` |
| `aws:cloudformation:logical-id` | `ApiServer1` |
| `aws:cloudformation:stack-id` | `arn:aws:cloudformation:us-east-1:...:stack/payments/...` |

The prefix `aws:` (with the colon) is reserved for AWS. Nobody can edit these
tags, and nobody writes them in the desired state, so they are **never** drift.
A tag like `aws-team` (dash, no colon) is an ordinary tag a person set.

### Checking how a string starts

`strings.HasPrefix(s, prefix)` reports whether `s` starts with `prefix`:

```go
strings.HasPrefix("kube-system", "kube-") // => true
strings.HasPrefix("kubelet", "kube-")     // => false
```

### Skipping one loop round with `continue`

`continue` jumps straight to the next round of the loop:

```go
for _, name := range files {
	if strings.HasPrefix(name, ".") {
		continue // hidden file: skip it
	}
	fmt.Println(name)
}
```

> **Why is it written like this?** `continue` is a Go keyword. Putting the
> "skip" check first, at the top of the loop body, is convention: it keeps the
> main work un-indented.

### Sorting a slice of structs

`slices.SortFunc(s, cmp)` sorts `s` using your compare function. The function
gets two elements and returns a negative number, 0, or a positive number;
`cmp.Compare` does that for strings and numbers:

```go
slices.SortFunc(pods, func(a, b Pod) int {
	return cmp.Compare(a.Name, b.Name)
})
```

> **Why is it written like this?** The signature `func(a, b T) int` is required
> by `SortFunc`. The names `a` and `b` are convention. Which field you compare
> is your choice.

### Files

| File | What it holds |
|---|---|
| `instances.go` | Complete. `Resource`, `ID()`, `Change`, `byID`, and `Actual` (the `DescribeInstances` paginator; skips terminated instances; `Name` tag becomes the name). |
| `exercise.go` | `ChangedKeys` and `Diff`, each with one bug to fix. |

## Your task

The on-call engineer wants a quiet night when nothing changed, and a report
that reads the same way every time. Fix `exercise.go`.

### 1. Ignore AWS-managed tags

**Goal:** `ChangedKeys(want, got)` never reports a key that starts with `aws:`,
on either side.

Everything else stays the same: different values and keys only one side has
still count, and the result is sorted. Two nil maps give nil. Add the `strings`
import.

```go
want := map[string]string{"Env": "prod", "Owner": "payments"}
got := map[string]string{"Env": "prod", "Owner": "bob", "aws:cloudformation:stack-name": "payments"}
ChangedKeys(want, got)
// => []string{"Owner"}   (the starter returns [Owner aws:cloudformation:stack-name])

ChangedKeys(want, map[string]string{"Env": "prod", "Owner": "payments", "aws-team": "sre", "aws:backup:source-resource": "pay-api-1"})
// => []string{"aws-team"}
```

**Think it through:**
- `ChangedKeys` has two loops. Which variable holds the key in each?
- What exact prefix marks an AWS tag? Compare it with `aws-team`.

### 2. Make the report order stable

**Goal:** `Diff` returns its changes sorted by `ID`.

Right now the order depends on random map order, and deletes always come last.
Sort `changes` by the `ID` field just before returning. Add the `cmp` import.

```go
desired: pay-api-1, pay-api-2, pay-worker-1 (all Env=prod, Owner=payments)
actual:  pay-api-1 (Owner=bob), pay-worker-1, batch-tmp, plus CloudFormation tags
// => [{delete instance/batch-tmp []} {update instance/pay-api-1 [Owner]} {create instance/pay-api-2 []}]
```

`TestDiff` feeds instances that carry CloudFormation tags, so it only passes
once task 1 is fixed too.

**Think it through:**
- Where is the last moment before `changes` leaves the function?
- Which field of `Change` do you compare?

## Learn more

- [AWS: Tag your Amazon EC2 resources (the `aws:` prefix)](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Tags.html)
- [strings.HasPrefix](https://pkg.go.dev/strings#HasPrefix)
- [slices.SortFunc](https://pkg.go.dev/slices#SortFunc)
