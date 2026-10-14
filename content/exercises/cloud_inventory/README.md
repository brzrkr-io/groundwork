# Practice: cloud inventory report

"What do we actually run, and where?" is the first question in every cost
review, security audit and incident. You build a small inventory tool that asks
several AWS regions at once and prints the answer as a table and as JSON.

**New here:** `s3.NewListBucketsPaginator` with `BucketRegion`. The rest is
practice, including `text/tabwriter` from *Project: ops inventory CLI* and
`cmp.Or` from *Project: inventory service*.

## Learn

### Recap

| You need | Where you saw it |
|---|---|
| SDK interfaces, fakes in tests | *SDK clients: small interfaces and fakes* |
| `HasMorePages` / `NextPage` loops | *Paginators: walk every page* |
| `aws.String`, `aws.ToString`, `State.Name`, `Tags` | *Tag compliance: audit and fix instance tags* |
| Wrapping errors with `%w` | *Wrapping errors* |
| `sync.WaitGroup`, `sync.Mutex`, buffered channel as a limit | *WaitGroups*, *Mutexes*, *Worker pools* |
| `context.WithCancel` | *Context and cancellation* |
| `json.MarshalIndent` | *Encoding JSON: a deployment status report* |

### Architecture

```text
Gather(regions)  ── one goroutine per region (at most maxParallel at once)
   └─ regionItems(region)
         ├─ Buckets(s3api, region)       shared S3 client, filtered by region
         └─ Instances(ec2For(region))    one EC2 client per region
   → merge, sort  → Table(items) or JSON(items)
```

| File | What it holds |
|---|---|
| `exercise.go` | `Item`, `Report`, `EC2Factory`, and the five functions you write. |
| `exercise_test.go` | Fake S3 (2 buckets per page, honours `BucketRegion`) and fake EC2 clients. Nothing talks to AWS. |

### Regional EC2, global S3 list

EC2 is *regional*: a client for `eu-west-1` only sees `eu-west-1`, so you need
one client per region. `EC2Factory` is a function that hands you that client.

S3's `ListBuckets` returns buckets from the whole account. The `BucketRegion`
field filters to one region, and `s3.NewListBucketsPaginator` pages through:

```go
p := s3.NewListBucketsPaginator(api, &s3.ListBucketsInput{
	BucketRegion: aws.String("ap-south-1"),
})
```

### Aligned tables with text/tabwriter

Write cells separated by `\t`, end rows with `\n`, then call `Flush`:

```go
var b strings.Builder
w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
fmt.Fprintln(w, "HOST\tPORT")
fmt.Fprintln(w, "db-1\t5432")
w.Flush()
// b.String() == "HOST  PORT\ndb-1  5432\n"
```

`0, 0, 2, ' ', 0` means: no minimum width, no tab stops, 2 spaces between
columns, pad with spaces, no flags. The last cell is never padded.

> **Why is it written like this?** Go requires `&b`, because the writer must
> add to your builder. Forgetting `Flush` is the classic bug: nothing is
> written until you call it.

### Sorting by several fields with cmp.Or

`cmp.Or` returns its first non-zero argument, so "sort by this, then that" is
one expression:

```go
slices.SortFunc(hosts, func(a, b Host) int {
	return cmp.Or(
		cmp.Compare(a.Rack, b.Rack),
		cmp.Compare(a.Name, b.Name),
	)
})
```

> **Why is it written like this?** The `func(a, b T) int` shape is required by
> `SortFunc`. The order of `cmp.Compare` calls is your choice: the first one
> wins, later ones only break ties.

## Your task

Your platform team wants a weekly report of every S3 bucket and EC2 instance in
its regions. An `Item` is one row:

| Field | Bucket | Instance |
|---|---|---|
| `Kind` | `"bucket"` | `"instance"` |
| `Region` | the `region` parameter | the `region` parameter |
| `ID` | bucket name | instance ID, such as `"i-0a1"` |
| `Name` | `""` | value of the `Name` tag, or `""` if none |
| `State` | `""` | `string(inst.State.Name)`, e.g. `"running"`; `""` if `State` is nil |

### 1. List the buckets of one region

**Goal:** `Buckets(ctx, api, region)` returns one `"bucket"` item per bucket in
`region`.

- Use `s3.NewListBucketsPaginator` with `BucketRegion` set to `region`; walk
  every page.
- Keep the order S3 returns. No buckets → empty or nil slice.
- On error return `nil` and the error wrapped as `list buckets: <error>` (`%w`).

