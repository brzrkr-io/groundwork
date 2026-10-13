## Task 1: three more fields

The `PutObjectInput` already has `Bucket` and `Key`. Add `Body`,
`ContentType` and `Metadata` in the same `Field: value,` style. String fields
are pointers, so `ContentType` needs `aws.String(...)`. `Metadata` is a plain
`map[string]string`, no pointers.

## Task 2: read, then compare

```go
data, err := io.ReadAll(out.Body)
if err != nil {
	return nil, fmt.Errorf("download s3://%s/%s: %w", bucket, key, err)
}
```

Then compare `Checksum(data)` with `out.Metadata["sha256"]`. Reading a
missing map key gives `""`, which never equals a real checksum, so one `!=`
check covers both the stale and the missing case.

## Task 3: prefix and folder markers

Add `Prefix: aws.String(prefix),` next to `Bucket`. In the inner loop, put the
key in a variable first, then skip markers with `continue`:

```go
key := aws.ToString(obj.Key)
if strings.HasSuffix(key, "/") {
	continue
}
keys = append(keys, key)
```

## Almost the answer

```go
// Task 1
Body:        bytes.NewReader(data),
ContentType: aws.String("application/json"),
Metadata:    map[string]string{"sha256": Checksum(data)},

// Task 2, after reading data
if Checksum(data) != out.Metadata["sha256"] {
	return nil, fmt.Errorf("s3://%s/%s: %w", bucket, key, ErrChecksumMismatch)
}
return data, nil
```
