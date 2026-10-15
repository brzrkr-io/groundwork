# Paginators: walk every page

A log bucket can hold millions of objects, and a big AWS account can run
thousands of instances. AWS never sends all of that in one reply. It sends
*pages*, and your code has to ask for each one. Forget that, and your cleanup
script quietly skips almost everything.

**New here:** pagination, continuation tokens, SDK paginators, the SDK's `XxxAPIClient` interfaces.

## Learn

### Why AWS returns pages

When you list an S3 bucket with `ListObjectsV2`, S3 sends at most 1,000 keys
per call. If more remain, the reply carries a *continuation token*: a bookmark
saying "carry on from here". You call again with that token. No token means
you reached the last page. This is *pagination*.

Every AWS list or describe call works like this. Only the field names differ:

| Operation | "More pages?" | Token in the reply | Token in the next request |
|---|---|---|---|
| S3 `ListObjectsV2` | `IsTruncated` is true | `NextContinuationToken` | `ContinuationToken` |
| EC2 `DescribeInstances` | the token is not empty | `NextToken` | `NextToken` |

Tokens are *opaque*: you never read or build one, you only pass it back.

### Paginators do the bookkeeping

The SDK gives you a *paginator* for each operation that pages. You create it
once, with the same input you would pass to the call, then loop. This one
adds up the bytes in a bucket:

```go
p := s3.NewListObjectsV2Paginator(api, &s3.ListObjectsV2Input{
	Bucket: aws.String("backups"),
})

var total int64
for p.HasMorePages() {
	page, err := p.NextPage(ctx)
	if err != nil {
		return 0, fmt.Errorf("size of backups: %w", err)
	}
	for _, obj := range page.Contents {
		total += aws.ToInt64(obj.Size)
	}
}
```

| Piece | What it does |
|---|---|
| `p.HasMorePages()` | true before the first page, then for as long as the last reply carried a token |
| `p.NextPage(ctx)` | makes one API call and returns that page (a normal `*s3.ListObjectsV2Output`); stores the next token |
| `page.Contents` | the objects on this page; each `Key` is a `*string` |

> **Why is it written like this?** `for condition { ... }` is Go's only
> "while" loop: Go requires `for`, there is no `while` keyword. Checking
> `err` straight after `NextPage` is convention, so you never touch a `nil`
> page. The names `p` and `page` are your choice.

EC2 has the same thing: `ec2.NewDescribeInstancesPaginator(api, input)`. Its
pages are `*ec2.DescribeInstancesOutput`, with `Reservations`, each holding
`Instances`, each with an `InstanceId *string` (as in *SDK clients: small
interfaces and fakes*).

### Interfaces that come with the SDK

In *SDK clients: small interfaces and fakes* you wrote a small interface with
just the methods you call. The SDK already ships one for every operation that
pages:

```go
type ListObjectsV2APIClient interface {
	ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}
```

`s3.ListObjectsV2APIClient` and `ec2.DescribeInstancesAPIClient` are what the
paginators accept. The real clients satisfy them; the tests pass fakes that
hand out small pages (see `fakes_test.go` if you're curious).

### Page size is not a total

`MaxKeys` on the input sets the *page size*: "at most N keys per page". It
does not stop the loop: the paginator keeps fetching until the bucket runs
out. To stop early, `return` (or `break`) out of the loop yourself. Every page
you skip is an API call you don't pay for.

```go
for _, line := range lines {
	found = append(found, line)
	if len(found) == 10 {
		return found // leaves the loop and the function at once
	}
}
```

### The context reaches every call

`NextPage(ctx)` hands `ctx` to the API call. If `ctx` is cancelled (Ctrl-C,
a deadline, a shutdown, as in *Context and cancellation*), the next call
fails with an error that wraps `context.Canceled`. Wrap it with `%w`, so
callers can still check `errors.Is(err, context.Canceled)`.

## Your task

Your team's ops tool lists log files in S3 and instances in EC2. It works in
the dev account, which is small, but misses things in production.

### 1. Read every page

**Goal:** `AllKeys(ctx, api, bucket, prefix)` returns the key of every object
starting with `prefix`, from every page, in the order S3 returns them.

The starter already creates the paginator and reads one page. Make that
page-reading code run once per page. With no matching keys, return `nil`. On
error, return `nil` and the error wrapped as `list s3://<bucket>/<prefix>: <err>`
(the starter already formats this).

```go
AllKeys(ctx, api, "app-logs", "web/") // the bucket sends 2 keys per page
// => ["web/2026-10-09.log" "web/2026-10-10.log" "web/2026-10-11.log"], nil

AllKeys(ctx, api, "app-logs", "web/") // page 2 fails with AccessDenied
// => nil, "list s3://app-logs/web/: api error AccessDenied: Access Denied"
```

**Think it through:**
- Which lines read one page? Those are what must repeat.
- Which paginator method answers "is there another page?"
- Where must `return keys, nil` end up: inside the loop or after it?

### 2. Stop once you have enough

**Goal:** `FirstKeys(ctx, api, bucket, limit)` returns the first `limit` keys
and asks S3 for no more pages than it needs.

Right now it fetches every page, because `MaxKeys` only sets the page size.
Return `keys` as soon as it holds `limit` keys. If the bucket has fewer keys
than `limit`, return all of them. `limit` is always at least 1.

```go
FirstKeys(ctx, api, "deploy-artifacts", 3) // the bucket sends 2 keys per page
// => the first 3 keys, after only 2 ListObjectsV2 calls
```

**Think it through:**
- Right after which line does the number of keys change?
- What do you compare to know you have enough?

### 3. List every instance

**Goal:** `InstanceIDs(ctx, api)` returns the `InstanceId` of every instance
in every reservation on every page, in order.

Use `ec2.NewDescribeInstancesPaginator` with an empty
`&ec2.DescribeInstancesInput{}`, and pass `ctx` to `NextPage`. With no
instances, return `nil`. On error, return `nil` and the error wrapped with
`%w` as `describe instances: <err>`.

```go
InstanceIDs(ctx, api) // three pages
// => ["i-0a1b2c3d4e5f00001" ... "i-0a1b2c3d4e5f00006"], nil

InstanceIDs(ctx, api) // ctx is cancelled after page 1
// => nil, error "describe instances: context canceled"
//    errors.Is(err, context.Canceled) == true, and no third page is requested
```

**Think it through:**
- Task 1's loop is your model. What is different about the EC2 output?
- How many nested `for` loops do you need: pages, then what, then what?
- Which verb in `fmt.Errorf` keeps `context.Canceled` findable?

## Learn more

- [Using paginators (AWS SDK for Go v2)](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/using.html#using-operation-paginators)
- [ListObjectsV2 (Amazon S3 API reference)](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjectsV2.html)
- [s3.ListObjectsV2Paginator on pkg.go.dev](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3#ListObjectsV2Paginator)
- [ec2.DescribeInstancesPaginator on pkg.go.dev](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#DescribeInstancesPaginator)
