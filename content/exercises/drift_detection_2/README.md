# Drift detection: ignore AWS-managed tags

The nightly drift check pages the on-call engineer every single night, even
though nobody touched the payments stack. Two bugs make the report useless;
fix them.

## Learn

**Recap.** As in *Drift detection: a tiny terraform plan*, you compare the
desired state with the actual state from AWS, keyed by `kind/name` IDs.
Resources missing in AWS are "create", differing tags are "update" with the
sorted differing keys, and unmanaged resources are "delete". Ranging over a map
visits keys in random order, so the list of changes must be sorted by ID.

**New: AWS-managed tags.** Some services tag your resources for you. The
payments stack was created by CloudFormation, so every instance in it carries
tags like these:

| Key | Value |
|---|---|
| `aws:cloudformation:stack-name` | `payments` |
| `aws:cloudformation:logical-id` | `ApiServer1` |
| `aws:cloudformation:stack-id` | `arn:aws:cloudformation:us-east-1:123456789012:stack/payments/...` |

The `aws:` prefix is reserved for AWS. You can't add, edit or delete these
tags, and nobody writes them in the desired state. So they are never drift.
Only the full prefix `aws:` (with the colon) is reserved: a tag such as
`aws-team` is an ordinary tag that a person set.

`strings.HasPrefix(s, prefix)` reports whether `s` starts with `prefix`:

```go
strings.HasPrefix("aws:cloudformation:stack-name", "aws:") // => true
strings.HasPrefix("aws-team", "aws:")                      // => false
```

`instances.go` is complete: it holds `Resource`, `Change`, `byID`, and
`Actual`, which gathers the instances with the `DescribeInstances` paginator.

## Your task

The on-call engineer wants a quiet night when nothing changed, and a report
that reads the same way every time. Fix `exercise.go`.

### 1. Ignore AWS-managed tags

`ChangedKeys(want, got)` reports CloudFormation's tags as drift. Change it so
that any key starting with `aws:` is skipped, on either side. Everything else
stays the same: real differences, and keys that only one side has, still count,
and the result is sorted. Add the `strings` import.

```go
want := map[string]string{"Env": "prod", "Owner": "payments"}
got := map[string]string{"Env": "prod", "Owner": "bob", "aws:cloudformation:stack-name": "payments"}
ChangedKeys(want, got)
// => []string{"Owner"}   (the starter returns []string{"Owner", "aws:cloudformation:stack-name"})

ChangedKeys(want, map[string]string{"Env": "prod", "Owner": "payments", "aws-team": "sre"})
// => []string{"aws-team"}
```

### 2. Make the report order stable

`Diff` returns the right changes, but in the order the maps happen to visit
them, and deletes always come last. Sort `changes` by `ID` before returning
them, so the report is the same every night:

```go
desired: pay-api-1, pay-api-2, pay-worker-1 (all Env=prod, Owner=payments)
actual:  pay-api-1 (Owner=bob), pay-worker-1, batch-tmp, plus CloudFormation tags
// => [{delete instance/batch-tmp []} {update instance/pay-api-1 [Owner]} {create instance/pay-api-2 []}]
```

`TestDiff` runs `Actual` on instances that carry CloudFormation tags, so it
only passes once task 1 is fixed too.

## Learn more

- [AWS: Tag your Amazon EC2 resources (tag restrictions, the `aws:` prefix)](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Tags.html)
- [AWS CloudFormation: Resource tags](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-resource-tags.html)
- [slices.SortFunc](https://pkg.go.dev/slices#SortFunc)
