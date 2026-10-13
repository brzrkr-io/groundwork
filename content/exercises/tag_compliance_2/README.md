# Tag compliance: untagged S3 buckets and bad values

S3 buckets need tags too, and a bucket nobody owns is worse than a stray
server: it may hold customer data. This time you audit buckets, and you check
that the tag *values* make sense, not just that the keys exist.

## Learn

Recap: tags are key/value labels such as `Owner`, `Env` and `CostCenter`,
and keys are case-sensitive. In the SDK they are a list of `types.Tag` with
`*string` fields, so copy them into a map with `aws.ToString`. Sort reports
so they come out the same on every run.

### S3 bucket tags: `GetBucketTagging`

S3 stores all of a bucket's tags as one *tag set*:

```go
out, err := api.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
	Bucket: aws.String("logs-archive"),
})
// out.TagSet is a []types.Tag (from service/s3/types)
```

One surprise: if a bucket has **no tags at all**, S3 doesn't send an empty
`TagSet`. The call fails with an API error whose code is `NoSuchTagSet`. For
an audit that is not a failure. It's the most important finding of all. Spot
the code with `errors.AsType` and `smithy.APIError`, as in *AWS API errors:
codes, faults and NoSuchKey*:

```go
if apiErr, ok := errors.AsType[smithy.APIError](err); ok && apiErr.ErrorCode() == "NoSuchTagSet" {
	// no tags: carry on with an empty map
}
```

Other errors, such as `AccessDenied` or `NoSuchBucket`, are still real
failures.

### Checking values

A tag that exists can still be wrong. `Env=production` and `Env=Prod` both
break a cost report that groups by `prod`. Keep the allowed values in a slice
and use `slices.Contains(list, v)`, which reports whether `v` is in `list`.

## Your task

The `buckettags` package audits S3 bucket tags. `RequiredTags` and
`AllowedEnvs` are at the top of `exercise.go`.

### 1. Fix the untagged-bucket bug

`BucketTags` fails for every bucket with no tags, so the audit crashes on
exactly the buckets it should report. Change it so that an API error with
code `"NoSuchTagSet"` returns an empty map and a `nil` error. Every other
error must still be returned, wrapped as it is now
(`get tags for bucket "<name>": ...`). Add `"errors"` and
`"github.com/aws/smithy-go"` to the imports.

```go
BucketTags(ctx, api, "logs-archive") // bucket has no tags
// => map[], nil
```

### 2. Check the Env value

`Problems` already reports missing keys as `<bucket>: missing <Key>`, in
`RequiredTags` order. Add one more check at the end. If `Env` is set (not
`""`) but isn't one of `AllowedEnvs`, append this line, which has the value
in double quotes (use `%q`):

```text
web-assets: Env "production" is not prod, staging or dev
```

The check is case-sensitive: `"Staging"` isn't allowed. Add `"slices"` to
the imports.

```go
Problems("ml-models", map[string]string{"Env": "Staging", "CostCenter": "CC-3300"})
// => ["ml-models: missing Owner"
//     "ml-models: Env \"Staging\" is not prod, staging or dev"]
```

### 3. Sort the report

`Audit` collects the lines for every bucket in the order the buckets were
given. Sort all the lines with `slices.Sort` before returning them, so that
two runs over the same account produce the same report. Errors from
`BucketTags` must still stop the audit (that part already works). `Audit`
calls `BucketTags` and `Problems`, so finish tasks 1 and 2 first.

```go
Audit(ctx, api, []string{"zz-backups", "aa-reports"})
// => ["aa-reports: missing Env" "zz-backups: missing CostCenter"]
```

## Learn more

- [Using cost allocation S3 bucket tags](https://docs.aws.amazon.com/AmazonS3/latest/userguide/CostAllocTagging.html)
- [GetBucketTagging API reference](https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketTagging.html)
- [smithy.APIError](https://pkg.go.dev/github.com/aws/smithy-go#APIError)
- [Package slices: Contains, Sort](https://pkg.go.dev/slices)
