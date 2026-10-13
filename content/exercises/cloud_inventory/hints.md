## Tasks 1 and 2: the paginator loop

Both listers have the same shape as in *Paginators: walk every page*:

```go
p := s3.NewListBucketsPaginator(api, &s3.ListBucketsInput{BucketRegion: aws.String(region)})
for p.HasMorePages() {
	page, err := p.NextPage(ctx)
	if err != nil {
		return nil, fmt.Errorf("list buckets: %w", err)
	}
	for _, b := range page.Buckets {
		// append an Item built from aws.ToString(b.Name)
	}
}
```

For instances, loop over `page.Reservations`, then over each reservation's
`Instances`. Look through `inst.Tags` for the tag with
`aws.ToString(tag.Key) == "Name"`.

## Tasks 3 and 4: rendering

For the table, a small helper keeps the row line short:

```go
func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", it.Kind, it.Region, it.ID, orDash(it.Name), orDash(it.State))
```

For JSON, start with `report := Report{Items: []Item{}}`, then loop over the
items, count by `it.Kind`, and append each item to `report.Items`.

## Task 5: one region

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

## Task 5: the goroutine skeleton

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
		sem <- struct{}{}        // wait for a free slot
		defer func() { <-sem }() // give the slot back

		found, err := regionItems(ctx, region, s3api, ec2For(region))
		// lock mu, then either record the first error and cancel,
		// or append found to items
	})
}
wg.Wait()
```

## Task 5: after the wait

```go
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

Inside the goroutine, record the error only once:

```go
if err != nil {
	if firstErr == nil {
		firstErr = fmt.Errorf("region %s: %w", region, err)
		cancel()
	}
	return
}
items = append(items, found...)
```
