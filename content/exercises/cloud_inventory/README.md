# Practice: cloud inventory report

"What do we actually run, and where?" is the first question in every cost
review, security audit and incident. In this practice you build a small
inventory tool that asks several AWS regions at once and prints the answer as
a table for humans and as JSON for other tools.

## Learn

### Recap

Everything here comes from earlier exercises:

| You need | Where you saw it |
|---|---|
| SDK interfaces such as `ec2.DescribeInstancesAPIClient`, fakes in tests | *SDK clients: small interfaces and fakes* |
| `HasMorePages` / `NextPage` loops | *Paginators: walk every page* |
| `aws.String`, `aws.ToString`, instance `State.Name` and `Tags` | *SDK clients: small interfaces and fakes*, *Tag compliance: audit and fix instance tags* |
| Wrapping errors with `%w` | *Wrapping errors* |
| `sync.WaitGroup`, `sync.Mutex`, a buffered channel as a limit | *WaitGroups*, *Mutexes*, *Worker pools* |
| `context.WithCancel` | *Context and cancellation* |
| `json.MarshalIndent` | *Encoding JSON: a deployment status report* |

### Buckets have a region, but ListBuckets is global

EC2 is *regional*: an EC2 client for `eu-west-1` only sees instances in
`eu-west-1`, so you need one client per region. In real code you build them
from one config with `cfg.Copy()`, as in *AWS SDK configuration: many
regions, no leaked keys*:

```go
func(region string) ec2.DescribeInstancesAPIClient {
	regional := cfg.Copy()
	regional.Region = region
	return ec2.NewFromConfig(regional)
}
```

S3 is different. `ListBuckets` returns the buckets of the whole account, no
matter which region you call. Each bucket still lives in one region, and you
can ask S3 to return only the buckets of one region with the `BucketRegion`
field. `ListBuckets` pages like other list calls, and the SDK ships
`s3.NewListBucketsPaginator`:

```go
p := s3.NewListBucketsPaginator(api, &s3.ListBucketsInput{
	BucketRegion: aws.String("eu-west-1"),
})
```

### Aligned tables with text/tabwriter

`text/tabwriter` turns tab-separated cells into aligned columns. You write
rows with `\t` between cells and `\n` at the end, then call `Flush`:

```go
var b strings.Builder
w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
fmt.Fprintln(w, "HOST\tPORT")
fmt.Fprintln(w, "db-1\t5432")
w.Flush()
// b.String() == "HOST  PORT\ndb-1  5432\n"
```

The arguments `0, 0, 2, ' ', 0` mean: no minimum width, ignore tab stops, put
2 spaces between columns, pad with spaces, no special flags. The last cell of
each row is not padded, so lines never end in spaces.

### Sorting by several fields with cmp.Or

`cmp.Compare(a, b)` returns -1, 0 or +1. `cmp.Or` returns its first argument
that is not zero, which makes "sort by this, then by that" one expression:

```go
slices.SortFunc(hosts, func(a, b Host) int {
	return cmp.Or(
		cmp.Compare(a.Region, b.Region),
		cmp.Compare(a.Name, b.Name),
	)
})
```

## Your task

Your platform team wants a weekly report of every S3 bucket and EC2 instance
in its regions. All the types are in `exercise.go`. The tests use fake S3 and
EC2 clients, so nothing talks to AWS.

An `Item` is one row of the report:

| Field | Bucket | Instance |
|---|---|---|
| `Kind` | `"bucket"` | `"instance"` |
| `Region` | the region asked for | the region asked for |
| `ID` | bucket name | instance ID, such as `"i-0a1"` |
| `Name` | `""` | value of the `Name` tag, or `""` if it has none |
| `State` | `""` | `string(inst.State.Name)`, such as `"running"`; `""` if `State` is nil |

### 1. List the buckets of one region

Write `Buckets`. Page through `ListBuckets` with `s3.NewListBucketsPaginator`,
passing `BucketRegion: aws.String(region)` so S3 only returns that region's
buckets. Return one `"bucket"` item per bucket, in the order S3 returns them.
Wrap errors as `list buckets: <error>` with `%w`.

```go
Buckets(ctx, api, "eu-west-1")
// => []Item{{Kind: "bucket", Region: "eu-west-1", ID: "acme-logs-euw1"}}
```

