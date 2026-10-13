## Task 1: List the buckets of one region

This is the paginator loop from *Paginators: walk every page*. Build the
paginator with the region filter, then turn each bucket on each page into an
`Item`.

### Closer

```go
p := s3.NewListBucketsPaginator(api, &s3.ListBucketsInput{BucketRegion: aws.String(region)})
for p.HasMorePages() {
	page, err := p.NextPage(ctx)
	// handle err, then range over page.Buckets
}
```

The bucket name is `aws.ToString(b.Name)`.

### Answer

```go
var items []Item
p := s3.NewListBucketsPaginator(api, &s3.ListBucketsInput{BucketRegion: aws.String(region)})
for p.HasMorePages() {
	page, err := p.NextPage(ctx)
	if err != nil {
		return nil, fmt.Errorf("list buckets: %w", err)
	}
	for _, b := range page.Buckets {
		items = append(items, Item{Kind: "bucket", Region: region, ID: aws.ToString(b.Name)})
	}
}
return items, nil
```

## Task 2: List the instances of one region

Same loop shape as task 1, but each page holds reservations, and each
reservation holds instances. The name is not a field: you find it among the
instance's tags.

### Closer

Use `ec2.NewDescribeInstancesPaginator(api, &ec2.DescribeInstancesInput{})`.
Loop `for _, r := range page.Reservations`, then `for _, inst := range r.Instances`.
Check `inst.State != nil` before reading `inst.State.Name`, and look for the tag
with `aws.ToString(tag.Key) == "Name"`.

### Answer

```go
for _, r := range page.Reservations {
	for _, inst := range r.Instances {
		item := Item{Kind: "instance", Region: region, ID: aws.ToString(inst.InstanceId)}
		if inst.State != nil {
			item.State = string(inst.State.Name)
		}
		for _, tag := range inst.Tags {
			if aws.ToString(tag.Key) == "Name" {
				item.Name = aws.ToString(tag.Value)
			}
		}
		items = append(items, item)
	}
}
```

Put this inside the same `HasMorePages` loop as task 1, wrapping errors as
`describe instances: %w`.

## Task 3: Render a table

Write the header and every row into the tabwriter with `\t` between cells and
`\n` at the end, then flush it and return the builder's text. Empty cells need
to become `-`.

### Closer

A small helper keeps the row short:

```go
func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
```

Rows look like `fmt.Fprintf(w, "%s\t%s\t...\n", it.Kind, ...)`.

### Answer

```go
var b strings.Builder
w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
fmt.Fprintln(w, "KIND\tREGION\tID\tNAME\tSTATE")
for _, it := range items {
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", it.Kind, it.Region, it.ID, orDash(it.Name), orDash(it.State))
}
w.Flush()
return b.String()
```

## Task 4: Render JSON

Build a `Report`, count each item by its `Kind`, and add the item to the
report's list. Remember that a nil slice encodes as `null`.

### Closer

Start from `report := Report{Items: []Item{}}`. A `switch it.Kind` with cases
`"bucket"` and `"instance"` does the counting. Finish with
`json.MarshalIndent(report, "", "  ")`.

### Answer

```go
report := Report{Items: []Item{}}
for _, it := range items {
	switch it.Kind {
	case "bucket":
		report.Buckets++
	case "instance":
		report.Instances++
	}
	report.Items = append(report.Items, it)
}
return json.MarshalIndent(report, "", "  ")
```

## Task 5: Gather every region concurrently

Start with `regionItems`: it just calls `Buckets`, then `Instances`, and joins
the results. In `Gather`, start one goroutine per region and wait for them all;
a buffered channel limits how many run, a mutex guards the shared slice and
the first error.

### Closer

```go
ctx, cancel := context.WithCancel(ctx)
defer cancel()
sem := make(chan struct{}, maxParallel)
// in each goroutine:
sem <- struct{}{}
defer func() { <-sem }()
found, err := regionItems(ctx, region, s3api, ec2For(region))
// lock mu; record the first error and cancel, or append found
```

After `wg.Wait()`, return the error if any, else sort with `cmp.Or` over
`Kind`, `Region` and `ID`.

### Answer

```go
func regionItems(ctx context.Context, region string, s3api s3.ListBucketsAPIClient, ec2api ec2.DescribeInstancesAPIClient) ([]Item, error) {
	buckets, err := Buckets(ctx, s3api, region)
	if err != nil {
		return nil, err
	}
	instances, err := Instances(ctx, ec2api, region)
	if err != nil {
		return nil, err
	}
	return append(buckets, instances...), nil
}
```

```go
ctx, cancel := context.WithCancel(ctx)
defer cancel()

sem := make(chan struct{}, maxParallel)
var (
	wg       sync.WaitGroup
	mu       sync.Mutex
	items    []Item
	firstErr error
)
for _, region := range regions {
	wg.Go(func() {
		sem <- struct{}{}
		defer func() { <-sem }()

		found, err := regionItems(ctx, region, s3api, ec2For(region))

		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("region %s: %w", region, err)
				cancel()
			}
			return
		}
		items = append(items, found...)
	})
}
wg.Wait()

if firstErr != nil {
	return nil, firstErr
}
slices.SortFunc(items, func(a, b Item) int {
	return cmp.Or(
		cmp.Compare(a.Kind, b.Kind),
		cmp.Compare(a.Region, b.Region),
		cmp.Compare(a.ID, b.ID),
	)
})
return items, nil
```

Cancelling on the first error makes the other regions stop early instead of
doing work nobody will read.
