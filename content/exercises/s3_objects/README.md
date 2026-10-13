# S3 objects: upload, verify and list configs

Amazon S3 is where most AWS teams keep files: build artifacts, backups, logs
and config. Here you upload a service config, download it again with a
checksum check, and list every config under a "folder".

**New here:** `PutObject`, `GetObject`, object keys and prefixes, user metadata, `bytes.NewReader`, `io.ReadAll`.

## Learn

### S3 in one minute

*S3* (Simple Storage Service) stores files. A *bucket* is a named container,
such as `ops-configs`. Inside it are *objects*: bytes plus a little metadata,
found by a *key* like `checkout/prod.json`.

S3 has **no real folders**. `checkout/prod.json` is one key that contains a
`/`. The part of a key you filter on is a *prefix*. The console's "Create
folder" stores an empty object whose key ends in `/` (like `checkout/`), a
*folder marker*. File listings usually skip those.

An object's full address is written `s3://bucket/key`, for example
`s3://ops-configs/checkout/prod.json`.

### Uploading: PutObject

Same shape as every SDK call in *SDK clients: small interfaces and fakes*:

```go
_, err := api.PutObject(ctx, &s3.PutObjectInput{
	Bucket:      aws.String("ops-backups"),
	Key:         aws.String("db/2026-10-01.txt"),
	Body:        bytes.NewReader(dump),
	ContentType: aws.String("text/plain"),
	Metadata:    map[string]string{"owner": "platform"},
})
```

- `Body` is an `io.Reader`, something S3 can read bytes from.
  `bytes.NewReader(b)` (package `bytes`) turns a `[]byte` into one.
- `ContentType` tells browsers and tools what the bytes are.
- `Metadata` holds your own *user metadata*: small key/value strings next to
  the object. S3 lowercases the keys, so use lowercase keys.

> **Why is it written like this?** `aws.String(...)` is required: string
> fields in SDK inputs are pointers. `Metadata` is a plain map, so no
> `aws.String` there. Lining up the values after the colons is convention
> (gofmt does it).

### Downloading: GetObject

```go
out, err := api.GetObject(ctx, &s3.GetObjectInput{
	Bucket: aws.String("ops-backups"),
	Key:    aws.String("db/2026-10-01.txt"),
})
if err != nil {
	return nil, err
}
defer out.Body.Close()

dump, err := io.ReadAll(out.Body)
```

`out.Body` is a stream over the network. You must close it, or the
connection leaks, so `defer out.Body.Close()` right after the error check.
`io.ReadAll` (package `io`) reads the stream to the end.

`out.Metadata` is the map you uploaded. Reading a missing map key gives
`""`. A missing object comes back as `*types.NoSuchKey`, as in *AWS API
errors: codes, faults and NoSuchKey*.

### Checking the bytes

Before a deploy applies a config, you want proof it's the file CI uploaded.
The starter has the `Checksum(data)` helper from *Checksums and archives:
verify and extract safely*: it returns the sha256 as 64 hex characters.

Upload stores the checksum in the metadata. Download recomputes it and
compares. (S3 can also do this for you via `ChecksumSHA256`, but by hand you
see every step.)

### Listing under a prefix

Listing is `ListObjectsV2` with a paginator, as in *Paginators: walk every
page*. Its input has a `Prefix` field: S3 then returns only keys that start
with it, in alphabetical order.

Keep the trailing `/` in a prefix: `checkout` would also match
`checkout-old/prod.json`.

`strings.HasSuffix(s, end)` reports whether `s` ends with `end`:

```go
strings.HasSuffix("report.csv", ".csv") // => true
```

## Your task

CI stores every service's config in the `ops-configs` bucket, and the deploy
tool reads them back. Finish the three functions in `exercise.go`. They get
an `ObjectAPI`, so the tests can pass in an in-memory fake bucket.

### 1. Upload a config

**Goal:** `UploadConfig` stores `data` as JSON with its checksum in the metadata.

Add three fields to its `PutObjectInput`:

- `Body`: the bytes in the `data` parameter, as a reader. Add the import `"bytes"`.
- `ContentType`: `"application/json"`.
- `Metadata`: one entry, key `"sha256"`, value the checksum of `data`.

```go
UploadConfig(ctx, api, "ops-configs", "checkout/prod.json", []byte(`{"replicas": 3}`))
// stores s3://ops-configs/checkout/prod.json with
// ContentType "application/json" and Metadata {"sha256": "<64 hex chars>"}
```

**Think it through:**
- Which fields are pointers (need `aws.String`) and which is a map?
- Which helper turns `data` into the checksum string?

### 2. Download and verify a config

**Goal:** `DownloadConfig` returns the object's bytes only if their checksum matches the stored one.

It already calls `GetObject` and closes the body. Finish it:

1. Read the whole body. Add the import `"io"`. If reading fails, return
   `nil` and the error wrapped the same way as the `GetObject` error above it
   (`download s3://BUCKET/KEY: <err>`).
2. Compare the checksum of the data with the metadata entry `"sha256"`. If
   they differ (a missing entry counts as different), return `nil` and
   `ErrChecksumMismatch` wrapped with `%w` as `s3://BUCKET/KEY: <err>`.
3. Otherwise return the data and `nil`.

```go
DownloadConfig(ctx, api, "ops-configs", "checkout/prod.json")
// => the bytes, nil
DownloadConfig(ctx, api, "ops-configs", "checkout/staging.json") // edited by hand
// => nil, error "s3://ops-configs/checkout/staging.json: checksum mismatch"
```

**Think it through:**
- Where does the stored checksum live on `out`?
- What does a missing map key read as, and can it ever equal a real checksum?

### 3. List configs under a prefix

**Goal:** `ListConfigs` returns only the keys under `prefix`, without folder markers.

It already loops over every page but returns every key in the bucket. Fix it:

- Ask S3 to list only keys under the `prefix` parameter.
- Skip folder markers: keys ending in `"/"`. Add the import `"strings"`.

Return keys in the order S3 lists them. An empty prefix lists the whole
bucket. No matches gives `nil`.

```go
ListConfigs(ctx, api, "ops-configs", "checkout/")
// => []string{"checkout/dev.json", "checkout/prod.json", "checkout/staging.json"}
//    ("checkout/" itself and "checkout-old/prod.json" are left out)
```

**Think it through:**
- Which field of `ListObjectsV2Input` filters by prefix?
- Inside the loop, which keyword skips to the next object?

## Learn more

- [Amazon S3: Naming objects (keys and prefixes)](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-keys.html)
- [Amazon S3: Organizing objects in the console by using folders](https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-folders.html)
- [Amazon S3: Working with object metadata](https://docs.aws.amazon.com/AmazonS3/latest/userguide/UsingMetadata.html)
- [pkg.go.dev: service/s3](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3)