A region with no buckets gives an empty (or nil) slice.

### 2. List the instances of one region

Write `Instances`. Page through `DescribeInstances` with
`ec2.NewDescribeInstancesPaginator`, and go through every reservation and
every instance in it. Return one `"instance"` item per instance, in the order
EC2 returns them. Find the name by looking for the tag whose key is `Name`
(an instance has other tags too). Wrap errors as
`describe instances: <error>`.

```go
Instances(ctx, api, "us-east-1")
// => []Item{
//      {Kind: "instance", Region: "us-east-1", ID: "i-0a1", Name: "web-1", State: "running"},
//      {Kind: "instance", Region: "us-east-1", ID: "i-0a3", State: "pending"}, // no Name tag
//    }
```

### 3. Render a table

Write `Table` with `tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)`. Write the
header row `KIND REGION ID NAME STATE`, then one row per item in the order
given. Show an empty `Name` or `State` as `-`. End every row with `\n`, and
remember to `Flush`.

```go
Table([]Item{
	{Kind: "bucket", Region: "eu-west-1", ID: "acme-logs-euw1"},
	{Kind: "instance", Region: "us-east-1", ID: "i-0a1", Name: "web-1", State: "running"},
})
```

```text
KIND      REGION     ID              NAME   STATE
bucket    eu-west-1  acme-logs-euw1  -      -
instance  us-east-1  i-0a1           web-1  running
```

With no items, the result is just the header:
`"KIND  REGION  ID  NAME  STATE\n"`.

### 4. Render JSON

Write `JSON`. Count the items whose `Kind` is `"bucket"` and `"instance"`, put
the counts and the items into a `Report`, and encode it with
`json.MarshalIndent(report, "", "  ")` (two spaces). The struct tags in
`exercise.go` already give the field names, and `omitempty` drops the empty
`name` and `state` of buckets.

```json
{
  "buckets": 1,
  "instances": 1,
  "items": [
    {
      "kind": "bucket",
      "region": "eu-west-1",
      "id": "acme-logs-euw1"
    },
    {
      "kind": "instance",
      "region": "us-east-1",
      "id": "i-0a1",
      "name": "web-1",
      "state": "running"
    }
  ]
}
```

With no items, `"items"` must be `[]`, not `null`. A nil slice encodes as
`null`, so start from an empty slice: `Items: []Item{}`.

### 5. Gather every region concurrently

Write `Gather`. First write a helper
`regionItems(ctx, region, s3api, ec2api) ([]Item, error)` that calls
`Buckets` and then `Instances` for one region and returns both lists in one
slice (return the first error unchanged). Then, in `Gather`, run
`regionItems` for every region in its own goroutine, with `ec2For(region)` as
the EC2 client and the one shared `s3api`. Then:

- **Limit:** at most `maxParallel` regions may be in progress at once (it is
  always at least 1). A buffered channel of size `maxParallel` works as a
  limit: send to it before the work and receive from it after.
- **Collect:** append each region's items to one slice. Guard it with a
  `sync.Mutex`, because several goroutines write to it.
- **Fail fast:** if a region fails, remember the first error as
  `region <name>: <error>` (wrapped with `%w`), and cancel a context made
  with `context.WithCancel(ctx)` so the other regions stop early. Gather then
  returns `nil` and that error.
- **Sort:** on success, sort the items by `Kind`, then `Region`, then `ID`,
  and return them.

```go
Gather(ctx, []string{"us-east-1", "eu-west-1"}, s3api, ec2For, 2)
// => buckets first (eu-west-1 before us-east-1), then instances, each group sorted by region and ID.

// If eu-west-1's DescribeInstances fails with "boom":
// => nil, "region eu-west-1: describe instances: boom"
```

This exercise runs with the race detector, as in *The race detector*: any
unguarded shared write fails the tests.

## Learn more

- [S3 ListBuckets API](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListBuckets.html)
- [AWS regions and availability zones](https://docs.aws.amazon.com/global-infrastructure/latest/regions/aws-regions.html)
- [text/tabwriter](https://pkg.go.dev/text/tabwriter)
- [cmp.Or](https://pkg.go.dev/cmp#Or)
