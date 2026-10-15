# Tag compliance: untagged S3 buckets and bad values

S3 buckets need tags too, and a bucket nobody owns is worse than a stray
server: it may hold customer data. This time you audit buckets and check that
tag *values* make sense, not just that keys exist.

**New here:** `GetBucketTagging`, the `NoSuchTagSet` error.

## Learn

### Recap

Tags are key/value labels like `Owner`, `Env` and `CostCenter`; keys are
case-sensitive. The SDK gives them as `[]types.Tag` with `*string` fields, so
copy them into a map with `aws.ToString`. Sort reports so every run looks the
same.

### S3 bucket tags: GetBucketTagging

S3 stores a bucket's tags as one *tag set*:

```go
out, err := api.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
	Bucket: aws.String("logs-archive"),
})
// out.TagSet is a []types.Tag (from service/s3/types)
```

The surprise: a bucket with **no tags at all** doesn't give an empty
`TagSet`. The call fails with an API error whose code is `NoSuchTagSet`. For
an audit that's not a failure; it's the most important finding.

Spot a code with `errors.AsType` and `smithy.APIError`, as in *AWS API
errors: codes, faults and NoSuchKey*:

```go
if apiErr, ok := errors.AsType[smithy.APIError](err); ok && apiErr.ErrorCode() == "AccessDenied" {
	log.Println("ask for s3:GetBucketTagging permission")
}
```

> **Why is it written like this?** `if x, ok := ...; cond {` (a short
> statement, a semicolon, then the condition) is Go syntax. Testing `ok`
> before calling `apiErr.ErrorCode()` is required: without an API error,
> `apiErr` is `nil`. `&&` skips the right side when the left is false.

### Checking values

A tag that exists can still be wrong: `Env=production` and `Env=Prod` both
break a report that groups by `prod`. `slices.Contains(list, v)` reports
whether `v` is in `list`:

```go
slices.Contains([]string{"amd64", "arm64"}, "arm64") // => true
```

The `%q` verb in `fmt.Sprintf` prints a string *with* double quotes around
it: `fmt.Sprintf("got %q", "x")` gives `got "x"`.

## Your task

The `buckettags` package audits S3 bucket tags. `RequiredTags` and
`AllowedEnvs` are at the top of `exercise.go`.

### 1. Fix the untagged-bucket bug

**Goal:** `BucketTags` returns an empty map and `nil` for a bucket with no tags.

Today it fails on every untagged bucket, so the audit stops on exactly the
buckets it should report.

- An API error with code `"NoSuchTagSet"`: return an empty (non-`nil`) map
  and a `nil` error.
- Every other error (like `AccessDenied` or `NoSuchBucket`): still return
  it, wrapped as now: `get tags for bucket "<name>": <err>`.

Add `"errors"` and `"github.com/aws/smithy-go"` to the imports.

```go
BucketTags(ctx, api, "logs-archive") // bucket has no tags
// => map[], nil
```

**Think it through:**
- Inside `if err != nil`, does your new check go before or after the existing `return`?
- Which code are you comparing against?

### 2. Check the Env value

**Goal:** `Problems` also reports an `Env` value that isn't in `AllowedEnvs`.

It already reports missing keys as `<bucket>: missing <Key>`, in
`RequiredTags` order. Add one check at the end: if `Env` is set (not `""`)
but isn't in `AllowedEnvs`, append this line:

```text
web-assets: Env "production" is not prod, staging or dev
```

The bucket name and the value come from your data; the value has double
quotes around it (use `%q`). The rest is fixed text. The check is
case-sensitive: `"Staging"` isn't allowed. Add `"slices"` to the imports.

```go
Problems("ml-models", map[string]string{"Env": "Staging", "CostCenter": "CC-3300"})
// => ["ml-models: missing Owner"
//     "ml-models: Env \"Staging\" is not prod, staging or dev"]
```

**Think it through:**
- Two conditions must both hold. What are they, and which operator joins them?
- An empty `Env` was already reported as missing. Why shouldn't it also get this line?

### 3. Sort the report

**Goal:** `Audit` returns all lines sorted alphabetically.

It collects lines in the order the buckets were given. Sort them with
`slices.Sort` before returning, so two runs over the same account give the
same report. Errors from `BucketTags` still stop the audit with `nil` lines
(already works). `Audit` uses tasks 1 and 2, so finish those first.

```go
Audit(ctx, api, []string{"zz-backups", "aa-reports"})
// => ["aa-reports: missing Env" "zz-backups: missing CostCenter"]
```

**Think it through:**
- Where is the last moment before `lines` is returned?

## Learn more

- [Using cost allocation S3 bucket tags](https://docs.aws.amazon.com/AmazonS3/latest/userguide/CostAllocTagging.html)
- [GetBucketTagging API reference](https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketTagging.html)
- [smithy.APIError](https://pkg.go.dev/github.com/aws/smithy-go#APIError)
- [Package slices: Contains, Sort](https://pkg.go.dev/slices)
