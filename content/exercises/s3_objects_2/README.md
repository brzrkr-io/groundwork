# S3 objects: fix the log cleanup

Log buckets grow forever unless something deletes old files. Here you finish
a retention job and fix two bugs that could wipe out data you must keep.

## Learn

**Recap.** An S3 bucket holds objects found by key, and "folders" are just key
prefixes such as `logs/checkout/`. You list keys under a prefix with the
`ListObjectsV2` paginator and `Prefix: aws.String(prefix)`, as in *S3 objects: upload, verify and list configs*.

**New: deleting and dating objects.** Deleting one object looks like every
other call:

```go
_, err := api.DeleteObject(ctx, &s3.DeleteObjectInput{
	Bucket: aws.String("ops-logs"),
	Key:    aws.String("logs/checkout/2026-08-01.log.gz"),
})
```

There is no undo and no "are you sure?". That is why cleanup jobs usually
have a *dry run* mode, which prints what would be deleted and deletes nothing.

Every listed `types.Object` has a `LastModified` field, a `*time.Time` saying
when it was written. It is a pointer, so it can be `nil`. As in *Time and
durations*, `now.Sub(t)` gives a `time.Duration`, the object's age. The
function receives `now` as a parameter instead of calling `time.Now()`, so the
tests can pick a fixed time.

## Your task

The `ops-logs` bucket keeps app logs under `logs/<service>/` and audit records
under `audit/`. The retention job should delete logs older than 30 days, one
service at a time, and must never touch anything else.

### 1. Find the expired keys

`ExpiredKeys` returns every key. Make it return only the keys of objects that
are **more than** `maxAge` old, in their original order:

- Skip objects whose `LastModified` is `nil`. You never delete what you can't
  date.
- An object is expired when `now.Sub(*obj.LastModified) > maxAge`. An object
  exactly `maxAge` old is kept.

```go
// now is 2026-10-11, maxAge is 30 days
ExpiredKeys(objects, now, 30*24*time.Hour)
// => []string{"logs/api/2026-08-01.log.gz", "logs/api/2026-09-01.log.gz"}
//    (the logs from 10 days and 1 day ago are kept)
```

### 2. Stay inside the prefix

`Cleanup` lists the **whole bucket**, so cleaning `logs/checkout/` also
deletes old billing logs and audit records. Fix it:

- If `prefix` is `""`, return `nil, ErrEmptyPrefix` straight away, before any
  call to S3. A cleanup with no prefix would clean the whole bucket.
- Add `Prefix: aws.String(prefix)` to the `ListObjectsV2Input`.

`Cleanup` returns the deleted keys in the order S3 lists them:

```go
Cleanup(ctx, api, "ops-logs", "logs/checkout/", now, 30*24*time.Hour, false)
// => []string{"logs/checkout/2026-08-01.log.gz", "logs/checkout/2026-09-01.log.gz"}
//    and only those two objects are deleted
```

Task 1 must pass first, because `Cleanup` uses `ExpiredKeys`.

### 3. Make dry run safe

`Cleanup` ignores `dryRun` and deletes anyway. When `dryRun` is `true`, skip
the `DeleteObject` call but still append the key to the result, so the
caller sees what would be deleted:

```go
Cleanup(ctx, api, "ops-logs", "logs/billing/", now, 30*24*time.Hour, true)
// => []string{"logs/billing/2026-07-15.log.gz"}, and nothing is deleted
```

## Learn more

- [Amazon S3: Deleting a single object](https://docs.aws.amazon.com/AmazonS3/latest/userguide/delete-objects.html)
- [Amazon S3: Managing the lifecycle of objects](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html)
- [pkg.go.dev: s3.Client.DeleteObject](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3#Client.DeleteObject)
- [pkg.go.dev: time.Time.Sub](https://pkg.go.dev/time#Time.Sub)
