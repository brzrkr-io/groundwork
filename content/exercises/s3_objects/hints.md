## Task 1: Upload a config

The `PutObjectInput` already has `Bucket` and `Key`. Add `Body`,
`ContentType` and `Metadata` in the same `Field: value,` style, and add the
`"bytes"` import.

### Closer

String fields are pointers, so `ContentType` needs `aws.String(...)`.
`Body` wants a reader: `bytes.NewReader(data)`. `Metadata` is a plain
`map[string]string{...}`, no pointers.

### Answer

```go
Body:        bytes.NewReader(data),
ContentType: aws.String("application/json"),
Metadata:    map[string]string{"sha256": Checksum(data)},
```

## Task 2: Download and verify a config

First read all the bytes from `out.Body`, then compute their checksum and
compare it with the one stored in the metadata. Add the `"io"` import.

### Closer

```go
data, err := io.ReadAll(out.Body)
if err != nil {
	return nil, fmt.Errorf("download s3://%s/%s: %w", bucket, key, err)
}
```

Then compare `Checksum(data)` with `out.Metadata["sha256"]` using `!=`.

### Answer

```go
data, err := io.ReadAll(out.Body)
if err != nil {
	return nil, fmt.Errorf("download s3://%s/%s: %w", bucket, key, err)
}
if Checksum(data) != out.Metadata["sha256"] {
	return nil, fmt.Errorf("s3://%s/%s: %w", bucket, key, ErrChecksumMismatch)
}
return data, nil
```

A missing map key reads as `""`, which never equals a real checksum, so one
check covers both the stale and the missing case.

## Task 3: List configs under a prefix

Two changes: tell S3 which prefix to list, and skip keys that end in `/`
inside the loop. Add the `"strings"` import.

### Closer

Add `Prefix: aws.String(prefix),` next to `Bucket`. In the inner loop, put
the key in a variable, then use `strings.HasSuffix(key, "/")` with
`continue`.

### Answer

```go
p := s3.NewListObjectsV2Paginator(api, &s3.ListObjectsV2Input{
	Bucket: aws.String(bucket),
	Prefix: aws.String(prefix),
})
// ...
for _, obj := range page.Contents {
	key := aws.ToString(obj.Key)
	if strings.HasSuffix(key, "/") {
		continue
	}
	keys = append(keys, key)
}
```
