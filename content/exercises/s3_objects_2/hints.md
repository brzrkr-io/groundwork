## Task 1: two checks before appending

Inside the loop, first `continue` when `obj.LastModified == nil`. Then only
append when the age, `now.Sub(*obj.LastModified)`, is greater than `maxAge`.
The `*` reads the time the pointer points to.

## Task 2: guard first, then filter

Put the empty-prefix check at the very top of `Cleanup`, before the
paginator is created:

```go
if prefix == "" {
	return nil, ErrEmptyPrefix
}
```

Then add `Prefix: aws.String(prefix),` next to `Bucket` in the input.

## Task 3: wrap the delete

Keep the `removed = append(removed, key)` line where it is, and wrap only the
`DeleteObject` call and its error check in `if !dryRun { ... }`.

## Almost the answer

```go
// Task 1
if obj.LastModified == nil {
	continue
}
if now.Sub(*obj.LastModified) > maxAge {
	keys = append(keys, aws.ToString(obj.Key))
}

// Task 3
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
```
