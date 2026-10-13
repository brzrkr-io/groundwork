## Task 1: Find the expired keys

Each object needs two checks before its key is appended: can you date it at
all, and is it old enough? Look at `obj.LastModified` and at `now.Sub`.

### Closer

First `continue` when `obj.LastModified == nil`. Then compare the age with
`maxAge`; the `*` reads the time the pointer points to:
`if now.Sub(*obj.LastModified) > maxAge { ... }`.

### Answer

```go
for _, obj := range objects {
	if obj.LastModified == nil {
		continue
	}
	if now.Sub(*obj.LastModified) > maxAge {
		keys = append(keys, aws.ToString(obj.Key))
	}
}
```

Using `>` (not `>=`) keeps an object that is exactly `maxAge` old.

## Task 2: Stay inside the prefix

Two changes: refuse an empty prefix before talking to S3 at all, and tell the
list call which prefix to use.

### Closer

The guard goes at the very top of `Cleanup`, before the paginator is created,
and returns `ErrEmptyPrefix`. The input struct needs one more field next to
`Bucket`: `Prefix: aws.String(...)`.

### Answer

```go
if prefix == "" {
	return nil, ErrEmptyPrefix
}
p := s3.NewListObjectsV2Paginator(api, &s3.ListObjectsV2Input{
	Bucket: aws.String(bucket),
	Prefix: aws.String(prefix),
})
```

## Task 3: Make dry run safe

The key should always be appended to `removed`; only the delete itself
depends on `dryRun`.

### Closer

Keep the `removed = append(removed, key)` line where it is, and wrap only the
`DeleteObject` call and its error check in `if !dryRun { ... }`.

### Answer

```go
for _, key := range ExpiredKeys(page.Contents, now, maxAge) {
	if !dryRun {
		_, err := api.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return removed, fmt.Errorf("delete s3://%s/%s: %w", bucket, key, err)
		}
	}
	removed = append(removed, key)
}
```
