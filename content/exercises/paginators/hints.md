## Task 1: wrap it in a loop

Everything from `page, err := p.NextPage(ctx)` down to the end of the inner
`for` loop has to run once per page. Put those lines inside
`for p.HasMorePages() { ... }` and keep `return keys, nil` after the loop.

## Task 2: count as you go

Inside the inner loop, right after appending a key, check
`len(keys) == limit`. If it is, `return keys, nil` straight away. Returning
from inside the loop means `NextPage` is never called again.

## Task 3: the shape of the EC2 loop

It is the same loop as task 1, with one more level of nesting:

```go
p := ec2.NewDescribeInstancesPaginator(api, &ec2.DescribeInstancesInput{})
var ids []string
for p.HasMorePages() {
	page, err := p.NextPage(ctx)
	// handle err
	for _, r := range page.Reservations {
		// loop over r.Instances
	}
}
```

## Task 3: almost the answer

```go
	if err != nil {
		return nil, fmt.Errorf("describe instances: %w", err)
	}
	for _, r := range page.Reservations {
		for _, inst := range r.Instances {
			ids = append(ids, aws.ToString(inst.InstanceId))
		}
	}
```

Return `ids, nil` after the outer loop.
