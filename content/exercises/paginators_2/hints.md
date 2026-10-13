## Task 1: Fix the size report

The paginator remembers which page comes next. Look at where the starter
creates it: a fresh one on every pass always starts again at page 1.

### Closer

Move `p := s3.NewListObjectsV2Paginator(api, input)` above the loop. Then
change `for {` to `for p.HasMorePages() {` and delete the
`if !p.HasMorePages() { break }` at the bottom, because the loop condition
now does that job.

### Answer

```go
p := s3.NewListObjectsV2Paginator(api, input)
for p.HasMorePages() {
	page, err := p.NextPage(ctx)
	if err != nil {
		return nil, fmt.Errorf("size report s3://%s: %w", bucket, err)
	}
	for _, obj := range page.Contents {
		prefix, _, _ := strings.Cut(aws.ToString(obj.Key), "/")
		sizes[prefix] += aws.ToInt64(obj.Size)
	}
}
return sizes, nil
```

One paginator keeps the next-page token, so each page is read exactly once.

## Task 2: Count running instances by type

This is the same loop as task 1, but for EC2. Each page holds reservations,
and each reservation holds instances, so you need two `range` loops inside
the page loop. Remember to add the `types` import.

### Closer

```go
p := ec2.NewDescribeInstancesPaginator(api, &ec2.DescribeInstancesInput{})
for p.HasMorePages() {
	page, err := p.NextPage(ctx)
	// handle err
	for _, r := range page.Reservations {
		for _, inst := range r.Instances {
			// count inst if it is running
		}
	}
}
```

### Answer

```go
p := ec2.NewDescribeInstancesPaginator(api, &ec2.DescribeInstancesInput{})
for p.HasMorePages() {
	page, err := p.NextPage(ctx)
	if err != nil {
		return nil, fmt.Errorf("running by type: %w", err)
	}
	for _, r := range page.Reservations {
		for _, inst := range r.Instances {
			if inst.State != nil && inst.State.Name == types.InstanceStateNameRunning {
				counts[string(inst.InstanceType)]++
			}
		}
	}
}
return counts, nil
```

An empty page just has no reservations to range over; `HasMorePages()` still
decides when to stop.
