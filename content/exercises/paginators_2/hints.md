## Task 1: one paginator

Move `p := s3.NewListObjectsV2Paginator(api, input)` above the loop. Then
change `for {` to `for p.HasMorePages() {` and delete the
`if !p.HasMorePages() { break }` at the bottom, because the loop condition
now does that job.

## Task 2: the same loop, for EC2

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

## Task 2: almost the answer

```go
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
```

Return `counts, nil` after the outer loop.
