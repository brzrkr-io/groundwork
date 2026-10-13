# Paginators: fix the stuck size report

Reports such as "how much are we storing?" or "what are we running?" have to
add up numbers across every page. In this exercise you fix a report that
never finishes, then write a fleet summary from scratch.

## Learn

**Recap.** AWS list and describe calls return results one *page* at a time,
with a token that points at the next page. A paginator such as
`s3.NewListObjectsV2Paginator(api, input)` stores that token for you.
`for p.HasMorePages() { page, err := p.NextPage(ctx) ... }` walks every page,
as in *Paginators: walk every page*.

### Create the paginator once

The paginator *is* the bookmark. Create it once, before the loop. A new
paginator starts again at page 1, so creating one inside the loop reads the
first page over and over and never reaches the end.

### Adding up across pages

To total something across pages, create the map before the loop and add to
it inside the loop. Each page adds to the same map:

```go
sizes := map[string]int64{}
for p.HasMorePages() {
	page, err := p.NextPage(ctx)
	// ...
	for _, obj := range page.Contents {
		sizes["all"] += aws.ToInt64(obj.Size) // Size is a *int64 (bytes)
	}
}
```

### Empty pages happen

EC2 can send a page with no reservations at all that still carries a
`NextToken`, especially when you filter. That is not the end of the list.
Only `HasMorePages()` decides when to stop. Never stop just because a page
was empty.

## Your task

The platform team wants two numbers for the monthly cost review: how much log
data each service keeps in the `central-logs` bucket, and how many instances
of each type are running.

### 1. Fix the size report

`PrefixSizes` adds up `Size` per top-level prefix: the part of the key before
the first `/`. A key with no `/` counts under its full name. That part
already works. The bug is in the loop: it creates a new paginator on every
pass. With one page it works. With more pages it re-reads page 1 forever.
(The test fake gives up after 20 calls, so you get an error instead of a
hang.) Rewrite the loop so it reads each page exactly once. Keep the error
wrapping as it is.

```go
PrefixSizes(ctx, api, "central-logs") // the bucket sends 2 objects per page
// => map[README.txt:120 api:2000 nginx:55100 worker:5000], nil
```

### 2. Count running instances by type

Write `RunningByType`. Use `ec2.NewDescribeInstancesPaginator` with an empty
`&ec2.DescribeInstancesInput{}`. Walk every reservation on every page. For
each running instance, add 1 to `counts[string(inst.InstanceType)]`. An instance
is running when `inst.State != nil && inst.State.Name ==
types.InstanceStateNameRunning`, with the import
`"github.com/aws/aws-sdk-go-v2/service/ec2/types"`. Check for `nil` first:
`State` is a pointer. `InstanceType` is a string type such as
`"t3.micro"`, so `string(...)` turns it into a plain string. Skip instances
that are stopped, pending or in any other state.

On error, return `nil` and the error wrapped with `%w` as
`running by type: <err>`.

```go
RunningByType(ctx, api) // page 2 of 3 is empty
// => map[m5.large:2 t3.micro:1], nil
```

## Learn more

- [Using paginators (AWS SDK for Go v2)](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/using.html#using-operation-paginators)
- [DescribeInstances (Amazon EC2 API reference)](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html)
- [ec2.DescribeInstancesPaginator on pkg.go.dev](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#DescribeInstancesPaginator)