```go
Buckets(ctx, api, "eu-west-1")
// => []Item{{Kind: "bucket", Region: "eu-west-1", ID: "acme-logs-euw1"}}
```

**Think it through:**
- Which field of each page holds the buckets, and which bucket field is the name?
- Is `region` a fixed string or a parameter here?

### 2. List the instances of one region

**Goal:** `Instances(ctx, api, region)` returns one `"instance"` item per
instance.

- Use `ec2.NewDescribeInstancesPaginator` with an empty input; walk every page,
  every reservation, every instance. Keep EC2's order.
- `Name` comes from the tag whose key is `Name` (instances have other tags too).
- On error return `nil` and `describe instances: <error>` (`%w`).

```go
Instances(ctx, api, "us-east-1")
// => []Item{
//      {Kind: "instance", Region: "us-east-1", ID: "i-0a1", Name: "web-1", State: "running"},
//      {Kind: "instance", Region: "us-east-1", ID: "i-0a3", State: "pending"}, // no Name tag
//    }
```

**Think it through:**
- How many nested loops do pages → reservations → instances need?
- What must you check before reading `inst.State.Name`?

### 3. Render a table

**Goal:** `Table(items)` returns an aligned text table with a header row.

- Use `tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)` writing into a `strings.Builder`.
- Header cells: `KIND`, `REGION`, `ID`, `NAME`, `STATE`. Then one row per item,
  in the given order, with the fields in that same order.
- An empty `Name` or `State` is shown as `-`. Every row ends with `\n`. Flush.

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

With no items, the result is just `"KIND  REGION  ID  NAME  STATE\n"`.

**Think it through:**
- Which character goes between cells, and which at the end of a row?
- Two fields need the same "empty → `-`" rule. A tiny helper function avoids repeating it.

### 4. Render JSON

**Goal:** `JSON(items)` returns a `Report` with counts and items, as indented JSON.

- `Buckets` counts items with `Kind == "bucket"`, `Instances` those with
  `"instance"`.
- Encode with `json.MarshalIndent(report, "", "  ")` (two spaces). The struct
  tags already name the fields; `omitempty` drops a bucket's empty `name` and
  `state`.
- No items → `"items": []`, not `null`. A nil slice encodes as `null`, so start
  `Items` as an empty, non-nil slice.

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

Empty input gives `{"buckets": 0, "instances": 0, "items": []}` in the same
indented layout.

**Think it through:**
- Which field of each item decides which counter grows?
- What is the difference between a nil slice and an empty slice literal?

### 5. Gather every region concurrently

**Goal:** `Gather` collects buckets and instances of every region, at most
`maxParallel` regions at once, sorted.

First write the helper
`regionItems(ctx, region, s3api, ec2api) ([]Item, error)`: call `Buckets`, then
`Instances`, return both lists in one slice. Return the first error unchanged.

Then in `Gather`, run `regionItems` for each region in its own goroutine, with
`ec2For(region)` as the EC2 client and the shared `s3api`:

- **Limit:** at most `maxParallel` regions in progress at once (always ≥ 1). A
  buffered channel of that size works: send before the work, receive after.
- **Collect:** append each region's items to one slice, guarded by a `sync.Mutex`.
- **Fail fast:** keep only the first error, wrapped as `region <name>: <error>`
  (`%w`), and cancel a context made with `context.WithCancel(ctx)` so other
  regions stop early. Then return `nil` and that error.
- **Sort:** on success, sort by `Kind`, then `Region`, then `ID`.

```go
Gather(ctx, []string{"us-east-1", "eu-west-1"}, s3api, ec2For, 2)
// => buckets first (eu-west-1 before us-east-1), then instances, each sorted by region, then ID.

// If eu-west-1's DescribeInstances fails with "boom":
// => nil, "region eu-west-1: describe instances: boom"
```

This exercise runs with the race detector (*The race detector*): any unguarded
shared write fails the tests.

**Think it through:**
- Which variables do several goroutines touch? Those need the mutex.
- How do you know an error is the *first* one?
- Why must you wait for every goroutine before sorting?

## Learn more

- [S3 ListBuckets API](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListBuckets.html)
- [text/tabwriter](https://pkg.go.dev/text/tabwriter)
- [cmp.Or](https://pkg.go.dev/cmp#Or)
- [Go by Example: WaitGroups](https://gobyexample.com/waitgroups)
