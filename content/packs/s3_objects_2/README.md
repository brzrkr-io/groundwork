# S3 objects: fix the log cleanup

Log buckets grow forever unless something deletes old files. Here you finish
a retention job and fix two bugs that could wipe out data you must keep.

**New here:** `DeleteObject`, `LastModified`, dry-run mode.

## Learn

### Recap: prefixes

An S3 bucket holds objects found by key. "Folders" are just key prefixes,
like `logs/checkout/`. The `ListObjectsV2` paginator lists keys under a prefix
when you set the input's `Prefix` field, as in *S3 objects: upload, verify and
list configs*.

### Deleting an object

Deleting looks like every other call:

```go
_, err := api.DeleteObject(ctx, &s3.DeleteObjectInput{
	Bucket: aws.String("ops-logs"),
	Key:    aws.String("logs/checkout/2026-08-01.log.gz"),
})
```

There is no undo and no "are you sure?". So cleanup jobs usually have a *dry
run* mode: report what would be deleted, delete nothing.

### Dating an object

Each listed `types.Object` has `LastModified`, a `*time.Time` (when it was
written). It is a pointer, so it can be `nil`. Put `*` in front of a pointer
to read the value it points to.

As in *Time and durations*, `later.Sub(earlier)` gives a `time.Duration`, and
durations compare with `<`, `>`, `==`:

```go
uptime := now.Sub(bootedAt)
if uptime > 7*24*time.Hour {
	fmt.Println("reboot soon")
}
```

> **Why is it written like this?** The `*` is required to read through a
> pointer, and reading a `nil` pointer crashes, so check `nil` first. Passing
> `now` in as a parameter instead of calling `time.Now()` is the author's
> choice: the tests can pick a fixed time.

## Your task

The `ops-logs` bucket keeps app logs under `logs/<service>/` and audit records
under `audit/`. The retention job should delete logs older than 30 days, one
service at a time, and must never touch anything else.

### 1. Find the expired keys

**Goal:** `ExpiredKeys` returns only the keys of objects **more than** `maxAge` old, in their original order.

- Skip objects whose `LastModified` is `nil`. Never delete what you can't date.
- An object's age is `now` minus its `LastModified`. Expired means age
  greater than `maxAge`. An object exactly `maxAge` old is kept.
- No expired objects gives `nil`.

```go
// now is 2026-10-11, maxAge is 30 days
ExpiredKeys(objects, now, 30*24*time.Hour)
// => []string{"logs/api/2026-08-01.log.gz", "logs/api/2026-09-01.log.gz"}
//    (the logs from 10 days and 1 day ago are kept)
```

**Think it through:**
- What two checks does each object pass before its key is appended, and in which order?
- "More than" or "at least": which comparison operator keeps the exact-`maxAge` object?

### 2. Stay inside the prefix

**Goal:** `Cleanup` refuses an empty prefix and only lists keys under `prefix`.

Right now it lists the **whole bucket**, so cleaning `logs/checkout/` also
deletes old billing logs and audit records. Fix it:

- `prefix` is `""`: return `nil, ErrEmptyPrefix` straight away, before any
  call to S3.
- Otherwise list only keys under `prefix`.

`Cleanup` returns the deleted keys in the order S3 lists them:

```go
Cleanup(ctx, api, "ops-logs", "logs/checkout/", now, 30*24*time.Hour, false)
// => []string{"logs/checkout/2026-08-01.log.gz", "logs/checkout/2026-09-01.log.gz"}
//    and only those two objects are deleted
```

Task 1 must pass first, because `Cleanup` uses `ExpiredKeys`.

**Think it through:**
- "Before any call to S3": which line must the check come before?
- Which field of `ListObjectsV2Input` is missing?

### 3. Make dry run safe

**Goal:** with `dryRun` set to `true`, `Cleanup` deletes nothing but still returns the keys it would delete.

Today it ignores `dryRun` and deletes anyway. Skip the `DeleteObject` call
(and its error check) in a dry run, but still append the key to the result.

```go
Cleanup(ctx, api, "ops-logs", "logs/billing/", now, 30*24*time.Hour, true)
// => []string{"logs/billing/2026-07-15.log.gz"}, and nothing is deleted
```

**Think it through:**
- Which lines in the inner loop should depend on `dryRun`, and which should always run?
- How do you write "if not dry run" with the `!` operator?

## Learn more

- [Amazon S3: Deleting a single object](https://docs.aws.amazon.com/AmazonS3/latest/userguide/delete-objects.html)
- [Amazon S3: Managing the lifecycle of objects](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html)
- [pkg.go.dev: s3.Client.DeleteObject](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3#Client.DeleteObject)
- [pkg.go.dev: time.Time.Sub](https://pkg.go.dev/time#Time.Sub)
