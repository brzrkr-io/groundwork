# S3 objects: upload, verify and list configs

Amazon S3 is where most AWS teams keep files: build artifacts, backups, logs
and config. In this exercise you upload a service config to S3, download it
again with a checksum check, and list every config under a "folder".

## Learn

### S3 in one minute

*S3* (Simple Storage Service) stores files. A *bucket* is a named container,
such as `ops-configs`. Bucket names are unique across all of AWS. Inside a
bucket you store *objects*. An object is a blob of bytes plus a little
metadata, and it is found by its *key*, a string like `checkout/prod.json`.

S3 has **no real folders**. `checkout/prod.json` is one key that happens to
contain a `/`. The AWS console splits keys on `/` and shows `checkout/` as a
folder, but that is only display. The part of a key you filter on is called
a *prefix*. When you click "Create folder" in the console, it stores an empty
object whose key ends in `/` (such as `checkout/`), a *folder marker*. Tools
that list files usually skip those.

People write an object's full address as `s3://bucket/key`, for example
`s3://ops-configs/checkout/prod.json`.

### Uploading: PutObject

The call has the same shape as every SDK call in *SDK clients: small interfaces and fakes*:

```go
_, err := api.PutObject(ctx, &s3.PutObjectInput{
	Bucket:      aws.String("ops-configs"),
	Key:         aws.String("checkout/prod.json"),
	Body:        bytes.NewReader(data),
	ContentType: aws.String("application/json"),
	Metadata:    map[string]string{"owner": "platform"},
})
```

- `Body` is an `io.Reader`, something S3 can read bytes from.
  `bytes.NewReader(data)` (package `bytes`) wraps a `[]byte` as a reader.
- `ContentType` tells browsers and tools what the bytes are.
- `Metadata` holds your own *user metadata*, small key/value strings stored
  next to the object. S3 lowercases the keys, so always use lowercase keys.

### Downloading: GetObject

```go
out, err := api.GetObject(ctx, &s3.GetObjectInput{
	Bucket: aws.String("ops-configs"),
	Key:    aws.String("checkout/prod.json"),
})
if err != nil {
	return nil, err
}
defer out.Body.Close()

data, err := io.ReadAll(out.Body)
```

`out.Body` is an `io.ReadCloser`: a stream coming over the network. You
must close it, or the connection leaks, so `defer out.Body.Close()` right
after the error check. `io.ReadAll` (package `io`) reads the stream to the
end. `out.Metadata` is the same map you uploaded. A missing key comes back as
`*types.NoSuchKey`, as in *AWS API errors: codes, faults and NoSuchKey*.

### Checking the bytes

Before a deploy applies a config, you want to be sure it's the file CI
uploaded, not a half-written or hand-edited one. The starter has the
`Checksum` helper from *Checksums and archives: verify and extract safely*:

```go
sum := sha256.Sum256(data)
return hex.EncodeToString(sum[:])
```

Upload stores `Checksum(data)` in the metadata. Download recomputes it and
compares. (S3 can also compute checksums for you, via `ChecksumSHA256`, but
doing it by hand shows every step.)

### Listing under a prefix

Listing is `ListObjectsV2` with a paginator, as in *Paginators: walk every page*. Setting
`Prefix: aws.String("checkout/")` returns only keys that start with
`checkout/`. Keep the trailing `/`: the prefix `checkout` would also match
`checkout-old/prod.json`. S3 returns keys in alphabetical order.

## Your task

CI stores every service's config in the `ops-configs` bucket, and the deploy
tool reads them back. Finish the three functions in `exercise.go`. Your code
gets an `ObjectAPI`, so the tests can pass in an in-memory fake bucket.

### 1. Upload a config

In `UploadConfig`, add three fields to the `PutObjectInput`:

- `Body`: the bytes in `data`, via `bytes.NewReader`. Add the import `"bytes"`.
- `ContentType`: `"application/json"`.
- `Metadata`: a map with one entry, `"sha256"` set to `Checksum(data)`.

```go
UploadConfig(ctx, api, "ops-configs", "checkout/prod.json", []byte(`{"replicas": 3}`))
// stores s3://ops-configs/checkout/prod.json with
// ContentType "application/json" and Metadata {"sha256": "<64 hex chars>"}
```

### 2. Download and verify a config

`DownloadConfig` already calls `GetObject` and closes the body. Finish it:

1. Read the whole body with `io.ReadAll`. Add the import `"io"`. If reading
   fails, return `nil` and the error, wrapped like the `GetObject` error above.
2. Compare `Checksum(data)` with `out.Metadata["sha256"]`. If they differ
   (a missing `"sha256"` entry counts as different), return `nil` and this
   error:

   ```go
   fmt.Errorf("s3://%s/%s: %w", bucket, key, ErrChecksumMismatch)
   ```

   which reads `s3://ops-configs/checkout/staging.json: checksum mismatch`.
3. Otherwise return the data and `nil`.

### 3. List configs under a prefix

`ListConfigs` already loops over every page but returns every key in the
bucket. Fix it:

- Add `Prefix: aws.String(prefix)` to the `ListObjectsV2Input`.
- Skip folder markers: keys ending in `"/"`. Use `strings.HasSuffix` and add
  the import `"strings"`.

Return the keys in the order S3 lists them. An empty prefix lists the whole
bucket.

```go
ListConfigs(ctx, api, "ops-configs", "checkout/")
// => []string{"checkout/dev.json", "checkout/prod.json", "checkout/staging.json"}
//    ("checkout/" itself and "checkout-old/prod.json" are left out)
```

## Learn more

- [Amazon S3: Naming objects (keys and prefixes)](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-keys.html)
- [Amazon S3: Organizing objects in the console by using folders](https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-folders.html)
- [Amazon S3: Working with object metadata](https://docs.aws.amazon.com/AmazonS3/latest/userguide/UsingMetadata.html)
- [pkg.go.dev: service/s3](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3)
