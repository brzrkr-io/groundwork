# Paginators: fix the stuck size report

Reports such as "how much are we storing?" or "what are we running?" have to
add up numbers across every page. In this exercise you fix a report that
never finishes, then write a fleet summary from scratch.

**New here:** totals across pages, converting SDK string types with `string(...)`; practice with paginators.

## Learn

### Recap

From *Paginators: walk every page*: AWS returns results one *page* at a time,
with a token pointing at the next page. A paginator such as
`s3.NewListObjectsV2Paginator(api, input)` stores that token for you, and
`for p.HasMorePages() { page, err := p.NextPage(ctx) ... }` walks every page.

### Create the paginator once

The paginator *is* the bookmark. Create it once, before the loop. A new
paginator starts again at page 1, so creating one inside the loop reads the
first page over and over and never reaches the end.

> **Why is it written like this?** Go runs everything inside `for { ... }`
> on every pass, including `:=` lines, so a value made there is brand new
> each time. Where you put a line is your choice, and here it decides
> whether the program works.

### Adding up across pages

To total something across pages, create the map *before* the loop and add to
it *inside* the loop. Every page adds to the same map:

```go
perClass := map[string]int64{}
for p.HasMorePages() {
	page, err := p.NextPage(ctx)
	// ... handle err ...
	for _, obj := range page.Contents {
		perClass[string(obj.StorageClass)] += aws.ToInt64(obj.Size)
	}
}
```

`m[k] += n` works even when `k` is new: a missing key reads as `0`. For a map
of counts, `m[k]++` adds 1.

### SDK string types

Many SDK fields are named string types, not plain `string`. For example
`obj.StorageClass` is an `s3types.ObjectStorageClass` and an EC2 instance's
`InstanceType` is an `ec2types.InstanceType` (values like `"t3.micro"`).
A `map[string]int` needs a plain `string` key, so convert with
`string(value)`.

> **Why is it written like this?** Go requires the explicit conversion: it
> never mixes two different types silently, even when both are strings
> underneath.

### Empty pages happen

EC2 can send a page with no reservations that still carries a `NextToken`,
especially when you filter. That is not the end of the list. Only
`HasMorePages()` decides when to stop; never stop because a page was empty.

## Your task

The platform team wants two numbers for the monthly cost review: how much log
data each service keeps in the `central-logs` bucket, and how many instances
of each type are running.

### 1. Fix the size report

**Goal:** `PrefixSizes(ctx, api, bucket)` reads each page exactly once and
returns the total `Size` per top-level prefix.

The prefix is the part of the key before the first `/`; a key with no `/`
counts under its full name. That part already works. The bug is in the loop:
it creates a new paginator on every pass, so with more than one page it
re-reads page 1 forever. (The test fake gives up after 20 calls, so you see
an error instead of a hang.) Keep the error wrapping
`size report s3://<bucket>: <err>` as it is. An empty bucket gives an empty
map.

```go
PrefixSizes(ctx, api, "central-logs") // the bucket sends 2 objects per page
// => map[README.txt:120 api:2000 nginx:55100 worker:5000], nil
```

**Think it through:**
- Which line in the loop makes a fresh paginator? Where should it be instead?
- Once the paginator lives outside, which paginator method belongs in the `for` line, and is the `break` at the bottom still needed?

### 2. Count running instances by type

**Goal:** `RunningByType(ctx, api)` returns how many *running* instances there
are of each instance type, across every page.

Use `ec2.NewDescribeInstancesPaginator` with an empty
`&ec2.DescribeInstancesInput{}`. Walk every reservation on every page. An
instance counts only if its `State` is not `nil` and its `State.Name` is
`types.InstanceStateNameRunning`; stopped, pending and other states are
skipped. The map key is the instance type as a plain string. Add the import
`"github.com/aws/aws-sdk-go-v2/service/ec2/types"`.

With nothing running, return the empty map. On error, return `nil` and the
error wrapped with `%w` as `running by type: <err>`.

```go
RunningByType(ctx, api) // page 2 of 3 is empty
// => map[m5.large:2 t3.micro:1], nil
```

**Think it through:**
- Task 1's fixed loop is your model: what is nested inside an EC2 page?
- Which check must come first so you never read `Name` through a `nil` `State`?
- Which field holds the type, and how do you turn it into a map key?

## Learn more

- [Using paginators (AWS SDK for Go v2)](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/using.html#using-operation-paginators)
- [DescribeInstances (Amazon EC2 API reference)](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html)
- [ec2.DescribeInstancesPaginator on pkg.go.dev](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#DescribeInstancesPaginator)
- [Go spec: Conversions](https://go.dev/ref/spec#Conversions)
