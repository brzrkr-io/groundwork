## Task 1: Read every page

The code that reads one page is already there. It just has to run once per
page, for as long as the paginator says more pages remain.

### Closer

Wrap everything from `page, err := p.NextPage(ctx)` to the end of the inner
`for` in `for p.HasMorePages() { ... }`. Keep `return keys, nil` after the
loop.

### Answer

```go
for p.HasMorePages() {
	page, err := p.NextPage(ctx)
	if err != nil {
		return nil, fmt.Errorf("list s3://%s/%s: %w", bucket, prefix, err)
	}
	for _, obj := range page.Contents {
		keys = append(keys, aws.ToString(obj.Key))
	}
}
return keys, nil
```

## Task 2: Stop once you have enough

`MaxKeys` doesn't stop the loop, so you have to. After each key you append,
check whether you now have enough.

### Closer

Right after the `append`, compare `len(keys)` with `limit`. When they are
equal, return from the function.

### Answer

```go
keys = append(keys, aws.ToString(obj.Key))
if len(keys) == limit {
	return keys, nil
}
```

Returning from inside the loop means `NextPage` is never called again.

## Task 3: List every instance

It's the same loop as task 1, using `ec2.NewDescribeInstancesPaginator`, with
one more level of nesting: pages, then reservations, then instances.

### Closer

```go
p := ec2.NewDescribeInstancesPaginator(api, &ec2.DescribeInstancesInput{})
var ids []string
for p.HasMorePages() {
	page, err := p.NextPage(ctx)
	// handle err, wrapped as "describe instances: %w"
	for _, r := range page.Reservations {
		// loop over r.Instances
	}
}
```

### Answer

```go
p := ec2.NewDescribeInstancesPaginator(api, &ec2.DescribeInstancesInput{})
var ids []string
for p.HasMorePages() {
	page, err := p.NextPage(ctx)
	if err != nil {
		return nil, fmt.Errorf("describe instances: %w", err)
	}
	for _, r := range page.Reservations {
		for _, inst := range r.Instances {
			ids = append(ids, aws.ToString(inst.InstanceId))
		}
	}
}
return ids, nil
```

`%w` keeps `context.Canceled` inside the error, so `errors.Is` still finds it.
